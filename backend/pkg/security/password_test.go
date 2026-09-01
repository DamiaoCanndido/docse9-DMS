package security_test

import (
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_And_CheckPasswordHash(t *testing.T) {
	password := "MinhaSenhaForte@2026"

	// Gerar Hash
	hash, err := security.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Verificar senha correta
	assert.True(t, security.CheckPasswordHash(password, hash))

	// Verificar senha incorreta
	assert.False(t, security.CheckPasswordHash("SenhaErrada@2026", hash))
	assert.False(t, security.CheckPasswordHash("", hash))

	// Verificar hash inválido/malformado
	assert.False(t, security.CheckPasswordHash(password, "invalid_hash_string"))
}

func TestGenerateRandomPassword(t *testing.T) {
	t.Run("generate password with standard length 16", func(t *testing.T) {
		pwd, err := security.GenerateRandomPassword(16)
		require.NoError(t, err)
		assert.Len(t, pwd, 16)

		// Deve conter pelo menos um caractere de cada conjunto
		hasLower := false
		hasUpper := false
		hasDigit := false
		hasSpecial := false

		for _, r := range pwd {
			switch {
			case r >= 'a' && r <= 'z':
				hasLower = true
			case r >= 'A' && r <= 'Z':
				hasUpper = true
			case r >= '0' && r <= '9':
				hasDigit = true
			default:
				hasSpecial = true
			}
		}

		assert.True(t, hasLower, "deve conter letra minúscula")
		assert.True(t, hasUpper, "deve conter letra maiúscula")
		assert.True(t, hasDigit, "deve conter dígito")
		assert.True(t, hasSpecial, "deve conter caractere especial")
	})

	t.Run("generate password with short length 8", func(t *testing.T) {
		pwd, err := security.GenerateRandomPassword(8)
		require.NoError(t, err)
		assert.Len(t, pwd, 8)
	})

	t.Run("multiple generated passwords are unique", func(t *testing.T) {
		pwd1, err1 := security.GenerateRandomPassword(12)
		pwd2, err2 := security.GenerateRandomPassword(12)
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, pwd1, pwd2)
	})
}
