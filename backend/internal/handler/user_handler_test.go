package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler"
	handlerMocks "github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupUserRouter monta o Gin com o handler e o mock de service injetado.
func setupUserRouter(svc domain.UserService) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		claims := &security.UserClaims{
			UserID:         uuid.New(),
			Username:       "admin_test",
			Role:           string(domain.RoleAdmin),
			MunicipalityID: uuid.New(),
		}
		c.Set("user", claims)
		c.Next()
	})
	h := handler.NewUserHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

// ═══════════════════════════════════════════════════════════════════════════════
// POST /api/v1/users
// ═══════════════════════════════════════════════════════════════════════════════

func TestCreateUser_Handler_201(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	input := domain.CreateUserInput{
		Username:        u.Username,
		Email:           u.Email,
		Role:            u.Role,
		MunicipalityID:  u.MunicipalityID,
	}

	svc.On("Create", input).Return(&u, "randomPassword123", nil)

	w := doRequest(setupUserRouter(svc), http.MethodPost, "/api/v1/users", input)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, u.Username, resp["data"].(map[string]any)["user"].(map[string]any)["username"])
	assert.Equal(t, "randomPassword123", resp["data"].(map[string]any)["randomPassword"])
	svc.AssertExpectations(t)
}

func TestCreateUser_Handler_400_Validation(t *testing.T) {
	svc := new(handlerMocks.UserService)
	// input sem username obrigatório
	input := domain.CreateUserInput{
		Email:           "invalid@example.com",
		Role:            domain.RoleCommon,
		MunicipalityID:  testhelper.MunPassagemID,
	}

	w := doRequest(setupUserRouter(svc), http.MethodPost, "/api/v1/users", input)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestCreateUser_Handler_409_EmailConflict(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	input := domain.CreateUserInput{
		Username:        u.Username,
		Email:           u.Email,
		Role:            u.Role,
		MunicipalityID:  u.MunicipalityID,
	}

	svc.On("Create", input).Return(nil, "", domain.ErrEmailAlreadyExists)

	w := doRequest(setupUserRouter(svc), http.MethodPost, "/api/v1/users", input)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateUser_Handler_400_MunicipalityNotFound(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.NonExistentID)
	input := domain.CreateUserInput{
		Username:        u.Username,
		Email:           u.Email,
		Role:            u.Role,
		MunicipalityID:  u.MunicipalityID,
	}

	svc.On("Create", input).Return(nil, "", service.ErrMunicipalityNotFound)

	w := doRequest(setupUserRouter(svc), http.MethodPost, "/api/v1/users", input)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GET /api/v1/users
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetAllUsers_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	users := []domain.User{
		testhelper.MakeUserAdmin(testhelper.MunPassagemID),
		testhelper.MakeUserCommon(testhelper.MunPassagemID),
	}

	svc.On("GetAll", 1, 20).Return(users, int64(2), nil)

	w := doRequest(setupUserRouter(svc), http.MethodGet, "/api/v1/users", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].([]any)
	assert.Len(t, data, 2)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GET /api/v1/users/trash
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetDeletedUsers_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	users := []domain.User{
		testhelper.MakeUserCommon(testhelper.MunPassagemID),
	}

	svc.On("GetDeleted", 1, 20).Return(users, int64(1), nil)

	w := doRequest(setupUserRouter(svc), http.MethodGet, "/api/v1/users/trash", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].([]any)
	assert.Len(t, data, 1)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GET /api/v1/users/:id
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetUserByID_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)

	svc.On("GetByID", u.ID).Return(&u, nil)

	path := fmt.Sprintf("/api/v1/users/%s", u.ID)
	w := doRequest(setupUserRouter(svc), http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, u.ID.String(), data["id"])
}

