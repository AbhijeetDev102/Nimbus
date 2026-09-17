package types

import "encoding/json"

type CreateJobRequest struct {
	ResourceID string          `json:"resourceID"`
	JobType    string          `json:"jobType"`
	Parameters json.RawMessage `json:"parameters"` // Captures the exact JSON bytes!
}

type CreateJobResponse struct {
	JobId  string `json:"jobID"`
	Status string `json:"status"`
}

type GetJobResponse struct {
	JobID            string          `json:"jobId"`
	ResourceID       *string         `json:"resourceId,omitempty"`
	JobType          string          `json:"jobType"`
	Status           string          `json:"status"`
	RetryCount       int32           `json:"retryCount"`
	MaxRetries       int32           `json:"maxRetries"`
	ErrorMessage     *string         `json:"errorMessage,omitempty"`
	OutputResourceID *string         `json:"outputResourceId,omitempty"`
	Parameters       json.RawMessage `json:"parameters,omitempty"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	CreatedAt        string          `json:"createdAt"`
	StartedAt        *string         `json:"startedAt,omitempty"`
	CompletedAt      *string         `json:"completedAt,omitempty"`
	WorkerID         *string         `json:"workerId,omitempty"`
	LeaseExpiresAt   *string         `json:"leaseExpiresAt,omitempty"`
}

type ListJobsResponse struct {
	Jobs       []GetJobResponse `json:"jobs"`
	TotalCount int64            `json:"totalCount"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
}

type JobStatsResponse struct {
	Total     int64 `json:"total"`
	Queued    int64 `json:"queued"`
	Running   int64 `json:"running"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
}

type WorkerInfo struct {
	WorkerID      string  `json:"workerId"`
	Hostname      string  `json:"hostname"`
	Status        string  `json:"status"` // "ACTIVE", "BUSY", "IDLE", "OFFLINE"
	CurrentJobID  *string `json:"currentJobId,omitempty"`
	LastHeartbeat string  `json:"lastHeartbeat"`
	StartedAt     string  `json:"startedAt"`
}

type ListWorkersResponse struct {
	Workers    []WorkerInfo `json:"workers"`
	TotalCount int          `json:"totalCount"`
	ActiveCount int         `json:"activeCount"`
}

