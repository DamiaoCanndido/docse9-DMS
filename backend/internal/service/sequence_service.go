package service

import (
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
)

type sequenceService struct {
	sequenceRepo domain.SequenceRepository
	docRepo      domain.DocumentRepository
}

// NewSequenceService cria uma nova instância do serviço de sequências numéricas.
func NewSequenceService(
	sequenceRepo domain.SequenceRepository,
	docRepo domain.DocumentRepository,
) domain.SequenceService {
	return &sequenceService{
		sequenceRepo: sequenceRepo,
		docRepo:      docRepo,
	}
}

func (s *sequenceService) GetSequences(municipalityID uuid.UUID, year *int) ([]domain.SequenceItemResponse, error) {
	currentYear := time.Now().Year()
	effectiveYear := currentYear
	if year != nil && *year > 0 {
		effectiveYear = *year
	}

	// 7 séries padrão do município:
	type seriesSpec struct {
		docType      domain.DocumentType
		contractType *domain.ContractType
		isPerpetual  bool
	}

	serviceContract := domain.ContractService
	biddingContract := domain.ContractBidding
	publicInterestContract := domain.ContractPublicInterest

	seriesList := []seriesSpec{
		{docType: domain.TypeNotice, isPerpetual: false},
		{docType: domain.TypeDecree, isPerpetual: false},
		{docType: domain.TypeOrdinance, isPerpetual: false},
		{docType: domain.TypeLaw, isPerpetual: true},
		{docType: domain.TypeContract, contractType: &serviceContract, isPerpetual: false},
		{docType: domain.TypeContract, contractType: &biddingContract, isPerpetual: false},
		{docType: domain.TypeContract, contractType: &publicInterestContract, isPerpetual: false},
	}

	var result []domain.SequenceItemResponse

	for _, spec := range seriesList {
		var seriesYear *int
		if !spec.isPerpetual {
			seriesYear = &effectiveYear
		}

		offset, err := s.sequenceRepo.FindOffset(municipalityID, spec.docType, spec.contractType, seriesYear)
		if err != nil {
			return nil, err
		}

		var id *uuid.UUID
		initialOrder := 1
		if offset != nil {
			id = &offset.ID
			if offset.InitialOrder >= 1 {
				initialOrder = offset.InitialOrder
			}
		}

		lastOrder, err := s.docRepo.GetLastOrder(municipalityID, spec.docType, spec.contractType, seriesYear)
		if err != nil {
			return nil, err
		}

		nextOrder := lastOrder
		if initialOrder-1 > nextOrder {
			nextOrder = initialOrder - 1
		}
		nextOrder++

		result = append(result, domain.SequenceItemResponse{
			ID:           id,
			Type:         spec.docType,
			ContractType: spec.contractType,
			Year:         seriesYear,
			InitialOrder: initialOrder,
			CurrentOrder: lastOrder,
			NextOrder:    nextOrder,
		})
	}

	return result, nil
}

func (s *sequenceService) SetOffset(municipalityID uuid.UUID, input domain.SetSequenceOffsetInput) (*domain.SequenceOffset, error) {
	if input.InitialOrder < 1 {
		return nil, domain.ErrInvalidInitialOrder
	}

	if input.Type == domain.TypeLaw {
		input.Year = nil
		input.ContractType = nil
	} else {
		if input.Year == nil || *input.Year <= 0 {
			return nil, domain.ErrInvalidSequenceYear
		}
		if input.Type == domain.TypeContract {
			if input.ContractType == nil {
				return nil, domain.ErrInvalidSequenceContractType
			}
			cType := *input.ContractType
			if cType != domain.ContractService && cType != domain.ContractBidding && cType != domain.ContractPublicInterest {
				return nil, domain.ErrInvalidContractType
			}
		} else {
			input.ContractType = nil
		}
	}

	offset := &domain.SequenceOffset{
		MunicipalityID: municipalityID,
		Type:           input.Type,
		ContractType:   input.ContractType,
		Year:           input.Year,
		InitialOrder:   input.InitialOrder,
	}

	if err := s.sequenceRepo.Upsert(offset); err != nil {
		return nil, err
	}

	return offset, nil
}
