package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// SequenceService é o mock da interface domain.SequenceService.
type SequenceService struct {
	mock.Mock
}

func (m *SequenceService) GetSequences(municipalityID uuid.UUID, year *int) ([]domain.SequenceItemResponse, error) {
	args := m.Called(municipalityID, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SequenceItemResponse), args.Error(1)
}

func (m *SequenceService) SetOffset(municipalityID uuid.UUID, input domain.SetSequenceOffsetInput) (*domain.SequenceOffset, error) {
	args := m.Called(municipalityID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SequenceOffset), args.Error(1)
}
