package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type UserPermissionRepository struct {
	mock.Mock
}

func (m *UserPermissionRepository) FindByUserID(userID uuid.UUID) (*domain.UserPermission, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserPermission), args.Error(1)
}

func (m *UserPermissionRepository) Create(p *domain.UserPermission) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *UserPermissionRepository) Update(p *domain.UserPermission) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *UserPermissionRepository) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}
