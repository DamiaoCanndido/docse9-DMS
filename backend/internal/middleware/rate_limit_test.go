package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("permite requisições dentro do limite", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimiterMiddleware(3, 1*time.Minute))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("bloqueia com 429 quando limite for excedido", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimiterMiddleware(2, 1*time.Minute))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// 2 permitidas
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// 3ª deve ser bloqueada
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.NotEmpty(t, w.Header().Get("Retry-After"))
	})

	t.Run("reseta tokens após expiração da janela", func(t *testing.T) {
		limiter := NewIPRateLimiter(1, 50*time.Millisecond)
		defer limiter.Stop()

		allowed, _ := limiter.Allow("127.0.0.1")
		assert.True(t, allowed)

		allowed, _ = limiter.Allow("127.0.0.1")
		assert.False(t, allowed)

		time.Sleep(60 * time.Millisecond)

		allowed, _ = limiter.Allow("127.0.0.1")
		assert.True(t, allowed)
	})
}
