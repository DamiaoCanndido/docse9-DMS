package service_test

import (
	"strings"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newUserService(t *testing.T) (domain.UserService, *mocks.UserRepository, *mocks.MunicipalityRepository) {
	t.Helper()
	userRepo := new(mocks.UserRepository)
	munRepo := new(mocks.MunicipalityRepository)
	permRepo := new(mocks.UserPermissionRepository)
	permRepo.On("DeleteByUserID", mock.Anything).Return(nil).Maybe()
	svc := service.NewUserService(userRepo, munRepo, permRepo)
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
		Role:            domain.RoleCommon,
		MunicipalityID:  mun.ID,
	}

	userRepo.On("ExistsByEmail", "new@example.com", (*uuid.UUID)(nil)).Return(false, nil)
	userRepo.On("ExistsByUsername", "new_user", (*uuid.UUID)(nil)).Return(false, nil)
	munRepo.On("FindByID", mun.ID).Return(&mun, nil)
	userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
	createdUserID := uuid.New()
	userRepo.On("FindByID", mock.Anything).Return(&domain.User{
		ID:             createdUserID,
		Username:       "new_user",
		Email:          "new@example.com",
		Role:           domain.RoleCommon,
		MunicipalityID: mun.ID,
		Municipality:   mun,
	}, nil)

	result, rawPassword, err := svc.Create(input)

	require.NoError(t, err)
	assert.NotEmpty(t, rawPassword)
	assert.Equal(t, "new_user", result.Username)
	assert.Equal(t, "new@example.com", result.Email)
	assert.Equal(t, domain.RoleCommon, result.Role)
	assert.Equal(t, mun.ID, result.MunicipalityID)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, mun.Name, result.Municipality.Name) // verify municipality is populated
	userRepo.AssertExpectations(t)
	munRepo.AssertExpectations(t)
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	mun := testhelper.MakePassagem()
	input := domain.CreateUserInput{
		Username:       "new_user",
		Email:          "existing@example.com",
		Role:           domain.RoleCommon,
		MunicipalityID: mun.ID,
	}

	userRepo.On("ExistsByEmail", "existing@example.com", (*uuid.UUID)(nil)).Return(true, nil)

	_, _, err := svc.Create(input)

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
		Role:           domain.RoleCommon,
		MunicipalityID: mun.ID,
	}

	userRepo.On("ExistsByEmail", "new@example.com", (*uuid.UUID)(nil)).Return(false, nil)
	userRepo.On("ExistsByUsername", "existing_user", (*uuid.UUID)(nil)).Return(true, nil)

	_, _, err := svc.Create(input)

	assert.ErrorIs(t, err, domain.ErrUsernameAlreadyExists)
	munRepo.AssertNotCalled(t, "FindByID")
	userRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_MunicipalityNotFound(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	input := domain.CreateUserInput{
		Username:       "new_user",
		Email:          "new@example.com",
		Role:           domain.RoleCommon,
		MunicipalityID: testhelper.NonExistentID,
	}

	userRepo.On("ExistsByEmail", "new@example.com", (*uuid.UUID)(nil)).Return(false, nil)
	userRepo.On("ExistsByUsername", "new_user", (*uuid.UUID)(nil)).Return(false, nil)
	munRepo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	_, _, err := svc.Create(input)

	assert.ErrorIs(t, err, service.ErrMunicipalityNotFound)
	userRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_InvalidUsername(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()

	cases := []struct {
		name     string
		username string
	}{
		{"spaces in middle", "hello world"},
		{"uppercase letters", "Hello"},
		{"special char @", "user@name"},
		{"special char !", "user!name"},
		{"space only", " "},
		{"leading space trimmed still uppercase", " Hello "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := domain.CreateUserInput{
				Username:       tc.username,
				Email:          "new@example.com",
				Role:           domain.RoleCommon,
				MunicipalityID: mun.ID,
			}
			_, _, err := svc.Create(input)
			assert.ErrorIs(t, err, domain.ErrInvalidUsername)
			userRepo.AssertNotCalled(t, "ExistsByEmail")
			userRepo.AssertNotCalled(t, "Create")
		})
	}
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()

	cases := []struct {
		name  string
		email string
	}{
		{"missing at and domain", "invalid-email"},
		{"missing domain", "user@"},
		{"missing local part", "@example.com"},
		{"missing top level domain", "user@domain"},
		{"spaces in email", "user name@example.com"},
		{"empty email", ""},
		{"invalid domain prefix dot", "user@.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := domain.CreateUserInput{
				Username:       "valid_user",
				Email:          tc.email,
				Role:           domain.RoleCommon,
				MunicipalityID: mun.ID,
			}
			_, _, err := svc.Create(input)
			assert.ErrorIs(t, err, domain.ErrInvalidEmail)
			userRepo.AssertNotCalled(t, "ExistsByEmail")
			userRepo.AssertNotCalled(t, "Create")
		})
	}
}

