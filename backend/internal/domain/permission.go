package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPermissionNotFound = errors.New("permissões do usuário não encontradas")
)

// UserPermission define as permissões individuais de um usuário comum (COMMON) para gerenciar documentos de um determinado tipo.
type UserPermission struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_user_doc_type" json:"userId"`
	DocumentType DocumentType    `gorm:"type:varchar(50);not null;uniqueIndex:idx_user_doc_type" json:"documentType"`
	Level        PermissionLevel `gorm:"type:varchar(50);not null;default:'NONE'" json:"level"`
	CreatedAt    time.Time       `gorm:"autoCreateTime"                                 json:"createdAt"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime"                                 json:"updatedAt"`
}

type PermissionLevel string

const (
	LevelNone   PermissionLevel = "NONE"
	LevelRead   PermissionLevel = "READ"
	LevelWrite  PermissionLevel = "WRITE"
	LevelDelete PermissionLevel = "DELETE"
)

type UpdateUserPermissionsInput struct {
	Permissions []UpdateUserPermissionItem `json:"permissions" binding:"required,dive"`
}

type UpdateUserPermissionItem struct {
	DocumentType DocumentType    `json:"documentType" binding:"required,oneof=NOTICE DECREE ORDINANCE LAW CONTRACT"`
	Level        PermissionLevel `json:"level"        binding:"required,oneof=NONE READ WRITE DELETE"`
}

type UserPermissionRepository interface {
	FindByUserID(userID uuid.UUID) ([]UserPermission, error)
	Create(p *UserPermission) error
	DeleteByUserID(userID uuid.UUID) error
}
