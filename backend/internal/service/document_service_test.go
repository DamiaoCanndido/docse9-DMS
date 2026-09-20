package service_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service/mocks"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newDocumentService(t *testing.T) (domain.DocumentService, *mocks.DocumentRepository, *mocks.UserRepository, *mocks.MunicipalityRepository) {
	t.Helper()
	docRepo := new(mocks.DocumentRepository)
	userRepo := new(mocks.UserRepository)
	munRepo := new(mocks.MunicipalityRepository)
	storageSvc := storage.NewMockStorageService()
	svc := service.NewDocumentService(docRepo, userRepo, munRepo, storageSvc)
	return svc, docRepo, userRepo, munRepo
}

func newDocumentServiceWithStorage(t *testing.T) (domain.DocumentService, *mocks.DocumentRepository, *mocks.UserRepository, *mocks.MunicipalityRepository, *storage.MockStorageService) {
	t.Helper()
	docRepo := new(mocks.DocumentRepository)
	userRepo := new(mocks.UserRepository)
	munRepo := new(mocks.MunicipalityRepository)
	storageSvc := storage.NewMockStorageService()
	svc := service.NewDocumentService(docRepo, userRepo, munRepo, storageSvc)
	return svc, docRepo, userRepo, munRepo, storageSvc
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
	docRepo.On("CreateWithNextOrder", mock.AnythingOfType("*domain.Document"), &currentYear).Return(nil)

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
	docRepo.On("CreateWithNextOrder", mock.AnythingOfType("*domain.Document"), (*int)(nil)).Return(nil)

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
	docRepo.On("CreateWithNextOrder", mock.AnythingOfType("*domain.Document"), &currentYear).Return(nil)

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
	docRepo.On("CreateWithNextOrder", mock.AnythingOfType("*domain.Document"), &currentYear).Return(nil)

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

func TestUpdateDocument_CreatedAt_NonContract_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	originalCreatedAt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	doc := &domain.Document{
		ID:          id,
		Description: "Portaria",
		Type:        domain.TypeOrdinance,
		CreatedAt:   originalCreatedAt,
	}

	newCreatedAt := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)
	input := domain.UpdateDocumentInput{
		CreatedAt: &newCreatedAt,
	}

	docRepo.On("FindByID", id).Return(doc, nil)
	docRepo.On("Update", mock.MatchedBy(func(d *domain.Document) bool {
		return d.CreatedAt.Equal(newCreatedAt)
	})).Return(nil)

	updatedDoc := &domain.Document{
		ID:          id,
		Description: "Portaria",
		Type:        domain.TypeOrdinance,
		CreatedAt:   newCreatedAt,
	}
	docRepo.On("FindByID", id).Return(updatedDoc, nil)

	result, err := svc.Update(id, input)

	require.NoError(t, err)
	assert.Equal(t, newCreatedAt, result.CreatedAt)
	docRepo.AssertExpectations(t)
}

func TestUpdateDocument_CreatedAt_Contract_Ignored(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	originalCreatedAt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	duration := 12
	cType := domain.ContractService
	val := 5000.0
	startIn := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)

	doc := &domain.Document{
		ID:           id,
		Description:  "Contrato",
		Type:         domain.TypeContract,
		Duration:     &duration,
		ContractType: &cType,
		Value:        &val,
		StartIn:      &startIn,
		CreatedAt:    originalCreatedAt,
	}

	attemptedCreatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	newStartIn := time.Date(2024, 2, 1, 15, 45, 0, 0, time.UTC)
	input := domain.UpdateDocumentInput{
		CreatedAt: &attemptedCreatedAt,
		StartIn:   &newStartIn,
	}

	docRepo.On("FindByID", id).Return(doc, nil)
	// Must keep originalCreatedAt in Update call
	docRepo.On("Update", mock.MatchedBy(func(d *domain.Document) bool {
		return d.CreatedAt.Equal(originalCreatedAt) && d.StartIn.Equal(newStartIn)
	})).Return(nil)

	updatedDoc := &domain.Document{
		ID:           id,
		Description:  "Contrato",
		Type:         domain.TypeContract,
		Duration:     &duration,
		ContractType: &cType,
		Value:        &val,
		StartIn:      &newStartIn,
		CreatedAt:    originalCreatedAt,
	}
	docRepo.On("FindByID", id).Return(updatedDoc, nil)

	result, err := svc.Update(id, input)

	require.NoError(t, err)
	assert.Equal(t, originalCreatedAt, result.CreatedAt)
	assert.Equal(t, &newStartIn, result.StartIn)
	docRepo.AssertExpectations(t)
}

