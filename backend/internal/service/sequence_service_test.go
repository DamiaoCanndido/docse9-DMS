package service_test

import (
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSequences_Success(t *testing.T) {
	seqRepo := new(mocks.SequenceRepository)
	docRepo := new(mocks.DocumentRepository)
	svc := service.NewSequenceService(seqRepo, docRepo)

	munID := testhelper.MunPassagemID
	year := 2026

	// Mock para DECREE configurado com marco inicial 85 e lastOrder 0 -> nextOrder deve ser 85
	decreeOffsetID := uuid.New()
	decreeOffset := &domain.SequenceOffset{
		ID:             decreeOffsetID,
		MunicipalityID: munID,
		Type:           domain.TypeDecree,
		Year:           &year,
		InitialOrder:   85,
	}

	seqRepo.On("FindOffset", munID, domain.TypeNotice, (*domain.ContractType)(nil), &year).Return((*domain.SequenceOffset)(nil), nil)
	docRepo.On("GetLastOrder", munID, domain.TypeNotice, (*domain.ContractType)(nil), &year).Return(10, nil)

	seqRepo.On("FindOffset", munID, domain.TypeDecree, (*domain.ContractType)(nil), &year).Return(decreeOffset, nil)
	docRepo.On("GetLastOrder", munID, domain.TypeDecree, (*domain.ContractType)(nil), &year).Return(0, nil)

	seqRepo.On("FindOffset", munID, domain.TypeOrdinance, (*domain.ContractType)(nil), &year).Return((*domain.SequenceOffset)(nil), nil)
	docRepo.On("GetLastOrder", munID, domain.TypeOrdinance, (*domain.ContractType)(nil), &year).Return(0, nil)

	seqRepo.On("FindOffset", munID, domain.TypeLaw, (*domain.ContractType)(nil), (*int)(nil)).Return((*domain.SequenceOffset)(nil), nil)
	docRepo.On("GetLastOrder", munID, domain.TypeLaw, (*domain.ContractType)(nil), (*int)(nil)).Return(5, nil)

	contractService := domain.ContractService
	contractBidding := domain.ContractBidding
	contractPublicInterest := domain.ContractPublicInterest

	seqRepo.On("FindOffset", munID, domain.TypeContract, &contractService, &year).Return((*domain.SequenceOffset)(nil), nil)
	docRepo.On("GetLastOrder", munID, domain.TypeContract, &contractService, &year).Return(0, nil)

	seqRepo.On("FindOffset", munID, domain.TypeContract, &contractBidding, &year).Return((*domain.SequenceOffset)(nil), nil)
	docRepo.On("GetLastOrder", munID, domain.TypeContract, &contractBidding, &year).Return(0, nil)

	seqRepo.On("FindOffset", munID, domain.TypeContract, &contractPublicInterest, &year).Return((*domain.SequenceOffset)(nil), nil)
	docRepo.On("GetLastOrder", munID, domain.TypeContract, &contractPublicInterest, &year).Return(0, nil)

	items, err := svc.GetSequences(munID, &year)
	require.NoError(t, err)
	require.Len(t, items, 7)

	// Verificar NOTICE: sem offset (padrão 1), lastOrder 10 -> nextOrder 11
	assert.Equal(t, domain.TypeNotice, items[0].Type)
	assert.Equal(t, 1, items[0].InitialOrder)
	assert.Equal(t, 10, items[0].CurrentOrder)
	assert.Equal(t, 11, items[0].NextOrder)

	// Verificar DECREE: offset 85, lastOrder 0 -> nextOrder 85
	assert.Equal(t, domain.TypeDecree, items[1].Type)
	assert.Equal(t, 85, items[1].InitialOrder)
	assert.Equal(t, 0, items[1].CurrentOrder)
	assert.Equal(t, 85, items[1].NextOrder)
	assert.Equal(t, &decreeOffsetID, items[1].ID)

	// Verificar LAW: perpétuo (year=nil), lastOrder 5 -> nextOrder 6
	assert.Equal(t, domain.TypeLaw, items[3].Type)
	assert.Nil(t, items[3].Year)
	assert.Equal(t, 1, items[3].InitialOrder)
	assert.Equal(t, 5, items[3].CurrentOrder)
	assert.Equal(t, 6, items[3].NextOrder)

	seqRepo.AssertExpectations(t)
	docRepo.AssertExpectations(t)
}

func TestSetOffset_Success_Decree(t *testing.T) {
	seqRepo := new(mocks.SequenceRepository)
	docRepo := new(mocks.DocumentRepository)
	svc := service.NewSequenceService(seqRepo, docRepo)

	munID := testhelper.MunPassagemID
	year := 2026

	input := domain.SetSequenceOffsetInput{
		Type:         domain.TypeDecree,
		Year:         &year,
		InitialOrder: 85,
	}

	seqRepo.On("Upsert", mock.MatchedBy(func(offset *domain.SequenceOffset) bool {
		return offset.MunicipalityID == munID &&
			offset.Type == domain.TypeDecree &&
			offset.Year != nil && *offset.Year == 2026 &&
			offset.InitialOrder == 85 &&
			offset.ContractType == nil
	})).Return(nil)

	res, err := svc.SetOffset(munID, input)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 85, res.InitialOrder)
	assert.Equal(t, domain.TypeDecree, res.Type)

	seqRepo.AssertExpectations(t)
}