func TestCreateUser_ValidEmailFormats(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	mun := testhelper.MakePassagem()

	cases := []string{
		"simple@example.com",
		"user.name+tag@sub.domain.org",
		"test_123-abc@domain.com",
	}

	for _, email := range cases {
		t.Run(email, func(t *testing.T) {
			input := domain.CreateUserInput{
				Username:       "valid_user",
				Email:          email,
				Role:           domain.RoleCommon,
				MunicipalityID: mun.ID,
			}

			userRepo.On("ExistsByEmail", strings.ToLower(email), (*uuid.UUID)(nil)).Return(false, nil).Once()
			userRepo.On("ExistsByUsername", "valid_user", (*uuid.UUID)(nil)).Return(false, nil).Once()
			munRepo.On("FindByID", mun.ID).Return(&mun, nil).Once()
			userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil).Once()
			userRepo.On("FindByID", mock.Anything).Return(&domain.User{
				ID:             uuid.New(),
				Username:       "valid_user",
				Email:          strings.ToLower(email),
				Role:           domain.RoleCommon,
				MunicipalityID: mun.ID,
			}, nil).Once()

			result, rawPassword, err := svc.Create(input)

			require.NoError(t, err)
			assert.NotEmpty(t, rawPassword)
			assert.Equal(t, strings.ToLower(email), result.Email)
		})
	}
}

func TestCreateUser_ValidUsernameFormats(t *testing.T) {
	svc, userRepo, munRepo := newUserService(t)
	mun := testhelper.MakePassagem()

	cases := []string{
		"valid.user-name_1",
		"simple",
		"user123",
		"a-b-c",
		"test_user",
		"u.s.e.r",
	}

	for _, username := range cases {
		t.Run(username, func(t *testing.T) {
			input := domain.CreateUserInput{
				Username:       username,
				Email:          username + "@example.com",
				Role:           domain.RoleCommon,
				MunicipalityID: mun.ID,
			}

			userRepo.On("ExistsByEmail", username+"@example.com", (*uuid.UUID)(nil)).Return(false, nil).Once()
			userRepo.On("ExistsByUsername", username, (*uuid.UUID)(nil)).Return(false, nil).Once()
			munRepo.On("FindByID", mun.ID).Return(&mun, nil).Once()
			userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil).Once()
			userRepo.On("FindByID", mock.Anything).Return(&domain.User{
				ID:             uuid.New(),
				Username:       username,
				Email:          username + "@example.com",
				Role:           domain.RoleCommon,
				MunicipalityID: mun.ID,
			}, nil).Once()

			result, rawPassword, err := svc.Create(input)

			require.NoError(t, err)
			assert.NotEmpty(t, rawPassword)
			assert.Equal(t, username, result.Username)
		})
	}
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
	filter := domain.UserFilter{}

	userRepo.On("FindAll", filter, 1, 10).Return(users, int64(2), nil)

	result, total, err := svc.GetAll(filter, 1, 10)

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
	filter := domain.UserFilter{}

	userRepo.On("FindDeleted", filter, 1, 10).Return(users, int64(1), nil)

	result, total, err := svc.GetDeleted(filter, 1, 10)

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

	userRepo.On("FindByID", u.ID).Return(&u, nil).Once()
	userRepo.On("ExistsByUsername", "updated_name", &u.ID).Return(false, nil)
	userRepo.On("ExistsByEmail", "updated@example.com", &u.ID).Return(false, nil)
	munRepo.On("FindByID", newMunID).Return(&patos, nil)
	userRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)
	userRepo.On("FindByID", u.ID).Return(&domain.User{
		ID:             u.ID,
		Username:       "updated_name",
		Email:          "updated@example.com",
		Role:           domain.RoleAdmin,
		MunicipalityID: newMunID,
		Municipality:   patos,
		LastLogin:      &lastLogin,
	}, nil).Once()

	result, rawPassword, err := svc.Update(u.ID, input)

	require.NoError(t, err)
	assert.Empty(t, rawPassword)
	assert.Equal(t, "updated_name", result.Username)
	assert.Equal(t, "updated@example.com", result.Email)
	assert.Equal(t, domain.RoleAdmin, result.Role)
	assert.Equal(t, newMunID, result.MunicipalityID)
	assert.Equal(t, &lastLogin, result.LastLogin)
	assert.Equal(t, patos.Name, result.Municipality.Name) // verify municipality was updated in reload
}

func TestUpdateUser_NotFound(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	newUsername := "new_name"
	input := domain.UpdateUserInput{Username: &newUsername}

	userRepo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	_, _, err := svc.Update(testhelper.NonExistentID, input)

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

	_, _, err := svc.Update(u.ID, input)

	assert.ErrorIs(t, err, domain.ErrUsernameAlreadyExists)
	userRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUser_InvalidUsername(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)

	cases := []struct {
		name     string
		username string
	}{
		{"spaces", "has space"},
		{"uppercase", "HasUpper"},
		{"special char", "user@name"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			newUsername := tc.username
			input := domain.UpdateUserInput{Username: &newUsername}

			userRepo.On("FindByID", u.ID).Return(&u, nil).Once()

			_, _, err := svc.Update(u.ID, input)

			assert.ErrorIs(t, err, domain.ErrInvalidUsername)
			userRepo.AssertNotCalled(t, "Update")
		})
	}
}

