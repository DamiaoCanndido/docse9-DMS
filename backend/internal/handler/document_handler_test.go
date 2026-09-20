package handler_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler"
	handlerMocks "github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
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

func TestDeleteDocument_Handler_204(t *testing.T) {
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

	svc.On("GetByID", docID).Return(doc, nil)
	svc.On("Delete", docID).Return(nil)
	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelDelete,
		},
	}, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodDelete, fmt.Sprintf("/api/v1/documents/%s", docID), nil)

	assert.Equal(t, http.StatusNoContent, w.Code)
	svc.AssertExpectations(t)
	permRepo.AssertExpectations(t)
}

func TestDeleteDocument_Handler_403_ForbiddenOtherUser(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		CreatorID:      otherUserID, // Created by someone else
		MunicipalityID: munID,
	}

	svc.On("GetByID", docID).Return(doc, nil)
	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{
		{
			UserID:       userID,
			DocumentType: domain.TypeNotice,
			Level:        domain.LevelDelete,
		},
	}, nil)

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodDelete, fmt.Sprintf("/api/v1/documents/%s", docID), nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "Delete")
}

func TestUpdateDocument_Handler_403_ForbiddenOtherUser(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		CreatorID:      otherUserID, // Created by someone else
		MunicipalityID: munID,
	}

	newDesc := "Tentativa de alteração"
	input := domain.UpdateDocumentInput{
		Description: &newDesc,
	}

	svc.On("GetByID", docID).Return(doc, nil)
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

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "Update")
}

