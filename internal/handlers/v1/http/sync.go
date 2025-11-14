package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/requestid"
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
	statuses, err := s.syncSrv.ListJobStatuses(c.Request.Context())
	if err != nil {
		logError(requestid.FromGin(c), "ListSyncJobs", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

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
	status, err := s.syncSrv.GetJobStatus(c.Request.Context(), id)
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
	err := s.syncSrv.StopJob(c.Request.Context(), id)
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

// ActionAllSyncJobs performs action on all sync jobs
func (s *Handler) ActionAllSyncJobs(c *gin.Context) {
	var request v1.SyncJobActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		zap.S().Errorw("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(c, "Invalid request body"))
		return
	}

	var affectedCount int
	var message string
	var responseAction v1.SyncJobActionResponseAction

	switch request.Action {
	case v1.SyncJobActionRequestActionCancel:
		// Get running jobs count before stopping
		runningJobStatuses, err := s.syncSrv.ListJobStatusesByStatus(c.Request.Context(), entity.StatusRunning)
		if err != nil {
			logError(requestid.FromGin(c), "ActionAllSyncJobs", err)
			c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
			return
		}
		pendingJobStatuses, err := s.syncSrv.ListJobStatusesByStatus(c.Request.Context(), entity.StatusPending)
		if err != nil {
			logError(requestid.FromGin(c), "ActionAllSyncJobs", err)
			c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
			return
		}
		affectedCount = len(runningJobStatuses) + len(pendingJobStatuses)

		// Stop all jobs using SyncService
		err = s.syncSrv.StopAllJobs(c.Request.Context())
		if err != nil {
			logError(requestid.FromGin(c), "ActionAllSyncJobs", err)
			c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
			return
		}
		message = "All sync jobs cancelled successfully"
		responseAction = v1.SyncJobActionResponseActionCancel

	case v1.SyncJobActionRequestActionPause:
		// Pause functionality not yet implemented for all jobs
		c.JSON(http.StatusBadRequest, errorResponse(c, "Pause all functionality not yet implemented"))
		return

	default:
		c.JSON(http.StatusBadRequest, errorResponse(c, "Invalid action. Supported actions: pause, cancel"))
		return
	}

	response := v1.SyncJobActionResponse{
		Message:       message,
		Action:        responseAction,
		AffectedCount: affectedCount,
	}
	c.JSON(http.StatusOK, response)
}

// ActionSyncJob performs action on a specific sync job
func (s *Handler) ActionSyncJob(c *gin.Context, id string) {
	var request v1.SyncJobActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		zap.S().Errorw("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(c, "Invalid request body"))
		return
	}

	var affectedCount int
	var message string
	var responseAction v1.SyncJobActionResponseAction

	switch request.Action {
	case v1.SyncJobActionRequestActionCancel:
		// Stop the job using SyncService
		err := s.syncSrv.StopJob(c.Request.Context(), id)
		if err != nil {
			logError(requestid.FromGin(c), "ActionSyncJob", err)
			c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
			return
		}
		affectedCount = 1
		message = "Sync job cancelled successfully"
		responseAction = v1.SyncJobActionResponseActionCancel

	case v1.SyncJobActionRequestActionPause:
		// Pause/resume the job using SyncService
		err := s.syncSrv.PauseJob(c.Request.Context(), id)
		if err != nil {
			logError(requestid.FromGin(c), "ActionSyncJob", err)
			c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
			return
		}
		affectedCount = 1
		message = "Sync job pause/resume toggled successfully"
		responseAction = v1.SyncJobActionResponseActionPause

	default:
		c.JSON(http.StatusBadRequest, errorResponse(c, "Invalid action. Supported actions: pause, cancel"))
		return
	}

	response := v1.SyncJobActionResponse{
		Message:       message,
		Action:        responseAction,
		AffectedCount: affectedCount,
	}
	c.JSON(http.StatusOK, response)
}

// ClearFinishedSyncJobs clears all finished sync jobs (completed, stopped, failed)
func (s *Handler) ClearFinishedSyncJobs(c *gin.Context) {
	// Get finished jobs count before clearing
	completedJobs, err := s.syncSrv.ListJobStatusesByStatus(c.Request.Context(), entity.StatusCompleted)
	if err != nil {
		logError(requestid.FromGin(c), "ClearFinishedSyncJobs", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}
	stoppedJobs, err := s.syncSrv.ListJobStatusesByStatus(c.Request.Context(), entity.StatusStopped)
	if err != nil {
		logError(requestid.FromGin(c), "ClearFinishedSyncJobs", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}
	failedJobs, err := s.syncSrv.ListJobStatusesByStatus(c.Request.Context(), entity.StatusFailed)
	if err != nil {
		logError(requestid.FromGin(c), "ClearFinishedSyncJobs", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	finishedJobStatuses := append(completedJobs, append(stoppedJobs, failedJobs...)...)
	clearedCount := len(finishedJobStatuses)

	// Clear all finished jobs using SyncService
	err = s.syncSrv.ClearFinishedJobs(c.Request.Context())
	if err != nil {
		logError(requestid.FromGin(c), "ClearFinishedSyncJobs", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	response := v1.ClearFinishedSyncJobsResponse{
		Message:      "Finished sync jobs cleared",
		ClearedCount: clearedCount,
	}
	c.JSON(http.StatusOK, response)
}
