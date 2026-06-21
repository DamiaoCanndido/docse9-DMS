package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/seu-usuario/doc-manager/internal/domain"
	"gorm.io/gorm"
)

type municipalityRepository struct {
	db *gorm.DB
}

// New retorna uma implementação de MunicipalityRepository.
func NewMunicipalityRepository(db *gorm.DB) domain.MunicipalityRepository {
	return &municipalityRepository{db: db}
}

func (r *municipalityRepository) Create(m *domain.Municipality) error {
	return r.db.Create(m).Error
}

func (r *municipalityRepository) FindAll(page, pageSize int) ([]domain.Municipality, int64, error) {
	var (
		municipalities []domain.Municipality
		total          int64
	)

	offset := (page - 1) * pageSize

	if err := r.db.Model(&domain.Municipality{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Order("name ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&municipalities).Error; err != nil {
		return nil, 0, err
	}

	return municipalities, total, nil
}

func (r *municipalityRepository) FindByID(id uuid.UUID) (*domain.Municipality, error) {
	var m domain.Municipality
	err := r.db.First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}

func (r *municipalityRepository) FindByUF(uf string, page, pageSize int) ([]domain.Municipality, int64, error) {
	var (
		municipalities []domain.Municipality
		total          int64
	)

	offset := (page - 1) * pageSize
	query := r.db.Model(&domain.Municipality{}).Where("uf = ?", uf)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("name ASC").Offset(offset).Limit(pageSize).Find(&municipalities).Error; err != nil {
		return nil, 0, err
	}

	return municipalities, total, nil
}

func (r *municipalityRepository) Update(m *domain.Municipality) error {
	return r.db.Save(m).Error
}

func (r *municipalityRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Municipality{}, "id = ?", id).Error
}

func (r *municipalityRepository) ExistsByName(name string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.Model(&domain.Municipality{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
