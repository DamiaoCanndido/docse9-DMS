package repository

import (
	"errors"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sequenceRepository struct {
	db *gorm.DB
}

// NewSequenceRepository cria uma nova instância do repositório de sequências.
func NewSequenceRepository(db *gorm.DB) domain.SequenceRepository {
	return &sequenceRepository{db: db}
}

func (r *sequenceRepository) GetByMunicipality(municipalityID uuid.UUID, year *int) ([]domain.SequenceOffset, error) {
	var offsets []domain.SequenceOffset
	query := r.db.Model(&domain.SequenceOffset{}).Where("municipality_id = ?", municipalityID)
	if year != nil {
		query = query.Where("year = ? OR year IS NULL", *year)
	}
	err := query.Order("type ASC, contract_type ASC").Find(&offsets).Error
	return offsets, err
}

func (r *sequenceRepository) FindOffset(municipalityID uuid.UUID, docType domain.DocumentType, contractType *domain.ContractType, year *int) (*domain.SequenceOffset, error) {
	var offset domain.SequenceOffset
	query := r.db.Model(&domain.SequenceOffset{}).
		Where("municipality_id = ? AND type = ?", municipalityID, docType)

	if contractType != nil {
		query = query.Where("contract_type = ?", *contractType)
	} else {
		query = query.Where("contract_type IS NULL")
	}

	if year != nil {
		query = query.Where("year = ?", *year)
	} else {
		query = query.Where("year IS NULL")
	}

	err := query.First(&offset).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &offset, err
}

func (r *sequenceRepository) Upsert(offset *domain.SequenceOffset) error {
	var existing domain.SequenceOffset
	query := r.db.Model(&domain.SequenceOffset{}).
		Where("municipality_id = ? AND type = ?", offset.MunicipalityID, offset.Type)

	if offset.ContractType != nil {
		query = query.Where("contract_type = ?", *offset.ContractType)
	} else {
		query = query.Where("contract_type IS NULL")
	}

	if offset.Year != nil {
		query = query.Where("year = ?", *offset.Year)
	} else {
		query = query.Where("year IS NULL")
	}

	err := query.First(&existing).Error
	if err == nil {
		existing.InitialOrder = offset.InitialOrder
		existing.UpdatedAt = time.Now()
		offset.ID = existing.ID
		offset.CreatedAt = existing.CreatedAt
		offset.UpdatedAt = existing.UpdatedAt
		return r.db.Save(&existing).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if offset.ID == uuid.Nil {
			offset.ID = uuid.New()
		}
		return r.db.Create(offset).Error
	}
	return err
}
