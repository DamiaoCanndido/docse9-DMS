package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler"
	handlerMocks "github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDocumentRouter(svc domain.DocumentService, claims *security.UserClaims) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if claims != nil {
			c.Set("user", claims)
		}
		c.Next()
	})
	h := handler.NewDocumentHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestCreateDocument_Handler_201(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	munID := uuid.New()
	userID := uuid.New()

	input := domain.CreateDocumentInput{
		Type:           domain.TypeNotice,
		Description:    "Oficio de Teste",
		OwnerID:        userID,
		MunicipalityID: munID,
	}

	doc := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio de Teste",
		OwnerID:        userID,
		MunicipalityID: munID,
	}

	svc.On("Create", input).Return(doc, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, claims), http.MethodPost, "/api/v1/documents", input)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestCreateDocument_Handler_403_ForbiddenTenant(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	munID := uuid.New()
	userID := uuid.New()

	input := domain.CreateDocumentInput{
		Type:           domain.TypeNotice,
		Description:    "Oficio de Teste",
		OwnerID:        userID,
		MunicipalityID: uuid.New(), // Outro municipio!
	}

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, claims), http.MethodPost, "/api/v1/documents", input)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestGetDocumentByID_Handler_200(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio de Teste",
		OwnerID:        userID,
		MunicipalityID: munID,
	}

	svc.On("GetByID", docID).Return(doc, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s", docID), nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestGetDocumentByID_Handler_403_ForbiddenTenant(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio de Teste",
		OwnerID:        userID,
		MunicipalityID: uuid.New(), // Outro municipio!
	}

	svc.On("GetByID", docID).Return(doc, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s", docID), nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAllDocuments_Handler_200(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	munID := uuid.New()
	userID := uuid.New()

	docs := []domain.Document{
		{ID: uuid.New(), Description: "Doc 1", Type: domain.TypeNotice, MunicipalityID: munID},
	}

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	// Como o usuário é COMMON, a rota deve forçar a busca pelo MunicipalityID dele nas claims
	svc.On("GetAll", mock.MatchedBy(func(filter domain.DocumentFilter) bool {
		return filter.MunicipalityID != nil && *filter.MunicipalityID == munID
	}), 1, 20).Return(docs, int64(1), nil)

	w := doRequest(setupDocumentRouter(svc, claims), http.MethodGet, "/api/v1/documents", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}
