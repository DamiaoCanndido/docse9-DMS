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
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupSequenceRouter(svc domain.SequenceService, claims *security.UserClaims) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if claims != nil {
			c.Set("user", claims)
		}
		c.Next()
	})
	h := handler.NewSequenceHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestGetSequences_Unauthorized_401(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	r := setupSequenceRouter(svc, nil)

	munID := uuid.New()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/municipalities/%s/sequences", munID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSequences_Forbidden_Admin_403(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	munID := uuid.New()
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleAdmin),
		MunicipalityID: munID,
	}
	r := setupSequenceRouter(svc, claims)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/municipalities/%s/sequences", munID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetSequences_Forbidden_Common_403(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	munID := uuid.New()
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}
	r := setupSequenceRouter(svc, claims)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/municipalities/%s/sequences", munID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetSequences_Forbidden_DifferentMunicipality_403(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	munID := uuid.New()
	otherMunID := uuid.New()
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleMod),
		MunicipalityID: munID,
	}
	r := setupSequenceRouter(svc, claims)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/municipalities/%s/sequences", otherMunID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetSequences_Success_Mod_200(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	munID := uuid.New()
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleMod),
		MunicipalityID: munID,
	}
	r := setupSequenceRouter(svc, claims)

	expectedItems := []domain.SequenceItemResponse{
		{
			Type:         domain.TypeDecree,
			InitialOrder: 85,
			CurrentOrder: 0,
			NextOrder:    85,
		},
	}
	year := 2026
	svc.On("GetSequences", munID, &year).Return(expectedItems, nil)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/municipalities/%s/sequences?year=2026", munID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSetOffset_Forbidden_Admin_403(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	munID := uuid.New()
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleAdmin),
		MunicipalityID: munID,
	}
	r := setupSequenceRouter(svc, claims)

	year := 2026
	body, _ := json.Marshal(domain.SetSequenceOffsetInput{
		Type:         domain.TypeDecree,
		Year:         &year,
		InitialOrder: 85,
	})

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/municipalities/%s/sequences", munID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSetOffset_Forbidden_Common_403(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	munID := uuid.New()
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}
	r := setupSequenceRouter(svc, claims)

	year := 2026
	body, _ := json.Marshal(domain.SetSequenceOffsetInput{
		Type:         domain.TypeDecree,
		Year:         &year,
		InitialOrder: 85,
	})

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/municipalities/%s/sequences", munID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSetOffset_Success_Mod_200(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	munID := uuid.New()
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleMod),
		MunicipalityID: munID,
	}
	r := setupSequenceRouter(svc, claims)

	year := 2026
	input := domain.SetSequenceOffsetInput{
		Type:         domain.TypeDecree,
		Year:         &year,
		InitialOrder: 85,
	}
	body, _ := json.Marshal(input)

	savedOffset := &domain.SequenceOffset{
		ID:             uuid.New(),
		MunicipalityID: munID,
		Type:           domain.TypeDecree,
		Year:           &year,
		InitialOrder:   85,
	}
	svc.On("SetOffset", munID, input).Return(savedOffset, nil)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/municipalities/%s/sequences", munID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSetOffset_BadRequest_InvalidOrder_400(t *testing.T) {
	svc := new(handlerMocks.SequenceService)
	munID := uuid.New()
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleMod),
		MunicipalityID: munID,
	}
	r := setupSequenceRouter(svc, claims)

	year := 2026
	// InitialOrder = 0 viola binding:"required,gte=1"
	body, _ := json.Marshal(map[string]any{
		"type":         "DECREE",
		"year":         year,
		"initialOrder": 0,
	})

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/municipalities/%s/sequences", munID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
