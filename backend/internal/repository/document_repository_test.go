package repository_test

import (
	"context"
	"fmt"
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
	if err != nil {
		s.T().Skip("Docker daemon não disponível, pulando suite de repositório:", err)
		return
	}
	s.container = container

	host, err := container.Host(ctx)
	s.Require().NoError(err)
	port, err := container.MappedPort(ctx, "5432")
	s.Require().NoError(err)

	dsn := "host=" + host + " user=test password=test dbname=test_db port=" + port.Port() + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	s.Require().NoError(err)

	s.Require().NoError(db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error)
	s.Require().NoError(db.AutoMigrate(&domain.Municipality{}, &domain.User{}, &domain.Document{}, &domain.SequenceOffset{}))
	_ = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_sequence_offsets_unique ON sequence_offsets (municipality_id, type, COALESCE(contract_type, ''), COALESCE(year, 0))")

	s.db = db
	s.munRepo = repository.NewMunicipalityRepository(db)
	s.userRepo = repository.NewUserRepository(db)
	s.repo = repository.NewDocumentRepository(db)
}

func (s *DocumentRepositorySuite) TearDownSuite() {
	if s.container != nil {
		_ = s.container.Terminate(context.Background())
	}
}

func (s *DocumentRepositorySuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE documents RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE sequence_offsets RESTART IDENTITY CASCADE")
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
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
	}

	err := s.repo.Create(doc)
	s.Require().NoError(err)

	found, err := s.repo.FindByID(doc.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(doc.Description, found.Description)
	s.Equal(doc.FileKey, found.FileKey)
	s.Equal(doc.CreatorID, found.CreatorID)
	s.Equal(doc.MunicipalityID, found.MunicipalityID)
	s.Equal(doc.Order, found.Order)
	s.Equal(s.user.Username, found.CreatedBy.Username)
	s.Equal(s.mun.Name, found.CreatedBy.Municipality.Name)
}

func (s *DocumentRepositorySuite) TestGetLastOrder() {
	// 1. Criar um ofício no ano passado
	lastYear := time.Now().Year() - 1
	docLastYear := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          5,
		Description:    "Oficio do ano passado",
		CreatorID:      s.user.ID,
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
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(currentYear, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.db.Create(doc1).Error)

	doc2 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          2,
		Description:    "Oficio 2",
		CreatorID:      s.user.ID,
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
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(lastYear, 1, 1, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.db.Create(law).Error)

	// Testar cálculo da ordem para ano atual
	orderNow, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeNotice, nil, &currentYear)
	s.Require().NoError(err)
	s.Equal(2, orderNow)

	// Testar cálculo da ordem para o ano passado
	orderPast, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeNotice, nil, &lastYear)
	s.Require().NoError(err)
	s.Equal(5, orderPast)

	// Testar cálculo da lei (sem ano)
	orderLaw, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeLaw, nil, nil)
	s.Require().NoError(err)
	s.Equal(10, orderLaw)
}

func (s *DocumentRepositorySuite) TestGetLastOrder_ContractTypes() {
	currentYear := time.Now().Year()

	cService := domain.ContractService
	cPublicInterest := domain.ContractPublicInterest
	cBidding := domain.ContractBidding

	duration := 12
	val := 1000.0
	startIn := time.Now()

	// 1. Criar contrato de serviço com ordem 3
	docService := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeContract,
		Order:          3,
		Description:    "Serviço",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		Duration:       &duration,
		ContractType:   &cService,
		Value:          &val,
		StartIn:        &startIn,
		CreatedAt:      time.Date(currentYear, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.db.Create(docService).Error)

	// 2. Criar contrato de licitação com ordem 5
	docBidding := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeContract,
		Order:          5,
		Description:    "Licitação",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		Duration:       &duration,
		ContractType:   &cBidding,
		Value:          &val,
		StartIn:        &startIn,
		CreatedAt:      time.Date(currentYear, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.db.Create(docBidding).Error)

	// 3. Obter última ordem para cada tipo de contrato no ano atual
	orderService, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeContract, &cService, &currentYear)
	s.Require().NoError(err)
	s.Equal(3, orderService)

	orderBidding, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeContract, &cBidding, &currentYear)
	s.Require().NoError(err)
	s.Equal(5, orderBidding)

	orderPI, err := s.repo.GetLastOrder(s.mun.ID, domain.TypeContract, &cPublicInterest, &currentYear)
	s.Require().NoError(err)
	s.Equal(0, orderPI) // nenhum criado
}

