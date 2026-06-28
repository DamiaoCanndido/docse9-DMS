package service

import (
	"strings"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
)

type userService struct {
	userRepo         domain.UserRepository
	municipalityRepo domain.MunicipalityRepository
}

// NewUserService cria uma nova instância do serviço de usuários.
func NewUserService(userRepo domain.UserRepository, municipalityRepo domain.MunicipalityRepository) domain.UserService {
	return &userService{
		userRepo:         userRepo,
		municipalityRepo: municipalityRepo,
	}
}

func (s *userService) Create(input domain.CreateUserInput) (*domain.User, error) {
	// 1. Validar e normalizar email e username
	email := strings.ToLower(strings.TrimSpace(input.Email))
	username := strings.TrimSpace(input.Username)

	// 2. Verificar se o e-mail já existe
	emailExists, err := s.userRepo.ExistsByEmail(email, nil)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, domain.ErrEmailAlreadyExists
	}

	// 3. Verificar se o username já existe
	usernameExists, err := s.userRepo.ExistsByUsername(username, nil)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, domain.ErrUsernameAlreadyExists
	}

	// 4. Verificar se o município existe e está ativo
	mun, err := s.municipalityRepo.FindByID(input.MunicipalityID)
	if err != nil {
		return nil, err
	}
	if mun == nil {
		return nil, ErrMunicipalityNotFound
	}

	// 5. Criar a entidade User
	u := &domain.User{
		ID:             uuid.New(),
		Username:       username,
		Email:          email,
		Password:       input.Password, // Fase 2 cuidará do hashing real com bcrypt
		Role:           input.Role,
		MunicipalityID: input.MunicipalityID,
	}

	if err := s.userRepo.Create(u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *userService) GetAll(page, pageSize int) ([]domain.User, int64, error) {
	return s.userRepo.FindAll(page, pageSize)
}

func (s *userService) GetDeleted(page, pageSize int) ([]domain.User, int64, error) {
	return s.userRepo.FindDeleted(page, pageSize)
}

func (s *userService) GetByID(id uuid.UUID) (*domain.User, error) {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (s *userService) Update(id uuid.UUID, input domain.UpdateUserInput) (*domain.User, error) {
	// 1. Verificar se o usuário existe
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, domain.ErrUserNotFound
	}

	// 2. Atualizar username se fornecido
	if input.Username != nil {
		username := strings.TrimSpace(*input.Username)
		exists, err := s.userRepo.ExistsByUsername(username, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrUsernameAlreadyExists
		}
		u.Username = username
	}

	// 3. Atualizar e-mail se fornecido
	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		exists, err := s.userRepo.ExistsByEmail(email, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrEmailAlreadyExists
		}
		u.Email = email
	}

	// 4. Atualizar senha se fornecida
	if input.Password != nil {
		u.Password = *input.Password // Fase 2 cuidará do hashing real com bcrypt
	}

	// 5. Atualizar role se fornecida
	if input.Role != nil {
		u.Role = *input.Role
	}

	// 6. Atualizar município se fornecido
	if input.MunicipalityID != nil {
		mun, err := s.municipalityRepo.FindByID(*input.MunicipalityID)
		if err != nil {
			return nil, err
		}
		if mun == nil {
			return nil, ErrMunicipalityNotFound
		}
		u.MunicipalityID = *input.MunicipalityID
	}

	// 7. Atualizar last login se fornecido
	if input.LastLogin != nil {
		u.LastLogin = input.LastLogin
	}

	// 8. Salvar no repositório
	if err := s.userRepo.Update(u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *userService) Delete(id uuid.UUID) error {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if u == nil {
		return domain.ErrUserNotFound
	}
	return s.userRepo.Delete(id)
}

func (s *userService) Restore(id uuid.UUID) (*domain.User, error) {
	u, err := s.userRepo.FindByIDUnscoped(id)
	if err != nil {
		return nil, err
	}
	if u == nil || !u.DeletedAt.Valid {
		return nil, domain.ErrUserNotFound
	}

	if err := s.userRepo.Restore(id); err != nil {
		return nil, err
	}

	u.DeletedAt.Valid = false
	return u, nil
}

func (s *userService) HardDelete(id uuid.UUID) error {
	u, err := s.userRepo.FindByIDUnscoped(id)
	if err != nil {
		return err
	}
	if u == nil {
		return domain.ErrUserNotFound
	}
	return s.userRepo.HardDelete(id)
}
