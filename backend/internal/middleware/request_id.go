package middleware

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// HeaderXRequestID é o nome padrão do cabeçalho HTTP para rastreamento de requisições.
	HeaderXRequestID = "X-Request-ID"

	// ContextKeyRequestID é a chave utilizada no contexto do Gin para armazenar o request ID.
	ContextKeyRequestID = "request_id"
)

// RequestIDMiddleware injeta ou propaga um identificador único de requisição (UUID v4)
// em todas as requisições HTTP, propagando-o nos headers de resposta e contextos da aplicação.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(HeaderXRequestID)
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// Injeta no cabeçalho da resposta
		c.Header(HeaderXRequestID, reqID)

		// Injeta no contexto do Gin
		c.Set(ContextKeyRequestID, reqID)

		// Injeta no contexto nativo de Go (context.Context) da requisição
		ctx := logger.WithRequestID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetRequestID extrai o ID da requisição a partir do contexto do Gin.
func GetRequestID(c *gin.Context) string {
	if reqID, exists := c.Get(ContextKeyRequestID); exists {
		if idStr, ok := reqID.(string); ok && idStr != "" {
			return idStr
		}
	}

	if idHeader := c.GetHeader(HeaderXRequestID); idHeader != "" {
		return idHeader
	}

	if c.Request != nil {
		return logger.GetRequestID(c.Request.Context())
	}

	return ""
}
