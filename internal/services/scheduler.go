package services

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	DEFAULT_KEEP_PERIOD          = 3600 // seconds
	MAX_NUMBER_RUNNING_PIPELINES = 2
)

var (
	once      sync.Once
	scheduler *Scheduler
)

type Job interface {
	GetId() uuid.UUID
	Start(ctx context.Context) error
	Stop() error
	Status() JobProgress
}

type Scheduler struct {
	m     sync.Mutex
	queue *list.List
	done  chan chan struct{}
}

func GetScheduler() *Scheduler {
	once.Do(func() {
		scheduler = &Scheduler{queue: list.New()}
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					scheduler.run()
				case d := <-scheduler.done:
					d <- struct{}{}
					return
				}
			}
		}()
	})
	return scheduler
}

func (s *Scheduler) Add(j Job) error {
	s.m.Lock()
	defer s.m.Unlock()

	s.queue.PushBack(j)
	return nil
}

func (s *Scheduler) Get(id string) Job {
	jobs := s.find(func(j Job) bool {
		return j.GetId().String() == id
	})
	if len(jobs) > 0 {
		return jobs[0]
	}
	return nil
}

func (s *Scheduler) GetByStatus(status JobStatus) []Job {
	return s.find(func(j Job) bool { return j.Status().Status == status })
}

func (s *Scheduler) GetAll() []Job {
	return s.find(func(j Job) bool { return true })
}

func (s *Scheduler) Stop() {
	runningJobs := s.GetByStatus(StatusRunning)
	for _, j := range runningJobs {
		j.Stop()
	}

	d := make(chan struct{})
	s.done <- d
	<-d
}

func (s *Scheduler) countByStatus(status JobStatus) int {
	return len(s.GetByStatus(status))
}

func (s *Scheduler) find(f func(j Job) bool) []Job {
	s.m.Lock()
	defer s.m.Unlock()

	jobs := []Job{}
	element := s.queue.Front()
	for {
		if element == nil {
			break
		}
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
	for {
		if e == nil {
			break
		}
		j := e.Value.(Job)
		switch j.Status().Status {
		case StatusPending:
			if s.countByStatus(StatusPending) < MAX_NUMBER_RUNNING_PIPELINES {
				go func(job Job) {
					ctx := context.Background()
					job.Start(ctx)
				}(j)
			}
		case StatusCompleted:
		case StatusStopped:
		case StatusFailed:
			if j.Status().CompletedAt.Add(DEFAULT_KEEP_PERIOD * time.Second).Before(time.Now()) {
				s.m.Lock()
				_ = s.queue.Remove(e)
				s.m.Unlock()
				goto start
			}
		}
		e = e.Next()
	}
}
