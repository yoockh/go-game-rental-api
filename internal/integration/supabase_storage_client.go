package integration

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type SupabaseStorageClient struct {
	baseURL   string
	apiKey    string
	bucket    string
	client    *http.Client
	publicURL string
}

func NewSupabaseStorageClient() *SupabaseStorageClient {
	baseURL := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_SERVICE_KEY")
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")

	if baseURL == "" || apiKey == "" || bucket == "" {
		log.Println("WARN: Supabase storage env vars not fully configured")
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
	}
}

func (s *SupabaseStorageClient) UploadFile(ctx context.Context, destinationPath string, fileName string, contentType string, data []byte) (string, error) {
	if s.baseURL == "" || s.apiKey == "" || s.bucket == "" {
		return "", fmt.Errorf("supabase storage not configured")
	}

	putURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, destinationPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, io.NopCloser(bytesReader(data)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("apiKey", s.apiKey)
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("x-upsert", "true") // overwrite if exists

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("ERROR: Supabase upload failed for %s: %v", destinationPath, err)
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("ERROR: Supabase upload error %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("upload failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	fullPublicURL := fmt.Sprintf("%s/%s", s.publicURL, destinationPath)
	log.Printf("INFO: File uploaded to Supabase: %s", fullPublicURL)
	return fullPublicURL, nil
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
		log.Printf("ERROR: Supabase delete failed for %s: %v", destinationPath, err)
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("ERROR: Supabase delete error %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("delete failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	log.Printf("INFO: File deleted from Supabase: %s", destinationPath)
	return nil
}

// helper to avoid importing bytes for a simple reader
func bytesReader(b []byte) *byteReader { return &byteReader{b: b} }

type byteReader struct{ b []byte }

func (r *byteReader) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

// GetPublicURL helper untuk mendapatkan public URL dari path
func (s *SupabaseStorageClient) GetPublicURL(path string) string {
	return fmt.Sprintf("%s/%s", s.publicURL, path)
}