func TestUpdateDocument_Contract_Fields_All(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	duration := 12
	cType := domain.ContractService
	val := 5000.0
	startIn := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)

	doc := &domain.Document{
		ID:           id,
		Description:  "Contrato Antigo",
		Type:         domain.TypeContract,
		Duration:     &duration,
		ContractType: &cType,
		Value:        &val,
		StartIn:      &startIn,
	}

	newDuration := 24
	newCType := domain.ContractBidding
	newVal := 10000.0
	newDesc := "Contrato Novo"

	input := domain.UpdateDocumentInput{
		Description:  &newDesc,
		Duration:     &newDuration,
		ContractType: &newCType,
		Value:        &newVal,
	}

	docRepo.On("FindByID", id).Return(doc, nil).Once()
	docRepo.On("Update", mock.AnythingOfType("*domain.Document")).Return(nil)
	docRepo.On("FindByID", id).Return(&domain.Document{
		ID:           id,
		Description:  newDesc,
		Type:         domain.TypeContract,
		Duration:     &newDuration,
		ContractType: &newCType,
		Value:        &newVal,
	}, nil).Once()

	res, err := svc.Update(id, input)
	require.NoError(t, err)
	assert.Equal(t, newDesc, res.Description)
	assert.Equal(t, &newDuration, res.Duration)
	assert.Equal(t, &newCType, res.ContractType)
	assert.Equal(t, &newVal, res.Value)
}

func TestUpdateDocument_Contract_InvalidContractType(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	doc := &domain.Document{
		ID:   id,
		Type: domain.TypeContract,
	}

	invalidType := domain.ContractType("INVALID_CONTRACT_TYPE")
	input := domain.UpdateDocumentInput{
		ContractType: &invalidType,
	}

	docRepo.On("FindByID", id).Return(doc, nil)

	_, err := svc.Update(id, input)
	assert.ErrorIs(t, err, domain.ErrInvalidContractType)
}

