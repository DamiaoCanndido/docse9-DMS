package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DatabasePinger define a interface para verificação de conectividade com o banco de dados.
type DatabasePinger interface {
	PingContext(ctx context.Context) error
}

// HealthHandler gerencia os endpoints de health check da aplicação.
type HealthHandler struct {
	db     *gorm.DB
	pinger DatabasePinger
}

// NewHealthHandler instancia o handler com a conexão do banco de dados GORM.
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// NewHealthHandlerWithPinger instancia o handler com um DatabasePinger customizado (ideal para testes unitários).
func NewHealthHandlerWithPinger(pinger DatabasePinger) *HealthHandler {
	return &HealthHandler{pinger: pinger}
}

// HealthCheck verifica a integridade da aplicação e a conectividade com o banco de dados.
// Retorna 200 OK com status "up" se saudável, ou 503 Service Unavailable se o banco estiver inacessível.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	var pinger DatabasePinger = h.pinger

	if pinger == nil && h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "health check: falha ao obter pool de conexões do banco", "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "unreachable",
			})
			return
		}
		pinger = sqlDB
	}

	if pinger == nil {
		slog.ErrorContext(c.Request.Context(), "health check: nenhum provedor de banco de dados configurado")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"database": "unreachable",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := pinger.PingContext(ctx); err != nil {
		slog.ErrorContext(c.Request.Context(), "health check: banco de dados indisponível", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"database": "down",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "up",
	})
}
