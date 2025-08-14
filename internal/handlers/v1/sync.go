package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Note: Sync jobs are now managed by the SyncService which uses the job scheduler

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

	// Start sync job using the SyncService
	jobID, err := s.syncSrv.StartSync(c.Request.Context(), request.Path)
	if err != nil {
		zap.S().Errorw("failed to start sync job", "path", request.Path, "error", err)
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: "Failed to start sync job",
		})
		return
	}

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
	// Get all job statuses from the SyncService
	statuses := s.syncSrv.ListSyncJobStatuses()

	// Convert to API response format
	var apiJobs []v1.SyncJob
	for _, status := range statuses {

		// Convert job progress to API format
		remaining := status.Remaining
		total := status.Total
		taskResults := v1.ConvertJobResultsToTaskResults(status.Results)

		apiJob := v1.SyncJob{
			Id:             status.Id.String(),
			Status:         v1.ConvertJobStatusToAPI(status.Status),
			RemainingTasks: remaining,
			TotalTasks:     total,
			CompletedTasks: taskResults,
			CreatedAt:      status.CreatedAt,
		}

		// Set timing fields based on job status
		if status.StartedAt != nil {
			apiJob.StartedAt = status.StartedAt
		} else {
			apiJob.StartedAt = &status.CreatedAt // Use created time if not started yet
		}

		if status.CompletedAt != nil {
			apiJob.FinishedAt = status.CompletedAt
		}
		// Don't set FinishedAt if job hasn't completed - let it be omitted

		apiJobs = append(apiJobs, apiJob)
	}

	response := v1.ListSyncJobsResponse{
		Jobs: apiJobs,
	}

	c.JSON(http.StatusOK, response)
}

// GetSyncJob handles GET /api/v1/sync/{id} requests to get details of a specific sync job.
// Returns HTTP 404 if job not found, HTTP 200 with job details on success.
func (s *ServerImpl) GetSyncJob(c *gin.Context, id string) {
	// Get job status from SyncService
	status, err := s.syncSrv.GetSyncJobStatus(id)
	if err != nil {
		zap.S().Warnw("sync job not found", "jobId", id, "error", err)
		c.JSON(http.StatusNotFound, v1.Error{
			Message: "Sync job not found",
		})
		return
	}

	// Convert to API response format
	remaining := status.Remaining
	total := status.Total
	taskResults := v1.ConvertJobResultsToTaskResults(status.Results)

	response := v1.SyncJob{
		Id:             status.Id.String(),
		Status:         v1.ConvertJobStatusToAPI(status.Status),
		RemainingTasks: remaining,
		TotalTasks:     total,
		CompletedTasks: taskResults,
		CreatedAt:      status.CreatedAt,
	}

	// Set timing fields based on job status
	if status.StartedAt != nil {
		response.StartedAt = status.StartedAt
	} else {
		response.StartedAt = &status.CreatedAt // Use created time if not started yet
	}

	if status.CompletedAt != nil {
		response.FinishedAt = status.CompletedAt
	}
	// Don't set FinishedAt if job hasn't completed - let it be omitted

	c.JSON(http.StatusOK, response)
}

// StopSyncJob stops a specific sync job by ID
func (s *ServerImpl) StopSyncJob(c *gin.Context, id string) {
	// Stop the job using SyncService
	err := s.syncSrv.StopSyncJob(id)
	if err != nil {
		zap.S().Warnw("failed to stop sync job", "jobId", id, "error", err)
		c.JSON(http.StatusNotFound, v1.Error{
			Message: "Sync job not found",
		})
		return
	}

	response := map[string]interface{}{
		"message": "Sync job stopped successfully",
		"jobId":   id,
	}
	c.JSON(http.StatusOK, response)
}

// StopAllSyncJobs stops all running sync jobs
func (s *ServerImpl) StopAllSyncJobs(c *gin.Context) {
	// Get running jobs count before stopping
	runningJobStatuses := s.syncSrv.ListSyncJobStatusesByStatus(services.StatusRunning)
	stoppedCount := len(runningJobStatuses)

	// Stop all jobs using SyncService
	err := s.syncSrv.StopAllSyncJobs()
	if err != nil {
		zap.S().Errorw("failed to stop all sync jobs", "error", err)
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: "Failed to stop all sync jobs",
		})
		return
	}

	response := map[string]interface{}{
		"message":      "All sync jobs stopped",
		"stoppedCount": stoppedCount,
	}
	c.JSON(http.StatusOK, response)
}
