package service

import (
	"strings"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
)

type authService struct {
	userRepo domain.UserRepository
}

func NewAuthService(userRepo domain.UserRepository) domain.AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Login(input domain.LoginInput) (*domain.LoginResponse, error) {
	identity := strings.TrimSpace(input.Username)

	// 1. Procurar usuário por username ou email
	var u *domain.User
	var err error

	if strings.Contains(identity, "@") {
		u, err = s.userRepo.FindByEmail(strings.ToLower(identity))
	} else {
		u, err = s.userRepo.FindByUsername(identity)
	}

	if err != nil {
		return nil, err
	}
	if u == nil {
		if strings.Contains(identity, "@") {
			u, err = s.userRepo.FindByUsername(identity)
		} else {
			u, err = s.userRepo.FindByEmail(strings.ToLower(identity))
		}
		if err != nil {
			return nil, err
		}
	}

	if u == nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 2. Verificar senha
	if !security.CheckPasswordHash(input.Password, u.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	// 3. Atualizar LastLogin
	now := time.Now()
	u.LastLogin = &now
	if err := s.userRepo.Update(u); err != nil {
		return nil, err
	}

	// 4. Gerar Token JWT (duração de 24 horas por padrão)
	token, err := security.GenerateToken(u.ID, u.Username, string(u.Role), u.MunicipalityID, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		Token: token,
		User:  *u,
	}, nil
}
