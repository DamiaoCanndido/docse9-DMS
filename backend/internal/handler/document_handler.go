package handler

import (
	"errors"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	svc      domain.DocumentService
	permRepo domain.UserPermissionRepository
}

// NewDocumentHandler cria um novo handler de documentos com suporte a permissões.
func NewDocumentHandler(svc domain.DocumentService, permRepo domain.UserPermissionRepository) *DocumentHandler {
	return &DocumentHandler{svc: svc, permRepo: permRepo}
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
		g.DELETE("/:id/hard", h.HardDelete)
		g.DELETE("/:id", h.Delete)
	}
}

// Create cria um novo documento
func (h *DocumentHandler) Create(c *gin.Context) {
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	actorRole := domain.Role(claims.Role)
	if actorRole == domain.RoleAdmin {
		response.Forbidden(c, "administradores não têm permissão para acessar documentos")
		return
	}

	var input domain.CreateDocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// MOD pode criar qualquer documento em seu município
	if actorRole == domain.RoleMod {
		if input.MunicipalityID != claims.MunicipalityID {
			response.Forbidden(c, "permissão insuficiente para criar documentos em outro município")
			return
		}
	}

	// COMMON requer permissão WRITE ou DELETE para o tipo de documento
	if actorRole == domain.RoleCommon {
		if input.MunicipalityID != claims.MunicipalityID {
			response.Forbidden(c, "permissão insuficiente para criar documentos em outro município")
			return
		}
		if input.OwnerID != claims.UserID {
			response.Forbidden(c, "permissão insuficiente para criar documentos em nome de outro usuário")
			return
		}

		permList, err := h.permRepo.FindByUserID(claims.UserID)
		if err != nil {
			response.InternalError(c)
			return
		}
		var hasWrite bool
		for _, p := range permList {
			if p.DocumentType == input.Type && (p.Level == domain.LevelWrite || p.Level == domain.LevelDelete) {
				hasWrite = true
				break
			}
		}
		if !hasWrite {
			response.Forbidden(c, "permissão insuficiente para criar documentos")
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
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	actorRole := domain.Role(claims.Role)
	if actorRole == domain.RoleAdmin {
		response.Forbidden(c, "administradores não têm permissão para acessar documentos")
		return
	}

	var filter domain.DocumentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if actorRole == domain.RoleCommon {
		ok, msg := h.buildAllowedTypesFilter(claims.UserID, &filter)
		if !ok {
			if msg == "erro interno" {
				response.InternalError(c)
			} else {
				response.Forbidden(c, msg)
			}
			return
		}
		filter.MunicipalityID = &claims.MunicipalityID
	} else if actorRole == domain.RoleMod {
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
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	actorRole := domain.Role(claims.Role)
	if actorRole == domain.RoleAdmin {
		response.Forbidden(c, "administradores não têm permissão para acessar documentos")
		return
	}

	var filter domain.DocumentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if actorRole == domain.RoleCommon {
		ok, msg := h.buildAllowedTypesFilter(claims.UserID, &filter)
		if !ok {
			if msg == "erro interno" {
				response.InternalError(c)
			} else {
				response.Forbidden(c, msg)
			}
			return
		}
		filter.MunicipalityID = &claims.MunicipalityID
	} else if actorRole == domain.RoleMod {
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

	d, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, d, "view") {
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

	d, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, d, "update") {
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

	d, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, d, "delete") {
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

	actorRole := domain.Role(claims.Role)
	if actorRole == domain.RoleAdmin {
		response.Forbidden(c, "administradores não têm permissão para acessar documentos")
		return
	}

	// Executa restauração temporária para validação
	restored, err := h.svc.Restore(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if !h.checkAccess(c, restored, "delete") {
		// Desfaz o restore caso o usuário não tenha permissão
		_ = h.svc.Delete(id)
		return
	}

	response.OK(c, restored)
}

// HardDelete remove permanentemente um documento
func (h *DocumentHandler) HardDelete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return
	}

	actorRole := domain.Role(claims.Role)
	if actorRole != domain.RoleMod {
		response.Forbidden(c, "apenas moderadores podem realizar hard delete de documentos")
		return
	}

	// Restaura temporariamente para obter o MunicipalityID
	restored, err := h.svc.Restore(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if restored.MunicipalityID != claims.MunicipalityID {
		_ = h.svc.Delete(id) // desfaz o restore
		response.Forbidden(c, "permissão insuficiente para gerenciar documentos de outro município")
		return
	}

	if err := h.svc.HardDelete(id); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *DocumentHandler) checkAccess(c *gin.Context, doc *domain.Document, action string) bool {
	claims := getClaims(c)
	if claims == nil {
		response.Unauthorized(c, "usuário não autenticado")
		return false
	}

	actorRole := domain.Role(claims.Role)

	if actorRole == domain.RoleAdmin {
		response.Forbidden(c, "administradores não têm permissão para acessar documentos")
		return false
	}

	// MOD tem acesso total no mesmo município
	if actorRole == domain.RoleMod {
		if doc.MunicipalityID != claims.MunicipalityID {
			response.Forbidden(c, "permissão insuficiente para gerenciar documentos de outro município")
			return false
		}
		return true
	}

	// COMMON requer validação da tabela UserPermission por tipo de documento
	if actorRole == domain.RoleCommon {
		if doc.MunicipalityID != claims.MunicipalityID {
			response.Forbidden(c, "permissão insuficiente para gerenciar documentos de outro município")
			return false
		}

		permList, err := h.permRepo.FindByUserID(claims.UserID)
		if err != nil {
			response.InternalError(c)
			return false
		}

		var level domain.PermissionLevel = domain.LevelNone
		for _, p := range permList {
			if p.DocumentType == doc.Type {
				level = p.Level
				break
			}
		}

		switch action {
		case "view":
			if level != domain.LevelRead && level != domain.LevelWrite && level != domain.LevelDelete {
				response.Forbidden(c, "permissão insuficiente para visualizar documentos")
				return false
			}
		case "update":
			if level != domain.LevelWrite && level != domain.LevelDelete {
				response.Forbidden(c, "permissão insuficiente para atualizar documentos")
				return false
			}
			if doc.OwnerID != claims.UserID {
				response.Forbidden(c, "permissão insuficiente para atualizar documentos de outro usuário")
				return false
			}
		case "delete":
			if level != domain.LevelDelete {
				response.Forbidden(c, "permissão insuficiente para deletar documentos")
				return false
			}
			if doc.OwnerID != claims.UserID {
				response.Forbidden(c, "permissão insuficiente para deletar documentos de outro usuário")
				return false
			}
		default:
			response.Forbidden(c, "ação inválida")
			return false
		}
		return true
	}

	response.Forbidden(c, "permissão insuficiente")
	return false
}

func (h *DocumentHandler) buildAllowedTypesFilter(userID uuid.UUID, filter *domain.DocumentFilter) (bool, string) {
	permList, err := h.permRepo.FindByUserID(userID)
	if err != nil {
		return false, "erro interno"
	}

	var allowedTypes []domain.DocumentType
	for _, p := range permList {
		if p.Level == domain.LevelRead || p.Level == domain.LevelWrite || p.Level == domain.LevelDelete {
			allowedTypes = append(allowedTypes, p.DocumentType)
		}
	}

	if filter.Type != nil {
		hasAccess := false
		for _, t := range allowedTypes {
			if t == *filter.Type {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			return false, "permissão insuficiente para visualizar documentos"
		}
	} else {
		if len(allowedTypes) == 0 {
			return false, "permissão insuficiente para visualizar documentos"
		}
		filter.AllowedTypes = allowedTypes
	}

	return true, ""
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
