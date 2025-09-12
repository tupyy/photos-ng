package services

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

const (
	defaultKeepPeriod   = 600 // seconds
	maxRunningPipelines = 1
)

var (
	once      sync.Once
	scheduler *Scheduler
)

type Job interface {
	GetID() uuid.UUID
	Start(ctx context.Context) error
	Pause()
	Cancel() error
	Status() entity.JobProgress
	Metadata() map[string]string
}

type Scheduler struct {
	m          sync.Mutex
	queue      *list.List
	done       chan chan struct{}
	logger     *logger.DebugLogger
	infoLogger *logger.StructuredLogger
}

func GetScheduler() *Scheduler {
	once.Do(func() {
		debugLogger := logger.NewDebugLogger("scheduler")
		tracer := debugLogger.StartOperation("initialize_scheduler").Build()

		infoLogger := logger.NewInfoLogger("scheduler")

		scheduler = &Scheduler{
			queue:      list.New(),
			logger:     debugLogger,
			infoLogger: infoLogger,
		}

		tracer.Success().
			WithInt("tick_interval_seconds", 2).
			WithInt("max_running_pipelines", maxRunningPipelines).
			WithInt("default_keep_period_seconds", defaultKeepPeriod).
			Log()

		go func() {
			scheduler.logger.BusinessLogic("starting_background_worker").Log()
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					scheduler.run()
				case d := <-scheduler.done:
					scheduler.logger.BusinessLogic("received_shutdown_signal").Log()
					d <- struct{}{}
					return
				}
			}
		}()
	})
	return scheduler
}

func (s *Scheduler) Add(j Job) error {
	tracer := s.logger.StartOperation("add_job").
		WithString(JobID, j.GetID().String()).
		Build()

	s.m.Lock()
	defer s.m.Unlock()

	s.queue.PushBack(j)

	tracer.Success().
		WithInt("queue_length", s.queue.Len()).
		Log()

	return nil
}

func (s *Scheduler) GetByAlbumID(albumID string) []Job {
	tracer := s.logger.StartOperation("get_jobs_by_album_id").
		WithString("album_id", albumID).
		Build()

	jobs := s.find(func(j Job) bool {
		v, ok := j.Metadata()["albumId"]
		if ok {
			return v == albumID
		}
		return false
	})

	tracer.Success().
		WithInt("found_jobs", len(jobs)).
		Log()

	return jobs
}

func (s *Scheduler) Get(id string) Job {
	tracer := s.logger.StartOperation("get_job_by_id").
		WithString(JobID, id).
		Build()

	jobs := s.find(func(j Job) bool {
		return j.GetID().String() == id
	})

	found := len(jobs) > 0
	tracer.Success().
		WithBool("found", found).
		Log()

	if found {
		return jobs[0]
	}
	return nil
}

func (s *Scheduler) GetByStatus(status entity.JobStatus) []Job {
	return s.find(func(j Job) bool { return j.Status().Status == status })
}

func (s *Scheduler) GetAll() []Job {
	return s.find(func(j Job) bool { return true })
}

func (s *Scheduler) StopJob(id string) error {
	job := s.Get(id)
	if job == nil {
		return nil
	}

	// Log info-level message for job state change only
	s.infoLogger.BusinessLogic("job_stopped").
		WithString(JobID, id).
		WithString("previous_status", string(job.Status().Status)).
		Log()

	err := job.Cancel()
	if err != nil {
		zap.S().Errorw("Failed to stop job", JobID, id, "error", err)
	}

	return err
}

func (s *Scheduler) PauseJob(id string) error {
	job := s.Get(id)
	if job == nil {
		return nil
	}

	currentStatus := job.Status().Status
	job.Pause()
	newStatus := job.Status().Status

	// Log info-level message only when status actually changed
	if currentStatus != newStatus {
		s.infoLogger.BusinessLogic("job_status_changed").
			WithString(JobID, id).
			WithString("from", string(currentStatus)).
			WithString("to", string(newStatus)).
			Log()
	}

	return nil
}

func (s *Scheduler) Remove(id string) error {
	tracer := s.logger.StartOperation("remove_job").
		WithString(JobID, id).
		Build()

	s.m.Lock()
	defer s.m.Unlock()

	element := s.queue.Front()
	for element != nil {
		j := element.Value.(Job)
		if j.GetID().String() == id {
			s.queue.Remove(element)

			tracer.Success().
				WithBool("job_found", true).
				WithInt("queue_length", s.queue.Len()).
				Log()

			return nil
		}
		element = element.Next()
	}

	tracer.Success().
		WithBool("job_found", false).
		WithInt("queue_length", s.queue.Len()).
		Log()

	return nil // Job not found, but that's ok
}

func (s *Scheduler) Stop() {
	jobs := s.GetAll()

	s.infoLogger.BusinessLogic("scheduler_shutting_down").
		WithInt("total_jobs", len(jobs)).
		Log()

	for _, j := range jobs {
		j.Cancel()
	}

	d := make(chan struct{})
	s.done <- d
	<-d
}

func (s *Scheduler) Pause() {
	jobs := s.GetAll()

	s.infoLogger.BusinessLogic("pausing_all_jobs").
		WithInt("total_jobs", len(jobs)).
		Log()

	for _, j := range jobs {
		j.Pause()
	}
}

func (s *Scheduler) countByStatus(status entity.JobStatus) int {
	return len(s.GetByStatus(status))
}

func (s *Scheduler) find(f func(j Job) bool) []Job {
	s.m.Lock()
	defer s.m.Unlock()

	jobs := []Job{}
	element := s.queue.Front()
	for element != nil {
		p := element.Value.(Job)
		if f(p) {
			jobs = append(jobs, element.Value.(Job))
		}
		element = element.Next()
	}

	return jobs
}

func (s *Scheduler) run() {
start:
	e := s.queue.Front()
	for e != nil {
		j := e.Value.(Job)

		switch j.Status().Status {
		case entity.StatusPending:
			currentRunning := s.countByStatus(entity.StatusRunning)
			if currentRunning < maxRunningPipelines {
				// Log info-level message for job state change
				s.infoLogger.BusinessLogic("starting_job").
					WithString(JobID, j.GetID().String()).
					WithInt("running_jobs", currentRunning).
					Log()

				go func(job Job) {
					ctx := context.Background()
					job.Start(ctx)
				}(j)
			}

			return
		case entity.StatusCompleted:
			fallthrough
		case entity.StatusStopped:
			fallthrough
		case entity.StatusFailed:
			completedAt := j.Status().CompletedAt
			if completedAt != nil && completedAt.Add(defaultKeepPeriod*time.Second).Before(time.Now()) {
				s.infoLogger.BusinessLogic("cleaning_up_old_job").
					WithString(JobID, j.GetID().String()).
					WithString(Status, string(j.Status().Status)).
					WithInt("age_seconds", int(time.Since(*completedAt).Seconds())).
					Log()

				s.m.Lock()
				_ = s.queue.Remove(e)
				s.m.Unlock()
				goto start
			}
		}
		e = e.Next()
	}
}
