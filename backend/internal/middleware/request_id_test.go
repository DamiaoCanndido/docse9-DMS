package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/middleware"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates new request ID when none provided", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.RequestIDMiddleware())

		var capturedInGin string
		var capturedInCtx string

		r.GET("/test", func(c *gin.Context) {
			capturedInGin = middleware.GetRequestID(c)
			capturedInCtx = logger.GetRequestID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		headerID := resp.Header().Get(middleware.HeaderXRequestID)
		require.NotEmpty(t, headerID)

		// Verifica se é um UUID válido
		_, err := uuid.Parse(headerID)
		assert.NoError(t, err)

		assert.Equal(t, headerID, capturedInGin)
		assert.Equal(t, headerID, capturedInCtx)
	})

	t.Run("propagates existing request ID header", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.RequestIDMiddleware())

		customID := "custom-trace-uuid-12345"
		var capturedInGin string
		var capturedInCtx string

		r.GET("/test", func(c *gin.Context) {
			capturedInGin = middleware.GetRequestID(c)
			capturedInCtx = logger.GetRequestID(c.Request.Context())
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(middleware.HeaderXRequestID, customID)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, customID, resp.Header().Get(middleware.HeaderXRequestID))
		assert.Equal(t, customID, capturedInGin)
		assert.Equal(t, customID, capturedInCtx)
	})
}
