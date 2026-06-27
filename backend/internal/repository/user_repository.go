package repository

import (
	"errors"
	"strings"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository cria uma nova instância do repositório de usuários.
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create insere um novo usuário no banco de dados.
func (r *userRepository) Create(u *domain.User) error {
	return translateUserPgError(r.db.Create(u).Error)
}

// FindAll retorna todos os usuários ativos paginados.
func (r *userRepository) FindAll(page, pageSize int) ([]domain.User, int64, error) {
	var (
		users []domain.User
		total int64
	)

	offset := (page - 1) * pageSize

	if err := r.db.Model(&domain.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Preload("Municipality").
		Order("username ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// FindDeleted retorna os usuários deletados de forma lógica (soft delete), paginados.
func (r *userRepository) FindDeleted(page, pageSize int) ([]domain.User, int64, error) {
	var (
		users []domain.User
		total int64
	)

	offset := (page - 1) * pageSize
	query := r.db.Unscoped().
		Model(&domain.User{}).
		Where("deleted_at IS NOT NULL")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Municipality").
		Order("deleted_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// FindByID busca um usuário ativo pelo seu ID.
func (r *userRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.Preload("Municipality").First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// FindByIDUnscoped busca um usuário (ativo ou deletado) pelo seu ID.
func (r *userRepository) FindByIDUnscoped(id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.Unscoped().Preload("Municipality").First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// FindByEmail busca um usuário ativo pelo seu endereço de e-mail.
func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	err := r.db.Preload("Municipality").First(&u, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// FindByUsername busca um usuário ativo pelo seu nome de usuário (username).
func (r *userRepository) FindByUsername(username string) (*domain.User, error) {
	var u domain.User
	err := r.db.Preload("Municipality").First(&u, "username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// Update atualiza as informações de um usuário existente.
func (r *userRepository) Update(u *domain.User) error {
	return translateUserPgError(r.db.Save(u).Error)
}

// Delete remove um usuário de forma lógica (soft delete) pelo seu ID.
func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.User{}, "id = ?", id).Error
}

// Restore restaura um usuário que sofreu soft delete.
func (r *userRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// HardDelete remove permanentemente um usuário do banco de dados pelo seu ID.
func (r *userRepository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&domain.User{}, "id = ?", id).Error
}

// ExistsByEmail verifica se já existe um usuário com o e-mail informado (podendo desconsiderar um ID específico).
func (r *userRepository) ExistsByEmail(email string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.Model(&domain.User{}).Where("LOWER(email) = LOWER(?)", email)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// ExistsByUsername verifica se já existe um usuário com o username informado (podendo desconsiderar um ID específico).
func (r *userRepository) ExistsByUsername(username string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.Model(&domain.User{}).Where("LOWER(username) = LOWER(?)", username)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// translateUserPgError traduz erros do driver PostgreSQL em erros de domínio para entidades de usuários.
func translateUserPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgErrUniqueViolation {
			constraint := strings.ToLower(pgErr.ConstraintName)
			detail := strings.ToLower(pgErr.Detail)
			if strings.Contains(constraint, "email") || strings.Contains(detail, "email") {
				return domain.ErrEmailAlreadyExists
			}
			if strings.Contains(constraint, "username") || strings.Contains(detail, "username") {
				return domain.ErrUsernameAlreadyExists
			}
		}
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key value violates unique constraint") {
		if strings.Contains(msg, "email") {
			return domain.ErrEmailAlreadyExists
		}
		if strings.Contains(msg, "username") {
			return domain.ErrUsernameAlreadyExists
		}
	}

	return err
}
