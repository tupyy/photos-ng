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
)

type TaskResult struct {
	ItemType string
	Name     string
	Err      error
}

// JobProgress tracks the progress of a job
type JobProgress struct {
	Id          uuid.UUID
	Status      JobStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Total       int
	Remaining   int
	Completed   []TaskResult
}
