package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// UserRepository é o mock da interface domain.UserRepository.
type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) Create(u *domain.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *UserRepository) FindAll(page, pageSize int) ([]domain.User, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *UserRepository) FindDeleted(page, pageSize int) ([]domain.User, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *UserRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) FindByIDUnscoped(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) FindByUsername(username string) (*domain.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) Update(u *domain.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *UserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *UserRepository) Restore(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *UserRepository) HardDelete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *UserRepository) ExistsByEmail(email string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(email, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *UserRepository) ExistsByUsername(username string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(username, excludeID)
	return args.Bool(0), args.Error(1)
}
