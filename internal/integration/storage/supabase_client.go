package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
	maxRetries  = 3
)

type SupabaseStorageClient struct {
	baseURL   string
	apiKey    string
	bucket    string
	client    *http.Client
	publicURL string
}

func NewSupabaseStorageClient() (*SupabaseStorageClient, error) {
	baseURL := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_SERVICE_KEY")
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")

	if baseURL == "" || apiKey == "" || bucket == "" {
		return nil, fmt.Errorf("supabase not configured: missing URL, SERVICE_KEY, or BUCKET")
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s", baseURL, bucket)

	return &SupabaseStorageClient{
		baseURL:   baseURL,
		apiKey:    apiKey,
		bucket:    bucket,
		publicURL: publicURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *SupabaseStorageClient) UploadFile(ctx context.Context, destinationPath string, fileName string, contentType string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty data provided")
	}
	if len(data) > maxFileSize {
		return "", fmt.Errorf("file too large: max %dMB", maxFileSize/(1024*1024))
	}

	// Sanitasi path
	if strings.Contains(destinationPath, "../") || strings.Contains(destinationPath, "..\\") {
		return "", fmt.Errorf("invalid path: contains directory traversal")
	}

	// Set default content type sebelum loop
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	putURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, destinationPath)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, bytes.NewReader(data))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("apiKey", s.apiKey)
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
		req.Header.Set("x-upsert", "true")

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			logrus.WithError(err).WithFields(logrus.Fields{
				"path":     destinationPath,
				"attempts": attempt,
			}).Error("Supabase upload failed")
			return "", fmt.Errorf("upload request failed: %w", err)
		}

		// Close response body immediately after use
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body))
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			logrus.WithFields(logrus.Fields{
				"status": resp.StatusCode,
				"body":   string(body),
				"path":   destinationPath,
			}).Error("Supabase upload error")
			return "", fmt.Errorf("upload failed: %w", lastErr)
		}

		resp.Body.Close()
		fullPublicURL := fmt.Sprintf("%s/%s", s.publicURL, destinationPath)
		logrus.WithField("url", fullPublicURL).Info("File uploaded to Supabase")
		return fullPublicURL, nil
	}
	return "", fmt.Errorf("upload failed after %d attempts: %w", maxRetries, lastErr)
}

func (s *SupabaseStorageClient) DeleteFile(ctx context.Context, destinationPath string) error {
	if s.baseURL == "" || s.apiKey == "" || s.bucket == "" {
		return fmt.Errorf("supabase storage not configured")
	}

	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, destinationPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("apiKey", s.apiKey)
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		logrus.WithError(err).WithField("path", destinationPath).Error("Supabase delete failed")
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
			"path":   destinationPath,
		}).Error("Supabase delete error")
		return fmt.Errorf("delete failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	logrus.WithField("path", destinationPath).Info("File deleted from Supabase")
	return nil
}



// GetPublicURL helper untuk mendapatkan public URL dari path
func (s *SupabaseStorageClient) GetPublicURL(path string) string {
	return fmt.Sprintf("%s/%s", s.publicURL, path)
}
