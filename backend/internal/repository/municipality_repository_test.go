package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/repository"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ─── Suite ───────────────────────────────────────────────────────────────────

// MunicipalityRepositorySuite sobe um PostgreSQL real via Docker e roda todos
// os testes do repositório nele, fazendo rollback entre cada teste.
type MunicipalityRepositorySuite struct {
	suite.Suite
	container testcontainers.Container
	db        *gorm.DB
	repo      domain.MunicipalityRepository
}

func TestMunicipalityRepositorySuite(t *testing.T) {
	suite.Run(t, new(MunicipalityRepositorySuite))
}

// SetupSuite sobe o container uma única vez para toda a suite.
func (s *MunicipalityRepositorySuite) SetupSuite() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test_db",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.container = container

	host, err := container.Host(ctx)
	s.Require().NoError(err)
	port, err := container.MappedPort(ctx, "5432")
	s.Require().NoError(err)

	dsn := "host=" + host + " user=test password=test dbname=test_db port=" + port.Port() + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	s.Require().NoError(err)

	// pgcrypto para gen_random_uuid
	s.Require().NoError(db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error)
	s.Require().NoError(db.AutoMigrate(&domain.Municipality{}))

	s.db = db
}

// TearDownSuite para o container ao fim da suite.
func (s *MunicipalityRepositorySuite) TearDownSuite() {
	_ = s.container.Terminate(context.Background())
}

