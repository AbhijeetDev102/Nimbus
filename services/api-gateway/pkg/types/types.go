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
