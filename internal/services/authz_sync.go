// Package services provides authorization-wrapped sync service implementations.
// This file contains the AuthzSyncService which wraps SyncService with authorization checks.
package services

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/user"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

// AuthzSyncService wraps SyncService with authorization checks.
// All sync operations require entity.SyncPermission on entity.LocalDatastore resource.
type AuthzSyncService struct {
	syncSrv  *SyncService
	authzSrv Authz
	logger   *logger.StructuredLogger
}

// NewAuthzSyncService creates a new authorization-wrapped sync service.
// It requires an authorization service, album service, media service, and filesystem datastore.
func NewAuthzSyncService(authzSrv Authz, albumService *AlbumService, mediaService *MediaService, fsDatastore *fs.Datastore) *AuthzSyncService {
	return &AuthzSyncService{
		syncSrv:  NewSyncService(albumService, mediaService, fsDatastore),
		authzSrv: authzSrv,
		logger:   logger.New("authz_sync_service"),
	}
}

// StartSync starts a new sync job for the given album path.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) StartSync(ctx context.Context, albumPath string) (string, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_start_sync").
		WithString(AlbumPath, albumPath).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return "", err
	}

	if !hasPermission {
		return "", NewForbiddenAccessError(ctx, "authz_start_sync", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "start_sync").Log()
	return s.syncSrv.StartSync(ctx, albumPath)
}

// GetJobStatus returns the status of a specific sync job.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) GetJobStatus(ctx context.Context, jobID string) (*entity.JobProgress, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_get_job_status").
		WithString(JobID, jobID).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_get_job_status", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "get_job_status").Log()
	return s.syncSrv.GetJobStatus(ctx, jobID)
}

// ListJobStatuses returns statuses of all sync jobs.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) ListJobStatuses(ctx context.Context) ([]entity.JobProgress, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_list_job_statuses").Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_list_job_statuses", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "list_job_statuses").Log()
	return s.syncSrv.ListJobStatuses(ctx)
}

// ListJobStatusesByStatus returns sync job statuses filtered by status.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) ListJobStatusesByStatus(ctx context.Context, status entity.JobStatus) ([]entity.JobProgress, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_list_job_statuses_by_status").
		WithParam("status", status).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_list_job_statuses_by_status", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "list_job_statuses_by_status").Log()
	return s.syncSrv.ListJobStatusesByStatus(ctx, status)
}

// StopJob stops a specific sync job.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) StopJob(ctx context.Context, jobID string) error {
	logger := s.logger.WithContext(ctx).Debug("authz_stop_job").
		WithString(JobID, jobID).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return err
	}

	if !hasPermission {
		return NewForbiddenAccessError(ctx, "authz_stop_job", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "stop_job").Log()
	return s.syncSrv.StopJob(ctx, jobID)
}

// PauseJob pauses or resumes a specific sync job.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) PauseJob(ctx context.Context, jobID string) error {
	logger := s.logger.WithContext(ctx).Debug("authz_pause_job").
		WithString(JobID, jobID).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return err
	}

	if !hasPermission {
		return NewForbiddenAccessError(ctx, "authz_pause_job", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "pause_job").Log()
	return s.syncSrv.PauseJob(ctx, jobID)
}

// StopAllJobs stops all running sync jobs.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) StopAllJobs(ctx context.Context) error {
	logger := s.logger.WithContext(ctx).Debug("authz_stop_all_jobs").Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return err
	}

	if !hasPermission {
		return NewForbiddenAccessError(ctx, "authz_stop_all_jobs", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "stop_all_jobs").Log()
	return s.syncSrv.StopAllJobs(ctx)
}

// ClearFinishedJobs removes all completed, stopped, and failed sync jobs.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) ClearFinishedJobs(ctx context.Context) error {
	logger := s.logger.WithContext(ctx).Debug("authz_clear_finished_jobs").Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return err
	}

	if !hasPermission {
		return NewForbiddenAccessError(ctx, "authz_clear_finished_jobs", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "clear_finished_jobs").Log()
	return s.syncSrv.ClearFinishedJobs(ctx)
}

// Shutdown gracefully shuts down the sync service.
// This is an admin/lifecycle operation that doesn't require authorization.
func (s *AuthzSyncService) Shutdown() {
	ctx := context.Background()
	logger := s.logger.WithContext(ctx).Debug("authz_shutdown").Build()

	// Shutdown doesn't require authorization - it's an admin/lifecycle operation
	logger.Step("delegating to sync service").Log()
	s.syncSrv.Shutdown()
}

// IsAlbumSyncing checks if there's an active sync job for the given album.
// This is a read-only status check that doesn't require authorization.
func (s *AuthzSyncService) IsAlbumSyncing(albumID string) bool {
	return s.syncSrv.IsAlbumSyncing(albumID)
}

// GetSchedulerStats returns statistics about the sync job scheduler.
// Requires entity.SyncPermission on entity.LocalDatastore.
func (s *AuthzSyncService) GetSchedulerStats(ctx context.Context) (map[string]int, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_get_scheduler_stats").Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_sync_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_get_scheduler_stats", entity.NewDatastoreResource(entity.LocalDatastore), entity.SyncPermission)
	}

	logger.Step("authorization granted").WithString("method", "get_scheduler_stats").Log()
	return s.syncSrv.GetSchedulerStats(), nil
}
