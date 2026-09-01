package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logEntry struct {
	Level     string  `json:"level"`
	Msg       string  `json:"msg"`
	RequestID string  `json:"request_id"`
	Method    string  `json:"method"`
	Path      string  `json:"path"`
	Status    int     `json:"status"`
	LatencyMs float64 `json:"latency_ms"`
	Query     string  `json:"query"`
	Errors    string  `json:"errors"`
	Error     string  `json:"error"`
	Stack     string  `json:"stack"`
}

func setupTestLoggerRouter(logBuf *bytes.Buffer) (*gin.Engine, *slog.Logger) {
	gin.SetMode(gin.TestMode)
	handler := slog.NewJSONHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug})
	testLogger := slog.New(handler)

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.StructuredLoggerMiddleware(testLogger))
	r.Use(middleware.RecoveryWithSlog(testLogger))

	return r, testLogger
}

func TestStructuredLoggerMiddleware(t *testing.T) {
	t.Run("logs successful 200 OK request at INFO level", func(t *testing.T) {
		var buf bytes.Buffer
		r, _ := setupTestLoggerRouter(&buf)

		r.GET("/api/v1/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/ping?foo=bar", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var entry logEntry
		err := json.Unmarshal(buf.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "GET", entry.Method)
		assert.Equal(t, "/api/v1/ping", entry.Path)
		assert.Equal(t, "foo=bar", entry.Query)
		assert.Equal(t, http.StatusOK, entry.Status)
		assert.NotEmpty(t, entry.RequestID)
		assert.GreaterOrEqual(t, entry.LatencyMs, float64(0))
	})

	t.Run("logs client error 400 at WARN level", func(t *testing.T) {
		var buf bytes.Buffer
		r, _ := setupTestLoggerRouter(&buf)

		r.POST("/api/v1/fail", func(c *gin.Context) {
			_ = c.Error(errors.New("validation failed"))
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		})

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/fail", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var entry logEntry
		err := json.Unmarshal(buf.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "WARN", entry.Level)
		assert.Equal(t, http.StatusBadRequest, entry.Status)
		assert.Contains(t, entry.Errors, "validation failed")
	})

	t.Run("logs server error 500 at ERROR level", func(t *testing.T) {
		var buf bytes.Buffer
		r, _ := setupTestLoggerRouter(&buf)

		r.GET("/api/v1/error", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/error", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)

		var entry logEntry
		err := json.Unmarshal(buf.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "ERROR", entry.Level)
		assert.Equal(t, http.StatusInternalServerError, entry.Status)
	})

	t.Run("recovers from panic and logs with stack trace", func(t *testing.T) {
		var buf bytes.Buffer
		r, _ := setupTestLoggerRouter(&buf)

		r.GET("/api/v1/panic", func(c *gin.Context) {
			panic("unexpected crash in handler")
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/panic", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)

		var respBody map[string]any
		err := json.Unmarshal(resp.Body.Bytes(), &respBody)
		require.NoError(t, err)
		assert.Equal(t, false, respBody["success"])

		// Verifica se registrou o log de panic
		output := buf.String()
		assert.Contains(t, output, "Panic recuperado no servidor HTTP")
		assert.Contains(t, output, "unexpected crash in handler")
	})

	t.Run("handles nil logger gracefully fallback to default", func(t *testing.T) {
		middleware1 := middleware.StructuredLoggerMiddleware(nil)
		assert.NotNil(t, middleware1)
		middleware2 := middleware.RecoveryWithSlog(nil)
		assert.NotNil(t, middleware2)
	})
}
