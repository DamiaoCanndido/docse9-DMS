package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/repository"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/testhelper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserRepositorySuite struct {
	suite.Suite
	container testcontainers.Container
	db        *gorm.DB
	munRepo   domain.MunicipalityRepository
	repo      domain.UserRepository
	mun       domain.Municipality
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}

func (s *UserRepositorySuite) SetupSuite() {
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

	s.Require().NoError(db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error)
	s.Require().NoError(db.AutoMigrate(&domain.Municipality{}, &domain.User{}))

	s.db = db
	s.munRepo = repository.NewMunicipalityRepository(db)
	s.repo = repository.NewUserRepository(db)
}

func (s *UserRepositorySuite) TearDownSuite() {
	_ = s.container.Terminate(context.Background())
}

func (s *UserRepositorySuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE municipalities RESTART IDENTITY CASCADE")

	// Criar um município padrão para associar aos usuários nos testes.
	s.mun = testhelper.MakePassagem()
	s.Require().NoError(s.munRepo.Create(&s.mun))
}

func (s *UserRepositorySuite) insertUserAdmin() domain.User {
	u := testhelper.MakeUserAdmin(s.mun.ID)
	s.Require().NoError(s.repo.Create(&u))
	return u
}

func (s *UserRepositorySuite) insertUserCommon() domain.User {
	u := testhelper.MakeUserCommon(s.mun.ID)
	s.Require().NoError(s.repo.Create(&u))
	return u
}

// ═══════════════════════════════════════════════════════════════════════════════
// Create
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestCreate_Persists() {
	u := testhelper.MakeUserAdmin(s.mun.ID)
	err := s.repo.Create(&u)

	s.NoError(err)
	s.NotEqual(uuid.Nil, u.ID)

	// Buscar no banco para verificar se persistiu e preencheu timestamps
	dbUser, err := s.repo.FindByID(u.ID)
	s.NoError(err)
	s.NotNil(dbUser)
	s.Equal(u.Username, dbUser.Username)
	s.Equal(u.Email, dbUser.Email)
	s.Equal(s.mun.ID, dbUser.MunicipalityID)
	s.Equal(s.mun.Name, dbUser.Municipality.Name)
}

func (s *UserRepositorySuite) TestCreate_UniqueEmailConstraint() {
	s.insertUserAdmin()

	duplicate := testhelper.MakeUserCommon(s.mun.ID)
	duplicate.ID = uuid.New()
	// Mesmo e-mail do admin
	duplicate.Email = "admin@example.com"

	err := s.repo.Create(&duplicate)
	s.ErrorIs(err, domain.ErrEmailAlreadyExists)
}

func (s *UserRepositorySuite) TestCreate_UniqueUsernameConstraint() {
	s.insertUserAdmin()

	duplicate := testhelper.MakeUserCommon(s.mun.ID)
	duplicate.ID = uuid.New()
	// Mesmo username do admin
	duplicate.Username = "admin_user"

	err := s.repo.Create(&duplicate)
	s.ErrorIs(err, domain.ErrUsernameAlreadyExists)
}

// ═══════════════════════════════════════════════════════════════════════════════
// FindAll
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestFindAll_Pagination() {
	s.insertUserAdmin()  // username: admin_user
	s.insertUserCommon() // username: common_user

	// Página 1, 1 item por página
	result, total, err := s.repo.FindAll(1, 1)
	s.NoError(err)
	s.Len(result, 1)
	s.Equal(int64(2), total)
	s.Equal("admin_user", result[0].Username) // ordenado por username ASC
	s.Equal(s.mun.Name, result[0].Municipality.Name)

	// Página 2
	result2, _, err := s.repo.FindAll(2, 1)
	s.NoError(err)
	s.Len(result2, 1)
	s.Equal("common_user", result2[0].Username)
}

// ═══════════════════════════════════════════════════════════════════════════════
// FindDeleted
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestFindDeleted_Pagination() {
	u1 := s.insertUserAdmin()
	u2 := s.insertUserCommon()

	s.NoError(s.repo.Delete(u1.ID))
	s.NoError(s.repo.Delete(u2.ID))

	result, total, err := s.repo.FindDeleted(1, 1)
	s.NoError(err)
	s.Len(result, 1)
	s.Equal(int64(2), total)
	s.Equal(s.mun.Name, result[0].Municipality.Name)
}

