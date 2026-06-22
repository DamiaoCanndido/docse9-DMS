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

// newSvc é um helper que instancia o service com o mock injetado.
func newSvc(t *testing.T) (domain.MunicipalityService, *mocks.MunicipalityRepository) {
	t.Helper()
	repo := new(mocks.MunicipalityRepository)
	svc := service.NewMunicipalityService(repo)
	return svc, repo
}

// ═══════════════════════════════════════════════════════════════════════════════
// Create
// ═══════════════════════════════════════════════════════════════════════════════

func TestCreate_Success(t *testing.T) {
	svc, repo := newSvc(t)
	input := testhelper.MakeCreateInput("Passagem", "PB", "https://example.com/img.png")

	repo.On("ExistsByName", "Passagem", (*uuid.UUID)(nil)).Return(false, nil)
	repo.On("Create", mock.AnythingOfType("*domain.Municipality")).Return(nil)

	result, err := svc.Create(input)

	require.NoError(t, err)
	assert.Equal(t, "Passagem", result.Name)
	assert.Equal(t, "PB", result.UF)
	assert.NotEqual(t, uuid.Nil, result.ID)
	repo.AssertExpectations(t)
}

func TestCreate_InvalidUF(t *testing.T) {
	svc, repo := newSvc(t)
	input := testhelper.MakeCreateInput("Passagem", "XX", "")

	_, err := svc.Create(input)

	assert.ErrorIs(t, err, service.ErrInvalidUF)
	repo.AssertNotCalled(t, "ExistsByName")
	repo.AssertNotCalled(t, "Create")
}

func TestCreate_UFCaseInsensitive(t *testing.T) {
	svc, repo := newSvc(t)
	input := testhelper.MakeCreateInput("Passagem", "pb", "") // minúscula

	repo.On("ExistsByName", "Passagem", (*uuid.UUID)(nil)).Return(false, nil)
	repo.On("Create", mock.AnythingOfType("*domain.Municipality")).Return(nil)

	result, err := svc.Create(input)

	require.NoError(t, err)
	assert.Equal(t, "PB", result.UF) // deve normalizar para maiúscula
}

func TestCreate_DuplicateName(t *testing.T) {
	svc, repo := newSvc(t)
	input := testhelper.MakeCreateInput("Passagem", "PB", "")

	repo.On("ExistsByName", "Passagem", (*uuid.UUID)(nil)).Return(true, nil)

	_, err := svc.Create(input)

	assert.ErrorIs(t, err, service.ErrMunicipalityNameConflict)
	repo.AssertNotCalled(t, "Create")
}

// ═══════════════════════════════════════════════════════════════════════════════
// GetByID
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetByID_Found(t *testing.T) {
	svc, repo := newSvc(t)
	m := testhelper.MakePassagem()

	repo.On("FindByID", m.ID).Return(&m, nil)

	result, err := svc.GetByID(m.ID)

	require.NoError(t, err)
	assert.Equal(t, m.ID, result.ID)
	assert.Equal(t, "Passagem", result.Name)
}

func TestGetByID_NotFound(t *testing.T) {
	svc, repo := newSvc(t)

	repo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	_, err := svc.GetByID(testhelper.NonExistentID)

	assert.ErrorIs(t, err, service.ErrMunicipalityNotFound)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GetAll
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetAll_ReturnsPaginatedResults(t *testing.T) {
	svc, repo := newSvc(t)
	municipalities := []domain.Municipality{testhelper.MakePassagem(), testhelper.MakePatos()}

	repo.On("FindAll", 1, 20).Return(municipalities, int64(2), nil)

	result, total, err := svc.GetAll(1, 20)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
}

func TestGetAll_EmptyResult(t *testing.T) {
	svc, repo := newSvc(t)

	repo.On("FindAll", 1, 20).Return([]domain.Municipality{}, int64(0), nil)

	result, total, err := svc.GetAll(1, 20)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.Equal(t, int64(0), total)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GetByUF
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetByUF_ValidUF(t *testing.T) {
	svc, repo := newSvc(t)
	municipalities := []domain.Municipality{testhelper.MakePassagem()}

	repo.On("FindByUF", "PB", 1, 20).Return(municipalities, int64(1), nil)

	result, total, err := svc.GetByUF("PB", 1, 20)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
}

func TestGetByUF_InvalidUF(t *testing.T) {
	svc, _ := newSvc(t)

	_, _, err := svc.GetByUF("ZZ", 1, 20)

	assert.ErrorIs(t, err, service.ErrInvalidUF)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Update
// ═══════════════════════════════════════════════════════════════════════════════

func TestUpdate_PartialUpdate_Name(t *testing.T) {
	svc, repo := newSvc(t)
	m := testhelper.MakePassagem()
	newName := "Passagem Nova"
	input := domain.UpdateMunicipalityInput{Name: &newName}

	repo.On("FindByID", m.ID).Return(&m, nil)
	repo.On("ExistsByName", newName, &m.ID).Return(false, nil)
	repo.On("Update", mock.AnythingOfType("*domain.Municipality")).Return(nil)

	result, err := svc.Update(m.ID, input)

	require.NoError(t, err)
	assert.Equal(t, "Passagem Nova", result.Name)
	assert.Equal(t, "PB", result.UF) // uf não mudou
}

func TestUpdate_NotFound(t *testing.T) {
	svc, repo := newSvc(t)
	newName := "Qualquer"
	input := domain.UpdateMunicipalityInput{Name: &newName}

	repo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	_, err := svc.Update(testhelper.NonExistentID, input)

	assert.ErrorIs(t, err, service.ErrMunicipalityNotFound)
	repo.AssertNotCalled(t, "Update")
}

func TestUpdate_NameConflict(t *testing.T) {
	svc, repo := newSvc(t)
	m := testhelper.MakePassagem()
	existingName := "Patos" // já existe no banco
	input := domain.UpdateMunicipalityInput{Name: &existingName}

	repo.On("FindByID", m.ID).Return(&m, nil)
	repo.On("ExistsByName", existingName, &m.ID).Return(true, nil)

	_, err := svc.Update(m.ID, input)

	assert.ErrorIs(t, err, service.ErrMunicipalityNameConflict)
	repo.AssertNotCalled(t, "Update")
}

// ═══════════════════════════════════════════════════════════════════════════════
// Delete
// ═══════════════════════════════════════════════════════════════════════════════

func TestDelete_Success(t *testing.T) {
	svc, repo := newSvc(t)
	m := testhelper.MakePassagem()

	repo.On("FindByID", m.ID).Return(&m, nil)
	repo.On("Delete", m.ID).Return(nil)

	err := svc.Delete(m.ID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	svc, repo := newSvc(t)

	repo.On("FindByID", testhelper.NonExistentID).Return(nil, nil)

	err := svc.Delete(testhelper.NonExistentID)

	assert.ErrorIs(t, err, service.ErrMunicipalityNotFound)
	repo.AssertNotCalled(t, "Delete")
}

func TestHardDelete_Success(t *testing.T) {
	svc, repo := newSvc(t)
	m := testhelper.MakePassagem()

	repo.On("FindByIDUnscoped", m.ID).Return(&m, nil)
	repo.On("HardDelete", m.ID).Return(nil)

	err := svc.HardDelete(m.ID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHardDelete_NotFound(t *testing.T) {
	svc, repo := newSvc(t)

	repo.On("FindByIDUnscoped", testhelper.NonExistentID).Return(nil, nil)

	err := svc.HardDelete(testhelper.NonExistentID)

	assert.ErrorIs(t, err, service.ErrMunicipalityNotFound)
	repo.AssertNotCalled(t, "HardDelete")
}
