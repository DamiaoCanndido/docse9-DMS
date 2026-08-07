package service

import (
	"strings"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
)

type documentService struct {
	docRepo          domain.DocumentRepository
	userRepo         domain.UserRepository
	municipalityRepo domain.MunicipalityRepository
}

// NewDocumentService cria uma nova instância do serviço de documentos.
func NewDocumentService(
	docRepo domain.DocumentRepository,
	userRepo domain.UserRepository,
	municipalityRepo domain.MunicipalityRepository,
) domain.DocumentService {
	return &documentService{
		docRepo:          docRepo,
		userRepo:         userRepo,
		municipalityRepo: municipalityRepo,
	}
}

func (s *documentService) Create(input domain.CreateDocumentInput) (*domain.Document, error) {
	// 1. Validar tipo de documento
	if input.Type != domain.TypeNotice &&
		input.Type != domain.TypeDecree &&
		input.Type != domain.TypeOrdinance &&
		input.Type != domain.TypeLaw &&
		input.Type != domain.TypeContract {
		return nil, domain.ErrInvalidDocumentType
	}

	// 2. Verificar se o município existe e está ativo
	mun, err := s.municipalityRepo.FindByID(input.MunicipalityID)
	if err != nil {
		return nil, err
	}
	if mun == nil {
		return nil, ErrMunicipalityNotFound
	}

	// 3. Verificar se o autor (owner) existe e está ativo
	owner, err := s.userRepo.FindByID(input.OwnerID)
	if err != nil {
		return nil, err
	}
	if owner == nil {
		return nil, domain.ErrUserNotFound
	}

	// 4. Validar campos específicos de contrato
	var duration *int
	var contractType *domain.ContractType
	var value *float64
	var startIn *time.Time

	if input.Type == domain.TypeContract {
		if input.Duration == nil || input.ContractType == nil || input.Value == nil || input.StartIn == nil {
			return nil, domain.ErrContractFieldsRequired
		}

		cType := *input.ContractType
		if cType != domain.ContractPublicInterest &&
			cType != domain.ContractBidding &&
			cType != domain.ContractService {
			return nil, domain.ErrInvalidContractType
		}

		duration = input.Duration
		contractType = input.ContractType
		value = input.Value
		startIn = input.StartIn
	}

	// 5. Calcular o próximo número sequencial (order)
	var year *int
	if input.Type != domain.TypeLaw {
		currentYear := time.Now().Year()
		year = &currentYear
	}

	lastOrder, err := s.docRepo.GetLastOrder(input.MunicipalityID, input.Type, year)
	if err != nil {
		return nil, err
	}
	nextOrder := lastOrder + 1

	// 6. Criar e persistir o documento
	doc := &domain.Document{
		ID:             uuid.New(),
		Type:           input.Type,
		Order:          nextOrder,
		Description:    strings.TrimSpace(input.Description),
		FileKey:        "", // em branco por padrão
		OwnerID:        input.OwnerID,
		MunicipalityID: input.MunicipalityID,
		Duration:       duration,
		ContractType:   contractType,
		Value:          value,
		StartIn:        startIn,
	}

	if err := s.docRepo.Create(doc); err != nil {
		return nil, err
	}

	// Recarrega relações
	return s.docRepo.FindByID(doc.ID)
}

func (s *documentService) GetAll(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	return s.docRepo.FindAll(filter, page, pageSize)
}

func (s *documentService) GetDeleted(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	return s.docRepo.FindDeleted(filter, page, pageSize)
}

func (s *documentService) GetByID(id uuid.UUID) (*domain.Document, error) {
	doc, err := s.docRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrDocumentNotFound
	}
	return doc, nil
}

func (s *documentService) Update(id uuid.UUID, input domain.UpdateDocumentInput) (*domain.Document, error) {
	// 1. Buscar documento existente
	doc, err := s.docRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrDocumentNotFound
	}

	// 2. Atualizar descrição se fornecida
	if input.Description != nil {
		doc.Description = strings.TrimSpace(*input.Description)
	}

	// 3. Atualizar fileKey se fornecido
	if input.FileKey != nil {
		doc.FileKey = strings.TrimSpace(*input.FileKey)
	}

	// 4. Se for contrato, atualiza campos específicos se fornecidos
	if doc.Type == domain.TypeContract {
		if input.Duration != nil {
			doc.Duration = input.Duration
		}
		if input.ContractType != nil {
			cType := *input.ContractType
			if cType != domain.ContractPublicInterest &&
				cType != domain.ContractBidding &&
				cType != domain.ContractService {
				return nil, domain.ErrInvalidContractType
			}
			doc.ContractType = input.ContractType
		}
		if input.Value != nil {
			doc.Value = input.Value
		}
		if input.StartIn != nil {
			doc.StartIn = input.StartIn
		}
	}

	// 5. Salvar alterações
	if err := s.docRepo.Update(doc); err != nil {
		return nil, err
	}

	return s.docRepo.FindByID(id)
}

func (s *documentService) Delete(id uuid.UUID) error {
	doc, err := s.docRepo.FindByID(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return domain.ErrDocumentNotFound
	}
	return s.docRepo.Delete(id)
}

func (s *documentService) Restore(id uuid.UUID) (*domain.Document, error) {
	doc, err := s.docRepo.FindByIDUnscoped(id)
	if err != nil {
		return nil, err
	}
	if doc == nil || !doc.DeletedAt.Valid {
		return nil, domain.ErrDocumentNotFound
	}

	if err := s.docRepo.Restore(id); err != nil {
		return nil, err
	}

	doc.DeletedAt.Valid = false
	return doc, nil
}

func (s *documentService) HardDelete(id uuid.UUID) error {
	doc, err := s.docRepo.FindByIDUnscoped(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return domain.ErrDocumentNotFound
	}
	return s.docRepo.HardDelete(id)
}
