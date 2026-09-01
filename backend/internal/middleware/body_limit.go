package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySizeMiddleware restringe o tamanho máximo do corpo da requisição HTTP.
func MaxBodySizeMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
