package services

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
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
	GetAlbumID() string
	Start(ctx context.Context) error
	Stop() error
	Status() entity.JobProgress
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

func (s *Scheduler) GetByAlbumID(albumID string) []Job {
	return s.find(func(j Job) bool {
		return j.GetAlbumID() == albumID
	})
}

func (s *Scheduler) Get(id string) Job {
	jobs := s.find(func(j Job) bool {
		return j.GetID().String() == id
	})
	if len(jobs) > 0 {
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
	return job.Stop()
}

func (s *Scheduler) Remove(id string) error {
	s.m.Lock()
	defer s.m.Unlock()

	element := s.queue.Front()
	for element != nil {
		j := element.Value.(Job)
		if j.GetID().String() == id {
			s.queue.Remove(element)
			return nil
		}
		element = element.Next()
	}
	return nil // Job not found, but that's ok
}

func (s *Scheduler) Stop() {
	jobs := s.GetAll()
	for _, j := range jobs {
		j.Stop()
	}

	d := make(chan struct{})
	s.done <- d
	<-d
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
			if s.countByStatus(entity.StatusRunning) < maxRunningPipelines {
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
			if j.Status().CompletedAt.Add(defaultKeepPeriod * time.Second).Before(time.Now()) {
				s.m.Lock()
				_ = s.queue.Remove(e)
				s.m.Unlock()
				goto start
			}
		}
		e = e.Next()
	}
}
