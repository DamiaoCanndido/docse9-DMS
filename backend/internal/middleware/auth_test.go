package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/middleware"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	r := gin.New()

	// Rota pública
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "public"})
	})

	// Rota protegida
	protected := r.Group("/protected")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("", func(c *gin.Context) {
			user, _ := c.Get("user")
			claims := user.(*security.UserClaims)
			c.JSON(http.StatusOK, gin.H{"username": claims.Username})
		})

		// Rota admin-only
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole(domain.RoleAdmin))
		admin.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "admin"})
		})
	}

	return r
}

func executeRequest(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	r := setupTestRouter()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)

	w := executeRequest(r, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	r := setupTestRouter()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token")

	w := executeRequest(r, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r := setupTestRouter()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-str")

	w := executeRequest(r, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_Success(t *testing.T) {
	r := setupTestRouter()
	userID := uuid.New()
	munID := uuid.New()
	token, err := security.GenerateToken(userID, "testuser", "COMMON", munID, time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := executeRequest(r, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"username":"testuser"`)
}

func TestRequireRole_AdminAccessingAdminRoute(t *testing.T) {
	r := setupTestRouter()
	userID := uuid.New()
	munID := uuid.New()
	token, err := security.GenerateToken(userID, "adminuser", "ADMIN", munID, time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := executeRequest(r, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_CommonUserAccessingAdminRoute(t *testing.T) {
	r := setupTestRouter()
	userID := uuid.New()
	munID := uuid.New()
	token, err := security.GenerateToken(userID, "commonuser", "COMMON", munID, time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := executeRequest(r, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
