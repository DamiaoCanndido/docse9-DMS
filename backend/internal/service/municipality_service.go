package service

import (
	"errors"
	"strings"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
)

// Erros de domínio — o handler os mapeia para HTTP status codes.
var (
	ErrMunicipalityNotFound     = errors.New("município não encontrado")
	ErrMunicipalityNameConflict = errors.New("já existe um município com este nome")
	ErrInvalidUF                = errors.New("UF inválida")
)

type municipalityService struct {
	repo domain.MunicipalityRepository
}

func NewMunicipalityService(repo domain.MunicipalityRepository) domain.MunicipalityService {
	return &municipalityService{repo: repo}
}

// validUFs contém as 27 unidades federativas do Brasil.
var validUFs = map[string]bool{
	"AC": true, "AL": true, "AP": true, "AM": true, "BA": true,
	"CE": true, "DF": true, "ES": true, "GO": true, "MA": true,
	"MT": true, "MS": true, "MG": true, "PA": true, "PB": true,
	"PR": true, "PE": true, "PI": true, "RJ": true, "RN": true,
	"RS": true, "RO": true, "RR": true, "SC": true, "SP": true,
	"SE": true, "TO": true,
}

func (s *municipalityService) Create(input domain.CreateMunicipalityInput) (*domain.Municipality, error) {
	uf := strings.ToUpper(strings.TrimSpace(input.UF))
	if !validUFs[uf] {
		return nil, ErrInvalidUF
	}

	exists, err := s.repo.ExistsByName(input.Name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrMunicipalityNameConflict
	}

	m := &domain.Municipality{
		ID:       uuid.New(),
		Name:     strings.TrimSpace(input.Name),
		UF:       uf,
		ImageURL: strings.TrimSpace(input.ImageURL),
	}

	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *municipalityService) GetAll(page, pageSize int) ([]domain.Municipality, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *municipalityService) GetByID(id uuid.UUID) (*domain.Municipality, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMunicipalityNotFound
	}
	return m, nil
}

func (s *municipalityService) GetByUF(uf string, page, pageSize int) ([]domain.Municipality, int64, error) {
	uf = strings.ToUpper(strings.TrimSpace(uf))
	if !validUFs[uf] {
		return nil, 0, ErrInvalidUF
	}
	return s.repo.FindByUF(uf, page, pageSize)
}

func (s *municipalityService) Update(id uuid.UUID, input domain.UpdateMunicipalityInput) (*domain.Municipality, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMunicipalityNotFound
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		exists, err := s.repo.ExistsByName(name, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrMunicipalityNameConflict
		}
		m.Name = name
	}

	if input.UF != nil {
		uf := strings.ToUpper(strings.TrimSpace(*input.UF))
		if !validUFs[uf] {
			return nil, ErrInvalidUF
		}
		m.UF = uf
	}

	if input.ImageURL != nil {
		m.ImageURL = strings.TrimSpace(*input.ImageURL)
	}

	if err := s.repo.Update(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *municipalityService) Delete(id uuid.UUID) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrMunicipalityNotFound
	}
	return s.repo.Delete(id)
}
