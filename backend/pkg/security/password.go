package security

import "golang.org/x/crypto/bcrypt"

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
