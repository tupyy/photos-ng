package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/requestid"
)

// Note: Sync jobs are now managed by the SyncService which uses the job scheduler

// StartSyncJob handles POST /api/v1/sync requests to start a new sync operation.
// Accepts a path in the request body and returns HTTP 202 with job ID on success.
func (s *Handler) StartSyncJob(c *gin.Context) {
	var request v1.StartSyncRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		zap.S().Errorw("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(c, "Invalid request body"))
		return
	}

	// Start sync job using the SyncService
	jobID, err := s.syncSrv.StartSync(c.Request.Context(), request.Path)
	if err != nil {
		logError(requestid.FromGin(c), "StartSyncJob", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	// Return job ID
	response := v1.StartSyncResponse{
		Id: jobID,
	}

	c.JSON(http.StatusAccepted, response)
}

// ListSyncJobs handles GET /api/v1/sync requests to list all sync jobs.
// Returns HTTP 200 with list of sync jobs on success.
func (s *Handler) ListSyncJobs(c *gin.Context) {
	// Get all job statuses from the SyncService
	statuses := s.syncSrv.ListSyncJobStatuses()

	// Convert to API response format
	var apiJobs []v1.SyncJob
	for _, status := range statuses {
		apiJobs = append(apiJobs, v1.NewSyncJob(status))
	}

	response := v1.ListSyncJobsResponse{
		Jobs: apiJobs,
	}

	c.JSON(http.StatusOK, response)
}

// GetSyncJob handles GET /api/v1/sync/{id} requests to get details of a specific sync job.
// Returns HTTP 404 if job not found, HTTP 200 with job details on success.
func (s *Handler) GetSyncJob(c *gin.Context, id string) {
	// Get job status from SyncService
	status, err := s.syncSrv.GetSyncJobStatus(id)
	if err != nil {
		logError(requestid.FromGin(c), "GetSyncJob", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	// Convert to API response format using centralized function
	response := v1.NewSyncJob(*status)

	c.JSON(http.StatusOK, response)
}

// StopSyncJob stops a specific sync job by ID
func (s *Handler) StopSyncJob(c *gin.Context, id string) {
	// Stop the job using SyncService
	err := s.syncSrv.StopSyncJob(id)
	if err != nil {
		logError(requestid.FromGin(c), "StopSyncJob", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	response := map[string]any{
		"message": "Sync job stopped successfully",
		"jobId":   id,
	}
	c.JSON(http.StatusOK, response)
}

// StopAllSyncJobs stops all running sync jobs
func (s *Handler) StopAllSyncJobs(c *gin.Context) {
	// Get running jobs count before stopping
	runningJobStatuses := s.syncSrv.ListSyncJobStatusesByStatus(services.StatusRunning)
	stoppedCount := len(runningJobStatuses)

	// Stop all jobs using SyncService
	err := s.syncSrv.StopAllSyncJobs()
	if err != nil {
		logError(requestid.FromGin(c), "StopAllSyncJobs", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	response := map[string]any{
		"message":      "All sync jobs stopped",
		"stoppedCount": stoppedCount,
	}
	c.JSON(http.StatusOK, response)
}
