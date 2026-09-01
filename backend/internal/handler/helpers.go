package handler

import (
	"strconv"

	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getClaims extrai os claims JWT do contexto Gin.
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

// parseUUID converte um parâmetro de rota em uuid.UUID com resposta automática em caso de erro.
func parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		response.BadRequest(c, "ID inválido")
		return uuid.Nil, false
	}
	return id, true
}

// parsePagination extrai e sanitiza os parâmetros de paginação `page` e `pageSize`.
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
