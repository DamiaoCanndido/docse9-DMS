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

type DocumentRepositorySuite struct {
	suite.Suite
	container testcontainers.Container
	db        *gorm.DB
	munRepo   domain.MunicipalityRepository
	userRepo  domain.UserRepository
	repo      domain.DocumentRepository
	mun       domain.Municipality
	user      domain.User
}

func TestDocumentRepositorySuite(t *testing.T) {
	suite.Run(t, new(DocumentRepositorySuite))
}

func (s *DocumentRepositorySuite) SetupSuite() {
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
	s.Require().NoError(db.AutoMigrate(&domain.Municipality{}, &domain.User{}, &domain.Document{}))

	s.db = db
	s.munRepo = repository.NewMunicipalityRepository(db)
	s.userRepo = repository.NewUserRepository(db)
	s.repo = repository.NewDocumentRepository(db)
}

func (s *DocumentRepositorySuite) TearDownSuite() {
	_ = s.container.Terminate(context.Background())
}

func (s *DocumentRepositorySuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE documents RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE municipalities RESTART IDENTITY CASCADE")

	s.mun = testhelper.MakePassagem()
	s.Require().NoError(s.munRepo.Create(&s.mun))

	s.user = testhelper.MakeUserCommon(s.mun.ID)
	s.Require().NoError(s.userRepo.Create(&s.user))
}

func (s *DocumentRepositorySuite) TestCreateAndFindByID() {
	doc := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio de Teste",
		FileKey:        "file-key-123",
		OwnerID:        s.user.ID,
		MunicipalityID: s.mun.ID,
	}

	err := s.repo.Create(doc)
	s.Require().NoError(err)

	found, err := s.repo.FindByID(doc.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(doc.Description, found.Description)
	s.Equal(doc.FileKey, found.FileKey)
	s.Equal(doc.OwnerID, found.OwnerID)
	s.Equal(doc.MunicipalityID, found.MunicipalityID)
	s.Equal(doc.Order, found.Order)
}

func (s *DocumentRepositorySuite) TestGetLastOrder() {
	// 1. Criar um ofício no ano passado
	lastYear := time.Now().Year() - 1
	docLastYear := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          5,
		Description:    "Oficio do ano passado",
		OwnerID:        s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(lastYear, 5, 10, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.db.Create(docLastYear).Error)

	// 2. Criar ofícios no ano atual
	currentYear := time.Now().Year()
	doc1 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio 1",
		OwnerID:        s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(currentYear, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.db.Create(doc1).Error)

	doc2 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          2,
		Description:    "Oficio 2",
		OwnerID:        s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(currentYear, 2, 20, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.db.Create(doc2).Error)

	// 3. Criar uma lei
	law := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Order:          10,
		Description:    "Lei 1",
		OwnerID:        s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(lastYear, 1, 1, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.db.Create(law).Error)

	// Testar cálculo da ordem para ano atual
	orderNow, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeNotice, &currentYear)
	s.Require().NoError(err)
	s.Equal(2, orderNow)

	// Testar cálculo da ordem para o ano passado
	orderPast, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeNotice, &lastYear)
	s.Require().NoError(err)
	s.Equal(5, orderPast)

	// Testar cálculo da lei (sem ano)
	orderLaw, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeLaw, nil)
	s.Require().NoError(err)
	s.Equal(10, orderLaw)
}

func (s *DocumentRepositorySuite) TestFindAll_Filters() {
	doc1 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio Importante",
		OwnerID:        s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Now(),
	}
	s.Require().NoError(s.repo.Create(doc1))

	duration := 12
	cType := domain.ContractPublicInterest
	val := 50000.00
	startIn := time.Now()

	doc2 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeContract,
		Order:          1,
		Description:    "Contrato de Aluguel",
		OwnerID:        s.user.ID,
		MunicipalityID: s.mun.ID,
		Duration:       &duration,
		ContractType:   &cType,
		Value:          &val,
		StartIn:        &startIn,
		CreatedAt:      time.Now(),
	}
	s.Require().NoError(s.repo.Create(doc2))

	// 1. Filtrar por Tipo
	tFilter := domain.TypeNotice
	docs, total, err := s.repo.FindAll(domain.DocumentFilter{Type: &tFilter}, 1, 10)
	s.Require().NoError(err)
	s.Equal(int64(1), total)
	s.Len(docs, 1)
	s.Equal(doc1.ID, docs[0].ID)

	// 2. Filtrar por Busca Textual (LOWER e parcial)
	docs, total, err = s.repo.FindAll(domain.DocumentFilter{Search: "aluguel"}, 1, 10)
	s.Require().NoError(err)
	s.Equal(int64(1), total)
	s.Equal(doc2.ID, docs[0].ID)
}
