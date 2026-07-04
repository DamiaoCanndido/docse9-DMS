package domain

import (
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("usuário ou senha incorretos")
)

type LoginInput struct {
	Username string `json:"username" binding:"required"` // Pode ser username ou email
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type AuthService interface {
	Login(input LoginInput) (*LoginResponse, error)
}
