package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type AuthService struct {
	mock.Mock
}

func (m *AuthService) Login(input domain.LoginInput) (*domain.LoginResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoginResponse), args.Error(1)
}
