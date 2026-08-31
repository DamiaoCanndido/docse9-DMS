package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// StructuredLoggerMiddleware retorna um middleware Gin que emite logs estruturados via slog
// em formato JSON com métricas de execução, status HTTP e request ID para rastreabilidade.
func StructuredLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		reqID := GetRequestID(c)
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		bytesOut := c.Writer.Size()
		bytesIn := c.Request.ContentLength

		attrs := []slog.Attr{
			slog.String("request_id", reqID),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.Float64("latency_ms", float64(latency.Microseconds())/1000.0),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.Int("bytes_out", bytesOut),
			slog.Int64("bytes_in", bytesIn),
		}

		if rawQuery != "" {
			attrs = append(attrs, slog.String("query", rawQuery))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		ctx := c.Request.Context()
		msg := fmt.Sprintf("HTTP %s %s -> %d", method, path, statusCode)

		switch {
		case statusCode >= http.StatusInternalServerError:
			logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
		case statusCode >= http.StatusBadRequest:
			logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
		default:
			logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
		}
	}
}

// RecoveryWithSlog intercepta panics na execução dos handlers HTTP, registrando stack traces
// estruturados no slog e retornando uma resposta 500 padronizada com segurança.
func RecoveryWithSlog(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID := GetRequestID(c)
				stack := string(debug.Stack())
				errStr := fmt.Sprintf("%v", r)

				logger.LogAttrs(
					c.Request.Context(),
					slog.LevelError,
					"Panic recuperado no servidor HTTP",
					slog.String("request_id", reqID),
					slog.String("error", errStr),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("stack", stack),
				)

				c.Abort()
				response.InternalError(c)
			}
		}()

		c.Next()
	}
}
