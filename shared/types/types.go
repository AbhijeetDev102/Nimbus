package types

type UploadUrlRequest struct {
	FileName     string `json:"fileName"`
	ContentType  string `json:"contentType"`
	FileSize     int64  `json:"fileSize"`
	ResourceType string `json:"resourceType"`
}
type UploadUrlResponse struct {
	UploadUrl string `json:"uploadUrl"`
	ObjectKey string `json:"objectKey"`
	ExpiresIn int64  `json:"expiresIn"`
}
