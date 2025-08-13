package services

import (
	"context"
	"fmt"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"

	"go.uber.org/zap"
)

// SyncService manages sync operations using a single scheduler instance
type SyncService struct {
	albumService *AlbumService
	mediaService *MediaService
	fsDatastore  *fs.Datastore
	scheduler    *Scheduler
}

// NewSyncService creates a new sync service that manages the job scheduler
func NewSyncService(albumService *AlbumService, mediaService *MediaService, fsDatastore *fs.Datastore) *SyncService {
	return &SyncService{
		albumService: albumService,
		mediaService: mediaService,
		fsDatastore:  fsDatastore,
		scheduler:    GetScheduler(),
	}
}

// StartSync starts a new sync job for the given album path
// Returns the job ID and any error that occurred during job creation
func (s *SyncService) StartSync(ctx context.Context, albumPath string) (string, error) {
	// Create a root album entity for the sync operation (the job will handle creation if needed)
	rootAlbum := entity.NewAlbum(albumPath)

	// Create the sync job - it will handle album creation during execution
	syncJob, err := NewSyncJob(rootAlbum, s.albumService, s.mediaService, s.fsDatastore)
	if err != nil {
		zap.S().Errorw("failed to create sync job", "path", albumPath, "error", err)
		return "", fmt.Errorf("failed to create sync job: %w", err)
	}

	// Add job to scheduler
	if err := s.scheduler.Add(syncJob); err != nil {
		zap.S().Errorw("failed to add job to scheduler", "job_id", syncJob.GetId(), "error", err)
		return "", fmt.Errorf("failed to schedule job: %w", err)
	}

	jobID := syncJob.GetId().String()
	zap.S().Infow("sync job created and scheduled", "job_id", jobID, "path", albumPath)

	return jobID, nil
}

// GetSyncJobStatus returns the status of a sync job by ID
func (s *SyncService) GetSyncJobStatus(jobID string) (*JobProgress, error) {

	syncJob := s.scheduler.Get(jobID)
	if syncJob == nil {
		return nil, fmt.Errorf("sync job not found: %s", jobID)
	}

	status := syncJob.Status()
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

	syncJob := s.scheduler.Get(jobID)
	if syncJob == nil {
		return fmt.Errorf("sync job not found: %s", jobID)
	}

	if err := syncJob.Stop(); err != nil {
		zap.S().Errorw("failed to stop sync job", "job_id", jobID, "error", err)
		return fmt.Errorf("failed to stop sync job: %w", err)
	}

	zap.S().Infow("sync job stopped", "job_id", jobID)
	return nil
}

// StopAllSyncJobs stops all running sync jobs
func (s *SyncService) StopAllSyncJobs() error {

	runningJobs := s.scheduler.GetByStatus(StatusRunning)
	var errors []error

	for _, syncJob := range runningJobs {
		if err := syncJob.Stop(); err != nil {
			zap.S().Errorw("failed to stop sync job", "job_id", syncJob.GetId(), "error", err)
			errors = append(errors, fmt.Errorf("failed to stop job %s: %w", syncJob.GetId(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop some jobs: %v", errors)
	}

	zap.S().Infow("all sync jobs stopped", "count", len(runningJobs))
	return nil
}

// Shutdown gracefully shuts down the sync service
func (s *SyncService) Shutdown() {

	zap.S().Info("shutting down sync service")

	// Stop all running jobs
	if err := s.StopAllSyncJobs(); err != nil {
		zap.S().Warnw("errors occurred while stopping jobs during shutdown", "error", err)
	}

	// Stop the scheduler
	s.scheduler.Stop()

	zap.S().Info("sync service shutdown complete")
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
