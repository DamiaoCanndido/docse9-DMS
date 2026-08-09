package mocks

import (
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// DocumentRepository é o mock da interface domain.DocumentRepository.
type DocumentRepository struct {
	mock.Mock
}

func (m *DocumentRepository) Create(d *domain.Document) error {
	args := m.Called(d)
	return args.Error(0)
}

func (m *DocumentRepository) FindAll(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	args := m.Called(filter, page, pageSize)
	var docs []domain.Document
	if args.Get(0) != nil {
		docs = args.Get(0).([]domain.Document)
	}
	return docs, args.Get(1).(int64), args.Error(2)
}

func (m *DocumentRepository) FindDeleted(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	args := m.Called(filter, page, pageSize)
	var docs []domain.Document
	if args.Get(0) != nil {
		docs = args.Get(0).([]domain.Document)
	}
	return docs, args.Get(1).(int64), args.Error(2)
}

func (m *DocumentRepository) FindByID(id uuid.UUID) (*domain.Document, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *DocumentRepository) FindByIDUnscoped(id uuid.UUID) (*domain.Document, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Document), args.Error(1)
}

func (m *DocumentRepository) Update(d *domain.Document) error {
	args := m.Called(d)
	return args.Error(0)
}

func (m *DocumentRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *DocumentRepository) Restore(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *DocumentRepository) HardDelete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *DocumentRepository) GetLastOrder(municipalityID uuid.UUID, docType domain.DocumentType, contractType *domain.ContractType, year *int) (int, error) {
	args := m.Called(municipalityID, docType, contractType, year)
	return args.Int(0), args.Error(1)
}
