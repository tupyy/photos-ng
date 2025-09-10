package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

type Task[R any] func(ctx context.Context) entity.Result[R]

// SyncAlbumJob represents a sync job that executes pre-generated tasks
type SyncAlbumJob struct {
	ID            uuid.UUID
	albumID       string
	albumPath     string
	albumTasks    []Task[entity.Album]
	mediaTasks    []Task[entity.Media]
	progressMeter *progressMeter
	doneCh        chan bool
}

func NewSyncJob(albumID string, albumPath string, albumTasks []Task[entity.Album], mediaTasks []Task[entity.Media]) (*SyncAlbumJob, error) {
	job := &SyncAlbumJob{
		ID:            uuid.New(),
		albumID:       albumID,
		albumPath:     albumPath,
		albumTasks:    albumTasks,
		mediaTasks:    mediaTasks,
		progressMeter: newProgressMeter(),
	}

	return job, nil
}

// Start begins the job execution with pre-generated tasks
func (j *SyncAlbumJob) Start(ctx context.Context) error {
	j.doneCh = make(chan bool)
	j.progressMeter.Start()

	defer func() {
		j.progressMeter.Stop(entity.StatusCompleted)
	}()

	// Set total task count from pre-generated tasks
	j.progressMeter.Total(len(j.albumTasks) + len(j.mediaTasks))

	zap.S().Debugw("creating albums", "job_id", j.ID, "tasks_count", len(j.albumTasks))

	// Execute album tasks
	for _, t := range j.albumTasks {
		start := time.Now()
		r := t(ctx)
		if r.Err != nil {
			zap.S().Warnw("failed to create album", "job_id", j.ID, "error", r.Err)
		}
		end := time.Now()

		j.progressMeter.Result(entity.JobResult{
			Result:      fmt.Sprintf("Album %s processed", r.Data.Path),
			Err:         r.Err,
			StartedAt:   start,
			CompletedAt: end,
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-j.doneCh:
			j.doneCh <- true
			return nil
		default:
		}
	}

	zap.S().Debugw("processing media", "job_id", j.ID, "tasks_count", len(j.mediaTasks))

	// Execute media tasks
	for _, t := range j.mediaTasks {
		start := time.Now()
		r := t(ctx)
		if r.Err != nil {
			zap.S().Warnw("failed to process media", "job_id", j.ID, "error", r.Err)
		}
		end := time.Now()

		j.progressMeter.Result(entity.JobResult{
			Result:      fmt.Sprintf("Media %s processed", r.Data.Filepath()),
			Err:         r.Err,
			StartedAt:   start,
			CompletedAt: end,
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-j.doneCh:
			j.doneCh <- true
			return nil
		default:
		}
	}

	return nil
}

func (j *SyncAlbumJob) GetID() uuid.UUID {
	return j.ID
}

func (j *SyncAlbumJob) GetAlbumID() string {
	return j.albumID
}

func (j *SyncAlbumJob) GetAlbumPath() string {
	return j.albumPath
}

// Stop cancels the job execution
func (j *SyncAlbumJob) Stop() error {
	j.progressMeter.status = entity.StatusStopping
	defer func() {
		j.progressMeter.Stop(entity.StatusStopped)
	}()
	if j.doneCh == nil {
		return nil
	}

	j.doneCh <- true
	<-j.doneCh

	zap.S().Infow("Job stopped", "id", j.ID)
	return nil
}

func (j *SyncAlbumJob) Status() entity.JobProgress {
	return j.progressMeter.Status(j.ID, j.albumPath)
}

type progressMeter struct {
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
	status      entity.JobStatus
	total       int
	remaining   int
	results     []entity.JobResult
	reason      string
	mu          sync.Mutex
}

func newProgressMeter() *progressMeter {
	return &progressMeter{
		createdAt: time.Now(),
		status:    entity.StatusPending,
		results:   []entity.JobResult{},
	}
}

func (j *progressMeter) Start() {
	now := time.Now()
	j.startedAt = &now
	j.status = entity.StatusRunning
}

func (j *progressMeter) Total(total int) {
	j.total, j.remaining = total, total
}

func (j *progressMeter) Stop(s entity.JobStatus) {
	now := time.Now()
	j.completedAt = &now
	j.status = s
}

func (j *progressMeter) Failed(err error) {
	now := time.Now()
	j.completedAt = &now
	j.reason = err.Error()
	j.status = entity.StatusFailed
}

func (j *progressMeter) Result(r entity.JobResult) {
	j.remaining--
	j.results = append(j.results, r)
}

func (j *progressMeter) Status(id uuid.UUID, path string) entity.JobProgress {
	j.mu.Lock()
	defer j.mu.Unlock()

	p := entity.JobProgress{
		Id:          id,
		Path:        path,
		Status:      j.status,
		CreatedAt:   j.createdAt,
		Total:       j.total,
		Remaining:   j.remaining,
		StartedAt:   j.startedAt,
		CompletedAt: j.completedAt,
		Results:     j.results,
		Reason:      j.reason,
	}

	return p
}
