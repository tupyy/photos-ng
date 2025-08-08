package v1

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// In-memory storage for sync jobs (in a real implementation, this would be in a database)
type SyncJobStatus struct {
	ID             string             `json:"id"`
	Path           string             `json:"path"`
	Status         string             `json:"status"` // "running", "completed", "failed"
	FilesRemaining int                `json:"filesRemaining"`
	TotalFiles     int                `json:"totalFiles"`
	FilesProcessed []v1.ProcessedFile `json:"filesProcessed"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

// Mock storage - in real implementation this would be in database/cache
var (
	syncJobs   = make(map[string]*SyncJobStatus)
	syncJobsMu sync.RWMutex // Protects syncJobs map for concurrent access
)

// StartSyncJob handles POST /api/v1/sync requests to start a new sync operation.
// Accepts a path in the request body and returns HTTP 202 with job ID on success.
func (s *ServerImpl) StartSyncJob(c *gin.Context) {
	var request v1.StartSyncRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		zap.S().Errorw("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Invalid request body",
		})
		return
	}

	// Generate unique job ID
	jobID := fmt.Sprintf("sync-%s", uuid.New().String())

	// Create job status entry
	job := &SyncJobStatus{
		ID:             jobID,
		Path:           request.Path,
		Status:         "running",
		FilesRemaining: 100, // Mock: simulate 100 files to process
		TotalFiles:     100,
		FilesProcessed: []v1.ProcessedFile{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Store job (in real implementation, this would be in database)
	syncJobsMu.Lock()
	syncJobs[jobID] = job
	syncJobsMu.Unlock()

	// Start background processing (mock implementation)
	go s.processSyncJob(jobID)

	zap.S().Infow("sync job started", "jobId", jobID, "path", request.Path)

	// Return job ID
	response := v1.StartSyncResponse{
		Id: jobID,
	}

	c.JSON(http.StatusAccepted, response)
}

// ListSyncJobs handles GET /api/v1/sync requests to list all sync jobs.
// Returns HTTP 200 with list of sync jobs on success.
func (s *ServerImpl) ListSyncJobs(c *gin.Context) {
	// Convert internal job status to API response format
	syncJobsMu.RLock()
	var jobs []v1.SyncJob
	for _, job := range syncJobs {
		apiJob := v1.SyncJob{
			Id:             job.ID,
			FilesRemaining: job.FilesRemaining,
			TotalFiles:     job.TotalFiles,
			FilesProcessed: job.FilesProcessed,
		}
		jobs = append(jobs, apiJob)
	}
	syncJobsMu.RUnlock()

	response := v1.ListSyncJobsResponse{
		Jobs: jobs,
	}

	c.JSON(http.StatusOK, response)
}

// GetSyncJob handles GET /api/v1/sync/{id} requests to get details of a specific sync job.
// Returns HTTP 404 if job not found, HTTP 200 with job details on success.
func (s *ServerImpl) GetSyncJob(c *gin.Context, id string) {
	syncJobsMu.RLock()
	job, exists := syncJobs[id]
	syncJobsMu.RUnlock()

	if !exists {
		zap.S().Warnw("sync job not found", "jobId", id)
		c.JSON(http.StatusNotFound, v1.Error{
			Message: "Sync job not found",
		})
		return
	}

	response := v1.SyncJob{
		Id:             job.ID,
		FilesRemaining: job.FilesRemaining,
		TotalFiles:     job.TotalFiles,
		FilesProcessed: job.FilesProcessed,
	}

	c.JSON(http.StatusOK, response)
}

// processSyncJob simulates background processing of a sync job
// In a real implementation, this would:
// 1. Walk the filesystem at the given path
// 2. Process each file (check if it's media, extract metadata, etc.)
// 3. Update database with new/updated media entries
// 4. Update job progress in real-time
func (s *ServerImpl) processSyncJob(jobID string) {
	syncJobsMu.RLock()
	job, exists := syncJobs[jobID]
	syncJobsMu.RUnlock()

	if !exists {
		return
	}

	zap.S().Infow("starting sync job processing", "jobId", jobID, "path", job.Path)

	// Simulate processing files over time
	for i := 0; i < job.TotalFiles; i++ {
		// Simulate processing time
		time.Sleep(100 * time.Millisecond)

		// Check if job still exists (could be cancelled)
		syncJobsMu.RLock()
		_, exists := syncJobs[jobID]
		syncJobsMu.RUnlock()
		if !exists {
			return
		}

		// Simulate file processing result
		filename := fmt.Sprintf("file_%03d.jpg", i+1)

		processedFile := v1.ProcessedFile{
			Filename: filename,
		}

		if i%10 == 9 { // Simulate 10% failure rate
			processedFile.Result.FromProcessedFileResult1("Failed to process: unsupported format")
		} else {
			processedFile.Result.FromProcessedFileResult0(v1.Ok)
		}

		// Update job status (need to lock for write access)
		syncJobsMu.Lock()
		if currentJob, exists := syncJobs[jobID]; exists {
			currentJob.FilesProcessed = append(currentJob.FilesProcessed, processedFile)
			currentJob.FilesRemaining = currentJob.TotalFiles - len(currentJob.FilesProcessed)
			currentJob.UpdatedAt = time.Now()

			if currentJob.FilesRemaining == 0 {
				currentJob.Status = "completed"
				zap.S().Infow("sync job completed", "jobId", jobID, "totalFiles", currentJob.TotalFiles)
			}
			job = currentJob // Update local reference
		}
		syncJobsMu.Unlock()

		if job.FilesRemaining == 0 {
			break
		}
	}
}

// StopSyncJob stops a specific sync job by ID
func (s *ServerImpl) StopSyncJob(c *gin.Context, id string) {
	syncJobsMu.Lock()
	job, exists := syncJobs[id]
	if !exists {
		syncJobsMu.Unlock()
		c.JSON(404, v1.Error{
			Message: "Sync job not found",
		})
		return
	}

	// Mark the job as stopped
	job.Status = "stopped"
	job.FilesRemaining = 0
	job.UpdatedAt = time.Now()
	syncJobs[id] = job
	syncJobsMu.Unlock()

	response := map[string]interface{}{
		"message": "Sync job stopped successfully",
		"jobId":   id,
	}
	c.JSON(200, response)
}

// StopAllSyncJobs stops all running sync jobs
func (s *ServerImpl) StopAllSyncJobs(c *gin.Context) {
	syncJobsMu.Lock()
	defer syncJobsMu.Unlock()

	stoppedCount := 0
	for id, job := range syncJobs {
		if job.Status == "running" && job.FilesRemaining > 0 {
			job.Status = "stopped"
			job.FilesRemaining = 0
			job.UpdatedAt = time.Now()
			syncJobs[id] = job
			stoppedCount++
		}
	}

	response := map[string]interface{}{
		"message":      "All sync jobs stopped",
		"stoppedCount": stoppedCount,
	}
	c.JSON(200, response)
}
