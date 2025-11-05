package storage

import "context"

// StorageClient interface for file storage abstraction
type StorageClient interface {
	UploadFile(ctx context.Context, destinationPath string, fileName string, contentType string, data []byte) (string, error)
	DeleteFile(ctx context.Context, destinationPath string) error
	GetPublicURL(path string) string
}
