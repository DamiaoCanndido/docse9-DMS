package handler

import (
	"errors"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/middleware"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	svc domain.DocumentService
}

// NewDocumentHandler cria um novo handler de documentos.
func NewDocumentHandler(svc domain.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

// RegisterRoutes registra todas as rotas do recurso Document.
func (h *DocumentHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/documents")
	{
		g.POST("", h.Create)
		g.GET("", h.GetAll)
		g.GET("/trash", h.GetDeleted)
		g.GET("/:id", h.GetByID)
		g.PATCH("/:id", h.Update)
		g.PATCH("/:id/restore", h.Restore)
		g.DELETE("/:id/hard", middleware.RequireRole(domain.RoleAdmin), h.HardDelete)
		g.DELETE("/:id", h.Delete)
	}
}

// Create cria um novo documento
func (h *DocumentHandler) Create(c *gin.Context) {
	var input domain.CreateDocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	// Restrições de Multi-inquilinato (Tenant Isolation)
	if domain.Role(claims.Role) == domain.RoleCommon {
		if input.MunicipalityID != claims.MunicipalityID {
			response.Forbidden(c, "permissão insuficiente para criar documentos em outro município")
			return
		}
		if input.OwnerID != claims.UserID {
			response.Forbidden(c, "permissão insuficiente para criar documentos em nome de outro usuário")
			return
		}
	}

	d, err := h.svc.Create(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, d)
}

// GetAll lista todos os documentos (paginado, filtrado)
func (h *DocumentHandler) GetAll(c *gin.Context) {
	var filter domain.DocumentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	// Força isolamento de dados para usuários comuns
	if domain.Role(claims.Role) == domain.RoleCommon {
		filter.MunicipalityID = &claims.MunicipalityID
	}

	page, pageSize := parsePagination(c)

	docs, total, err := h.svc.GetAll(filter, page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Paginated(c, docs, total, page, pageSize)
}

// GetDeleted lista os documentos na lixeira (soft-deleted)
func (h *DocumentHandler) GetDeleted(c *gin.Context) {
	var filter domain.DocumentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	// Força isolamento de dados para usuários comuns
	if domain.Role(claims.Role) == domain.RoleCommon {
		filter.MunicipalityID = &claims.MunicipalityID
	}

	page, pageSize := parsePagination(c)

	docs, total, err := h.svc.GetDeleted(filter, page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Paginated(c, docs, total, page, pageSize)
}

// GetByID busca um documento por ID
func (h *DocumentHandler) GetByID(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	d, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Força isolamento de dados para usuários comuns
	if domain.Role(claims.Role) == domain.RoleCommon && d.MunicipalityID != claims.MunicipalityID {
		response.Forbidden(c, "permissão insuficiente para acessar este recurso")
		return
	}

	response.OK(c, d)
}

// Update atualiza parcialmente um documento
func (h *DocumentHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	// Busca o documento para validar permissões antes de atualizar
	d, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if domain.Role(claims.Role) == domain.RoleCommon && d.MunicipalityID != claims.MunicipalityID {
		response.Forbidden(c, "permissão insuficiente para acessar este recurso")
		return
	}

	var input domain.UpdateDocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.svc.Update(id, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, updated)
}

// Delete remove logicamente um documento
func (h *DocumentHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	d, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if domain.Role(claims.Role) == domain.RoleCommon && d.MunicipalityID != claims.MunicipalityID {
		response.Forbidden(c, "permissão insuficiente para acessar este recurso")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.NoContent(c)
}

// Restore restaura um documento da lixeira
func (h *DocumentHandler) Restore(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	// Busca sem escopo de soft-delete para restaurar
	// Como a camada de serviço do Restore já valida, buscamos antes apenas para validar tenant
	// Mas a entidade retornada no GetByIDUnscoped ainda é comparada.
	// Vamos buscar usando o Restore diretamente e, caso ocorra erro, tratar.
	// Se for COMMON, validamos após o restore no serviço (mas para evitar restaurar e depois proibir,
	// podemos tratar no próprio serviço ou fazer validação prévia).
	// A melhor abordagem é restaurar na service e verificar se o retornado bate com o tenant.
	// Se não bater, podemos dar rollback? GORM faz soft restoration.
	// De forma elegante, faremos a chamada direta e validamos o resultado.
	d, err := h.svc.Restore(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if domain.Role(claims.Role) == domain.RoleCommon && d.MunicipalityID != claims.MunicipalityID {
		// Desfaz o restore caso o usuário não tenha permissão
		_ = h.svc.Delete(id)
		response.Forbidden(c, "permissão insuficiente para acessar este recurso")
		return
	}

	response.OK(c, d)
}

// HardDelete remove permanentemente um documento
func (h *DocumentHandler) HardDelete(c *gin.Context) {
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

func (h *DocumentHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrDocumentNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrInvalidDocumentType) ||
		errors.Is(err, domain.ErrInvalidContractType) ||
		errors.Is(err, domain.ErrContractFieldsRequired):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrMunicipalityNotFound) ||
		errors.Is(err, domain.ErrUserNotFound):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c)
	}
}

func getClaims(c *gin.Context) *security.UserClaims {
	userVal, exists := c.Get("user")
	if !exists {
		return nil
	}
	claims, ok := userVal.(*security.UserClaims)
	if !ok {
		return nil
	}
	return claims
}
