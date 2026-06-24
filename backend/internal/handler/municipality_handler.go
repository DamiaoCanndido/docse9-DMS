package handler

import (
	"errors"
	"strconv"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MunicipalityHandler struct {
	svc domain.MunicipalityService
}

func NewMunicipalityHandler(svc domain.MunicipalityService) *MunicipalityHandler {
	return &MunicipalityHandler{svc: svc}
}

// RegisterRoutes registra todas as rotas do recurso Municipality.
func (h *MunicipalityHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/municipalities")
	{
		g.POST("", h.Create)
		g.GET("", h.GetAll)
		g.GET("/trash", h.GetDeleted)
		g.GET("/uf/:uf", h.GetByUF)
		g.GET("/:id", h.GetByID)
		g.PATCH("/:id", h.Update)
		g.PATCH("/:id/restore", h.Restore)
		g.DELETE("/:id/hard", h.HardDelete)
		g.DELETE("/:id", h.Delete)
	}
}

// ──────────────────────────────────────────────
// POST /municipalities
// ──────────────────────────────────────────────

// @Summary     Cria um município
// @Tags        municipalities
// @Accept      json
// @Produce     json
// @Param       body body domain.CreateMunicipalityInput true "Dados do município"
// @Success     201 {object} domain.Municipality
// @Failure     400 {object} response.errorResponse
// @Failure     409 {object} response.errorResponse
// @Router      /municipalities [post]
func (h *MunicipalityHandler) Create(c *gin.Context) {
	var input domain.CreateMunicipalityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	m, err := h.svc.Create(input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, m)
}

// ──────────────────────────────────────────────
// GET /municipalities
// ──────────────────────────────────────────────

// @Summary     Lista todos os municípios (paginado)
// @Tags        municipalities
// @Produce     json
// @Param       page     query int false "Página (default 1)"
// @Param       pageSize query int false "Itens por página (default 20, max 100)"
// @Success     200 {object} response.paginatedResponse
// @Router      /municipalities [get]
func (h *MunicipalityHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)

	municipalities, total, err := h.svc.GetAll(page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Paginated(c, municipalities, total, page, pageSize)
}

// ──────────────────────────────────────────────
// GET /municipalities/trash
// ──────────────────────────────────────────────

// @Summary     Lista municípios removidos (lixeira)
// @Tags        municipalities
// @Produce     json
// @Param       page     query int false "Página (default 1)"
// @Param       pageSize query int false "Itens por página (default 20, max 100)"
// @Success     200 {object} response.paginatedResponse
// @Router      /municipalities/trash [get]
func (h *MunicipalityHandler) GetDeleted(c *gin.Context) {
	page, pageSize := parsePagination(c)

	municipalities, total, err := h.svc.GetDeleted(page, pageSize)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Paginated(c, municipalities, total, page, pageSize)
}

// ──────────────────────────────────────────────
// GET /municipalities/:id
// ──────────────────────────────────────────────

func (h *MunicipalityHandler) GetByID(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	m, err := h.svc.GetByID(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, m)
}

// ──────────────────────────────────────────────
// GET /municipalities/uf/:uf
// ──────────────────────────────────────────────

func (h *MunicipalityHandler) GetByUF(c *gin.Context) {
	uf := c.Param("uf")
	page, pageSize := parsePagination(c)

	municipalities, total, err := h.svc.GetByUF(uf, page, pageSize)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Paginated(c, municipalities, total, page, pageSize)
}

// ──────────────────────────────────────────────
// PATCH /municipalities/:id
// ──────────────────────────────────────────────

func (h *MunicipalityHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	var input domain.UpdateMunicipalityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	m, err := h.svc.Update(id, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, m)
}

// ──────────────────────────────────────────────
// DELETE /municipalities/:id
// ──────────────────────────────────────────────

func (h *MunicipalityHandler) Delete(c *gin.Context) {
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

// ──────────────────────────────────────────────
// PATCH /municipalities/:id/restore
// ──────────────────────────────────────────────

// @Summary     Restaura um município removido da lixeira
// @Tags        municipalities
// @Produce     json
// @Param       id path string true "UUID do município"
// @Success     200 {object} domain.Municipality
// @Failure     400 {object} response.errorResponse
// @Failure     404 {object} response.errorResponse
// @Router      /municipalities/{id}/restore [patch]
func (h *MunicipalityHandler) Restore(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	m, err := h.svc.Restore(id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, m)
}

// ──────────────────────────────────────────────
// DELETE /municipalities/:id/hard
// ──────────────────────────────────────────────

func (h *MunicipalityHandler) HardDelete(c *gin.Context) {
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

// ──────────────────────────────────────────────
// Helpers privados
// ──────────────────────────────────────────────

func (h *MunicipalityHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrMunicipalityNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrMunicipalityNameConflict):
		response.Conflict(c, err.Error())
	case errors.Is(err, service.ErrInvalidUF):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c)
	}
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		response.BadRequest(c, "ID inválido")
		return uuid.Nil, false
	}
	return id, true
}

func parsePagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return
}
