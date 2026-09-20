package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockStorageService_GeneratePresignedUploadURL(t *testing.T) {
	mockSvc := storage.NewMockStorageService()
	url, err := mockSvc.GeneratePresignedUploadURL(context.Background(), "tenants/1/NOTICE/2026/1/file.pdf", "application/pdf", 1024*1024, 10*time.Minute)
	require.NoError(t, err)
	assert.Contains(t, url, "contentLength=1048576")
	assert.Contains(t, url, "tenants/1/NOTICE/2026/1/file.pdf")
}

func TestMockStorageService_GeneratePresignedDownloadURL(t *testing.T) {
	mockSvc := storage.NewMockStorageService()
	url, err := mockSvc.GeneratePresignedDownloadURL(context.Background(), "tenants/1/file.pdf", "doc.pdf", false, 15*time.Minute)
	require.NoError(t, err)
	assert.Contains(t, url, "filename=doc.pdf")
	assert.Contains(t, url, "download=false")
}

func TestMockStorageService_ValidatePDFOCR_NotFound(t *testing.T) {
	mockSvc := storage.NewMockStorageService()
	valid, chars, err := mockSvc.ValidatePDFOCR(context.Background(), "non-existent-key", 50)
	assert.ErrorIs(t, err, storage.ErrObjectNotFound)
	assert.False(t, valid)
	assert.Equal(t, 0, chars)
}