// SetupTest recria o repo e limpa a tabela antes de cada teste.
func (s *MunicipalityRepositorySuite) SetupTest() {
	s.repo = repository.NewMunicipalityRepository(s.db)
	s.db.Exec("TRUNCATE TABLE municipalities RESTART IDENTITY CASCADE")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (s *MunicipalityRepositorySuite) insertPassagem() domain.Municipality {
	m := testhelper.MakePassagem()
	s.Require().NoError(s.repo.Create(&m))
	return m
}

// ═══════════════════════════════════════════════════════════════════════════════
// Create
// ═══════════════════════════════════════════════════════════════════════════════

func (s *MunicipalityRepositorySuite) TestCreate_Persists() {
	m := testhelper.MakePassagem()
	err := s.repo.Create(&m)

	s.NoError(err)
	s.NotEqual(uuid.Nil, m.ID)
}

func (s *MunicipalityRepositorySuite) TestCreate_UniqueNameConstraint() {
	s.insertPassagem()

	duplicate := testhelper.MakePassagem()
	duplicate.ID = uuid.New()

	err := s.repo.Create(&duplicate)
	s.Error(err) // deve violar unique index
}

// ═══════════════════════════════════════════════════════════════════════════════
// FindAll
// ═══════════════════════════════════════════════════════════════════════════════

func (s *MunicipalityRepositorySuite) TestFindAll_Pagination() {
	s.insertPassagem()

	patos := testhelper.MakePatos()
	s.Require().NoError(s.repo.Create(&patos))

	// Página 1, 1 item por página
	result, total, err := s.repo.FindAll(1, 1)
	s.NoError(err)
	s.Len(result, 1)
	s.Equal(int64(2), total)

	// Página 2
	result2, _, err := s.repo.FindAll(2, 1)
	s.NoError(err)
	s.Len(result2, 1)

	// IDs distintos entre páginas
	s.NotEqual(result[0].ID, result2[0].ID)
}

func (s *MunicipalityRepositorySuite) TestFindAll_OrderedByName() {
	patos := testhelper.MakePatos()
	s.Require().NoError(s.repo.Create(&patos))
	s.insertPassagem()

	result, _, err := s.repo.FindAll(1, 20)
	s.NoError(err)

	// "Passagem" < "Patos" alfabeticamente
	s.Equal("Passagem", result[0].Name)
	s.Equal("Patos", result[1].Name)
}

// ═══════════════════════════════════════════════════════════════════════════════
// FindByID
// ═══════════════════════════════════════════════════════════════════════════════

func (s *MunicipalityRepositorySuite) TestFindByID_Found() {
	inserted := s.insertPassagem()

	found, err := s.repo.FindByID(inserted.ID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(inserted.ID, found.ID)
	s.Equal("Passagem", found.Name)
}

func (s *MunicipalityRepositorySuite) TestFindByID_NotFound_ReturnsNil() {
	found, err := s.repo.FindByID(testhelper.NonExistentID)

	s.NoError(err)   // sem erro, apenas nil
	s.Nil(found)
}

// ═══════════════════════════════════════════════════════════════════════════════
// FindByUF
// ═══════════════════════════════════════════════════════════════════════════════

func (s *MunicipalityRepositorySuite) TestFindByUF_FiltersCorrectly() {
	s.insertPassagem() // PB

	sp := domain.Municipality{ID: uuid.New(), Name: "São Paulo", UF: "SP"}
	s.Require().NoError(s.repo.Create(&sp))

	result, total, err := s.repo.FindByUF("PB", 1, 20)
	s.NoError(err)
	s.Len(result, 1)
	s.Equal(int64(1), total)
	s.Equal("PB", result[0].UF)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Update
// ═══════════════════════════════════════════════════════════════════════════════

func (s *MunicipalityRepositorySuite) TestUpdate_PersistsChanges() {
	inserted := s.insertPassagem()
	inserted.Name = "Passagem Editada"

	err := s.repo.Update(&inserted)
	s.NoError(err)

	found, _ := s.repo.FindByID(inserted.ID)
	s.Equal("Passagem Editada", found.Name)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Delete
// ═══════════════════════════════════════════════════════════════════════════════

func (s *MunicipalityRepositorySuite) TestDelete_SoftDelete() {
	inserted := s.insertPassagem()

	err := s.repo.Delete(inserted.ID)
	s.NoError(err)

	// FindByID não deve retornar registros soft-deletados
	found, err := s.repo.FindByID(inserted.ID)
	s.NoError(err)
	s.Nil(found)

	// O registro ainda existe fisicamente no banco
	var count int64
	s.db.Unscoped().Model(&domain.Municipality{}).
		Where("id = ?", inserted.ID).
		Count(&count)
	s.Equal(int64(1), count)
}

func (s *MunicipalityRepositorySuite) TestHardDelete_RemovesPhysically() {
	inserted := s.insertPassagem()

	err := s.repo.HardDelete(inserted.ID)
	s.NoError(err)

	var count int64
	s.db.Unscoped().Model(&domain.Municipality{}).
		Where("id = ?", inserted.ID).
		Count(&count)
	s.Equal(int64(0), count)
}

func (s *MunicipalityRepositorySuite) TestFindByIDUnscoped_FindsSoftDeleted() {
	inserted := s.insertPassagem()
	s.Require().NoError(s.repo.Delete(inserted.ID))

	found, err := s.repo.FindByIDUnscoped(inserted.ID)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(inserted.ID, found.ID)
	s.True(found.DeletedAt.Valid)
}

// ═══════════════════════════════════════════════════════════════════════════════
// ExistsByName
// ═══════════════════════════════════════════════════════════════════════════════

func (s *MunicipalityRepositorySuite) TestExistsByName_CaseInsensitive() {
	s.insertPassagem()

	exists, err := s.repo.ExistsByName("passagem", nil) // minúscula
	s.NoError(err)
	s.True(exists)
}

func (s *MunicipalityRepositorySuite) TestExistsByName_ExcludesOwnID() {
	inserted := s.insertPassagem()

	// Mesmo nome, mas excluindo o próprio ID = não deve colidir
	exists, err := s.repo.ExistsByName("Passagem", &inserted.ID)
	s.NoError(err)
	s.False(exists)
}

func (s *MunicipalityRepositorySuite) TestExistsByName_NotFound() {
	exists, err := s.repo.ExistsByName("Inexistente", nil)
	s.NoError(err)
	s.False(exists)
}

// ─── Standalone (sem suite) para compatibilidade com -run ────────────────────

func TestFindByID_Standalone(t *testing.T) {
	// Útil para rodar um teste isolado sem subir a suite inteira.
	// Requer que DATABASE_URL esteja setado no ambiente.
	t.Skip("use a suite: go test ./internal/repository/...")
}

// ─── Helpers testify require dentro da suite ─────────────────────────────────

func (s *MunicipalityRepositorySuite) NoError(err error, msgAndArgs ...any) {
	require.NoError(s.T(), err, msgAndArgs...)
}
