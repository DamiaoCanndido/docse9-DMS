package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// SequenceRepository é o mock da interface domain.SequenceRepository.
type SequenceRepository struct {
	mock.Mock
}

func (m *SequenceRepository) GetByMunicipality(municipalityID uuid.UUID, year *int) ([]domain.SequenceOffset, error) {
	args := m.Called(municipalityID, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SequenceOffset), args.Error(1)
}

func (m *SequenceRepository) FindOffset(municipalityID uuid.UUID, docType domain.DocumentType, contractType *domain.ContractType, year *int) (*domain.SequenceOffset, error) {
	args := m.Called(municipalityID, docType, contractType, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SequenceOffset), args.Error(1)
}

func (m *SequenceRepository) Upsert(offset *domain.SequenceOffset) error {
	args := m.Called(offset)
	return args.Error(0)
}
