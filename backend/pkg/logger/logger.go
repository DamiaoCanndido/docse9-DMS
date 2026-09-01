package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

// InitLogger inicializa o logger global slog estruturado em formato JSON.
// Se appEnv for "development" ou "test", pode registrar níveis DEBUG se configurado.
func InitLogger(appEnv string) *slog.Logger {
	var level slog.Level

	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		if appEnv == "development" {
			level = slog.LevelDebug
		} else {
			level = slog.LevelInfo
		}
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: appEnv != "production", // Adiciona arquivo:linha em dev/test
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// WithRequestID adiciona o ID da requisição ao contexto do logger.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID recupera o ID da requisição a partir do contexto.
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}
