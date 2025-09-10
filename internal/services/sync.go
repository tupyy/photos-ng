package services

import (
	"context"
	"fmt"
	"strings"
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

	// Create a root album entity for the sync operation
	tracer.Step("create_album_entity").
		WithString("album_path", albumPath).
		Log()

	album := entity.NewAlbum(strings.TrimSuffix(albumPath, "/"))

	debug.BusinessLogic("album entity created for sync").
		WithString("album_id", album.ID).
		WithString("album_path", album.Path).
		Log()

	// Generate sync jobs using JobGenerator
	tracer.Step("generate_sync_jobs").
		WithString("album_id", album.ID).
		Log()

	start := time.Now()
	generator := NewJobGenerator(s.albumService, s.mediaService, s.fsDatastore)
	syncJobs, err := generator.Generate(ctx, albumPath)
	jobGenerationDuration := time.Since(start)
	if err != nil {
		// Return ServiceError (handlers will log the error)
		return "", NewSyncJobError(ctx, "generate_jobs", "", err).
			WithContext("album_path", albumPath)
	}

	tracer.Performance("job_generation", jobGenerationDuration)

	debug.BusinessLogic("sync jobs generated").
		WithString("album_id", album.ID).
		WithInt("job_count", len(syncJobs)).
		Log()

	if len(syncJobs) == 0 {
		return "", NewSyncJobError(ctx, "generate_jobs", "", fmt.Errorf("no sync jobs generated for path: %s", albumPath)).
			WithContext("album_path", albumPath)
	}

	// Add all jobs to scheduler and log each one
	tracer.Step("schedule_jobs").
		WithInt("job_count", len(syncJobs)).
		WithString("scheduler", "background").
		Log()

	start = time.Now()
	var jobIDs []string
	for i, syncJob := range syncJobs {
		if err := s.scheduler.Add(syncJob); err != nil {
			// Return ServiceError (handlers will log the error)
			return "", NewSyncJobError(ctx, "schedule_job", syncJob.GetID().String(), err).
				WithContext("album_path", albumPath).
				WithContext("job_index", i)
		}
		
		jobID := syncJob.GetID().String()
		jobIDs = append(jobIDs, jobID)
		
		debug.BusinessLogic("sync job scheduled").
			WithString("job_id", jobID).
			WithString("album_id", syncJob.GetAlbumID()).
			WithInt("job_index", i).
			Log()
	}
	schedulingDuration := time.Since(start)

	tracer.Performance("scheduling", schedulingDuration)

	debug.BusinessLogic("all sync jobs scheduled successfully").
		WithInt("total_jobs", len(syncJobs)).
		WithString("album_path", albumPath).
		WithString("album_id", album.ID).
		Log()

	tracer.Success().
		WithInt("total_jobs", len(syncJobs)).
		WithString("album_path", albumPath).
		Log()

	// Return a batch identifier or the album path since we have multiple jobs
	return albumPath, nil
}

