package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler"
	handlerMocks "github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupRouter monta o Gin com o handler e o mock de service injetado.
func setupRouter(svc domain.MunicipalityService) *gin.Engine {
	r := gin.New()
	h := handler.NewMunicipalityHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

// doRequest dispara uma requisição HTTP de teste e retorna o recorder.
func doRequest(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// parseBody é um helper genérico para desserializar a resposta.
func parseBody(t *testing.T, w *httptest.ResponseRecorder, target any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(w.Body).Decode(target))
}

// ═══════════════════════════════════════════════════════════════════════════════
// POST /api/v1/municipalities
// ═══════════════════════════════════════════════════════════════════════════════

func TestCreate_Handler_201(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	m := testhelper.MakePassagem()
	input := testhelper.MakeCreateInput("Passagem", "PB", "https://example.com/img.png")

	svc.On("Create", input).Return(&m, nil)

	w := doRequest(setupRouter(svc), http.MethodPost, "/api/v1/municipalities", input)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "Passagem", resp["data"].(map[string]any)["name"])
	svc.AssertExpectations(t)
}

func TestCreate_Handler_400_MissingName(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	body := map[string]string{"uf": "PB"} // sem name

	w := doRequest(setupRouter(svc), http.MethodPost, "/api/v1/municipalities", body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestCreate_Handler_400_InvalidUF(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	input := testhelper.MakeCreateInput("Passagem", "XX", "")

	svc.On("Create", input).Return(nil, service.ErrInvalidUF)

	w := doRequest(setupRouter(svc), http.MethodPost, "/api/v1/municipalities", input)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.False(t, resp["success"].(bool))
	assert.Equal(t, "UF inválida", resp["error"])
}

func TestCreate_Handler_409_NameConflict(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	input := testhelper.MakeCreateInput("Passagem", "PB", "")

	svc.On("Create", input).Return(nil, service.ErrMunicipalityNameConflict)

	w := doRequest(setupRouter(svc), http.MethodPost, "/api/v1/municipalities", input)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GET /api/v1/municipalities
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetAll_Handler_200_DefaultPagination(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	municipalities := []domain.Municipality{testhelper.MakePassagem(), testhelper.MakePatos()}

	// page=1, pageSize=20 são os defaults aplicados pelo handler
	svc.On("GetAll", 1, 20).Return(municipalities, int64(2), nil)

	w := doRequest(setupRouter(svc), http.MethodGet, "/api/v1/municipalities", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].([]any)
	assert.Len(t, data, 2)

	pagination := resp["pagination"].(map[string]any)
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(20), pagination["pageSize"])
	assert.Equal(t, float64(2), pagination["total"])
	assert.Equal(t, float64(1), pagination["totalPages"])
}

func TestGetAll_Handler_200_CustomPagination(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)

	svc.On("GetAll", 2, 5).Return([]domain.Municipality{}, int64(0), nil)

	w := doRequest(setupRouter(svc), http.MethodGet, "/api/v1/municipalities?page=2&pageSize=5", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertCalled(t, "GetAll", 2, 5)
}

func TestGetAll_Handler_PageSizeCappedAt100(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)

	// pageSize=999 deve ser normalizado para 20 (default ao ultrapassar max)
	svc.On("GetAll", 1, 20).Return([]domain.Municipality{}, int64(0), nil)

	w := doRequest(setupRouter(svc), http.MethodGet, "/api/v1/municipalities?pageSize=999", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertCalled(t, "GetAll", 1, 20)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GET /api/v1/municipalities/:id
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetByID_Handler_200(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	m := testhelper.MakePassagem()

	svc.On("GetByID", m.ID).Return(&m, nil)

	path := fmt.Sprintf("/api/v1/municipalities/%s", m.ID)
	w := doRequest(setupRouter(svc), http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, m.ID.String(), data["id"])
}

func TestGetByID_Handler_404(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)

	svc.On("GetByID", testhelper.NonExistentID).Return(nil, service.ErrMunicipalityNotFound)

	path := fmt.Sprintf("/api/v1/municipalities/%s", testhelper.NonExistentID)
	w := doRequest(setupRouter(svc), http.MethodGet, path, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetByID_Handler_400_InvalidUUID(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)

	w := doRequest(setupRouter(svc), http.MethodGet, "/api/v1/municipalities/nao-e-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "GetByID")
}

// ═══════════════════════════════════════════════════════════════════════════════
// GET /api/v1/municipalities/uf/:uf
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetByUF_Handler_200(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	municipalities := []domain.Municipality{testhelper.MakePassagem()}

	svc.On("GetByUF", "PB", 1, 20).Return(municipalities, int64(1), nil)

	w := doRequest(setupRouter(svc), http.MethodGet, "/api/v1/municipalities/uf/PB", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
}

func TestGetByUF_Handler_400_InvalidUF(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)

	svc.On("GetByUF", "ZZ", 1, 20).Return([]domain.Municipality{}, int64(0), service.ErrInvalidUF)

	w := doRequest(setupRouter(svc), http.MethodGet, "/api/v1/municipalities/uf/ZZ", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// PATCH /api/v1/municipalities/:id
// ═══════════════════════════════════════════════════════════════════════════════

func TestUpdate_Handler_200(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	m := testhelper.MakePassagem()
	newName := "Passagem Nova"
	input := domain.UpdateMunicipalityInput{Name: &newName}

	updated := m
	updated.Name = "Passagem Nova"
	svc.On("Update", m.ID, input).Return(&updated, nil)

	path := fmt.Sprintf("/api/v1/municipalities/%s", m.ID)
	w := doRequest(setupRouter(svc), http.MethodPatch, path, input)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.Equal(t, "Passagem Nova", resp["data"].(map[string]any)["name"])
}

func TestUpdate_Handler_404(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	newName := "Qualquer"
	input := domain.UpdateMunicipalityInput{Name: &newName}

	svc.On("Update", testhelper.NonExistentID, input).Return(nil, service.ErrMunicipalityNotFound)

	path := fmt.Sprintf("/api/v1/municipalities/%s", testhelper.NonExistentID)
	w := doRequest(setupRouter(svc), http.MethodPatch, path, input)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// DELETE /api/v1/municipalities/:id
// ═══════════════════════════════════════════════════════════════════════════════

func TestDelete_Handler_204(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	m := testhelper.MakePassagem()

	svc.On("Delete", m.ID).Return(nil)

	path := fmt.Sprintf("/api/v1/municipalities/%s", m.ID)
	w := doRequest(setupRouter(svc), http.MethodDelete, path, nil)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestDelete_Handler_404(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)

	svc.On("Delete", testhelper.NonExistentID).Return(service.ErrMunicipalityNotFound)

	path := fmt.Sprintf("/api/v1/municipalities/%s", testhelper.NonExistentID)
	w := doRequest(setupRouter(svc), http.MethodDelete, path, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHardDelete_Handler_204(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)
	m := testhelper.MakePassagem()

	svc.On("HardDelete", m.ID).Return(nil)

	path := fmt.Sprintf("/api/v1/municipalities/%s/hard", m.ID)
	w := doRequest(setupRouter(svc), http.MethodDelete, path, nil)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
	svc.AssertExpectations(t)
}

func TestHardDelete_Handler_404(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)

	svc.On("HardDelete", testhelper.NonExistentID).Return(service.ErrMunicipalityNotFound)

	path := fmt.Sprintf("/api/v1/municipalities/%s/hard", testhelper.NonExistentID)
	w := doRequest(setupRouter(svc), http.MethodDelete, path, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHardDelete_Handler_400_InvalidUUID(t *testing.T) {
	svc := new(handlerMocks.MunicipalityService)

	w := doRequest(setupRouter(svc), http.MethodDelete, "/api/v1/municipalities/nao-e-uuid/hard", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "HardDelete")
}
