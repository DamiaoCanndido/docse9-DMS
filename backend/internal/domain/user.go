package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Erros de domínio da entidade User.
var (
	ErrEmailAlreadyExists    = errors.New("já existe um usuário com este e-mail")
	ErrUsernameAlreadyExists = errors.New("já existe um usuário com este nome de usuário")
	ErrUserNotFound          = errors.New("usuário não encontrado")
)

type Role string

const (
	RoleAdmin  Role = "ADMIN"
	RoleCommon Role = "COMMON"
)

// User representa um usuário do sistema.
type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username       string         `gorm:"type:varchar(255);not null;uniqueIndex"         json:"username"`
	Email          string         `gorm:"type:varchar(255);not null;uniqueIndex"         json:"email"`
	Password       string         `gorm:"type:varchar(255);not null"                     json:"-"`
	Role           Role           `gorm:"type:varchar(50);not null;default:'COMMON'"      json:"role"`
	MunicipalityID uuid.UUID      `gorm:"type:uuid;not null"                             json:"municipalityId"`
	Municipality   Municipality   `gorm:"foreignKey:MunicipalityID"                      json:"municipality,omitempty"`
	LastLogin      *time.Time     `gorm:"type:timestamp"                                 json:"lastLogin,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                                 json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                                 json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                                          json:"-"` // soft-delete
}

// ──────────────────────────────────────────────
// DTOs
// ──────────────────────────────────────────────

type CreateUserInput struct {
	Username        string    `json:"username"        binding:"required,min=3,max=255"`
	Email           string    `json:"email"           binding:"required,email,max=255"`
	Password        string    `json:"password"        binding:"required,min=6,max=255"`
	ConfirmPassword string    `json:"confirmPassword" binding:"required,eqfield=Password"`
	Role            Role      `json:"role"            binding:"required,oneof=ADMIN COMMON"`
	MunicipalityID  uuid.UUID `json:"municipalityId"  binding:"required"`
}

type UpdateUserInput struct {
	Username       *string    `json:"username"       binding:"omitempty,min=3,max=255"`
	Email          *string    `json:"email"          binding:"omitempty,email,max=255"`
	Password       *string    `json:"password"       binding:"omitempty,min=6,max=255"`
	Role           *Role      `json:"role"           binding:"omitempty,oneof=ADMIN COMMON"`
	MunicipalityID *uuid.UUID `json:"municipalityId" binding:"omitempty"`
	LastLogin      *time.Time `json:"lastLogin"      binding:"omitempty"`
}

// ──────────────────────────────────────────────
// Repository interface (porta de saída)
// ──────────────────────────────────────────────

type UserRepository interface {
	Create(u *User) error
	FindAll(page, pageSize int) ([]User, int64, error)
	FindDeleted(page, pageSize int) ([]User, int64, error)
	FindByID(id uuid.UUID) (*User, error)
	FindByIDUnscoped(id uuid.UUID) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByUsername(username string) (*User, error)
	Update(u *User) error
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) error
	HardDelete(id uuid.UUID) error
	ExistsByEmail(email string, excludeID *uuid.UUID) (bool, error)
	ExistsByUsername(username string, excludeID *uuid.UUID) (bool, error)
}

// ──────────────────────────────────────────────
// Service interface (porta de entrada)
// ──────────────────────────────────────────────

type UserService interface {
	Create(input CreateUserInput) (*User, error)
	GetAll(page, pageSize int) ([]User, int64, error)
	GetDeleted(page, pageSize int) ([]User, int64, error)
	GetByID(id uuid.UUID) (*User, error)
	Update(id uuid.UUID, input UpdateUserInput) (*User, error)
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) (*User, error)
	HardDelete(id uuid.UUID) error
}