// GetSyncJobStatus returns the status of a sync job by ID
func (s *SyncService) GetSyncJobStatus(jobID string) (*entity.JobProgress, error) {
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
func (s *SyncService) ListSyncJobStatuses() []entity.JobProgress {
	jobs := s.scheduler.GetAll()
	statuses := make([]entity.JobProgress, len(jobs))
	for i, syncJob := range jobs {
		statuses[i] = syncJob.Status()
	}
	return statuses
}

// ListSyncJobStatusesByStatus returns job statuses filtered by status
func (s *SyncService) ListSyncJobStatusesByStatus(status entity.JobStatus) []entity.JobProgress {
	jobs := s.scheduler.GetByStatus(status)
	statuses := make([]entity.JobProgress, len(jobs))
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

	if err := s.scheduler.StopJob(jobID); err != nil {
		return NewSyncJobError(ctx, "stop_sync_job", jobID, err)
	}

	debug.BusinessLogic("sync job stopped successfully").
		WithString("job_id", jobID).
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

	// Get all active jobs (running and pending)
	tracer.Step("get_active_jobs").Log()

	runningJobs := s.scheduler.GetByStatus(entity.StatusRunning)
	pendingJobs := s.scheduler.GetByStatus(entity.StatusPending)
	activeJobs := append(runningJobs, pendingJobs...)

	debug.BusinessLogic("found active jobs to stop").
		WithInt("running_count", len(runningJobs)).
		WithInt("pending_count", len(pendingJobs)).
		WithInt("total_active", len(activeJobs)).
		Log()

	if len(activeJobs) == 0 {
		tracer.Success().
			WithInt("stopped_count", 0).
			WithInt("active_count", 0).
			Log()
		return nil
	}

	// Stop each job
	tracer.Step("stop_jobs").
		WithInt("job_count", len(activeJobs)).
		Log()

	var errors []error
	successCount := 0

	start := time.Now()
	for _, syncJob := range activeJobs {
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
		WithInt("total_jobs", len(activeJobs)).
		WithInt("success_count", successCount).
		WithInt("failed_count", len(errors)).
		WithParam("duration", stopDuration).
		Log()

	if len(errors) > 0 {
		return NewInternalError(ctx, "stop_all_sync_jobs", "multiple_job_failures", fmt.Errorf("failed to stop some jobs: %v", errors)).
			WithContext("failed_count", len(errors)).
			WithContext("total_count", len(activeJobs))
	}

	tracer.Success().
		WithInt("stopped_count", len(activeJobs)).
		Log()

	return nil
}

// ClearFinishedSyncJobs removes all completed, stopped, and failed sync jobs
func (s *SyncService) ClearFinishedSyncJobs() error {
	ctx := context.Background()
	debug := s.debug.WithContext(ctx)
	tracer := debug.StartOperation("clear_finished_sync_jobs").Build()

	// Get all finished jobs (completed, stopped, failed)
	tracer.Step("get_finished_jobs").Log()

	completedJobs := s.scheduler.GetByStatus(entity.StatusCompleted)
	stoppedJobs := s.scheduler.GetByStatus(entity.StatusStopped)
	failedJobs := s.scheduler.GetByStatus(entity.StatusFailed)
	
	finishedJobs := append(completedJobs, stoppedJobs...)
	finishedJobs = append(finishedJobs, failedJobs...)

	debug.BusinessLogic("found finished jobs to clear").
		WithInt("completed_count", len(completedJobs)).
		WithInt("stopped_count", len(stoppedJobs)).
		WithInt("failed_count", len(failedJobs)).
		WithInt("total_finished", len(finishedJobs)).
		Log()

	if len(finishedJobs) == 0 {
		tracer.Success().
			WithInt("cleared_count", 0).
			Log()
		return nil
	}

	// Remove each finished job from scheduler
	tracer.Step("clear_jobs").
		WithInt("job_count", len(finishedJobs)).
		Log()

	var errors []error
	successCount := 0

	start := time.Now()
	for _, syncJob := range finishedJobs {
		if err := s.scheduler.Remove(syncJob.GetID().String()); err != nil {
			// Collect error for return (handlers will log the error)
			errors = append(errors, NewSyncJobError(ctx, "clear_job", syncJob.GetID().String(), err))
		} else {
			successCount++
		}
	}
	clearDuration := time.Since(start)

	tracer.Performance("clear_all_duration", clearDuration)

	debug.BusinessLogic("bulk job clear completed").
		WithInt("total_jobs", len(finishedJobs)).
		WithInt("success_count", successCount).
		WithInt("failed_count", len(errors)).
		WithParam("duration", clearDuration).
		Log()

	if len(errors) > 0 {
		return NewInternalError(ctx, "clear_finished_sync_jobs", "multiple_job_failures", fmt.Errorf("failed to clear some jobs: %v", errors)).
			WithContext("failed_count", len(errors)).
			WithContext("total_count", len(finishedJobs))
	}

	tracer.Success().
		WithInt("cleared_count", len(finishedJobs)).
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

// IsAlbumSyncing checks if there's an active sync job for the given album ID
func (s *SyncService) IsAlbumSyncing(albumID string) bool {
	jobs := s.scheduler.GetByAlbumID(albumID)
	if len(jobs) == 0 {
		return false
	}

	for _, j := range jobs {
		if j.Status().Status == entity.StatusRunning {
			return true
		}
	}

	return false
}

// GetSchedulerStats returns statistics about the scheduler
func (s *SyncService) GetSchedulerStats() map[string]int {
	stats := make(map[string]int)
	stats["total"] = len(s.scheduler.GetAll())
	stats["pending"] = len(s.scheduler.GetByStatus(entity.StatusPending))
	stats["running"] = len(s.scheduler.GetByStatus(entity.StatusRunning))
	stats["completed"] = len(s.scheduler.GetByStatus(entity.StatusCompleted))
	stats["failed"] = len(s.scheduler.GetByStatus(entity.StatusFailed))
	stats["stopped"] = len(s.scheduler.GetByStatus(entity.StatusStopped))

	return stats
}
