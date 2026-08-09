package security

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword gera o hash bcrypt de uma senha.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compara uma senha em texto puro com o seu hash bcrypt.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomPassword gera uma senha aleatória forte contendo letras maiúsculas, minúsculas, números e caracteres especiais.
func GenerateRandomPassword(length int) (string, error) {
	const (
		lower   = "abcdefghijklmnopqrstuvwxyz"
		upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits  = "0123456789"
		special = "!@#$%&*+?_-"
	)
	all := lower + upper + digits + special

	var password []byte
	// Garantir pelo menos um caractere de cada tipo
	for _, charset := range []string{lower, upper, digits, special} {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password = append(password, charset[n.Int64()])
	}

	// Preencher o restante
	for i := len(password); i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(all))))
		if err != nil {
			return "", err
		}
		password = append(password, all[n.Int64()])
	}

	// Embaralhar (Fisher-Yates)
	for i := len(password) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := n.Int64()
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}

