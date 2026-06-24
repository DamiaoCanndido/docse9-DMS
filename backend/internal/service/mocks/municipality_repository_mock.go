package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MunicipalityRepository é o mock da interface domain.MunicipalityRepository.
type MunicipalityRepository struct {
	mock.Mock
}

func (m *MunicipalityRepository) Create(municipality *domain.Municipality) error {
	args := m.Called(municipality)
	return args.Error(0)
}

func (m *MunicipalityRepository) FindAll(page, pageSize int) ([]domain.Municipality, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]domain.Municipality), args.Get(1).(int64), args.Error(2)
}

func (m *MunicipalityRepository) FindDeleted(page, pageSize int) ([]domain.Municipality, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]domain.Municipality), args.Get(1).(int64), args.Error(2)
}

func (m *MunicipalityRepository) FindByID(id uuid.UUID) (*domain.Municipality, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Municipality), args.Error(1)
}

func (m *MunicipalityRepository) FindByIDUnscoped(id uuid.UUID) (*domain.Municipality, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Municipality), args.Error(1)
}

func (m *MunicipalityRepository) FindByUF(uf string, page, pageSize int) ([]domain.Municipality, int64, error) {
	args := m.Called(uf, page, pageSize)
	return args.Get(0).([]domain.Municipality), args.Get(1).(int64), args.Error(2)
}

func (m *MunicipalityRepository) Update(municipality *domain.Municipality) error {
	args := m.Called(municipality)
	return args.Error(0)
}

func (m *MunicipalityRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MunicipalityRepository) Restore(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MunicipalityRepository) HardDelete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MunicipalityRepository) ExistsByName(name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(name, excludeID)
	return args.Bool(0), args.Error(1)
}