func TestUpdateDocument_NotFound(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	newDesc := "Novo"
	input := domain.UpdateDocumentInput{Description: &newDesc}

	docRepo.On("FindByID", id).Return((*domain.Document)(nil), nil)

	_, err := svc.Update(id, input)
	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
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

func TestGetAllDocuments_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	filter := domain.DocumentFilter{}
	docs := []domain.Document{{ID: uuid.New(), Description: "Doc 1"}}

	docRepo.On("FindAll", filter, 1, 20).Return(docs, int64(1), nil)

	result, total, err := svc.GetAll(filter, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestGetDeletedDocuments_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	filter := domain.DocumentFilter{}
	docs := []domain.Document{{ID: uuid.New(), Description: "Deleted Doc"}}

	docRepo.On("FindDeleted", filter, 1, 20).Return(docs, int64(1), nil)

	result, total, err := svc.GetDeleted(filter, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestDeleteDocument_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	doc := &domain.Document{ID: id}

	docRepo.On("FindByID", id).Return(doc, nil)
	docRepo.On("Delete", id).Return(nil)

	err := svc.Delete(id)
	require.NoError(t, err)
}

func TestDeleteDocument_NotFound(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()

	docRepo.On("FindByID", id).Return((*domain.Document)(nil), nil)

	err := svc.Delete(id)
	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}

func TestRestoreDocument_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	deletedDoc := &domain.Document{
		ID:        id,
		DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
	}

	docRepo.On("FindByIDUnscoped", id).Return(deletedDoc, nil)
	docRepo.On("Restore", id).Return(nil)

	result, err := svc.Restore(id)
	require.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.False(t, result.DeletedAt.Valid)
}

func TestRestoreDocument_NotFound(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()

	docRepo.On("FindByIDUnscoped", id).Return((*domain.Document)(nil), nil)

	_, err := svc.Restore(id)
	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}

func TestRestoreDocument_NotDeleted(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	activeDoc := &domain.Document{ID: id} // DeletedAt is not set (Valid: false)

	docRepo.On("FindByIDUnscoped", id).Return(activeDoc, nil)

	_, err := svc.Restore(id)
	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}

func TestHardDeleteDocument_Success(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()
	doc := &domain.Document{ID: id}

	docRepo.On("FindByIDUnscoped", id).Return(doc, nil)
	docRepo.On("HardDelete", id).Return(nil)

	err := svc.HardDelete(id)
	require.NoError(t, err)
}

func TestHardDeleteDocument_NotFound(t *testing.T) {
	svc, docRepo, _, _ := newDocumentService(t)
	id := uuid.New()

	docRepo.On("FindByIDUnscoped", id).Return((*domain.Document)(nil), nil)

	err := svc.HardDelete(id)
	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}

// ──────────────────────────────────────────────
// Testes de Anexo e Storage (Cloudflare R2 / OCR)
// ──────────────────────────────────────────────

func buildTestPDFBytes(textContent string) []byte {
	var body bytes.Buffer
	body.WriteString("%PDF-1.4\n")

	var offsets []int

	offsets = append(offsets, body.Len())
	body.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	offsets = append(offsets, body.Len())
	body.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	if textContent != "" {
		offsets = append(offsets, body.Len())
		body.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n")

		offsets = append(offsets, body.Len())
		body.WriteString("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

		streamContent := fmt.Sprintf("BT\n/F1 12 Tf\n72 712 Td\n(%s) Tj\nET\n", textContent)
		offsets = append(offsets, body.Len())
		body.WriteString(fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(streamContent), streamContent))
	} else {
		offsets = append(offsets, body.Len())
		body.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n")

		offsets = append(offsets, body.Len())
		body.WriteString("4 0 obj\n<< /Length 0 >>\nstream\n\nendstream\nendobj\n")
	}

	startXref := body.Len()
	numObjs := len(offsets) + 1

	body.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", numObjs))
	for _, off := range offsets {
		body.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}

	body.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", numObjs, startXref))

	return body.Bytes()
}

func TestGenerateUploadURL_Success(t *testing.T) {
	svc, docRepo, _, _, _ := newDocumentServiceWithStorage(t)
	docID := uuid.New()
	munID := uuid.New()
	doc := &domain.Document{
		ID:             docID,
		MunicipalityID: munID,
		Type:           domain.TypeNotice,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	docRepo.On("FindByID", docID).Return(doc, nil)

	input := domain.UploadURLInput{
		FileName:    "oficio-01.pdf",
		FileSize:    1024 * 1024, // 1MB
		ContentType: "application/pdf",
	}

	res, err := svc.GenerateUploadURL(context.Background(), docID, input)
	require.NoError(t, err)
	assert.NotEmpty(t, res.UploadURL)
	assert.Contains(t, res.FileKey, fmt.Sprintf("tenants/%s/NOTICE/2026/%s/", munID, docID))
	assert.Equal(t, 600, res.ExpiresInSeconds)
}

func TestGenerateUploadURL_InvalidContentType(t *testing.T) {
	svc, docRepo, _, _, _ := newDocumentServiceWithStorage(t)
	docID := uuid.New()
	doc := &domain.Document{ID: docID}

	docRepo.On("FindByID", docID).Return(doc, nil)

	input := domain.UploadURLInput{
		FileName:    "foto.png",
		FileSize:    1024,
		ContentType: "image/png",
	}

	res, err := svc.GenerateUploadURL(context.Background(), docID, input)
	assert.ErrorIs(t, err, domain.ErrInvalidContentType)
	assert.Nil(t, res)
}

func TestGenerateUploadURL_FileTooLarge(t *testing.T) {
	svc, docRepo, _, _, _ := newDocumentServiceWithStorage(t)
	docID := uuid.New()
	doc := &domain.Document{ID: docID}

	docRepo.On("FindByID", docID).Return(doc, nil)

	input := domain.UploadURLInput{
		FileName:    "grande.pdf",
		FileSize:    30 * 1024 * 1024, // 30MB (> 25MB)
		ContentType: "application/pdf",
	}

	res, err := svc.GenerateUploadURL(context.Background(), docID, input)
	assert.ErrorIs(t, err, domain.ErrFileTooLarge)
	assert.Nil(t, res)
}

func TestConfirmUpload_Success_WithOCR(t *testing.T) {
	svc, docRepo, _, _, storageSvc := newDocumentServiceWithStorage(t)
	munID := uuid.New()
	docID := uuid.New()
	doc := &domain.Document{
		ID:             docID,
		MunicipalityID: munID,
		Type:           domain.TypeNotice,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		FileKey:        "",
	}

	key := fmt.Sprintf("tenants/%s/NOTICE/2026/%s/arquivo.pdf", munID, docID)
	validPDF := buildTestPDFBytes("Prefeitura Municipal de Teste Estado de Sergipe Publicacao Oficial Comprovada")
	storageSvc.PutTestObject(key, validPDF)

	docRepo.On("FindByID", docID).Return(doc, nil)
	docRepo.On("Update", mock.MatchedBy(func(d *domain.Document) bool {
		return d.FileKey == key
	})).Return(nil)

	input := domain.ConfirmUploadInput{FileKey: key}
	res, err := svc.ConfirmUpload(context.Background(), docID, input)

	require.NoError(t, err)
	assert.Equal(t, key, res.FileKey)
}

func TestConfirmUpload_Fails_MissingOCR(t *testing.T) {
	svc, docRepo, _, _, storageSvc := newDocumentServiceWithStorage(t)
	munID := uuid.New()
	docID := uuid.New()
	doc := &domain.Document{
		ID:             docID,
		MunicipalityID: munID,
		Type:           domain.TypeNotice,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		FileKey:        "",
	}

	key := fmt.Sprintf("tenants/%s/NOTICE/2026/%s/scanned_sem_ocr.pdf", munID, docID)
	scannedWithoutOCR := buildTestPDFBytes("") // PDF vazio sem texto
	storageSvc.PutTestObject(key, scannedWithoutOCR)

	docRepo.On("FindByID", docID).Return(doc, nil)

	input := domain.ConfirmUploadInput{FileKey: key}
	res, err := svc.ConfirmUpload(context.Background(), docID, input)

	assert.ErrorIs(t, err, domain.ErrPDFMissingOCR)
	assert.Nil(t, res)

	// Garante que o arquivo sem OCR foi excluído do storage
	exists, _ := storageSvc.ObjectExists(context.Background(), key)
	assert.False(t, exists)
}

func TestConfirmUpload_Fails_CrossTenantKey(t *testing.T) {
	svc, docRepo, _, _, storageSvc := newDocumentServiceWithStorage(t)
	munID1 := uuid.New()
	munID2 := uuid.New()
	docID1 := uuid.New()
	docID2 := uuid.New()

	doc := &domain.Document{
		ID:             docID1,
		MunicipalityID: munID1,
		Type:           domain.TypeNotice,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		FileKey:        "",
	}

	// Chave que pertence a outro município (munID2)
	crossTenantKey := fmt.Sprintf("tenants/%s/NOTICE/2026/%s/arquivo.pdf", munID2, docID2)
	validPDF := buildTestPDFBytes("Documento confidencial de outro municipio")
	storageSvc.PutTestObject(crossTenantKey, validPDF)

	docRepo.On("FindByID", docID1).Return(doc, nil)

	input := domain.ConfirmUploadInput{FileKey: crossTenantKey}
	res, err := svc.ConfirmUpload(context.Background(), docID1, input)

	assert.ErrorIs(t, err, domain.ErrInvalidFileKey)
	assert.Nil(t, res)

	// Garante que o arquivo de outro tenant NÃO foi deletado
	exists, _ := storageSvc.ObjectExists(context.Background(), crossTenantKey)
	assert.True(t, exists)
}

func TestConfirmUpload_Fails_InvalidExtension(t *testing.T) {
	svc, docRepo, _, _, _ := newDocumentServiceWithStorage(t)
	munID := uuid.New()
	docID := uuid.New()
	doc := &domain.Document{
		ID:             docID,
		MunicipalityID: munID,
		Type:           domain.TypeNotice,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		FileKey:        "",
	}

	docRepo.On("FindByID", docID).Return(doc, nil)

	invalidKey := fmt.Sprintf("tenants/%s/NOTICE/2026/%s/malware.exe", munID, docID)
	input := domain.ConfirmUploadInput{FileKey: invalidKey}
	res, err := svc.ConfirmUpload(context.Background(), docID, input)

	assert.ErrorIs(t, err, domain.ErrInvalidFileKey)
	assert.Nil(t, res)
}

func TestConfirmUpload_Fails_WrongDocumentKey(t *testing.T) {
	svc, docRepo, _, _, _ := newDocumentServiceWithStorage(t)
	munID := uuid.New()
	docID1 := uuid.New()
	docID2 := uuid.New()

	doc := &domain.Document{
		ID:             docID1,
		MunicipalityID: munID,
		Type:           domain.TypeNotice,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		FileKey:        "",
	}

	docRepo.On("FindByID", docID1).Return(doc, nil)

	// Mesmo município, mas ID de outro documento
	wrongDocKey := fmt.Sprintf("tenants/%s/NOTICE/2026/%s/arquivo.pdf", munID, docID2)
	input := domain.ConfirmUploadInput{FileKey: wrongDocKey}
	res, err := svc.ConfirmUpload(context.Background(), docID1, input)

	assert.ErrorIs(t, err, domain.ErrInvalidFileKey)
	assert.Nil(t, res)
}

func TestGenerateFileURL_Success(t *testing.T) {
	svc, docRepo, _, _, _ := newDocumentServiceWithStorage(t)
	docID := uuid.New()
	doc := &domain.Document{
		ID:        docID,
		Type:      domain.TypeNotice,
		Order:     42,
		CreatedAt: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		FileKey:   "tenants/mun1/NOTICE/2026/doc1/arquivo.pdf",
	}

	docRepo.On("FindByIDUnscoped", docID).Return(doc, nil)

	res, err := svc.GenerateFileURL(context.Background(), docID, false)
	require.NoError(t, err)
	assert.NotEmpty(t, res.URL)
	assert.Equal(t, 900, res.ExpiresInSeconds)
}

func TestGenerateFileURL_SoftDeletedInTrash_Success(t *testing.T) {
	svc, docRepo, _, _, _ := newDocumentServiceWithStorage(t)
	docID := uuid.New()
	deletedAt := gorm.DeletedAt{Time: time.Now(), Valid: true}
	doc := &domain.Document{
		ID:        docID,
		Type:      domain.TypeNotice,
		Order:     42,
		CreatedAt: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		DeletedAt: deletedAt,
		FileKey:   "tenants/mun1/NOTICE/2026/doc1/arquivo.pdf",
	}

	docRepo.On("FindByIDUnscoped", docID).Return(doc, nil)

	res, err := svc.GenerateFileURL(context.Background(), docID, false)
	require.NoError(t, err)
	assert.NotEmpty(t, res.URL)
	assert.Equal(t, 900, res.ExpiresInSeconds)
}

func TestGenerateFileURL_FileNotFound(t *testing.T) {
	svc, docRepo, _, _, _ := newDocumentServiceWithStorage(t)
	docID := uuid.New()
	doc := &domain.Document{
		ID:      docID,
		FileKey: "", // sem anexo
	}

	docRepo.On("FindByIDUnscoped", docID).Return(doc, nil)

	res, err := svc.GenerateFileURL(context.Background(), docID, false)
	assert.ErrorIs(t, err, domain.ErrFileNotFound)
	assert.Nil(t, res)
}


