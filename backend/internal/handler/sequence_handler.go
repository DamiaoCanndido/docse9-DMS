package handler

import (
	"strconv"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SequenceHandler struct {
	svc domain.SequenceService
}

// NewSequenceHandler cria um novo handler de sequências documentais.
func NewSequenceHandler(svc domain.SequenceService) *SequenceHandler {
	return &SequenceHandler{svc: svc}
}

// RegisterRoutes registra as rotas de sequências sob /municipalities/:id/sequences.
func (h *SequenceHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/municipalities/:id/sequences")
	{
		g.GET("", h.GetSequences)
		g.PUT("", h.SetOffset)
	}
}

// GetSequences lista o status de sequências e marcos iniciais do município.
func (h *SequenceHandler) GetSequences(c *gin.Context) {
	munID, ok := parseUUID(c, "id")
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
		response.Forbidden(c, "apenas moderadores têm permissão para acessar sequências numéricas")
		return
	}

	if claims.MunicipalityID != munID {
		response.Forbidden(c, "permissão insuficiente para gerenciar sequências de outro município")
		return
	}

	var year *int
	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
			year = &y
		}
	}

	items, err := h.svc.GetSequences(munID, year)
	if err != nil {
		if handleDomainError(c, err) {
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, items)
}

// SetOffset define ou atualiza o marco inicial de um tipo de ato e ano.
func (h *SequenceHandler) SetOffset(c *gin.Context) {
	munID, ok := parseUUID(c, "id")
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
		response.Forbidden(c, "apenas moderadores têm permissão para configurar sequências numéricas")
		return
	}

	if claims.MunicipalityID != munID {
		response.Forbidden(c, "permissão insuficiente para gerenciar sequências de outro município")
		return
	}

	var input domain.SetSequenceOffsetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	offset, err := h.svc.SetOffset(munID, input)
	if err != nil {
		if handleDomainError(c, err) {
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, offset)
}
