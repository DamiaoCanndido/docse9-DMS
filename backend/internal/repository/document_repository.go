package repository

import (
	"errors"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository cria uma nova instância do repositório de documentos.
func NewDocumentRepository(db *gorm.DB) domain.DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(d *domain.Document) error {
	return r.db.Create(d).Error
}

func (r *documentRepository) FindAll(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	var (
		documents []domain.Document
		total     int64
	)

	offset := (page - 1) * pageSize
	query := r.db.Model(&domain.Document{})

	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("CreatedBy").Preload("Municipality").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (r *documentRepository) FindDeleted(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	var (
		documents []domain.Document
		total     int64
	)

	offset := (page - 1) * pageSize
	query := r.db.Unscoped().
		Model(&domain.Document{}).
		Where("deleted_at IS NOT NULL")

	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("CreatedBy").Preload("Municipality").
		Order("deleted_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (r *documentRepository) FindByID(id uuid.UUID) (*domain.Document, error) {
	var d domain.Document
	err := r.db.Preload("CreatedBy").Preload("Municipality").First(&d, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func (r *documentRepository) FindByIDUnscoped(id uuid.UUID) (*domain.Document, error) {
	var d domain.Document
	err := r.db.Unscoped().Preload("CreatedBy").Preload("Municipality").First(&d, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func (r *documentRepository) Update(d *domain.Document) error {
	return r.db.Save(d).Error
}

func (r *documentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Document{}, "id = ?", id).Error
}

func (r *documentRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().
		Model(&domain.Document{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *documentRepository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&domain.Document{}, "id = ?", id).Error
}

func (r *documentRepository) GetLastOrder(municipalityID uuid.UUID, docType domain.DocumentType, contractType *domain.ContractType, year *int) (int, error) {
	var lastOrder int
	query := r.db.Unscoped().Model(&domain.Document{}).
		Select("COALESCE(MAX(documents.order), 0)").
		Where("municipality_id = ? AND type = ?", municipalityID, docType)

	if docType == domain.TypeContract && contractType != nil {
		query = query.Where("contract_type = ?", *contractType)
	}

	if year != nil {
		query = query.Where("EXTRACT(YEAR FROM created_at) = ?", *year)
	}

	err := query.Row().Scan(&lastOrder)
	return lastOrder, err
}

// applyFilters aplica filtros de busca à query.
func applyFilters(query *gorm.DB, filter domain.DocumentFilter) *gorm.DB {
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if len(filter.AllowedTypes) > 0 {
		query = query.Where("type IN ?", filter.AllowedTypes)
	}
	if filter.MunicipalityID != nil {
		query = query.Where("municipality_id = ?", *filter.MunicipalityID)
	}
	if filter.CreatorID != nil {
		query = query.Where("creator_id = ?", *filter.CreatorID)
	}
	if filter.ContractType != nil {
		query = query.Where("contract_type = ?", *filter.ContractType)
	}
	if filter.Year != nil {
		query = query.Where("EXTRACT(YEAR FROM created_at) = ?", *filter.Year)
	}
	if filter.Search != "" {
		query = query.Where("LOWER(description) LIKE LOWER(?)", "%"+filter.Search+"%")
	}
	return query
}
