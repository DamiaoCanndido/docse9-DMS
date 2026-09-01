package handler

import (
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
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	g := rg.Group("/auth")
	for _, m := range middlewares {
		g.Use(m)
	}
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
		handleAuthError(c, err)
		return
	}

	response.OK(c, resp)
}

