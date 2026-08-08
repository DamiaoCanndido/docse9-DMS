package repository

import (
	"errors"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userPermissionRepository struct {
	db *gorm.DB
}

// NewUserPermissionRepository cria uma nova instância do repositório de permissões de usuário.
func NewUserPermissionRepository(db *gorm.DB) domain.UserPermissionRepository {
	return &userPermissionRepository{db: db}
}

func (r *userPermissionRepository) FindByUserID(userID uuid.UUID) (*domain.UserPermission, error) {
	var p domain.UserPermission
	err := r.db.First(&p, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *userPermissionRepository) Create(p *domain.UserPermission) error {
	return r.db.Create(p).Error
}

func (r *userPermissionRepository) Update(p *domain.UserPermission) error {
	return r.db.Save(p).Error
}

func (r *userPermissionRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Delete(&domain.UserPermission{}, "user_id = ?", userID).Error
}
