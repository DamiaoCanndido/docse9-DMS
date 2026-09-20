package storage

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"
)

// MockStorageService é uma implementação em memória de StorageService para testes unitários e dev local.
type MockStorageService struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

// NewMockStorageService instancia um mock em memória do StorageService.
func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		objects: make(map[string][]byte),
	}
}

func (m *MockStorageService) PutTestObject(key string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[key] = data
}

func (m *MockStorageService) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiresIn time.Duration) (string, error) {
	return fmt.Sprintf("https://mock-r2.cloudflarestorage.com/upload?key=%s&expires=%d", key, time.Now().Add(expiresIn).Unix()), nil
}

func (m *MockStorageService) GeneratePresignedDownloadURL(ctx context.Context, key string, filename string, download bool, expiresIn time.Duration) (string, error) {
	return fmt.Sprintf("https://mock-r2.cloudflarestorage.com/download?key=%s&filename=%s&download=%t&expires=%d", key, filename, download, time.Now().Add(expiresIn).Unix()), nil
}

func (m *MockStorageService) ValidatePDFOCR(ctx context.Context, key string, minChars int) (bool, int, error) {
	m.mu.RLock()
	data, ok := m.objects[key]
	m.mu.RUnlock()

	if !ok {
		return false, 0, ErrObjectNotFound
	}

	reader := bytes.NewReader(data)
	return ValidatePDFText(reader, int64(len(data)), minChars)
}

func (m *MockStorageService) DeleteObject(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
	return nil
}

func (m *MockStorageService) ObjectExists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.objects[key]
	return ok, nil
}