func TestGetUserByID_Handler_404(t *testing.T) {
	svc := new(handlerMocks.UserService)
	id := uuid.New()

	svc.On("GetByID", id).Return(nil, domain.ErrUserNotFound)

	path := fmt.Sprintf("/api/v1/users/%s", id)
	w := doRequest(setupUserRouter(svc), http.MethodGet, path, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// PATCH /api/v1/users/:id
// ═══════════════════════════════════════════════════════════════════════════════

func TestUpdateUser_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	newEmail := "updated@example.com"
	input := domain.UpdateUserInput{Email: &newEmail}

	updated := u
	updated.Email = newEmail

	svc.On("GetByID", u.ID).Return(&u, nil)
	svc.On("Update", u.ID, input).Return(&updated, "", nil)

	path := fmt.Sprintf("/api/v1/users/%s", u.ID)
	w := doRequest(setupUserRouter(svc), http.MethodPatch, path, input)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.Equal(t, newEmail, resp["data"].(map[string]any)["email"])
}

func TestUpdateUser_Municipality_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	newMunID := testhelper.MunPatosID
	patos := testhelper.MakePatos()

	input := domain.UpdateUserInput{MunicipalityID: &newMunID}

	updated := u
	updated.MunicipalityID = newMunID
	updated.Municipality = patos

	svc.On("GetByID", u.ID).Return(&u, nil)
	svc.On("Update", u.ID, input).Return(&updated, "", nil)

	path := fmt.Sprintf("/api/v1/users/%s", u.ID)
	w := doRequest(setupUserRouter(svc), http.MethodPatch, path, input)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseBody(t, w, &resp)
	assert.Equal(t, newMunID.String(), resp["data"].(map[string]any)["municipalityId"])
}

// ═══════════════════════════════════════════════════════════════════════════════
// DELETE /api/v1/users/:id
// ═══════════════════════════════════════════════════════════════════════════════

func TestDeleteUser_Handler_204(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)

	svc.On("GetByID", u.ID).Return(&u, nil)
	svc.On("Delete", u.ID).Return(nil)

	path := fmt.Sprintf("/api/v1/users/%s", u.ID)
	w := doRequest(setupUserRouter(svc), http.MethodDelete, path, nil)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// PATCH /api/v1/users/:id/restore
// ═══════════════════════════════════════════════════════════════════════════════

func TestRestoreUser_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)

	svc.On("GetByIDUnscoped", u.ID).Return(&u, nil)
	svc.On("Restore", u.ID).Return(&u, nil)

	path := fmt.Sprintf("/api/v1/users/%s/restore", u.ID)
	w := doRequest(setupUserRouter(svc), http.MethodPatch, path, nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// DELETE /api/v1/users/:id/hard
// ═══════════════════════════════════════════════════════════════════════════════

func TestHardDeleteUser_Handler_204(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)

	svc.On("GetByIDUnscoped", u.ID).Return(&u, nil)
	svc.On("HardDelete", u.ID).Return(nil)

	path := fmt.Sprintf("/api/v1/users/%s/hard", u.ID)
	w := doRequest(setupUserRouter(svc), http.MethodDelete, path, nil)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GET /api/v1/users/:id/permissions
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetPermissions_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)

	perms := []domain.UserPermission{
		{UserID: u.ID, DocumentType: domain.TypeNotice, Level: domain.LevelRead},
	}

	svc.On("GetByID", u.ID).Return(&u, nil)
	svc.On("GetPermissions", u.ID).Return(perms, nil)

	path := fmt.Sprintf("/api/v1/users/%s/permissions", u.ID)
	w := doRequest(setupUserRouter(svc), http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ═══════════════════════════════════════════════════════════════════════════════
// PUT /api/v1/users/:id/permissions
// ═══════════════════════════════════════════════════════════════════════════════

func TestUpdatePermissions_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)

	input := domain.UpdateUserPermissionsInput{
		Permissions: []domain.UpdateUserPermissionItem{
			{DocumentType: domain.TypeNotice, Level: domain.LevelWrite},
		},
	}

	perms := []domain.UserPermission{
		{UserID: u.ID, DocumentType: domain.TypeNotice, Level: domain.LevelWrite},
	}

	svc.On("GetByID", u.ID).Return(&u, nil)
	svc.On("UpdatePermissions", u.ID, input).Return(perms, nil)

	path := fmt.Sprintf("/api/v1/users/%s/permissions", u.ID)
	w := doRequest(setupUserRouter(svc), http.MethodPut, path, input)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChangePassword_Handler_200(t *testing.T) {
	svc := new(handlerMocks.UserService)
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)

	input := domain.ChangePasswordInput{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	svc.On("ChangePassword", mock.Anything, input).Return(&u, nil)

	w := doRequest(setupUserRouter(svc), http.MethodPost, "/api/v1/users/me/change-password", input)

	assert.Equal(t, http.StatusOK, w.Code)
}
