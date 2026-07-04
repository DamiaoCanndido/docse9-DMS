package handler

import (
	"errors"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc domain.AuthService
}

func NewAuthHandler(svc domain.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// RegisterRoutes registra todas as rotas de autenticação.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	{
		g.POST("/login", h.Login)
	}
}

// Login realiza o login do usuário
func (h *AuthHandler) Login(c *gin.Context) {
	var input domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Login(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, resp)
}

func (h *AuthHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Unauthorized(c, err.Error())
	default:
		response.InternalError(c)
	}
}
