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
	logger       *logger.StructuredLogger
}

// NewSyncService creates a new sync service that manages the job scheduler
func NewSyncService(albumService *AlbumService, mediaService *MediaService, fsDatastore *fs.Datastore) *SyncService {
	return &SyncService{
		albumService: albumService,
		mediaService: mediaService,
		fsDatastore:  fsDatastore,
		scheduler:    GetScheduler(),
		logger:       logger.New("sync_service"),
	}
}

// StartSync starts a new sync job for the given album path
// Returns the job ID and any error that occurred during job creation
func (s *SyncService) StartSync(ctx context.Context, albumPath string) (string, error) {
	logger := s.logger.WithContext(ctx).Operation("start_sync").
		WithString(AlbumPath, albumPath).
		Build()

	// Create a root album entity for the sync operation
	logger.Step("create_album_entity").
		WithString(AlbumPath, albumPath).
		Log()

	album := entity.NewAlbum(strings.TrimSuffix(albumPath, "/"))

	logger.Step("album entity created for sync").
		WithString(AlbumID, album.ID).
		WithString(AlbumPath, album.Path).
		Log()

	// Generate sync jobs using JobGenerator
	logger.Step("generate_sync_jobs").
		WithString(AlbumID, album.ID).
		Log()

	generator := NewJobGenerator(s.albumService, s.mediaService, s.fsDatastore)
	syncJobs, err := generator.Generate(ctx, albumPath)
	if err != nil {
		// Return ServiceError (handlers will log the error)
		return "", NewSyncJobError(ctx, "generate_jobs", "", err).
			WithContext(AlbumPath, albumPath)
	}


	logger.Step("sync jobs generated").
		WithString(AlbumID, album.ID).
		WithInt("job_count", len(syncJobs)).
		Log()

	if len(syncJobs) == 0 {
		return "", NewSyncJobError(ctx, "generate_jobs", "", fmt.Errorf("no sync jobs generated for path: %s", albumPath)).
			WithContext(AlbumPath, albumPath)
	}

	// Add all jobs to scheduler and log each one
	logger.Step("schedule_jobs").
		WithInt("job_count", len(syncJobs)).
		WithString("scheduler", "background").
		Log()

	for i, syncJob := range syncJobs {
		if err := s.scheduler.Add(syncJob); err != nil {
			// Return ServiceError (handlers will log the error)
			return "", NewSyncJobError(ctx, "schedule_job", syncJob.GetID().String(), err).
				WithContext(AlbumPath, albumPath).
				WithContext("job_index", i)
		}

		jobID := syncJob.GetID().String()

		logger.Step("sync job scheduled").
			WithString(JobID, jobID).
			WithInt("job_index", i).
			Log()
	}


	logger.Step("all sync jobs scheduled successfully").
		WithInt("total_jobs", len(syncJobs)).
		WithString(AlbumPath, albumPath).
		WithString(AlbumID, album.ID).
		Log()

	logger.Success().
		WithInt("total_jobs", len(syncJobs)).
		WithString(AlbumPath, albumPath).
		Log()

	// Return a batch identifier or the album path since we have multiple jobs
	return albumPath, nil
}

// GetJobStatus returns the status of a job by ID
func (s *SyncService) GetJobStatus(jobID string) (*entity.JobProgress, error) {
	ctx := context.Background()
	logger := s.logger.WithContext(ctx).Operation("get_sync_job_status").
		WithString(JobID, jobID).
		Build()

	// Input validation (return ServiceError, no logging)
	if jobID == "" {
		return nil, NewValidationError(ctx, "get_sync_job_status", "invalid_input").
			WithContext("validation_error", "empty_job_id")
	}

	// Scheduler query with debug timing
	logger.Step("scheduler_lookup").
		WithString(JobID, jobID).
		Log()

	syncJob := s.scheduler.Get(jobID)

	logger.Step("scheduler job lookup").
		WithString(JobID, jobID).
		WithBool("found", syncJob != nil).
		Log()

	if syncJob == nil {
		// Return ServiceError (handlers will log the error)
		return nil, NewNotFoundError(ctx, "get_sync_job_status", "job_not_found").
			WithContext(JobID, jobID)
	}

	status := syncJob.Status()

	logger.Success().
		WithString(JobID, jobID).
		WithParam("status", status.Status).
		WithInt("total", status.Total).
		WithInt("remaining", status.Remaining).
		Log()

	return &status, nil
}

