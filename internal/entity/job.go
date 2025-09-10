package entity

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusStopped   JobStatus = "stopped"
	StatusStopping  JobStatus = "stopping"
)

// JobProgress tracks the progress of a job
type JobProgress struct {
	Id          uuid.UUID
	Status      JobStatus
	Reason      string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Total       int
	Remaining   int
	Results     []JobResult
	Path        string // The folder path being synchronized
}

// JobResult represents the result of processing a single task within a job
type JobResult struct {
	Result      string
	Err         error
	StartedAt   time.Time
	CompletedAt time.Time
}
