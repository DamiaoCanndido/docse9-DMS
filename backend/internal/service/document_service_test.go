package service_test

import (
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newDocumentService(t *testing.T) (domain.DocumentService, *mocks.DocumentRepository, *mocks.UserRepository, *mocks.MunicipalityRepository) {
	t.Helper()
	docRepo := new(mocks.DocumentRepository)
	userRepo := new(mocks.UserRepository)
	munRepo := new(mocks.MunicipalityRepository)
	svc := service.NewDocumentService(docRepo, userRepo, munRepo)
	return svc, docRepo, userRepo, munRepo
}

func TestCreateDocument_Success_Notice(t *testing.T) {
	svc, docRepo, userRepo, munRepo := newDocumentService(t)

	mun := testhelper.MakePassagem()
	user := testhelper.MakeUserCommon(mun.ID)

	input := domain.CreateDocumentInput{
		Type:           domain.TypeNotice,
		Description:    "Ofício de teste",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
	}

	munRepo.On("FindByID", mun.ID).Return(&mun, nil)
	userRepo.On("FindByID", user.ID).Return(&user, nil)
	
	// Para NOTICE, o ano atual deve ser passado no cálculo da ordem
	currentYear := time.Now().Year()
	docRepo.On("GetLastOrder", mun.ID, domain.TypeNotice, (*domain.ContractType)(nil), &currentYear).Return(5, nil)
	
	docRepo.On("Create", mock.AnythingOfType("*domain.Document")).Return(nil)

	expectedDoc := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          6,
		Description:    "Ofício de teste",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
	}
	docRepo.On("FindByID", mock.Anything).Return(expectedDoc, nil)

	result, err := svc.Create(input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TypeNotice, result.Type)
	assert.Equal(t, 6, result.Order)
	assert.Equal(t, "Ofício de teste", result.Description)
	assert.Equal(t, user.ID, result.CreatorID)

	munRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	docRepo.AssertExpectations(t)
}

func TestCreateDocument_Success_Law(t *testing.T) {
	svc, docRepo, userRepo, munRepo := newDocumentService(t)

	mun := testhelper.MakePassagem()
	user := testhelper.MakeUserCommon(mun.ID)

	input := domain.CreateDocumentInput{
		Type:           domain.TypeLaw,
		Description:    "Lei de teste",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
	}

	munRepo.On("FindByID", mun.ID).Return(&mun, nil)
	userRepo.On("FindByID", user.ID).Return(&user, nil)
	
	// Para LAW, o ano deve ser nil (ordem nunca reseta)
	docRepo.On("GetLastOrder", mun.ID, domain.TypeLaw, (*domain.ContractType)(nil), (*int)(nil)).Return(12, nil)
	
	docRepo.On("Create", mock.AnythingOfType("*domain.Document")).Return(nil)

	expectedDoc := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Order:          13,
		Description:    "Lei de teste",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
	}
	docRepo.On("FindByID", mock.Anything).Return(expectedDoc, nil)

	result, err := svc.Create(input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TypeLaw, result.Type)
	assert.Equal(t, 13, result.Order)

	munRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	docRepo.AssertExpectations(t)
}

func TestCreateDocument_Success_Contract(t *testing.T) {
	svc, docRepo, userRepo, munRepo := newDocumentService(t)

	mun := testhelper.MakePassagem()
	user := testhelper.MakeUserCommon(mun.ID)

	duration := 12
	cType := domain.ContractService
	val := 15000.50
	startIn := time.Now()

	input := domain.CreateDocumentInput{
		Type:           domain.TypeContract,
		Description:    "Contrato de teste",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
		Duration:       &duration,
		ContractType:   &cType,
		Value:          &val,
		StartIn:        &startIn,
	}

	munRepo.On("FindByID", mun.ID).Return(&mun, nil)
	userRepo.On("FindByID", user.ID).Return(&user, nil)
	
	currentYear := time.Now().Year()
	docRepo.On("GetLastOrder", mun.ID, domain.TypeContract, &cType, &currentYear).Return(0, nil)
	
	docRepo.On("Create", mock.AnythingOfType("*domain.Document")).Return(nil)

	expectedDoc := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeContract,
		Order:          1,
		Description:    "Contrato de teste",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
		Duration:       &duration,
		ContractType:   &cType,
		Value:          &val,
		StartIn:        &startIn,
	}
	docRepo.On("FindByID", mock.Anything).Return(expectedDoc, nil)

	result, err := svc.Create(input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TypeContract, result.Type)
	assert.Equal(t, 1, result.Order)
	assert.Equal(t, &duration, result.Duration)
	assert.Equal(t, &cType, result.ContractType)
	assert.Equal(t, &val, result.Value)

	munRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	docRepo.AssertExpectations(t)
}

