package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMaxBodySizeMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(MaxBodySizeMiddleware(1024)) // 1KB limit
	r.POST("/upload-test", func(c *gin.Context) {
		var req struct {
			Data string `json:"data"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "payload too large"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	t.Run("permite payload dentro do limite", func(t *testing.T) {
		payload := `{"data":"pequeno"}`
		req := httptest.NewRequest(http.MethodPost, "/upload-test", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bloqueia payload que excede o limite", func(t *testing.T) {
		largeData := strings.Repeat("A", 2048)
		payload := `{"data":"` + largeData + `"}`
		req := httptest.NewRequest(http.MethodPost, "/upload-test", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})
}
