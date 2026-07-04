package handler_test

import (
	"net/http"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler"
	handlerMocks "github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthRouter(svc domain.AuthService) *gin.Engine {
	r := gin.New()
	h := handler.NewAuthHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestLogin_Handler_200(t *testing.T) {
	svc := new(handlerMocks.AuthService)
	input := domain.LoginInput{
		Username: "user",
		Password: "password",
	}

	u := testhelper.MakeUserCommon(testhelper.MunPassagemID)
	resp := domain.LoginResponse{
		Token: "jwt_token_example",
		User:  u,
	}

	svc.On("Login", input).Return(&resp, nil)

	w := doRequest(setupAuthRouter(svc), http.MethodPost, "/api/v1/auth/login", input)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	parseBody(t, w, &result)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "jwt_token_example", result["data"].(map[string]any)["token"])
	svc.AssertExpectations(t)
}

func TestLogin_Handler_400_Validation(t *testing.T) {
	svc := new(handlerMocks.AuthService)
	// input sem password
	input := map[string]string{
		"username": "user",
	}

	w := doRequest(setupAuthRouter(svc), http.MethodPost, "/api/v1/auth/login", input)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "Login")
}

func TestLogin_Handler_401_InvalidCredentials(t *testing.T) {
	svc := new(handlerMocks.AuthService)
	input := domain.LoginInput{
		Username: "user",
		Password: "wrongpassword",
	}

	svc.On("Login", input).Return(nil, domain.ErrInvalidCredentials)

	w := doRequest(setupAuthRouter(svc), http.MethodPost, "/api/v1/auth/login", input)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