func TestCreateDocument_ContractTypes_IndependentSequences(t *testing.T) {
	svc, docRepo, userRepo, munRepo := newDocumentService(t)

	mun := testhelper.MakePassagem()
	user := testhelper.MakeUserCommon(mun.ID)

	duration := 12
	val := 15000.50
	startIn := time.Now()

	// 1. Contrato tipo Bidding
	cBidding := domain.ContractBidding
	inputBidding := domain.CreateDocumentInput{
		Type:           domain.TypeContract,
		Description:    "Contrato Licitação",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
		Duration:       &duration,
		ContractType:   &cBidding,
		Value:          &val,
		StartIn:        &startIn,
	}

	munRepo.On("FindByID", mun.ID).Return(&mun, nil)
	userRepo.On("FindByID", user.ID).Return(&user, nil)

	currentYear := time.Now().Year()
	docRepo.On("GetLastOrder", mun.ID, domain.TypeContract, &cBidding, &currentYear).Return(10, nil)

	docRepo.On("Create", mock.AnythingOfType("*domain.Document")).Return(nil)

	expectedDoc := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeContract,
		Order:          11,
		Description:    "Contrato Licitação",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
		Duration:       &duration,
		ContractType:   &cBidding,
		Value:          &val,
		StartIn:        &startIn,
	}
	docRepo.On("FindByID", mock.Anything).Return(expectedDoc, nil)

	result, err := svc.Create(inputBidding)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 11, result.Order)
	assert.Equal(t, &cBidding, result.ContractType)

	munRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	docRepo.AssertExpectations(t)
}

func TestCreateDocument_MissingContractFields(t *testing.T) {
	svc, _, userRepo, munRepo := newDocumentService(t)

	mun := testhelper.MakePassagem()
	user := testhelper.MakeUserCommon(mun.ID)

	input := domain.CreateDocumentInput{
		Type:           domain.TypeContract,
		Description:    "Contrato incompleto",
		CreatorID:      user.ID,
		MunicipalityID: mun.ID,
	}

	munRepo.On("FindByID", mun.ID).Return(&mun, nil)
	userRepo.On("FindByID", user.ID).Return(&user, nil)

	_, err := svc.Create(input)

	assert.ErrorIs(t, err, domain.ErrContractFieldsRequired)
}

func TestCreateDocument_MunicipalityNotFound(t *testing.T) {
	svc, _, _, munRepo := newDocumentService(t)

	input := domain.CreateDocumentInput{
		Type:           domain.TypeNotice,
		Description:    "Documento sem municipio",
		CreatorID:      uuid.New(),
		MunicipalityID: uuid.New(),
	}

	munRepo.On("FindByID", input.MunicipalityID).Return((*domain.Municipality)(nil), nil)

	_, err := svc.Create(input)

	assert.ErrorIs(t, err, service.ErrMunicipalityNotFound)
}

func TestCreateDocument_UserNotFound(t *testing.T) {
	svc, _, userRepo, munRepo := newDocumentService(t)

	mun := testhelper.MakePassagem()
	input := domain.CreateDocumentInput{
		Type:           domain.TypeNotice,
		Description:    "Documento sem autor",
		CreatorID:      uuid.New(),
		MunicipalityID: mun.ID,
	}

	munRepo.On("FindByID", mun.ID).Return(&mun, nil)
	userRepo.On("FindByID", input.CreatorID).Return((*domain.User)(nil), nil)

	_, err := svc.Create(input)

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestGetDocumentByID_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	doc := &domain.Document{ID: id, Description: "Doc"}

	docRepo.On("FindByID", id).Return(doc, nil)

	result, err := svc.GetByID(id)

	require.NoError(t, err)
	assert.Equal(t, id, result.ID)
}

func TestGetDocumentByID_NotFound(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()

	docRepo.On("FindByID", id).Return((*domain.Document)(nil), nil)

	_, err := svc.GetByID(id)

	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}

func TestUpdateDocument_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	doc := &domain.Document{ID: id, Description: "Old Description", Type: domain.TypeNotice}

	docRepo.On("FindByID", id).Return(doc, nil)
	docRepo.On("Update", doc).Return(nil)

	newDesc := "New Description"
	input := domain.UpdateDocumentInput{
		Description: &newDesc,
	}

	docRepo.On("FindByID", id).Return(&domain.Document{ID: id, Description: "New Description", Type: domain.TypeNotice}, nil)

	result, err := svc.Update(id, input)

	require.NoError(t, err)
	assert.Equal(t, "New Description", result.Description)
}

func TestGetDocumentByIDUnscoped_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	doc := &domain.Document{ID: id, Description: "Doc"}

	docRepo.On("FindByIDUnscoped", id).Return(doc, nil)

	result, err := svc.GetByIDUnscoped(id)

	require.NoError(t, err)
	assert.Equal(t, id, result.ID)
}

func TestGetDocumentByIDUnscoped_NotFound(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()

	docRepo.On("FindByIDUnscoped", id).Return((*domain.Document)(nil), nil)

	_, err := svc.GetByIDUnscoped(id)

	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}
