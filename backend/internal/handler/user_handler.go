package handler

import (
	"errors"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/middleware"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
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
	rg.POST("/users/me/change-password", h.ChangePassword)

	g := rg.Group("/users")
	g.Use(middleware.RequireRole(domain.RoleAdmin, domain.RoleMod))
	{
		g.POST("", h.Create)
		g.GET("", h.GetAll)
		g.GET("/trash", h.GetDeleted)
		g.GET("/:id", h.GetByID)
		g.PATCH("/:id", h.Update)
		g.PATCH("/:id/restore", h.Restore)
		g.DELETE("/:id/hard", h.HardDelete)
		g.DELETE("/:id", h.Delete)

		// Rotas de permissões do usuário
		g.GET("/:id/permissions", h.GetPermissions)
		g.PUT("/:id/permissions", h.UpdatePermissions)
	}
}

// Create cria um novo usuário
func (h *UserHandler) Create(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	var input domain.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	actorRole := domain.Role(claims.Role)
	if actorRole == domain.RoleMod {
		if input.MunicipalityID != claims.MunicipalityID {
			response.Forbidden(c, "permissão insuficiente para criar usuários em outro município")
			return
		}
		if input.Role == domain.RoleAdmin {
			response.Forbidden(c, "permissão insuficiente para criar usuários administradores")
			return
		}
	} else if actorRole == domain.RoleAdmin {
		if input.Role == domain.RoleAdmin {
			response.Forbidden(c, "permissão insuficiente para criar usuários administradores")
			return
		}
	}

	u, randomPassword, err := h.svc.Create(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, gin.H{
		"user":           u,
		"randomPassword": randomPassword,
	})
}

// GetAll lista todos os usuários ativos (paginado)
func (h *UserHandler) GetAll(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	page, pageSize := parsePagination(c)

	users, _, err := h.svc.GetAll(page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	actorRole := domain.Role(claims.Role)
	var filtered []domain.User
	for _, u := range users {
		if actorRole == domain.RoleAdmin {
			filtered = append(filtered, u)
		}
		if actorRole == domain.RoleMod && u.Role != domain.RoleAdmin && u.MunicipalityID == claims.MunicipalityID {
			filtered = append(filtered, u)
		}
	}

	response.Paginated(c, filtered, int64(len(filtered)), page, pageSize)
}

// GetDeleted lista os usuários deletados (lixeira)
func (h *UserHandler) GetDeleted(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	page, pageSize := parsePagination(c)

	users, _, err := h.svc.GetDeleted(page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	actorRole := domain.Role(claims.Role)
	var filtered []domain.User
	for _, u := range users {
		if actorRole == domain.RoleAdmin {
			filtered = append(filtered, u)
		}
		if actorRole == domain.RoleMod && u.Role != domain.RoleAdmin && u.MunicipalityID == claims.MunicipalityID {
			filtered = append(filtered, u)
		}
	}

	response.Paginated(c, filtered, int64(len(filtered)), page, pageSize)
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

	if !h.checkAccess(c, u) {
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

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	u, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, u) {
		return
	}

	var input domain.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	actorRole := domain.Role(claims.Role)
	if actorRole == domain.RoleMod {
		if input.MunicipalityID != nil && *input.MunicipalityID != claims.MunicipalityID {
			response.Forbidden(c, "permissão insuficiente para alterar o município do usuário")
			return
		}
		if input.Role != nil && *input.Role == domain.RoleAdmin {
			response.Forbidden(c, "permissão insuficiente para definir a role como administrador")
			return
		}
	} else if actorRole == domain.RoleAdmin {
		if input.Role != nil && *input.Role == domain.RoleAdmin {
			response.Forbidden(c, "permissão insuficiente para definir a role como administrador")
			return
		}
	}

	updated, randomPassword, err := h.svc.Update(id, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if randomPassword != "" {
		response.OK(c, gin.H{
			"user":           updated,
			"randomPassword": randomPassword,
		})
	} else {
		response.OK(c, updated)
	}
}

// Delete remove um usuário (soft delete)
func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	u, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, u) {
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

	u, err := h.svc.GetByIDUnscoped(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, u) {
		return
	}

	restored, err := h.svc.Restore(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, restored)
}

// HardDelete remove um usuário definitivamente
func (h *UserHandler) HardDelete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	u, err := h.svc.GetByIDUnscoped(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, u) {
		return
	}

	if err := h.svc.HardDelete(id); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.NoContent(c)
}

// GetPermissions retorna as permissões de documento do usuário
func (h *UserHandler) GetPermissions(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	u, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, u) {
		return
	}

	p, err := h.svc.GetPermissions(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, p)
}

// UpdatePermissions atualiza as permissões de documento do usuário
func (h *UserHandler) UpdatePermissions(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	u, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, u) {
		return
	}

	var input domain.UpdateUserPermissionsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.svc.UpdatePermissions(id, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, p)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	var input domain.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	u, err := h.svc.ChangePassword(claims.UserID, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Gerar novo Token JWT com MustChangePassword = false
	token, err := security.GenerateToken(u.ID, u.Username, string(u.Role), u.MunicipalityID, false, 24*time.Hour)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{
		"token": token,
		"user":  u,
	})
}

func (h *UserHandler) checkAccess(c *gin.Context, targetUser *domain.User) bool {
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return false
	}

	actorRole := domain.Role(claims.Role)

	// Se for MOD, não pode gerenciar administradores e só pode gerenciar usuários do mesmo município
	if actorRole == domain.RoleMod {
		if targetUser.Role == domain.RoleAdmin {
			response.Forbidden(c, "permissão insuficiente para gerenciar administradores")
			return false
		}
		if targetUser.MunicipalityID != claims.MunicipalityID {
			response.Forbidden(c, "permissão insuficiente para acessar usuários de outro município")
			return false
		}
	}

	return true
}

func (h *UserHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		response.Conflict(c, err.Error())
	case errors.Is(err, domain.ErrUsernameAlreadyExists):
		response.Conflict(c, err.Error())
	case errors.Is(err, domain.ErrIncorrectCurrentPassword):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrMunicipalityNotFound):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c)
	}
}
