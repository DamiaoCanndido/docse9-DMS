package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) PingContext(ctx context.Context) error {
	return m.err
}

func TestHealthHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 200 OK when database is reachable", func(t *testing.T) {
		h := handler.NewHealthHandlerWithPinger(&mockPinger{err: nil})

		r := gin.New()
		r.GET("/health", h.HealthCheck)

		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.JSONEq(t, `{"status":"ok","database":"up"}`, resp.Body.String())
	})

	t.Run("returns 503 Service Unavailable when database ping fails", func(t *testing.T) {
		h := handler.NewHealthHandlerWithPinger(&mockPinger{err: errors.New("connection refused")})

		r := gin.New()
		r.GET("/health", h.HealthCheck)

		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
		assert.JSONEq(t, `{"status":"error","database":"down"}`, resp.Body.String())
	})

	t.Run("returns 503 Service Unavailable when no database is configured", func(t *testing.T) {
		h := handler.NewHealthHandler(nil)

		r := gin.New()
		r.GET("/health", h.HealthCheck)

		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
		assert.JSONEq(t, `{"status":"error","database":"unreachable"}`, resp.Body.String())
	})
}
