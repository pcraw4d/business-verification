package jobs

import (
	"context"
	"time"
)

// JobStatus represents the status of an analysis job
type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
	StatusSkipped    JobStatus = "skipped"
)

// Job represents a generic analysis job
type Job interface {
	GetID() string
	GetMerchantID() string
	GetType() string
	GetStatus() JobStatus
	SetStatus(status JobStatus)
	Process(ctx context.Context) error
}

// JobResult represents the result of processing a job
type JobResult struct {
	JobID      string
	MerchantID string
	Type       string
	Status     JobStatus
	Result     interface{}
	Error      error
	Duration   time.Duration
	CompletedAt time.Time
}

