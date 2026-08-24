package types

type UploadUrlRequest struct {
	FileName     string `json:"fileName"`
	ContentType  string `json:"contentType"`
	FileSize     int64  `json:"fileSize"`
	ResourceType string `json:"resourceType"`
}
type UploadUrlResponse struct {
	UploadUrl  string `json:"uploadUrl"`
	ObjectKey  string `json:"objectKey"`
	ExpiresIn  int64  `json:"expiresIn"`
	ResourceID string `json:"resourceID"`
}

type JobType string

const (
	VideoTranscode JobType = "VIDEO_TRANSCODE"
	ImageResize    JobType = "IMAGE_RESIZE"
)

type JobStatus string

const (
	JobQueued    JobStatus = "QUEUED"
	JobRunning   JobStatus = "RUNNING"
	JobCompleted JobStatus = "COMPLETED"
	JobFailed    JobStatus = "FAILED"
	JobCancelled JobStatus = "CANCELLED"
	JobRetrying  JobStatus = "RETRYING"
)

type EventType string

const (
	EventJobCreated   EventType = "JobCreated"
	EventJobCompleted EventType = "JobCompleted"
	EventJobFailed    EventType = "JobFailed"
)

type ResourceType string

const (
	ResourceVideo    ResourceType = "VIDEO"
	ResourceImage    ResourceType = "IMAGE"
	ResourceDocument ResourceType = "DOCUMENT"
)
