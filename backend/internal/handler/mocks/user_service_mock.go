package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// UserService é o mock da interface domain.UserService.
type UserService struct {
	mock.Mock
}

func (m *UserService) Create(input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserService) GetAll(page, pageSize int) ([]domain.User, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *UserService) GetDeleted(page, pageSize int) ([]domain.User, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *UserService) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserService) GetByIDUnscoped(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserService) Update(id uuid.UUID, input domain.UpdateUserInput) (*domain.User, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserService) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *UserService) Restore(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserService) HardDelete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *UserService) GetPermissions(userID uuid.UUID) (*domain.UserPermission, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserPermission), args.Error(1)
}

func (m *UserService) UpdatePermissions(userID uuid.UUID, input domain.UpdateUserPermissionInput) (*domain.UserPermission, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserPermission), args.Error(1)
}
