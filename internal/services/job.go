package services

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

type Task[R any] func(ctx context.Context) entity.Result[R]

// Object contains general metadata about a job that doesn't change during execution
type Object struct {
	ID          uuid.UUID         // Unique identifier for the job
	reason      string            // Reason for job failure or completion details
	createdAt   time.Time         // When the job was created
	startedAt   *time.Time        // When the job started execution (nil if not started)
	completedAt *time.Time        // When the job completed (nil if not completed)
	total       int               // Total number of tasks in the job
	opts        map[string]string // Job configuration options and metadata
}

// SyncJob represents a sync job that executes pre-generated tasks
type SyncJob struct {
	Object
	tasks        *entity.LinkedList[Task[string]] // Queue of tasks to execute
	doneCh       chan bool                        // Channel to signal job cancellation
	stopResumeCh chan bool                        // Channel to handle pause/resume operations
	status       entity.JobStatus                 // Current job status
	results      []entity.JobResult               // Results from completed tasks
	logger       *logger.StructuredLogger         // Logger for job operations
	mu           sync.Mutex                       // Mutex for thread-safe access
}

func NewSyncJob(tasks *entity.LinkedList[Task[string]], opts map[string]string) (*SyncJob, error) {
	if opts == nil {
		opts = make(map[string]string)
	}

	job := &SyncJob{
		Object: Object{
			ID:        uuid.New(),
			createdAt: time.Now(),
			total:     tasks.Len(),
			opts:      opts,
		},
		tasks:   tasks,
		status:  entity.StatusPending,
		results: []entity.JobResult{},
		logger:  logger.NewDebugLogger("sync-job"),
	}

	return job, nil
}

// Start begins the job execution with pre-generated tasks
func (j *SyncJob) Start(ctx context.Context) error {
	j.doneCh = make(chan bool)
	j.stopResumeCh = make(chan bool)
	defer func() {
		j.stopResumeCh = nil
		j.doneCh = nil
	}()

	// Start the job
	now := time.Now()
	j.startedAt = &now
	j.status = entity.StatusRunning

	defer func() {
		// Stop the job
		now := time.Now()
		j.completedAt = &now
		j.status = entity.StatusCompleted
	}()

	tracer := j.logger.StartOperation("process_tasks").
		WithString("job_id", j.ID.String()).
		WithInt("tasks_count", j.tasks.Len()).
		Build()

	next, _, cancel := entity.TaskIterator(j.iter(ctx, j.tasks))
	taskIndex := 0
	for {
		start := time.Now()
		result, ok := next()

		if result.Err != nil {
			j.logger.BusinessLogic("task_failed").
				WithString("job_id", j.ID.String()).
				WithInt("task_index", taskIndex).
				WithString("error", result.Err.Error()).
				Log()
		}

		end := time.Now()

		// Record result directly in job
		j.results = append(j.results, entity.JobResult{
			Result:      result.Data,
			Err:         result.Err,
			StartedAt:   start,
			CompletedAt: end,
		})

		if !ok {
			break
		}

		taskIndex++

		select {
		case <-ctx.Done():
			cancel()
			return ctx.Err()
		case <-j.doneCh:
			cancel()
			j.doneCh <- true
			return nil
		case <-j.stopResumeCh:
			<-j.stopResumeCh
		default:
		}
	}

	tracer.Success().
		WithInt("completed_tasks", taskIndex).
		Log()

	return nil
}

func (j *SyncJob) GetID() uuid.UUID {
	return j.ID
}

func (j *SyncJob) Pause() {
	if j.stopResumeCh == nil {
		return
	}

	switch j.status {
	case entity.StatusPause:
		j.status = entity.StatusRunning
	case entity.StatusRunning:
		j.status = entity.StatusPause
	}
	j.stopResumeCh <- true
}

func (j *SyncJob) Cancel() error {
	j.status = entity.StatusStopping
	defer func() {
		// Stop the job
		now := time.Now()
		j.completedAt = &now
		j.status = entity.StatusStopped
	}()
	if j.doneCh == nil {
		return nil
	}

	j.doneCh <- true
	<-j.doneCh

	j.logger.BusinessLogic("job_cancelled").
		WithString("job_id", j.ID.String()).
		Log()

	return nil
}

func (j *SyncJob) Status() entity.JobProgress {
	j.mu.Lock()
	defer j.mu.Unlock()

	return entity.JobProgress{
		Id:          j.ID,
		Path:        j.opts["path"],
		Status:      j.status,
		CreatedAt:   j.createdAt,
		Total:       j.total,
		Remaining:   j.tasks.Len(),
		StartedAt:   j.startedAt,
		CompletedAt: j.completedAt,
		Results:     j.results,
		Reason:      j.reason,
	}
}

func (j *SyncJob) Metadata() map[string]string {
	return j.opts
}

func (j *SyncJob) iter(ctx context.Context, tasks *entity.LinkedList[Task[string]]) func(yield func(result entity.Result[string]) bool) {
	return func(yield func(result entity.Result[string]) bool) {
		for tasks.Len() > 0 {
			task, _ := tasks.Pop()
			r := task(ctx)
			yield(r)
		}
	}
}
