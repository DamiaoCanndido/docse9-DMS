package service_test

import (
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newUserService(t *testing.T) (domain.UserService, *mocks.UserRepository, *mocks.MunicipalityRepository) {
	t.Helper()
	userRepo := new(mocks.UserRepository)
	munRepo := new(mocks.MunicipalityRepository)
	svc := service.NewUserService(userRepo, munRepo)
	return svc, userRepo, munRepo
}

// ═══════════════════════════════════════════════════════════════════════════════
// Create
// ═══════════════════════════════════════════════════════════════════════════════

func TestCreateUser_Success(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	mun := testhelper.MakePassagem()
	input := domain.CreateUserInput{
		Username:        "new_user",
		Email:           "new@example.com",
		Password:        "secret123",
		ConfirmPassword: "secret123",
		Role:            domain.RoleCommon,
		MunicipalityID:  mun.ID,
	}

	userRepo.On("ExistsByEmail", "new@example.com", (*uuid.UUID)(nil)).Return(false, nil)
	userRepo.On("ExistsByUsername", "new_user", (*uuid.UUID)(nil)).Return(false, nil)
	munRepo.On("FindByID", mun.ID).Return(&mun, nil)
	userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

	result, err := svc.Create(input)

	require.NoError(t, err)
	assert.Equal(t, "new_user", result.Username)
	assert.Equal(t, "new@example.com", result.Email)
	assert.Equal(t, "secret123", result.Password)
	assert.Equal(t, domain.RoleCommon, result.Role)
	assert.Equal(t, mun.ID, result.MunicipalityID)
	assert.NotEqual(t, uuid.Nil, result.ID)
	userRepo.AssertExpectations(t)
	munRepo.AssertExpectations(t)
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	mun := testhelper.MakePassagem()
	input := domain.CreateUserInput{
		Username:       "new_user",
		Email:          "existing@example.com",
		Password:       "secret123",
		Role:           domain.RoleCommon,
		MunicipalityID: mun.ID,
	}

	userRepo.On("ExistsByEmail", "existing@example.com", (*uuid.UUID)(nil)).Return(true, nil)

	_, err := svc.Create(input)

	assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	userRepo.AssertNotCalled(t, "ExistsByUsername")
	munRepo.AssertNotCalled(t, "FindByID")
	userRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_UsernameAlreadyExists(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	mun := testhelper.MakePassagem()
	input := domain.CreateUserInput{
		Username:       "existing_user",
		Email:          "new@example.com",
		Password:       "secret123",
		Role:           domain.RoleCommon,
		MunicipalityID: mun.ID,
	}

	userRepo.On("ExistsByEmail", "new@example.com", (*uuid.UUID)(nil)).Return(false, nil)
	userRepo.On("ExistsByUsername", "existing_user", (*uuid.UUID)(nil)).Return(true, nil)

	_, err := svc.Create(input)

	assert.ErrorIs(t, err, domain.ErrUsernameAlreadyExists)
	munRepo.AssertNotCalled(t, "FindByID")
	userRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_MunicipalityNotFound(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	input := domain.CreateUserInput{
		Username:       "new_user",
		Email:          "new@example.com",
		Password:       "secret123",
		Role:           domain.RoleCommon,
		MunicipalityID: testhelper.NonExistentID,
	}

	userRepo.On("ExistsByEmail", "new@example.com", (*uuid.UUID)(nil)).Return(false, nil)
	userRepo.On("ExistsByUsername", "new_user", (*uuid.UUID)(nil)).Return(false, nil)
	munRepo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	_, err := svc.Create(input)

	assert.ErrorIs(t, err, service.ErrMunicipalityNotFound)
	userRepo.AssertNotCalled(t, "Create")
}

// ═══════════════════════════════════════════════════════════════════════════════
// GetByID
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetUserByID_Found(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)

	userRepo.On("FindByID", u.ID).Return(&u, nil)

	result, err := svc.GetByID(u.ID)

	require.NoError(t, err)
	assert.Equal(t, u.ID, result.ID)
	assert.Equal(t, u.Username, result.Username)
}

func TestGetUserByID_NotFound(t *testing.T) {
	svc, userRepo, _ := newUserService(t)

	userRepo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	_, err := svc.GetByID(testhelper.NonExistentID)

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GetAll
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetAllUsers_Success(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	users := []domain.User{testhelper.MakeUserCommon(mun.ID), testhelper.MakeUserAdmin(mun.ID)}

	userRepo.On("FindAll", 1, 10).Return(users, int64(2), nil)

	result, total, err := svc.GetAll(1, 10)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GetDeleted
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetDeletedUsers_Success(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	users := []domain.User{testhelper.MakeUserCommon(mun.ID)}

	userRepo.On("FindDeleted", 1, 10).Return(users, int64(1), nil)

	result, total, err := svc.GetDeleted(1, 10)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Update
// ═══════════════════════════════════════════════════════════════════════════════

func TestUpdateUser_PartialSuccess(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)

	newUsername := "updated_name"
	newEmail := "updated@example.com"
	newPassword := "newpassword123"
	newRole := domain.RoleAdmin
	newMunID := testhelper.MunPatosID
	lastLogin := time.Now()

	input := domain.UpdateUserInput{
		Username:       &newUsername,
		Email:          &newEmail,
		Password:       &newPassword,
		Role:           &newRole,
		MunicipalityID: &newMunID,
		LastLogin:      &lastLogin,
	}

	patos := testhelper.MakePatos()

	userRepo.On("FindByID", u.ID).Return(&u, nil)
	userRepo.On("ExistsByUsername", "updated_name", &u.ID).Return(false, nil)
	userRepo.On("ExistsByEmail", "updated@example.com", &u.ID).Return(false, nil)
	munRepo.On("FindByID", newMunID).Return(&patos, nil)
	userRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

	result, err := svc.Update(u.ID, input)

	require.NoError(t, err)
	assert.Equal(t, "updated_name", result.Username)
	assert.Equal(t, "updated@example.com", result.Email)
	assert.Equal(t, "newpassword123", result.Password)
	assert.Equal(t, domain.RoleAdmin, result.Role)
	assert.Equal(t, newMunID, result.MunicipalityID)
	assert.Equal(t, &lastLogin, result.LastLogin)
}

func TestUpdateUser_NotFound(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	newUsername := "new_name"
	input := domain.UpdateUserInput{Username: &newUsername}

	userRepo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	_, err := svc.Update(testhelper.NonExistentID, input)

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	userRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUser_UsernameConflict(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)
	newUsername := "taken_name"
	input := domain.UpdateUserInput{Username: &newUsername}

	userRepo.On("FindByID", u.ID).Return(&u, nil)
	userRepo.On("ExistsByUsername", "taken_name", &u.ID).Return(true, nil)

	_, err := svc.Update(u.ID, input)

	assert.ErrorIs(t, err, domain.ErrUsernameAlreadyExists)
	userRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)
	newEmail := "taken@example.com"
	input := domain.UpdateUserInput{Email: &newEmail}

	userRepo.On("FindByID", u.ID).Return(&u, nil)
	userRepo.On("ExistsByEmail", "taken@example.com", &u.ID).Return(true, nil)

	_, err := svc.Update(u.ID, input)

	assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	userRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUser_MunicipalityNotFound(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)
	newMunID := testhelper.NonExistentID
	input := domain.UpdateUserInput{MunicipalityID: &newMunID}

	userRepo.On("FindByID", u.ID).Return(&u, nil)
	munRepo.On("FindByID", newMunID).Return(nil, nil)

	_, err := svc.Update(u.ID, input)

	assert.ErrorIs(t, err, service.ErrMunicipalityNotFound)
	userRepo.AssertNotCalled(t, "Update")
}

// ═══════════════════════════════════════════════════════════════════════════════
// Delete
// ═══════════════════════════════════════════════════════════════════════════════

func TestDeleteUser_Success(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)

	userRepo.On("FindByID", u.ID).Return(&u, nil)
	userRepo.On("Delete", u.ID).Return(nil)

	err := svc.Delete(u.ID)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestDeleteUser_NotFound(t *testing.T) {
	svc, userRepo, _ := newUserService(t)

	userRepo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	err := svc.Delete(testhelper.NonExistentID)

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	userRepo.AssertNotCalled(t, "Delete")
}

// ═══════════════════════════════════════════════════════════════════════════════
// Restore
// ═══════════════════════════════════════════════════════════════════════════════

func TestRestoreUser_Success(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)
	u.DeletedAt.Valid = true

	userRepo.On("FindByIDUnscoped", u.ID).Return(&u, nil)
	userRepo.On("Restore", u.ID).Return(nil)

	result, err := svc.Restore(u.ID)

	require.NoError(t, err)
	assert.Equal(t, u.ID, result.ID)
	assert.False(t, result.DeletedAt.Valid)
	userRepo.AssertExpectations(t)
}

func TestRestoreUser_NotFound(t *testing.T) {
	svc, userRepo, _ := newUserService(t)

	userRepo.On("FindByIDUnscoped", testhelper.NonExistentID).Return(nil, nil)

	_, err := svc.Restore(testhelper.NonExistentID)

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	userRepo.AssertNotCalled(t, "Restore")
}

func TestRestoreUser_ActiveUser(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID) // not deleted

	userRepo.On("FindByIDUnscoped", u.ID).Return(&u, nil)

	_, err := svc.Restore(u.ID)

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	userRepo.AssertNotCalled(t, "Restore")
}

// ═══════════════════════════════════════════════════════════════════════════════
// HardDelete
// ═══════════════════════════════════════════════════════════════════════════════

func TestHardDeleteUser_Success(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)

	userRepo.On("FindByIDUnscoped", u.ID).Return(&u, nil)
	userRepo.On("HardDelete", u.ID).Return(nil)

	err := svc.HardDelete(u.ID)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestHardDeleteUser_NotFound(t *testing.T) {
	svc, userRepo, _ := newUserService(t)

	userRepo.On("FindByIDUnscoped", testhelper.NonExistentID).Return(nil, nil)

	err := svc.HardDelete(testhelper.NonExistentID)

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	userRepo.AssertNotCalled(t, "HardDelete")
}
