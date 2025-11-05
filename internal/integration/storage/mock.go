package storage

import "context"

// MockStorageClient for testing
type MockStorageClient struct {
	UploadedFiles []MockFile
}

type MockFile struct {
	Path        string
	FileName    string
	ContentType string
	Data        []byte
}

func (m *MockStorageClient) UploadFile(ctx context.Context, destinationPath string, fileName string, contentType string, data []byte) (string, error) {
	m.UploadedFiles = append(m.UploadedFiles, MockFile{
		Path:        destinationPath,
		FileName:    fileName,
		ContentType: contentType,
		Data:        data,
	})
	return "https://mock-storage.com/" + destinationPath, nil
}

func (m *MockStorageClient) DeleteFile(ctx context.Context, destinationPath string) error {
	return nil
}

func (m *MockStorageClient) GetPublicURL(path string) string {
	return "https://mock-storage.com/" + path
}
