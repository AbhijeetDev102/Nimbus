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
}
