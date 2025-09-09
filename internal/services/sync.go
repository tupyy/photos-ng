package services

import (
	"context"
	"fmt"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

// SyncService manages sync operations using a single scheduler instance
type SyncService struct {
	albumService *AlbumService
	mediaService *MediaService
	fsDatastore  *fs.Datastore
	scheduler    *Scheduler
	debug        *logger.DebugLogger
}

// NewSyncService creates a new sync service that manages the job scheduler
func NewSyncService(albumService *AlbumService, mediaService *MediaService, fsDatastore *fs.Datastore) *SyncService {
	return &SyncService{
		albumService: albumService,
		mediaService: mediaService,
		fsDatastore:  fsDatastore,
		scheduler:    GetScheduler(),
		debug:        logger.NewDebugLogger("sync_service"),
	}
}

// StartSync starts a new sync job for the given album path
// Returns the job ID and any error that occurred during job creation
func (s *SyncService) StartSync(ctx context.Context, albumPath string) (string, error) {
	debug := s.debug.WithContext(ctx)
	tracer := debug.StartOperation("start_sync").
		WithString("album_path", albumPath).
		Build()

	// Create a root album entity for the sync operation (the job will handle creation if needed)
	tracer.Step("create_album_entity").
		WithString("album_path", albumPath).
		Log()

	rootAlbum := entity.NewAlbum(albumPath)

	debug.BusinessLogic("root album entity created for sync").
		WithString("album_id", rootAlbum.ID).
		WithString("album_path", rootAlbum.Path).
		Log()

	// Create the sync job - it will handle album creation during execution
	tracer.Step("create_sync_job").
		WithString("album_id", rootAlbum.ID).
		Log()

	start := time.Now()
	syncJob, err := NewSyncJob(rootAlbum, s.albumService, s.mediaService, s.fsDatastore)
	jobCreationDuration := time.Since(start)
	if err != nil {
		// Return ServiceError (handlers will log the error)
		return "", NewSyncJobError(ctx, "create_job", "", err).
			WithContext("album_path", albumPath)
	}

	tracer.Performance("job_creation", jobCreationDuration)

	// Add job to scheduler
	tracer.Step("schedule_job").
		WithString("job_id", syncJob.GetID().String()).
		WithString("scheduler", "background").
		Log()

	if err := s.scheduler.Add(syncJob); err != nil {
		// Return ServiceError (handlers will log the error)
		return "", NewSyncJobError(ctx, "schedule_job", syncJob.GetID().String(), err).
			WithContext("album_path", albumPath)
	}

	jobID := syncJob.GetID().String()

	debug.BusinessLogic("sync job created and scheduled successfully").
		WithString("job_id", jobID).
		WithString("album_path", albumPath).
		WithString("album_id", rootAlbum.ID).
		Log()

	tracer.Success().
		WithString("job_id", jobID).
		WithString("album_path", albumPath).
		Log()

	return jobID, nil
}

// GetSyncJobStatus returns the status of a sync job by ID
func (s *SyncService) GetSyncJobStatus(jobID string) (*JobProgress, error) {
	ctx := context.Background()
	debug := s.debug.WithContext(ctx)
	tracer := debug.StartOperation("get_sync_job_status").
		WithString("job_id", jobID).
		Build()

	// Input validation (return ServiceError, no logging)
	if jobID == "" {
		return nil, NewValidationError(ctx, "get_sync_job_status", "invalid_input").
			WithContext("validation_error", "empty_job_id")
	}

	// Scheduler query with debug timing
	tracer.Step("scheduler_lookup").
		WithString("job_id", jobID).
		Log()

	start := time.Now()
	syncJob := s.scheduler.Get(jobID)
	lookupDuration := time.Since(start)

	debug.BusinessLogic("scheduler job lookup").
		WithString("job_id", jobID).
		WithBool("found", syncJob != nil).
		WithParam("duration", lookupDuration).
		Log()

	if syncJob == nil {
		// Return ServiceError (handlers will log the error)
		return nil, NewNotFoundError(ctx, "get_sync_job_status", "job_not_found").
			WithContext("job_id", jobID)
	}

	status := syncJob.Status()

	tracer.Success().
		WithString("job_id", jobID).
		WithParam("status", status.Status).
		WithInt("total", status.Total).
		WithInt("remaining", status.Remaining).
		Log()

	return &status, nil
}

// ListSyncJobStatuses returns statuses of all sync jobs
func (s *SyncService) ListSyncJobStatuses() []JobProgress {
	jobs := s.scheduler.GetAll()
	statuses := make([]JobProgress, len(jobs))
	for i, syncJob := range jobs {
		statuses[i] = syncJob.Status()
	}
	return statuses
}

// ListSyncJobStatusesByStatus returns job statuses filtered by status
func (s *SyncService) ListSyncJobStatusesByStatus(status JobStatus) []JobProgress {
	jobs := s.scheduler.GetByStatus(status)
	statuses := make([]JobProgress, len(jobs))
	for i, syncJob := range jobs {
		statuses[i] = syncJob.Status()
	}
	return statuses
}

// StopSyncJob stops a specific sync job
func (s *SyncService) StopSyncJob(jobID string) error {
	ctx := context.Background()
	debug := s.debug.WithContext(ctx)
	tracer := debug.StartOperation("stop_sync_job").
		WithString("job_id", jobID).
		Build()

	// Input validation (return ServiceError, no logging)
	if jobID == "" {
		return NewValidationError(ctx, "stop_sync_job", "invalid_input").
			WithContext("validation_error", "empty_job_id")
	}

	// Scheduler lookup
	tracer.Step("scheduler_lookup").
		WithString("job_id", jobID).
		Log()

	syncJob := s.scheduler.Get(jobID)
	if syncJob == nil {
		return NewNotFoundError(ctx, "stop_sync_job", "job_not_found").
			WithContext("job_id", jobID)
	}

	debug.BusinessLogic("job found, attempting to stop").
		WithString("job_id", jobID).
		WithParam("current_status", syncJob.Status().Status).
		Log()

	// Stop the job
	tracer.Step("stop_job").
		WithString("job_id", jobID).
		Log()

	start := time.Now()
	if err := syncJob.Stop(); err != nil {
		// Return ServiceError (handlers will log the error)
		return NewSyncJobError(ctx, "stop_job", jobID, err)
	}
	stopDuration := time.Since(start)

	debug.BusinessLogic("sync job stopped successfully").
		WithString("job_id", jobID).
		WithParam("duration", stopDuration).
		Log()

	tracer.Success().
		WithString("job_id", jobID).
		Log()

	return nil
}

// StopAllSyncJobs stops all running sync jobs
func (s *SyncService) StopAllSyncJobs() error {
	ctx := context.Background()
	debug := s.debug.WithContext(ctx)
	tracer := debug.StartOperation("stop_all_sync_jobs").Build()

	// Get all running jobs
	tracer.Step("get_running_jobs").Log()

	runningJobs := s.scheduler.GetByStatus(StatusRunning)

	debug.BusinessLogic("found running jobs to stop").
		WithInt("running_count", len(runningJobs)).
		Log()

	if len(runningJobs) == 0 {
		tracer.Success().
			WithInt("stopped_count", 0).
			WithInt("running_count", 0).
			Log()
		return nil
	}

	// Stop each job
	tracer.Step("stop_jobs").
		WithInt("job_count", len(runningJobs)).
		Log()

	var errors []error
	successCount := 0

	start := time.Now()
	for _, syncJob := range runningJobs {
		if err := syncJob.Stop(); err != nil {
			// Collect error for return (handlers will log the error)
			errors = append(errors, NewSyncJobError(ctx, "stop_job", syncJob.GetID().String(), err))
		} else {
			successCount++
		}
	}
	stopDuration := time.Since(start)

	tracer.Performance("stop_all_duration", stopDuration)

	debug.BusinessLogic("bulk job stop completed").
		WithInt("total_jobs", len(runningJobs)).
		WithInt("success_count", successCount).
		WithInt("failed_count", len(errors)).
		WithParam("duration", stopDuration).
		Log()

	if len(errors) > 0 {
		return NewInternalError(ctx, "stop_all_sync_jobs", "multiple_job_failures", fmt.Errorf("failed to stop some jobs: %v", errors)).
			WithContext("failed_count", len(errors)).
			WithContext("total_count", len(runningJobs))
	}

	tracer.Success().
		WithInt("stopped_count", len(runningJobs)).
		Log()

	return nil
}

// Shutdown gracefully shuts down the sync service
func (s *SyncService) Shutdown() {
	ctx := context.Background()
	debug := s.debug.WithContext(ctx)
	tracer := debug.StartOperation("shutdown").Build()

	debug.BusinessLogic("shutting down sync service").Log()

	// Stop all running jobs
	tracer.Step("stop_all_jobs").Log()

	start := time.Now()
	if err := s.StopAllSyncJobs(); err != nil {
		debug.BusinessLogic("errors occurred while stopping jobs during shutdown").
			WithString("error", err.Error()).
			Log()
	}
	stopJobsDuration := time.Since(start)

	// Stop the scheduler
	tracer.Step("stop_scheduler").Log()

	schedStart := time.Now()
	s.scheduler.Stop()
	stopSchedulerDuration := time.Since(schedStart)

	debug.BusinessLogic("sync service shutdown complete").
		WithParam("stop_jobs_duration", stopJobsDuration).
		WithParam("stop_scheduler_duration", stopSchedulerDuration).
		Log()

	tracer.Success().
		WithBool("shutdown_completed", true).
		Log()
}

// GetSchedulerStats returns statistics about the scheduler
func (s *SyncService) GetSchedulerStats() map[string]int {
	stats := make(map[string]int)
	stats["total"] = len(s.scheduler.GetAll())
	stats["pending"] = len(s.scheduler.GetByStatus(StatusPending))
	stats["running"] = len(s.scheduler.GetByStatus(StatusRunning))
	stats["completed"] = len(s.scheduler.GetByStatus(StatusCompleted))
	stats["failed"] = len(s.scheduler.GetByStatus(StatusFailed))
	stats["stopped"] = len(s.scheduler.GetByStatus(StatusStopped))

	return stats
}