// ListJobStatuses returns statuses of all jobs
func (s *SyncService) ListJobStatuses() []entity.JobProgress {
	jobs := s.scheduler.GetAll()
	statuses := make([]entity.JobProgress, len(jobs))
	for i, syncJob := range jobs {
		statuses[i] = syncJob.Status()
	}
	return statuses
}

// ListJobStatusesByStatus returns job statuses filtered by status
func (s *SyncService) ListJobStatusesByStatus(status entity.JobStatus) []entity.JobProgress {
	jobs := s.scheduler.GetByStatus(status)
	statuses := make([]entity.JobProgress, len(jobs))
	for i, syncJob := range jobs {
		statuses[i] = syncJob.Status()
	}
	return statuses
}

// StopJob stops a specific job
func (s *SyncService) StopJob(jobID string) error {
	ctx := context.Background()
	logger := s.logger.WithContext(ctx).Operation("stop_sync_job").
		WithString(JobID, jobID).
		Build()

	// Input validation (return ServiceError, no logging)
	if jobID == "" {
		return NewValidationError(ctx, "stop_sync_job", "invalid_input").
			WithContext("validation_error", "empty_job_id")
	}

	// Scheduler lookup
	logger.Step("scheduler_lookup").
		WithString(JobID, jobID).
		Log()

	if err := s.scheduler.StopJob(jobID); err != nil {
		return NewSyncJobError(ctx, "stop_sync_job", jobID, err)
	}

	logger.Step("sync job stopped successfully").
		WithString(JobID, jobID).
		Log()

	logger.Success().
		WithString(JobID, jobID).
		Log()

	return nil
}

// PauseJob pauses or resumes a specific job
func (s *SyncService) PauseJob(jobID string) error {
	ctx := context.Background()
	logger := s.logger.WithContext(ctx).Operation("pause_sync_job").
		WithString(JobID, jobID).
		Build()

	// Input validation (return ServiceError, no logging)
	if jobID == "" {
		return NewValidationError(ctx, "pause_sync_job", "invalid_input").
			WithContext("validation_error", "empty_job_id")
	}

	// Scheduler lookup
	logger.Step("scheduler_lookup").
		WithString(JobID, jobID).
		Log()

	if err := s.scheduler.PauseJob(jobID); err != nil {
		return NewSyncJobError(ctx, "pause_sync_job", jobID, err)
	}

	logger.Step("sync job pause/resume toggled successfully").
		WithString(JobID, jobID).
		Log()

	logger.Success().
		WithString(JobID, jobID).
		Log()

	return nil
}

// StopAllJobs stops all running jobs
func (s *SyncService) StopAllJobs() error {
	ctx := context.Background()
	logger := s.logger.WithContext(ctx).Operation("stop_all_sync_jobs").Build()

	// Get all active jobs (running and pending)
	logger.Step("get_active_jobs").Log()

	runningJobs := s.scheduler.GetByStatus(entity.StatusRunning)
	pendingJobs := s.scheduler.GetByStatus(entity.StatusPending)
	activeJobs := append(runningJobs, pendingJobs...)

	logger.Step("found active jobs to stop").
		WithInt("running_count", len(runningJobs)).
		WithInt("pending_count", len(pendingJobs)).
		WithInt("total_active", len(activeJobs)).
		Log()

	if len(activeJobs) == 0 {
		logger.Success().
			WithInt("stopped_count", 0).
			WithInt("active_count", 0).
			Log()
		return nil
	}

	// Stop each job
	logger.Step("stop_jobs").
		WithInt("job_count", len(activeJobs)).
		Log()

	var errors []error
	successCount := 0

	for _, syncJob := range activeJobs {
		if err := s.scheduler.StopJob(syncJob.GetID().String()); err != nil {
			// Collect error for return (handlers will log the error)
			errors = append(errors, NewSyncJobError(ctx, "stop_job", syncJob.GetID().String(), err))
		} else {
			successCount++
		}
	}


	logger.Step("bulk job stop completed").
		WithInt("total_jobs", len(activeJobs)).
		WithInt("success_count", successCount).
		WithInt("failed_count", len(errors)).
		Log()

	if len(errors) > 0 {
		return NewInternalError(ctx, "stop_all_sync_jobs", "multiple_job_failures", fmt.Errorf("failed to stop some jobs: %v", errors)).
			WithContext("failed_count", len(errors)).
			WithContext("total_count", len(activeJobs))
	}

	logger.Success().
		WithInt("stopped_count", len(activeJobs)).
		Log()

	return nil
}

