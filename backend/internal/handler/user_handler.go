package handler

import (
	"errors"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/middleware"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc domain.UserService
}

func NewUserHandler(svc domain.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes registra todas as rotas do recurso User.
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/users")
	g.Use(middleware.RequireRole(domain.RoleAdmin))
	{
		g.POST("", h.Create)
		g.GET("", h.GetAll)
		g.GET("/trash", h.GetDeleted)
		g.GET("/:id", h.GetByID)
		g.PATCH("/:id", h.Update)
		g.PATCH("/:id/restore", h.Restore)
		g.DELETE("/:id/hard", h.HardDelete)
		g.DELETE("/:id", h.Delete)
	}
}

// Create cria um novo usuário
func (h *UserHandler) Create(c *gin.Context) {
	var input domain.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	u, err := h.svc.Create(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, u)
}

// GetAll lista todos os usuários ativos (paginado)
func (h *UserHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)

	users, total, err := h.svc.GetAll(page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Paginated(c, users, total, page, pageSize)
}

// GetDeleted lista os usuários deletados (lixeira)
func (h *UserHandler) GetDeleted(c *gin.Context) {
	page, pageSize := parsePagination(c)

	users, total, err := h.svc.GetDeleted(page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Paginated(c, users, total, page, pageSize)
}

// GetByID busca um usuário por ID
func (h *UserHandler) GetByID(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	u, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, u)
}

// Update atualiza um usuário parcialmente
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	var input domain.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	u, err := h.svc.Update(id, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, u)
}

// Delete remove um usuário (soft delete)
func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	if err := h.svc.Delete(id); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.NoContent(c)
}

// Restore restaura um usuário deletado da lixeira
func (h *UserHandler) Restore(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	u, err := h.svc.Restore(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, u)
}

// HardDelete remove um usuário definitivamente
func (h *UserHandler) HardDelete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	if err := h.svc.HardDelete(id); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *UserHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		response.Conflict(c, err.Error())
	case errors.Is(err, domain.ErrUsernameAlreadyExists):
		response.Conflict(c, err.Error())
	case errors.Is(err, service.ErrMunicipalityNotFound):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c)
	}
}
