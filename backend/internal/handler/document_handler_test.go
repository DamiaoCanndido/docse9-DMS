package handler_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler"
	handlerMocks "github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupDocumentRouter(svc domain.DocumentService, permRepo domain.UserPermissionRepository, claims *security.UserClaims) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if claims != nil {
			c.Set("user", claims)
		}
		c.Next()
	})
	h := handler.NewDocumentHandler(svc, permRepo)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestCreateDocument_Handler_201(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	munID := uuid.New()
	userID := uuid.New()

	input := domain.CreateDocumentInput{
		Type:           domain.TypeNotice,
		Description:    "Oficio de Teste",
		CreatorID:      userID,
		MunicipalityID: munID,
	}

	doc := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio de Teste",
		CreatorID:      userID,
		MunicipalityID: munID,
	}

	svc.On("Create", input).Return(doc, nil)
	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelWrite,
		},
	}, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, "/api/v1/documents", input)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
	permRepo.AssertExpectations(t)
}

func TestCreateDocument_Handler_403_ForbiddenPermission(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	munID := uuid.New()
	userID := uuid.New()

	input := domain.CreateDocumentInput{
		Type:        domain.TypeNotice,
		Description: "Oficio de Teste",
	}

	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelRead, // Apenas READ! Não WRITE/DELETE!
		},
	}, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, "/api/v1/documents", input)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "Create")
	permRepo.AssertExpectations(t)
}

func TestGetDocumentByID_Handler_200(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio de Teste",
		CreatorID:      userID,
		MunicipalityID: munID,
	}

	svc.On("GetByID", docID).Return(doc, nil)
	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelRead,
		},
	}, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s", docID), nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
	permRepo.AssertExpectations(t)
}

func TestGetDocumentByID_Handler_403_ForbiddenTenant(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio de Teste",
		CreatorID:      userID,
		MunicipalityID: uuid.New(), // Outro municipio!
	}

	svc.On("GetByID", docID).Return(doc, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s", docID), nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAllDocuments_Handler_200(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
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

	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelRead,
		},
	}, nil)

	// Como o usuário é COMMON, a rota deve forçar a busca pelo MunicipalityID dele nas claims
	svc.On("GetAll", mock.MatchedBy(func(filter domain.DocumentFilter) bool {
		return filter.MunicipalityID != nil && *filter.MunicipalityID == munID &&
			len(filter.AllowedTypes) == 1 && filter.AllowedTypes[0] == domain.TypeNotice
	}), 1, 20).Return(docs, int64(1), nil)

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, "/api/v1/documents", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
	permRepo.AssertExpectations(t)
}

func TestRestoreDocument_Handler_200(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()

	deletedDoc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		CreatorID:      userID,
		MunicipalityID: munID,
		DeletedAt:      gorm.DeletedAt{Time: time.Now(), Valid: true},
	}

	restoredDoc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		CreatorID:      userID,
		MunicipalityID: munID,
	}

	svc.On("GetByIDUnscoped", docID).Return(deletedDoc, nil)
	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelDelete,
		},
	}, nil)
	svc.On("Restore", docID).Return(restoredDoc, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPatch, fmt.Sprintf("/api/v1/documents/%s/restore", docID), nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
	permRepo.AssertExpectations(t)
}

func TestRestoreDocument_Handler_403_ForbiddenPermission(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()

	deletedDoc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		CreatorID:      userID,
		MunicipalityID: munID,
		DeletedAt:      gorm.DeletedAt{Time: time.Now(), Valid: true},
	}

	svc.On("GetByIDUnscoped", docID).Return(deletedDoc, nil)
	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelRead, // Only Read, not Delete!
		},
	}, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPatch, fmt.Sprintf("/api/v1/documents/%s/restore", docID), nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "Restore")
}

func TestHardDeleteDocument_Handler_204(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		CreatorID:      userID,
		MunicipalityID: munID,
	}

	svc.On("GetByIDUnscoped", docID).Return(doc, nil)
	svc.On("HardDelete", docID).Return(nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleMod),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodDelete, fmt.Sprintf("/api/v1/documents/%s/hard", docID), nil)

	assert.Equal(t, http.StatusNoContent, w.Code)
	svc.AssertExpectations(t)
}

func TestHardDeleteDocument_Handler_403_ForbiddenTenant(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	otherMunID := uuid.New()
	userID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		CreatorID:      userID,
		MunicipalityID: otherMunID, // Different municipality!
	}

	svc.On("GetByIDUnscoped", docID).Return(doc, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleMod),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodDelete, fmt.Sprintf("/api/v1/documents/%s/hard", docID), nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "HardDelete")
}

func TestUpdateDocument_Handler_200(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()

	existingDoc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		Description:    "Oficio Original",
		CreatorID:      userID,
		MunicipalityID: munID,
		CreatedAt:      time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
	}

	newDesc := "Oficio Atualizado"
	newCreatedAt := time.Date(2023, 5, 20, 15, 30, 0, 0, time.UTC)
	input := domain.UpdateDocumentInput{
		Description: &newDesc,
		CreatedAt:   &newCreatedAt,
	}

	updatedDoc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		Description:    newDesc,
		CreatorID:      userID,
		MunicipalityID: munID,
		CreatedAt:      newCreatedAt,
	}

	svc.On("GetByID", docID).Return(existingDoc, nil)
	svc.On("Update", docID, input).Return(updatedDoc, nil)
	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelWrite,
		},
	}, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPatch, fmt.Sprintf("/api/v1/documents/%s", docID), input)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
	permRepo.AssertExpectations(t)
}
