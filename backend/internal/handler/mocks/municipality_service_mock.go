package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MunicipalityService é o mock da interface domain.MunicipalityService.
type MunicipalityService struct {
	mock.Mock
}

func (m *MunicipalityService) Create(input domain.CreateMunicipalityInput) (*domain.Municipality, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Municipality), args.Error(1)
}

func (m *MunicipalityService) GetAll(page, pageSize int) ([]domain.Municipality, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]domain.Municipality), args.Get(1).(int64), args.Error(2)
}

func (m *MunicipalityService) GetByID(id uuid.UUID) (*domain.Municipality, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Municipality), args.Error(1)
}

func (m *MunicipalityService) GetByUF(uf string, page, pageSize int) ([]domain.Municipality, int64, error) {
	args := m.Called(uf, page, pageSize)
	return args.Get(0).([]domain.Municipality), args.Get(1).(int64), args.Error(2)
}

func (m *MunicipalityService) Update(id uuid.UUID, input domain.UpdateMunicipalityInput) (*domain.Municipality, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Municipality), args.Error(1)
}

func (m *MunicipalityService) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MunicipalityService) HardDelete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