func (s *DocumentRepositorySuite) TestFindAll_Filters() {
	doc1 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          1,
		Description:    "Oficio Importante",
		CreatorID:      s.user.ID,
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
		CreatorID:      s.user.ID,
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

func (s *DocumentRepositorySuite) TestCreateWithNextOrder_Atomic_Concurrency() {
	year := 2026
	concurrency := 50
	errChan := make(chan error, concurrency)
	createdOrders := make(chan int, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			doc := &domain.Document{
				ID:             uuid.New(),
				Type:           domain.TypeDecree,
				Description:    "Decreto Concorrente 50 Goroutines",
				CreatorID:      s.user.ID,
				MunicipalityID: s.mun.ID,
			}
			err := s.repo.CreateWithNextOrder(doc, &year)
			if err != nil {
				errChan <- err
				return
			}
			createdOrders <- doc.Order
			errChan <- nil
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		err := <-errChan
		s.Require().NoError(err)
	}

	orderMap := make(map[int]bool)
	for i := 0; i < concurrency; i++ {
		ord := <-createdOrders
		s.False(orderMap[ord], "Order %d gerado em duplicidade durante concorrência!", ord)
		orderMap[ord] = true
	}
	s.Len(orderMap, concurrency)
}

func (s *DocumentRepositorySuite) TestCreateWithNextOrder_WithOffset() {
	year := 2026

	// 1. Configurar marco inicial de Decretos 2026 no número 85
	offset := &domain.SequenceOffset{
		ID:             uuid.New(),
		MunicipalityID: s.mun.ID,
		Type:           domain.TypeDecree,
		Year:           &year,
		InitialOrder:   85,
	}
	s.Require().NoError(s.db.Create(offset).Error)

	// 2. Primeiro decreto criado automaticamente deve receber order 85
	doc1 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeDecree,
		Description:    "Primeiro decreto após marco inicial",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(doc1, &year))
	s.Equal(85, doc1.Order)

	// 3. Segundo decreto criado automaticamente deve receber order 86
	doc2 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeDecree,
		Description:    "Segundo decreto subsequente",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(doc2, &year))
	s.Equal(86, doc2.Order)
}

func (s *DocumentRepositorySuite) TestCreateWithNextOrder_ManualOrder_AndConflict() {
	year := 2026

	// 1. Inserir documento com número manual (ex: 14)
	docManual := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          14,
		Description:    "Ofício legado papel nº 14",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(docManual, &year))
	s.Equal(14, docManual.Order)

	// 2. Tentar inserir outro documento manual com o mesmo número 14 -> Conflito 409
	docDup := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          14,
		Description:    "Tentativa duplicada do nº 14",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
	}
	err := s.repo.CreateWithNextOrder(docDup, &year)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrOrderAlreadyExists)

	// 3. Geração automática subsequente deve saltar para o próximo livre (15)
	docAuto := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          0, // automático
		Description:    "Ofício automático seguinte",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(docAuto, &year))
	s.Equal(15, docAuto.Order)
}

func (s *DocumentRepositorySuite) TestCreateWithNextOrder_LawPerpetual_WithOffset() {
	// 1. Configurar marco inicial de Leis (perpétuo, year = nil) no número 50
	offset := &domain.SequenceOffset{
		ID:             uuid.New(),
		MunicipalityID: s.mun.ID,
		Type:           domain.TypeLaw,
		Year:           nil,
		InitialOrder:   50,
	}
	s.Require().NoError(s.db.Create(offset).Error)

	// 2. Lei no ano de 2025
	law1 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Description:    "Lei Municipal em 2025",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(law1, nil))
	s.Equal(50, law1.Order)

	// 3. Próxima lei no ano de 2026 deve continuar a sequência perpétua (51)
	law2 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Description:    "Lei Municipal em 2026",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(law2, nil))
	s.Equal(51, law2.Order)
}

