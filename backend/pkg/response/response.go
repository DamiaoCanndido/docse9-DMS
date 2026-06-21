package response

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ──────────────────────────────────────────────
// Envelopes de resposta
// ──────────────────────────────────────────────

type successResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type paginatedResponse struct {
	Success    bool `json:"success"`
	Data       any  `json:"data"`
	Pagination Meta `json:"pagination"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successResponse{Success: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, successResponse{Success: true, Data: data})
}

func Paginated(c *gin.Context, data any, total int64, page, pageSize int) {
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	c.JSON(http.StatusOK, paginatedResponse{
		Success: true,
		Data:    data,
		Pagination: Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, errorResponse{Success: false, Error: msg})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, errorResponse{Success: false, Error: msg})
}

func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, errorResponse{Success: false, Error: msg})
}

func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, errorResponse{
		Success: false,
		Error:   "erro interno no servidor",
	})
}
