package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Erros de domínio da entidade Municipality.
var ErrNameAlreadyExists = errors.New("já existe um município com este nome")

// Municipality representa um município (organização base do sistema)
type Municipality struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex"         json:"name"`
	UF        string         `gorm:"type:char(2);not null"                          json:"uf"`
	ImageURL  string         `gorm:"type:text"                                      json:"imageUrl,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                                 json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                                 json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"` // soft-delete
}

// ──────────────────────────────────────────────
// DTOs
// ──────────────────────────────────────────────

type CreateMunicipalityInput struct {
	Name     string `json:"name"     binding:"required,min=2,max=255"`
	UF       string `json:"uf"       binding:"required,len=2"`
	ImageURL string `json:"imageUrl" binding:"omitempty,url"`
}

type UpdateMunicipalityInput struct {
	Name     *string `json:"name"     binding:"omitempty,min=2,max=255"`
	UF       *string `json:"uf"       binding:"omitempty,len=2"`
	ImageURL *string `json:"imageUrl" binding:"omitempty,url"`
}

// ──────────────────────────────────────────────
// Repository interface (porta de saída)
// ──────────────────────────────────────────────

type MunicipalityRepository interface {
	Create(m *Municipality) error
	FindAll(page, pageSize int) ([]Municipality, int64, error)
	FindDeleted(page, pageSize int) ([]Municipality, int64, error)
	FindByID(id uuid.UUID) (*Municipality, error)
	FindByIDUnscoped(id uuid.UUID) (*Municipality, error)
	FindByUF(uf string, page, pageSize int) ([]Municipality, int64, error)
	Update(m *Municipality) error
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) error
	HardDelete(id uuid.UUID) error
	ExistsByName(name string, excludeID *uuid.UUID) (bool, error)
}

// ──────────────────────────────────────────────
// Service interface (porta de entrada)
// ──────────────────────────────────────────────

type MunicipalityService interface {
	Create(input CreateMunicipalityInput) (*Municipality, error)
	GetAll(page, pageSize int) ([]Municipality, int64, error)
	GetDeleted(page, pageSize int) ([]Municipality, int64, error)
	GetByID(id uuid.UUID) (*Municipality, error)
	GetByUF(uf string, page, pageSize int) ([]Municipality, int64, error)
	Update(id uuid.UUID, input UpdateMunicipalityInput) (*Municipality, error)
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) (*Municipality, error)
	HardDelete(id uuid.UUID) error
}
