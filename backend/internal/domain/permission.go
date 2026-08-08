package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPermissionNotFound = errors.New("permissões do usuário não encontradas")
)

// UserPermission define as permissões individuais de um usuário comum (COMMON) para gerenciar documentos.
type UserPermission struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"                 json:"userId"`
	CanView   bool      `gorm:"type:boolean;not null;default:false"            json:"canView"`
	CanCreate bool      `gorm:"type:boolean;not null;default:false"            json:"canCreate"`
	CanUpdate bool      `gorm:"type:boolean;not null;default:false"            json:"canUpdate"`
	CanDelete bool      `gorm:"type:boolean;not null;default:false"            json:"canDelete"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updatedAt"`
}

type UpdateUserPermissionInput struct {
	CanView   *bool `json:"canView"   binding:"omitempty"`
	CanCreate *bool `json:"canCreate" binding:"omitempty"`
	CanUpdate *bool `json:"canUpdate" binding:"omitempty"`
	CanDelete *bool `json:"canDelete" binding:"omitempty"`
}

type UserPermissionRepository interface {
	FindByUserID(userID uuid.UUID) (*UserPermission, error)
	Create(p *UserPermission) error
	Update(p *UserPermission) error
	DeleteByUserID(userID uuid.UUID) error
}