func (s *DocumentRepositorySuite) TestCreateWithNextOrder_ChronologicalIntegrity() {
	// Cenário do usuário:
	// 1. Lei criada em 1970 com a numeração 400
	law1970 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Order:          400,
		Description:    "Lei Municipal Histórica de 1970",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(1970, 5, 10, 12, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(law1970, nil))
	s.Equal(400, law1970.Order)

	// 2. Tentativa de criar Lei em 1990 com a numeração 300 (< 400, data posterior) -> Erro de inconsistência cronológica
	law1990Invalid := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Order:          300,
		Description:    "Lei de 1990 inválida com número inferior à de 1970",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(1990, 8, 15, 12, 0, 0, 0, time.UTC),
	}
	err := s.repo.CreateWithNextOrder(law1990Invalid, nil)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrChronologicalOrderInvalid)

	// 3. Criar Lei em 1990 com numeração 450 (> 400) -> Sucesso
	law1990Valid := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Order:          450,
		Description:    "Lei de 1990 válida com número superior à de 1970",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(1990, 8, 15, 12, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(law1990Valid, nil))
	s.Equal(450, law1990Valid.Order)

	// 4. Tentativa reversa: Criar Lei em 1960 com numeração 500 (> 450 de 1990, data anterior) -> Erro
	law1960Invalid := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Order:          500,
		Description:    "Lei de 1960 inválida com número superior à de 1990",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(1960, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	err = s.repo.CreateWithNextOrder(law1960Invalid, nil)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrChronologicalOrderInvalid)

	// 5. Mesma data (mesmo dia): Múltiplos atos no mesmo dia permitidos
	sameDayDate := time.Date(1980, 3, 15, 10, 0, 0, 0, time.UTC)
	law1 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Order:          410,
		Description:    "Lei 410 no dia 15/03/1980",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      sameDayDate,
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(law1, nil))

	law2 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeLaw,
		Order:          411,
		Description:    "Lei 411 no mesmo dia 15/03/1980",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      sameDayDate,
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(law2, nil))

	// 6. Atos Anuais (ex: DECREE): anos diferentes são independentes
	yr1970 := 1970
	yr1990 := 1990
	dec1970 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeDecree,
		Order:          400,
		Description:    "Decreto 400 de 1970",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(1970, 5, 10, 12, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(dec1970, &yr1970))

	dec1990 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeDecree,
		Order:          300,
		Description:    "Decreto 300 de 1990 (permitido pois o ciclo é anual)",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(1990, 8, 15, 12, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(dec1990, &yr1990))

	// 7. Atos Anuais: dentro do mesmo exercício anual (2026), inconsistência de ordem e data é bloqueada
	yr2026 := 2026
	decJan := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeDecree,
		Order:          50,
		Description:    "Decreto 50 em 10/01/2026",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(decJan, &yr2026))

	decFebInvalid := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeDecree,
		Order:          30, // < 50 em data posterior (fevereiro) -> inconsistência
		Description:    "Decreto 30 em 20/02/2026 inválido",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC),
	}
	err = s.repo.CreateWithNextOrder(decFebInvalid, &yr2026)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrChronologicalOrderInvalid)

	// 8. Atos no mesmo dia com horários distintos (Cenário do Ofício 12 vs 20/22):
	// Ofício 10 (11:56), Ofício 11 (12:00), Ofício 20 (12:00), Ofício 22 (13:46), Ofício 23 (26/09 18:53)
	// Tentativa de criar Ofício 12 com horário 20:04 (posterior aos ofícios 20 e 22) -> Erro
	// Criar Ofício 12 com horário 12:00 (consistente com 11 e 20) -> Sucesso
	recifeLoc, err := time.LoadLocation("America/Recife")
	s.Require().NoError(err)

	notice10 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          10,
		Description:    "teste de numeral",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 9, 25, 11, 56, 0, 0, recifeLoc),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(notice10, &yr2026))

	notice11 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          11,
		Description:    "testando offset",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 9, 25, 12, 0, 0, 0, recifeLoc),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(notice11, &yr2026))

	notice20 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          20,
		Description:    "testando offset 20",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 9, 25, 12, 0, 0, 0, recifeLoc),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(notice20, &yr2026))

	notice22 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          22,
		Description:    "teste 22",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 9, 25, 13, 46, 0, 0, recifeLoc),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(notice22, &yr2026))

	notice23 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          23,
		Description:    "oficio 23",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 9, 26, 18, 53, 0, 0, recifeLoc),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(notice23, &yr2026))

	// Tentativa inválida: Ofício 12 às 20:04 (data/hora posterior a 20 e 22)
	notice12Invalid := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          12,
		Description:    "oficio 12 invalido",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 9, 25, 20, 4, 0, 0, recifeLoc),
	}
	err = s.repo.CreateWithNextOrder(notice12Invalid, &yr2026)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrChronologicalOrderInvalid)

	// Criação válida: Ofício 12 às 12:00 (consistente entre 11 e 20)
	notice12Valid := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          12,
		Description:    "oficio 12 valido",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2026, 9, 25, 12, 0, 0, 0, recifeLoc),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(notice12Valid, &yr2026))
}