// ClearFinishedJobs removes all completed, stopped, and failed jobs
func (s *SyncService) ClearFinishedJobs() error {
	ctx := context.Background()
	logger := s.logger.WithContext(ctx).Operation("clear_finished_sync_jobs").Build()

	// Get all finished jobs (completed, stopped, failed)
	logger.Step("get_finished_jobs").Log()

	completedJobs := s.scheduler.GetByStatus(entity.StatusCompleted)
	stoppedJobs := s.scheduler.GetByStatus(entity.StatusStopped)
	failedJobs := s.scheduler.GetByStatus(entity.StatusFailed)

	finishedJobs := append(completedJobs, stoppedJobs...)
	finishedJobs = append(finishedJobs, failedJobs...)

	logger.Step("found finished jobs to clear").
		WithInt("completed_count", len(completedJobs)).
		WithInt("stopped_count", len(stoppedJobs)).
		WithInt("failed_count", len(failedJobs)).
		WithInt("total_finished", len(finishedJobs)).
		Log()

	if len(finishedJobs) == 0 {
		logger.Success().
			WithInt("cleared_count", 0).
			Log()
		return nil
	}

	// Remove each finished job from scheduler
	logger.Step("clear_jobs").
		WithInt("job_count", len(finishedJobs)).
		Log()

	var errors []error
	successCount := 0

	for _, syncJob := range finishedJobs {
		if err := s.scheduler.Remove(syncJob.GetID().String()); err != nil {
			// Collect error for return (handlers will log the error)
			errors = append(errors, NewSyncJobError(ctx, "clear_job", syncJob.GetID().String(), err))
		} else {
			successCount++
		}
	}


	logger.Step("bulk job clear completed").
		WithInt("total_jobs", len(finishedJobs)).
		WithInt("success_count", successCount).
		WithInt("failed_count", len(errors)).
		Log()

	if len(errors) > 0 {
		return NewInternalError(ctx, "clear_finished_sync_jobs", "multiple_job_failures", fmt.Errorf("failed to clear some jobs: %v", errors)).
			WithContext("failed_count", len(errors)).
			WithContext("total_count", len(finishedJobs))
	}

	logger.Success().
		WithInt("cleared_count", len(finishedJobs)).
		Log()

	return nil
}

// Shutdown gracefully shuts down the sync service
func (s *SyncService) Shutdown() {
	ctx := context.Background()
	logger := s.logger.WithContext(ctx).Operation("shutdown").Build()

	logger.Step("shutting down sync service").Log()

	// Stop all running jobs
	logger.Step("stop_all_jobs").Log()

	if err := s.StopAllJobs(); err != nil {
		logger.Step("errors occurred while stopping jobs during shutdown").
			WithString("error", err.Error()).
			Log()
	}

	// Stop the scheduler
	logger.Step("stop_scheduler").Log()

	schedStart := time.Now()
	s.scheduler.Stop()
	_  = time.Since(schedStart)

	logger.Step("sync service shutdown complete").
		Log()

	logger.Success().
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
