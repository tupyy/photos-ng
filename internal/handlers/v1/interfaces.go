package v1

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
)

type AlbumService interface {
	List(ctx context.Context, opts *services.ListOptions) ([]entity.Album, error)
	Count(ctx context.Context, hasParent bool) (int, error)
	Get(ctx context.Context, id string) (*entity.Album, error)
	Create(ctx context.Context, album entity.Album) (*entity.Album, error)
	Update(ctx context.Context, album entity.Album) (*entity.Album, error)
	Delete(ctx context.Context, id string) error
}

type MediaService interface {
	List(ctx context.Context, filter *services.MediaOptions) ([]entity.Media, *services.PaginationCursor, error)
	Get(ctx context.Context, id string) (*entity.Media, error)
	WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error)
	Update(ctx context.Context, media entity.Media) (*entity.Media, error)
	Delete(ctx context.Context, id string) error
}

type SyncService interface {
	StartSync(ctx context.Context, albumPath string) (string, error)
	GetJobStatus(ctx context.Context, jobID string) (*entity.JobProgress, error)
	ListJobStatuses(ctx context.Context) ([]entity.JobProgress, error)
	ListJobStatusesByStatus(ctx context.Context, status entity.JobStatus) ([]entity.JobProgress, error)
	StopJob(ctx context.Context, jobID string) error
	PauseJob(ctx context.Context, jobID string) error
	StopAllJobs(ctx context.Context) error
	ClearFinishedJobs(ctx context.Context) error
	IsAlbumSyncing(albumID string) bool
}
