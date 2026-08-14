package security_test

import (
	"os"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateToken_Success(t *testing.T) {
	userID := uuid.New()
	munID := uuid.New()

	token, err := security.GenerateToken(userID, "testuser", "ADMIN", munID, false, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := security.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "ADMIN", claims.Role)
	assert.Equal(t, munID, claims.MunicipalityID)
	assert.False(t, claims.MustChangePassword)
}

func TestValidateToken_Expired(t *testing.T) {
	userID := uuid.New()
	munID := uuid.New()

	// Token expired 1 hour ago
	token, err := security.GenerateToken(userID, "testuser", "ADMIN", munID, false, -1*time.Hour)
	require.NoError(t, err)

	_, err = security.ValidateToken(token)
	assert.ErrorIs(t, err, security.ErrInvalidToken)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	userID := uuid.New()
	munID := uuid.New()

	// Generate with secret 1
	os.Setenv("JWT_SECRET", "secret-one-12345678901234567890")
	token, err := security.GenerateToken(userID, "testuser", "ADMIN", munID, false, time.Hour)
	require.NoError(t, err)

	// Validate with secret 2
	os.Setenv("JWT_SECRET", "secret-two-12345678901234567890")
	_, err = security.ValidateToken(token)
	assert.ErrorIs(t, err, security.ErrInvalidToken)

	os.Unsetenv("JWT_SECRET")
}

func TestValidateJWTConfig(t *testing.T) {
	t.Run("development allows default secret", func(t *testing.T) {
		os.Setenv("APP_ENV", "development")
		os.Unsetenv("JWT_SECRET")
		assert.NoError(t, security.ValidateJWTConfig())
	})

	t.Run("production fails with empty secret", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Unsetenv("JWT_SECRET")
		assert.Error(t, security.ValidateJWTConfig())
	})

	t.Run("production fails with default factory secret", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Setenv("JWT_SECRET", "docseq-secret-key-change-in-production")
		assert.Error(t, security.ValidateJWTConfig())
	})

	t.Run("production succeeds with strong custom secret", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Setenv("JWT_SECRET", "super-secure-production-jwt-key-999")
		assert.NoError(t, security.ValidateJWTConfig())
	})

	os.Unsetenv("APP_ENV")
	os.Unsetenv("JWT_SECRET")
}