func TestDocument_AdminAccessDenied(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	docID := uuid.New()
	munID := uuid.New()
	adminID := uuid.New()

	doc := &domain.Document{
		ID:             docID,
		Type:           domain.TypeNotice,
		CreatorID:      adminID,
		MunicipalityID: munID,
	}

	svc.On("GetByID", docID).Return(doc, nil)

	claims := &security.UserClaims{
		UserID:         adminID,
		Role:           string(domain.RoleAdmin),
		MunicipalityID: munID,
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s", docID), nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetDeletedDocuments_Handler_200(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	munID := uuid.New()
	modID := uuid.New()

	docs := []domain.Document{
		{ID: uuid.New(), Description: "Doc Lixeira", Type: domain.TypeNotice, MunicipalityID: munID},
	}

	claims := &security.UserClaims{
		UserID:         modID,
		Role:           string(domain.RoleMod),
		MunicipalityID: munID,
	}

	svc.On("GetDeleted", mock.MatchedBy(func(filter domain.DocumentFilter) bool {
		return filter.MunicipalityID != nil && *filter.MunicipalityID == munID
	}), 1, 20).Return(docs, int64(1), nil)

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, "/api/v1/documents/trash", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestGetAllDocuments_Handler_CommonUser_NoPermissions_403(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	munID := uuid.New()
	userID := uuid.New()

	claims := &security.UserClaims{
		UserID:         userID,
		Role:           string(domain.RoleCommon),
		MunicipalityID: munID,
	}

	permRepo.On("FindByUserID", userID).Return([]domain.UserPermission{}, nil) // 0 permissões

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, "/api/v1/documents", nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "GetAll")
}

func TestGetDocumentByID_Handler_400_InvalidUUID(t *testing.T) {
	svc := new(handlerMocks.DocumentService)
	permRepo := new(handlerMocks.UserPermissionRepository)
	claims := &security.UserClaims{
		UserID:         uuid.New(),
		Role:           string(domain.RoleMod),
		MunicipalityID: uuid.New(),
	}

	w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, "/api/v1/documents/not-a-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocument_Handler_Errors(t *testing.T) {
	munID := uuid.New()
	modID := uuid.New()
	claims := &security.UserClaims{
		UserID:         modID,
		Role:           string(domain.RoleMod),
		MunicipalityID: munID,
	}

	t.Run("CreateDocument 400 InvalidDocumentType", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		input := domain.CreateDocumentInput{Type: "invalid_type", Description: "Test", MunicipalityID: munID, CreatorID: modID}
		svc.On("Create", input).Return(nil, domain.ErrInvalidDocumentType)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, "/api/v1/documents", input)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateDocument 400 MunicipalityNotFound", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		input := domain.CreateDocumentInput{Type: domain.TypeNotice, Description: "Test", MunicipalityID: munID, CreatorID: modID}
		svc.On("Create", input).Return(nil, service.ErrMunicipalityNotFound)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, "/api/v1/documents", input)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateDocument 500 InternalError", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		input := domain.CreateDocumentInput{Type: domain.TypeNotice, Description: "Test", MunicipalityID: munID, CreatorID: modID}
		svc.On("Create", input).Return(nil, assert.AnError)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, "/api/v1/documents", input)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GetByID 404 NotFound", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		svc.On("GetByID", docID).Return(nil, domain.ErrDocumentNotFound)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s", docID), nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update 404 NotFound", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		svc.On("GetByID", docID).Return(nil, domain.ErrDocumentNotFound)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPatch, fmt.Sprintf("/api/v1/documents/%s", docID), map[string]string{"description": "new"})
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Delete 404 NotFound", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		svc.On("GetByID", docID).Return(nil, domain.ErrDocumentNotFound)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodDelete, fmt.Sprintf("/api/v1/documents/%s", docID), nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Restore 404 NotFound", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		svc.On("GetByIDUnscoped", docID).Return(nil, domain.ErrDocumentNotFound)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPatch, fmt.Sprintf("/api/v1/documents/%s/restore", docID), nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("HardDelete 404 NotFound", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		svc.On("GetByIDUnscoped", docID).Return(nil, domain.ErrDocumentNotFound)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodDelete, fmt.Sprintf("/api/v1/documents/%s/hard", docID), nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGenerateUploadURL_Handler(t *testing.T) {
	munID := uuid.New()
	userID := uuid.New()
	claims := &security.UserClaims{
		UserID:         userID,
		MunicipalityID: munID,
		Role:           string(domain.RoleMod),
	}

	t.Run("200 OK", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		doc := &domain.Document{
			ID:             docID,
			MunicipalityID: munID,
			Type:           domain.TypeNotice,
		}

		svc.On("GetByID", docID).Return(doc, nil)
		input := domain.UploadURLInput{
			FileName:    "decreto.pdf",
			FileSize:    1024 * 500,
			ContentType: "application/pdf",
		}
		expectedRes := &domain.UploadURLResponse{
			UploadURL:        "https://r2.example.com/upload",
			FileKey:          "tenants/mun/file.pdf",
			ExpiresInSeconds: 600,
		}
		svc.On("GenerateUploadURL", mock.Anything, docID, input).Return(expectedRes, nil)

		body := map[string]any{
			"fileName":    "decreto.pdf",
			"fileSize":    1024 * 500,
			"contentType": "application/pdf",
		}
		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, fmt.Sprintf("/api/v1/documents/%s/upload-url", docID), body)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("400 Bad Request - Non PDF", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		doc := &domain.Document{
			ID:             docID,
			MunicipalityID: munID,
			Type:           domain.TypeNotice,
		}

		svc.On("GetByID", docID).Return(doc, nil)

		body := map[string]any{
			"fileName":    "imagem.png",
			"fileSize":    1024,
			"contentType": "image/png",
		}
		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, fmt.Sprintf("/api/v1/documents/%s/upload-url", docID), body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestConfirmUpload_Handler(t *testing.T) {
	munID := uuid.New()
	userID := uuid.New()
	claims := &security.UserClaims{
		UserID:         userID,
		MunicipalityID: munID,
		Role:           string(domain.RoleMod),
	}

	t.Run("200 OK", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		doc := &domain.Document{
			ID:             docID,
			MunicipalityID: munID,
			Type:           domain.TypeNotice,
		}

		svc.On("GetByID", docID).Return(doc, nil)
		input := domain.ConfirmUploadInput{FileKey: "tenants/key.pdf"}
		updatedDoc := &domain.Document{
			ID:             docID,
			MunicipalityID: munID,
			FileKey:        "tenants/key.pdf",
		}
		svc.On("ConfirmUpload", mock.Anything, docID, input).Return(updatedDoc, nil)

		body := map[string]any{"fileKey": "tenants/key.pdf"}
		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, fmt.Sprintf("/api/v1/documents/%s/confirm-upload", docID), body)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("422 Unprocessable Entity - Missing OCR", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		doc := &domain.Document{
			ID:             docID,
			MunicipalityID: munID,
			Type:           domain.TypeNotice,
		}

		svc.On("GetByID", docID).Return(doc, nil)
		input := domain.ConfirmUploadInput{FileKey: "tenants/scanned_sem_ocr.pdf"}
		svc.On("ConfirmUpload", mock.Anything, docID, input).Return(nil, domain.ErrPDFMissingOCR)

		body := map[string]any{"fileKey": "tenants/scanned_sem_ocr.pdf"}
		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodPost, fmt.Sprintf("/api/v1/documents/%s/confirm-upload", docID), body)
		assert.Equal(t, 422, w.Code)
	})
}

func TestGenerateFileURL_Handler(t *testing.T) {
	munID := uuid.New()
	userID := uuid.New()
	claims := &security.UserClaims{
		UserID:         userID,
		MunicipalityID: munID,
		Role:           string(domain.RoleMod),
	}

	t.Run("200 OK", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		doc := &domain.Document{
			ID:             docID,
			MunicipalityID: munID,
			Type:           domain.TypeNotice,
			FileKey:        "tenants/key.pdf",
		}

		svc.On("GetByIDUnscoped", docID).Return(doc, nil)
		res := &domain.FileURLResponse{
			URL:              "https://r2.example.com/download",
			ExpiresInSeconds: 900,
		}
		svc.On("GenerateFileURL", mock.Anything, docID, false).Return(res, nil)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s/file-url", docID), nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("200 OK - Soft Deleted Document in Trash", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		doc := &domain.Document{
			ID:             docID,
			MunicipalityID: munID,
			Type:           domain.TypeNotice,
			FileKey:        "tenants/trashed.pdf",
			DeletedAt:      gorm.DeletedAt{Time: time.Now(), Valid: true},
		}

		svc.On("GetByIDUnscoped", docID).Return(doc, nil)
		res := &domain.FileURLResponse{
			URL:              "https://r2.example.com/download-trashed",
			ExpiresInSeconds: 900,
		}
		svc.On("GenerateFileURL", mock.Anything, docID, false).Return(res, nil)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s/file-url", docID), nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("404 NotFound - No File", func(t *testing.T) {
		svc := new(handlerMocks.DocumentService)
		permRepo := new(handlerMocks.UserPermissionRepository)
		docID := uuid.New()
		doc := &domain.Document{
			ID:             docID,
			MunicipalityID: munID,
			Type:           domain.TypeNotice,
			FileKey:        "",
		}

		svc.On("GetByIDUnscoped", docID).Return(doc, nil)
		svc.On("GenerateFileURL", mock.Anything, docID, false).Return(nil, domain.ErrFileNotFound)

		w := doRequest(setupDocumentRouter(svc, permRepo, claims), http.MethodGet, fmt.Sprintf("/api/v1/documents/%s/file-url", docID), nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}



