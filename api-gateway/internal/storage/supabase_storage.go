package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Uploader uploads report artifacts to object storage.
type Uploader interface {
	Upload(ctx context.Context, bucket, objectPath, contentType string, payload []byte) error
}

type SupabaseUploader struct {
	baseURL        string
	serviceRoleKey string
	httpClient     *http.Client
}

func NewSupabaseUploader(baseURL, serviceRoleKey string) *SupabaseUploader {
	return &SupabaseUploader{
		baseURL:        strings.TrimRight(baseURL, "/"),
		serviceRoleKey: serviceRoleKey,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (s *SupabaseUploader) Upload(ctx context.Context, bucket, objectPath, contentType string, payload []byte) error {
	if bucket == "" || objectPath == "" {
		return fmt.Errorf("bucket and object path are required")
	}
	if len(payload) == 0 {
		return fmt.Errorf("empty payload")
	}

	escapedPath := path.Clean(strings.TrimPrefix(objectPath, "/"))
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, url.PathEscape(bucket), escapedPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	req.Header.Set("apikey", s.serviceRoleKey)
	req.Header.Set("x-upsert", "true")
	req.Header.Set("Content-Type", contentType)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}
