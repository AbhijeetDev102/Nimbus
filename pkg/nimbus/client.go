package nimbus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client is the official Go SDK client for interacting with the Nimbus API Gateway
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Nimbus API Gateway client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type UploadURLResponse struct {
	ResourceID string `json:"resourceID"`
	UploadURL  string `json:"uploadUrl"`
}

type DownloadURLResponse struct {
	DownloadURL string `json:"download_url"`
}

// 1. GetUploadURL generates a Presigned S3 PUT URL for direct-to-storage uploads
func (c *Client) GetUploadURL(ctx context.Context, filename, contentType string) (*UploadURLResponse, error) {
	reqBody, err := json.Marshal(map[string]string{
		"fileName":    filename,
		"contentType": contentType,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/resource/upload-url", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to get upload url, status code: %d", resp.StatusCode)
	}

	var result UploadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// 2. SubmitJob queues a new task into the Nimbus Transactional Outbox
func (c *Client) SubmitJob(ctx context.Context, jobType JobType, resourceID *uuid.UUID, parameters any) (*Job, error) {
	payload := map[string]any{
		"jobType":    jobType,
		"parameters": parameters,
	}
	if resourceID != nil && *resourceID != uuid.Nil {
		payload["resourceID"] = resourceID.String()
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/jobs/create", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to submit job, status code: %d", resp.StatusCode)
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, err
	}
	return &job, nil
}

// 3. GetJob queries the latest status, metadata, and error details of a job
func (c *Client) GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/jobs/%s", c.baseURL, jobID.String()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get job, status code: %d", resp.StatusCode)
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, err
	}
	return &job, nil
}

// 4. GetDownloadURL generates a Presigned S3 GET URL to download processed media
func (c *Client) GetDownloadURL(ctx context.Context, resourceID uuid.UUID) (*DownloadURLResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/resource/%s/download", c.baseURL, resourceID.String()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get download url, status code: %d", resp.StatusCode)
	}

	var result DownloadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// 5. StreamProgress connects to the real-time WebSocket and streams live progress updates
func (c *Client) StreamProgress(ctx context.Context, jobID uuid.UUID, onUpdate func(update ProgressUpdate)) error {
	wsURL := strings.Replace(c.baseURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	fullURL := fmt.Sprintf("%s/ws/jobs/%s", wsURL, jobID.String())

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to progress websocket: %w", err)
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var update ProgressUpdate
			if err := conn.ReadJSON(&update); err != nil {
				return nil // Connection closed by server on job completion
			}
			if onUpdate != nil {
				onUpdate(update)
			}
		}
	}
}