// ═══════════════════════════════════════════════════════════════════════════════
// FindByID & FindByIDUnscoped
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestFindByID() {
	u := s.insertUserAdmin()

	res, err := s.repo.FindByID(u.ID)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(u.ID, res.ID)

	// Se deletar logicamente
	s.NoError(s.repo.Delete(u.ID))

	resDeleted, err := s.repo.FindByID(u.ID)
	s.NoError(err)
	s.Nil(resDeleted)

	// Unscoped deve achar
	resUnscoped, err := s.repo.FindByIDUnscoped(u.ID)
	s.NoError(err)
	s.NotNil(resUnscoped)
	s.Equal(u.ID, resUnscoped.ID)
}

// ═══════════════════════════════════════════════════════════════════════════════
// FindByEmail & FindByUsername
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestFindByEmail() {
	u := s.insertUserAdmin()

	res, err := s.repo.FindByEmail(u.Email)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(u.ID, res.ID)

	resNotFound, err := s.repo.FindByEmail("nonexistent@example.com")
	s.NoError(err)
	s.Nil(resNotFound)
}

func (s *UserRepositorySuite) TestFindByUsername() {
	u := s.insertUserAdmin()

	res, err := s.repo.FindByUsername(u.Username)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(u.ID, res.ID)

	resNotFound, err := s.repo.FindByUsername("nonexistent")
	s.NoError(err)
	s.Nil(resNotFound)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Update
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestUpdate() {
	u := s.insertUserAdmin()

	u.Username = "new_username"
	err := s.repo.Update(&u)
	s.NoError(err)

	res, err := s.repo.FindByID(u.ID)
	s.NoError(err)
	s.Equal("new_username", res.Username)
}

func (s *UserRepositorySuite) TestUpdate_Conflict() {
	u1 := s.insertUserAdmin()
	u2 := s.insertUserCommon()

	u2.Email = u1.Email // conflito de email
	err := s.repo.Update(&u2)
	s.ErrorIs(err, domain.ErrEmailAlreadyExists)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Restore
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestRestore() {
	u := s.insertUserAdmin()
	s.NoError(s.repo.Delete(u.ID))

	res, err := s.repo.FindByID(u.ID)
	s.NoError(err)
	s.Nil(res)

	s.NoError(s.repo.Restore(u.ID))

	resRestored, err := s.repo.FindByID(u.ID)
	s.NoError(err)
	s.NotNil(resRestored)
	s.Equal(u.ID, resRestored.ID)
}

// ═══════════════════════════════════════════════════════════════════════════════
// HardDelete
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestHardDelete() {
	u := s.insertUserAdmin()

	s.NoError(s.repo.HardDelete(u.ID))

	res, err := s.repo.FindByIDUnscoped(u.ID)
	s.NoError(err)
	s.Nil(res)
}

// ═══════════════════════════════════════════════════════════════════════════════
// ExistsByEmail & ExistsByUsername
// ═══════════════════════════════════════════════════════════════════════════════

func (s *UserRepositorySuite) TestExistsByEmail() {
	u := s.insertUserAdmin()

	exists, err := s.repo.ExistsByEmail("ADMIN@example.com", nil) // Case insensitive check
	s.NoError(err)
	s.True(exists)

	existsExclude, err := s.repo.ExistsByEmail("admin@example.com", &u.ID)
	s.NoError(err)
	s.False(existsExclude)

	existsNotFound, err := s.repo.ExistsByEmail("notfound@example.com", nil)
	s.NoError(err)
	s.False(existsNotFound)
}

func (s *UserRepositorySuite) TestExistsByUsername() {
	u := s.insertUserAdmin()

	exists, err := s.repo.ExistsByUsername("ADMIN_USER", nil) // Case insensitive check
	s.NoError(err)
	s.True(exists)

	existsExclude, err := s.repo.ExistsByUsername("admin_user", &u.ID)
	s.NoError(err)
	s.False(existsExclude)

	existsNotFound, err := s.repo.ExistsByUsername("notfound", nil)
	s.NoError(err)
	s.False(existsNotFound)
}