func TestSetOffset_Success_Law_ForcesNilYear(t *testing.T) {
	seqRepo := new(mocks.SequenceRepository)
	docRepo := new(mocks.DocumentRepository)
	svc := service.NewSequenceService(seqRepo, docRepo)

	munID := testhelper.MunPassagemID
	someYear := 2026

	input := domain.SetSequenceOffsetInput{
		Type:         domain.TypeLaw,
		Year:         &someYear, // fornecido, mas deve ser forçado a nil para leis
		InitialOrder: 100,
	}

	seqRepo.On("Upsert", mock.MatchedBy(func(offset *domain.SequenceOffset) bool {
		return offset.MunicipalityID == munID &&
			offset.Type == domain.TypeLaw &&
			offset.Year == nil &&
			offset.InitialOrder == 100
	})).Return(nil)

	res, err := svc.SetOffset(munID, input)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Nil(t, res.Year)
	assert.Equal(t, 100, res.InitialOrder)

	seqRepo.AssertExpectations(t)
}

func TestSetOffset_Success_Contract(t *testing.T) {
	seqRepo := new(mocks.SequenceRepository)
	docRepo := new(mocks.DocumentRepository)
	svc := service.NewSequenceService(seqRepo, docRepo)

	munID := testhelper.MunPassagemID
	year := 2026
	cType := domain.ContractService

	input := domain.SetSequenceOffsetInput{
		Type:         domain.TypeContract,
		ContractType: &cType,
		Year:         &year,
		InitialOrder: 50,
	}

	seqRepo.On("Upsert", mock.MatchedBy(func(offset *domain.SequenceOffset) bool {
		return offset.MunicipalityID == munID &&
			offset.Type == domain.TypeContract &&
			offset.ContractType != nil && *offset.ContractType == domain.ContractService &&
			offset.Year != nil && *offset.Year == 2026 &&
			offset.InitialOrder == 50
	})).Return(nil)

	res, err := svc.SetOffset(munID, input)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 50, res.InitialOrder)

	seqRepo.AssertExpectations(t)
}

func TestSetOffset_ValidationErrors(t *testing.T) {
	seqRepo := new(mocks.SequenceRepository)
	docRepo := new(mocks.DocumentRepository)
	svc := service.NewSequenceService(seqRepo, docRepo)

	munID := testhelper.MunPassagemID
	year := 2026

	// InitialOrder < 1
	_, err := svc.SetOffset(munID, domain.SetSequenceOffsetInput{
		Type:         domain.TypeNotice,
		Year:         &year,
		InitialOrder: 0,
	})
	assert.ErrorIs(t, err, domain.ErrInvalidInitialOrder)

	// Missing year for NOTICE
	_, err = svc.SetOffset(munID, domain.SetSequenceOffsetInput{
		Type:         domain.TypeNotice,
		Year:         nil,
		InitialOrder: 10,
	})
	assert.ErrorIs(t, err, domain.ErrInvalidSequenceYear)

	// Missing contractType for CONTRACT
	_, err = svc.SetOffset(munID, domain.SetSequenceOffsetInput{
		Type:         domain.TypeContract,
		Year:         &year,
		InitialOrder: 10,
	})
	assert.ErrorIs(t, err, domain.ErrInvalidSequenceContractType)
}
