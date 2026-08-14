package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// DocumentService é o mock da interface domain.DocumentService.
type DocumentService struct {
	mock.Mock
}

func (m *DocumentService) Create(input domain.CreateDocumentInput) (*domain.Document, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *DocumentService) GetAll(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	args := m.Called(filter, page, pageSize)
	var docs []domain.Document
	if args.Get(0) != nil {
		docs = args.Get(0).([]domain.Document)
	}
	return docs, args.Get(1).(int64), args.Error(2)
}

func (m *DocumentService) GetDeleted(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	args := m.Called(filter, page, pageSize)
	var docs []domain.Document
	if args.Get(0) != nil {
		docs = args.Get(0).([]domain.Document)
	}
	return docs, args.Get(1).(int64), args.Error(2)
}

func (m *DocumentService) GetByID(id uuid.UUID) (*domain.Document, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *DocumentService) GetByIDUnscoped(id uuid.UUID) (*domain.Document, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *DocumentService) Update(id uuid.UUID, input domain.UpdateDocumentInput) (*domain.Document, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *DocumentService) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *DocumentService) Restore(id uuid.UUID) (*domain.Document, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *DocumentService) HardDelete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
