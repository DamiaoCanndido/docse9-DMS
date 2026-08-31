package logger_test

import (
	"context"
	"os"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	t.Run("development environment", func(t *testing.T) {
		l := logger.InitLogger("development")
		assert.NotNil(t, l)
	})

	t.Run("production environment", func(t *testing.T) {
		l := logger.InitLogger("production")
		assert.NotNil(t, l)
	})

	t.Run("custom LOG_LEVEL environment", func(t *testing.T) {
		_ = os.Setenv("LOG_LEVEL", "debug")
		defer os.Unsetenv("LOG_LEVEL")
		l := logger.InitLogger("production")
		assert.NotNil(t, l)
	})

	t.Run("warn LOG_LEVEL", func(t *testing.T) {
		_ = os.Setenv("LOG_LEVEL", "warn")
		defer os.Unsetenv("LOG_LEVEL")
		l := logger.InitLogger("production")
		assert.NotNil(t, l)
	})

	t.Run("error LOG_LEVEL", func(t *testing.T) {
		_ = os.Setenv("LOG_LEVEL", "error")
		defer os.Unsetenv("LOG_LEVEL")
		l := logger.InitLogger("production")
		assert.NotNil(t, l)
	})
}

func TestContextRequestID(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, logger.GetRequestID(ctx))

	reqID := "req-test-12345"
	ctxWithID := logger.WithRequestID(ctx, reqID)
	assert.Equal(t, reqID, logger.GetRequestID(ctxWithID))
}
