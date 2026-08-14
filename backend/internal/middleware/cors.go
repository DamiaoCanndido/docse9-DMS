package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// defaultAllowedOrigins são as origens locais permitidas por padrão caso CORS_ALLOWED_ORIGINS não seja definida.
var defaultAllowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:5173",
	"http://127.0.0.1:3000",
	"http://127.0.0.1:5173",
}

// CORSMiddleware adiciona cabeçalhos CORS seguros verificando uma whitelist de origens permitidas.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			if isOriginAllowed(origin) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(origin string) bool {
	rawOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedList []string
	if rawOrigins != "" {
		for _, o := range strings.Split(rawOrigins, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				allowedList = append(allowedList, trimmed)
			}
		}
	} else {
		allowedList = defaultAllowedOrigins
	}

	for _, allowed := range allowedList {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}