func TestUpdateUser_InvalidEmail(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)

	cases := []struct {
		name  string
		email string
	}{
		{"missing domain", "invalid-email"},
		{"spaces in email", "user name@example.com"},
		{"missing top level domain", "user@domain"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			newEmail := tc.email
			input := domain.UpdateUserInput{Email: &newEmail}

			userRepo.On("FindByID", u.ID).Return(&u, nil).Once()

			_, _, err := svc.Update(u.ID, input)

			assert.ErrorIs(t, err, domain.ErrInvalidEmail)
			userRepo.AssertNotCalled(t, "Update")
		})
	}
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)
	newEmail := "taken@example.com"
	input := domain.UpdateUserInput{Email: &newEmail}

	userRepo.On("FindByID", u.ID).Return(&u, nil)
	userRepo.On("ExistsByEmail", "taken@example.com", &u.ID).Return(true, nil)

	_, _, err := svc.Update(u.ID, input)

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

	_, _, err := svc.Update(u.ID, input)

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

func TestGetPermissions_Success(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	munRepo := new(mocks.MunicipalityRepository)
	permRepo := new(mocks.UserPermissionRepository)
	svc := service.NewUserService(userRepo, munRepo, permRepo)

	userID := uuid.New()
	u := &domain.User{ID: userID, Role: domain.RoleCommon}
	p := []domain.UserPermission{
		{UserID: userID, DocumentType: domain.TypeNotice, Level: domain.LevelRead},
	}

	userRepo.On("FindByID", userID).Return(u, nil)
	permRepo.On("FindByUserID", userID).Return(p, nil)

	res, err := svc.GetPermissions(userID)
	require.NoError(t, err)
	assert.Len(t, res, 5)

	var noticePerm *domain.UserPermission
	for i := range res {
		if res[i].DocumentType == domain.TypeNotice {
			noticePerm = &res[i]
		}
	}
	require.NotNil(t, noticePerm)
	assert.Equal(t, domain.LevelRead, noticePerm.Level)
}

func TestUpdatePermissions_Success(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	munRepo := new(mocks.MunicipalityRepository)
	permRepo := new(mocks.UserPermissionRepository)
	svc := service.NewUserService(userRepo, munRepo, permRepo)

	userID := uuid.New()
	u := &domain.User{ID: userID, Role: domain.RoleCommon}

	userRepo.On("FindByID", userID).Return(u, nil)
	permRepo.On("DeleteByUserID", userID).Return(nil)
	permRepo.On("Create", mock.AnythingOfType("*domain.UserPermission")).Return(nil)

	pAfter := []domain.UserPermission{
		{UserID: userID, DocumentType: domain.TypeNotice, Level: domain.LevelWrite},
	}
	permRepo.On("FindByUserID", userID).Return(pAfter, nil)

	input := domain.UpdateUserPermissionsInput{
		Permissions: []domain.UpdateUserPermissionItem{
			{DocumentType: domain.TypeNotice, Level: domain.LevelWrite},
		},
	}

	res, err := svc.UpdatePermissions(userID, input)
	require.NoError(t, err)
	assert.Len(t, res, 5)

	var noticePerm *domain.UserPermission
	for i := range res {
		if res[i].DocumentType == domain.TypeNotice {
			noticePerm = &res[i]
		}
	}
	require.NotNil(t, noticePerm)
	assert.Equal(t, domain.LevelWrite, noticePerm.Level)
}

func TestChangePassword_Success(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)
	// Setup user with a hashed password for "secret123"
	hashed, _ := security.HashPassword("secret123")
	u.Password = hashed

	input := domain.ChangePasswordInput{
		CurrentPassword: "secret123",
		NewPassword:     "newsecret456",
		ConfirmPassword: "newsecret456",
	}

	userRepo.On("FindByID", u.ID).Return(&u, nil)
	userRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

	_, err := svc.ChangePassword(u.ID, input)
	assert.NoError(t, err)
}

func TestChangePassword_IncorrectCurrentPassword(t *testing.T) {
	svc, userRepo, _ := newUserService(t)
	mun := testhelper.MakePassagem()
	u := testhelper.MakeUserCommon(mun.ID)
	hashed, _ := security.HashPassword("secret123")
	u.Password = hashed

	input := domain.ChangePasswordInput{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newsecret456",
		ConfirmPassword: "newsecret456",
	}

	userRepo.On("FindByID", u.ID).Return(&u, nil)

	_, err := svc.ChangePassword(u.ID, input)
	assert.ErrorIs(t, err, domain.ErrIncorrectCurrentPassword)
	userRepo.AssertNotCalled(t, "Update")
}
