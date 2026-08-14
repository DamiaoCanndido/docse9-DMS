package security

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token inválido ou expirado")
	defaultSecret   = "docseq-secret-key-change-in-production"
)

// getJWTSecret obtém a chave secreta de assinatura JWT em tempo de execução.
func getJWTSecret() []byte {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}
	return []byte(defaultSecret)
}

// ValidateJWTConfig valida se as configurações de JWT são seguras para o ambiente atual.
func ValidateJWTConfig() error {
	secret := os.Getenv("JWT_SECRET")
	if os.Getenv("APP_ENV") == "production" {
		if secret == "" || secret == defaultSecret {
			return errors.New("JWT_SECRET não definida ou utilizando valor padrão de fábrica em ambiente de produção — defina uma chave forte")
		}
	}
	return nil
}

// UserClaims representa as claims customizadas do token JWT.
type UserClaims struct {
	UserID             uuid.UUID `json:"user_id"`
	Username           string    `json:"username"`
	Role               string    `json:"role"`
	MunicipalityID     uuid.UUID `json:"municipality_id"`
	MustChangePassword bool      `json:"must_change_password"`
	jwt.RegisteredClaims
}

// GenerateToken gera um token JWT assinado para um determinado usuário.
func GenerateToken(userID uuid.UUID, username, role string, municipalityID uuid.UUID, mustChangePassword bool, duration time.Duration) (string, error) {
	claims := UserClaims{
		UserID:             userID,
		Username:           username,
		Role:               role,
		MunicipalityID:     municipalityID,
		MustChangePassword: mustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ValidateToken valida um token JWT e retorna suas claims se for válido.
func ValidateToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