func (s *DocumentRepositorySuite) TestFindAll_StrictOrderSorting() {
	recifeLoc, err := time.LoadLocation("America/Recife")
	s.Require().NoError(err)

	yr2026 := 2026
	yr2025 := 2025

	items2026 := []struct {
		order int
		t     time.Time
	}{
		{10, time.Date(2026, 9, 25, 11, 56, 0, 0, recifeLoc)},
		{11, time.Date(2026, 9, 25, 12, 0, 0, 0, recifeLoc)},
		{12, time.Date(2026, 9, 25, 12, 0, 0, 0, recifeLoc)},
		{20, time.Date(2026, 9, 25, 12, 0, 0, 0, recifeLoc)},
		{22, time.Date(2026, 9, 25, 13, 46, 0, 0, recifeLoc)},
		{23, time.Date(2026, 9, 26, 18, 53, 0, 0, recifeLoc)},
	}

	for _, item := range items2026 {
		doc := &domain.Document{
			ID:             uuid.New(),
			Type:           domain.TypeNotice,
			Order:          item.order,
			Description:    fmt.Sprintf("Oficio %d", item.order),
			CreatorID:      s.user.ID,
			MunicipalityID: s.mun.ID,
			CreatedAt:      item.t,
		}
		s.Require().NoError(s.repo.CreateWithNextOrder(doc, &yr2026))
	}

	// Documento de 2025 com número mais alto (nº 50) para verificar ordenação inter-anual
	doc2025 := &domain.Document{
		ID:             uuid.New(),
		Type:           domain.TypeNotice,
		Order:          50,
		Description:    "Oficio 50 de 2025",
		CreatorID:      s.user.ID,
		MunicipalityID: s.mun.ID,
		CreatedAt:      time.Date(2025, 12, 30, 10, 0, 0, 0, recifeLoc),
	}
	s.Require().NoError(s.repo.CreateWithNextOrder(doc2025, &yr2025))

	tNotice := domain.TypeNotice

	// 1. Consulta com filtro de ano (2026): deve listar estritamente [23, 22, 20, 12, 11, 10]
	docs, total, err := s.repo.FindAll(domain.DocumentFilter{
		Type: &tNotice,
		Year: &yr2026,
	}, 1, 10)
	s.Require().NoError(err)
	s.Equal(int64(6), total)

	orders := make([]int, len(docs))
	for i, d := range docs {
		orders[i] = d.Order
	}
	s.Equal([]int{23, 22, 20, 12, 11, 10}, orders)

	// 2. Consulta sem filtro de ano (todos os anos): ano mais recente primeiro (2026), depois 2025
	docsAll, totalAll, err := s.repo.FindAll(domain.DocumentFilter{
		Type: &tNotice,
	}, 1, 10)
	s.Require().NoError(err)
	s.Equal(int64(7), totalAll)

	ordersAll := make([]int, len(docsAll))
	for i, d := range docsAll {
		ordersAll[i] = d.Order
	}
	// 2026 (23, 22, 20, 12, 11, 10) seguido de 2025 (50)
	s.Equal([]int{23, 22, 20, 12, 11, 10, 50}, ordersAll)
}

