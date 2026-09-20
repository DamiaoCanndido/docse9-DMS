package service_test

import (
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogin_Success_ByUsername(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	svc := service.NewAuthService(userRepo)

	hashedPassword, _ := security.HashPassword("secret123")
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	u.Password = hashedPassword

	input := domain.LoginInput{
		Username: u.Username,
		Password: "secret123",
	}

	userRepo.On("FindByUsername", u.Username).Return(&u, nil)
	userRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

	resp, err := svc.Login(input)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, u.Username, resp.User.Username)
	assert.NotNil(t, resp.User.LastLogin)

	// Validar que o token expira em 1 semana (~7 dias)
	claims, err := security.ValidateToken(resp.Token)
	require.NoError(t, err)
	diff := time.Until(claims.ExpiresAt.Time)
	assert.True(t, diff > 6*24*time.Hour && diff <= 7*24*time.Hour)

	userRepo.AssertExpectations(t)
}

func TestLogin_Success_ByEmail(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	svc := service.NewAuthService(userRepo)

	hashedPassword, _ := security.HashPassword("secret123")
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	u.Password = hashedPassword

	input := domain.LoginInput{
		Username: u.Email,
		Password: "secret123",
	}

	userRepo.On("FindByEmail", u.Email).Return(&u, nil)
	userRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

	resp, err := svc.Login(input)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, u.Email, resp.User.Email)
	userRepo.AssertExpectations(t)
}

func TestLogin_Fail_InvalidUser(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	svc := service.NewAuthService(userRepo)

	input := domain.LoginInput{
		Username: "nonexistent",
		Password: "password",
	}

	userRepo.On("FindByUsername", "nonexistent").Return(nil, nil)
	userRepo.On("FindByEmail", "nonexistent").Return(nil, nil)

	_, err := svc.Login(input)

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_Fail_InvalidPassword(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	svc := service.NewAuthService(userRepo)

	hashedPassword, _ := security.HashPassword("secret123")
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	u.Password = hashedPassword

	input := domain.LoginInput{
		Username: u.Username,
		Password: "wrongpassword",
	}

	userRepo.On("FindByUsername", u.Username).Return(&u, nil)

	_, err := svc.Login(input)

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_Success_FallbackLookup(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	svc := service.NewAuthService(userRepo)

	hashedPassword, _ := security.HashPassword("secret123")
	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	u.Password = hashedPassword

	t.Run("Identity with @ falls back to Username", func(t *testing.T) {
		input := domain.LoginInput{Username: "user@domain", Password: "secret123"}
		userRepo.On("FindByEmail", "user@domain").Return(nil, nil).Once()
		userRepo.On("FindByUsername", "user@domain").Return(&u, nil).Once()
		userRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil).Once()

		resp, err := svc.Login(input)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
	})
}

func TestLogin_Fail_RepoErrors(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	svc := service.NewAuthService(userRepo)

	input := domain.LoginInput{Username: "user", Password: "password"}
	userRepo.On("FindByUsername", "user").Return(nil, assert.AnError).Once()

	_, err := svc.Login(input)
	assert.ErrorIs(t, err, assert.AnError)
}

