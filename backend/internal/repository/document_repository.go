package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository cria uma nova instância do repositório de documentos.
func NewDocumentRepository(db *gorm.DB) domain.DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(d *domain.Document) error {
	return r.db.Create(d).Error
}

func (r *documentRepository) CreateWithNextOrder(d *domain.Document, year *int) error {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var lockKey string
		if d.ContractType != nil {
			lockKey = fmt.Sprintf("order:%s:%s:%s:%v", d.MunicipalityID, d.Type, *d.ContractType, year)
		} else {
			lockKey = fmt.Sprintf("order:%s:%s:%v", d.MunicipalityID, d.Type, year)
		}

		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
				return err
			}
		}

		if d.Order > 0 {
			// Lançamento com número manual informado
			var count int64
			dupQuery := tx.Unscoped().Model(&domain.Document{}).
				Where("municipality_id = ? AND type = ? AND documents.order = ?", d.MunicipalityID, d.Type, d.Order)

			if d.Type == domain.TypeContract && d.ContractType != nil {
				dupQuery = dupQuery.Where("contract_type = ?", *d.ContractType)
			}

			if year != nil {
				dupQuery = dupQuery.Where("EXTRACT(YEAR FROM created_at AT TIME ZONE 'America/Recife') = ?", *year)
			}

			if err := dupQuery.Count(&count).Error; err != nil {
				return err
			}

			if count > 0 {
				return fmt.Errorf("%w: o número %d já está cadastrado para este tipo e ano", domain.ErrOrderAlreadyExists, d.Order)
			}
		} else {
			// Geração automática de número sequencial
			var lastOrder int
			query := tx.Unscoped().Model(&domain.Document{}).
				Select("COALESCE(MAX(documents.order), 0)").
				Where("municipality_id = ? AND type = ?", d.MunicipalityID, d.Type)

			if d.Type == domain.TypeContract && d.ContractType != nil {
				query = query.Where("contract_type = ?", *d.ContractType)
			}

			if year != nil {
				query = query.Where("EXTRACT(YEAR FROM created_at AT TIME ZONE 'America/Recife') = ?", *year)
			}

			if err := query.Row().Scan(&lastOrder); err != nil {
				return err
			}

			// Busca marco inicial (initial_order) configurado em sequence_offsets
			initialOrder := 1
			offsetQuery := tx.Model(&domain.SequenceOffset{}).
				Select("initial_order").
				Where("municipality_id = ? AND type = ?", d.MunicipalityID, d.Type)

			if d.Type == domain.TypeContract && d.ContractType != nil {
				offsetQuery = offsetQuery.Where("contract_type = ?", *d.ContractType)
			} else {
				offsetQuery = offsetQuery.Where("contract_type IS NULL")
			}

			if year != nil {
				offsetQuery = offsetQuery.Where("year = ?", *year)
			} else {
				offsetQuery = offsetQuery.Where("year IS NULL")
			}

			var foundInitial int
			if err := offsetQuery.Row().Scan(&foundInitial); err == nil && foundInitial >= 1 {
				initialOrder = foundInitial
			}

			nextOrder := lastOrder
			if initialOrder-1 > nextOrder {
				nextOrder = initialOrder - 1
			}
			d.Order = nextOrder + 1
		}

		// Validação de consistência cronológica estrita da numeração
		if err := validateChronology(tx, d, year); err != nil {
			return err
		}

		return tx.Create(d).Error
	})
}

func (r *documentRepository) FindAll(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	var (
		documents []domain.Document
		total     int64
	)

	offset := (page - 1) * pageSize
	query := r.db.Model(&domain.Document{})

	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	yearExpr := "EXTRACT(YEAR FROM created_at AT TIME ZONE 'America/Recife')"
	if r.db.Dialector.Name() != "postgres" {
		yearExpr = "CAST(strftime('%Y', created_at) AS INTEGER)"
	}
	orderClause := fmt.Sprintf("%s DESC, documents.order DESC, created_at DESC", yearExpr)
	if filter.Type != nil && *filter.Type == domain.TypeLaw {
		orderClause = "documents.order DESC, created_at DESC"
	}

	if err := query.Preload("CreatedBy").Preload("CreatedBy.Municipality").Preload("Municipality").
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (r *documentRepository) FindDeleted(filter domain.DocumentFilter, page, pageSize int) ([]domain.Document, int64, error) {
	var (
		documents []domain.Document
		total     int64
	)

	offset := (page - 1) * pageSize
	query := r.db.Unscoped().
		Model(&domain.Document{}).
		Where("deleted_at IS NOT NULL")

	query = applyFilters(query, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("CreatedBy").Preload("CreatedBy.Municipality").Preload("Municipality").
		Order("deleted_at DESC, documents.order DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (r *documentRepository) FindByID(id uuid.UUID) (*domain.Document, error) {
	var d domain.Document
	err := r.db.Preload("CreatedBy").Preload("CreatedBy.Municipality").Preload("Municipality").First(&d, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func (r *documentRepository) FindByIDUnscoped(id uuid.UUID) (*domain.Document, error) {
	var d domain.Document
	err := r.db.Unscoped().Preload("CreatedBy").Preload("CreatedBy.Municipality").Preload("Municipality").First(&d, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func (r *documentRepository) Update(d *domain.Document) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var year *int
		if d.Type != domain.TypeLaw {
			recifeLoc, err := time.LoadLocation("America/Recife")
			if err != nil {
				recifeLoc = time.FixedZone("BRT", -3*3600)
			}
			y := d.CreatedAt.In(recifeLoc).Year()
			year = &y
		}

		if err := validateChronology(tx, d, year); err != nil {
			return err
		}

		return tx.Omit("CreatedBy", "Municipality").Save(d).Error
	})
}

func formatChronoDate(t time.Time, loc *time.Location) string {
	inLoc := t.In(loc)
	if inLoc.Hour() == 0 && inLoc.Minute() == 0 && inLoc.Second() == 0 {
		return inLoc.Format("02/01/2006")
	}
	return inLoc.Format("02/01/2006 15:04")
}

func validateChronology(tx *gorm.DB, d *domain.Document, year *int) error {
	recifeLoc, err := time.LoadLocation("America/Recife")
	if err != nil {
		recifeLoc = time.FixedZone("BRT", -3*3600)
	}

	docTimeTrunc := d.CreatedAt.Truncate(time.Minute)
	docMinuteStr := d.CreatedAt.In(recifeLoc).Format("2006-01-02 15:04:00")
	docDateFormatted := formatChronoDate(d.CreatedAt, recifeLoc)

	type chronoRecord struct {
		Order     int
		CreatedAt time.Time
	}

	yearExpr := "EXTRACT(YEAR FROM created_at AT TIME ZONE 'America/Recife')"
	if tx.Dialector.Name() != "postgres" {
		yearExpr = "CAST(strftime('%Y', created_at) AS INTEGER)"
	}

	var typeName string
	switch d.Type {
	case domain.TypeLaw:
		typeName = "a Lei"
	case domain.TypeDecree:
		typeName = "o Decreto"
	case domain.TypeOrdinance:
		typeName = "a Portaria"
	case domain.TypeNotice:
		typeName = "o Ofício"
	case domain.TypeContract:
		typeName = "o Contrato"
	default:
		typeName = "o ato"
	}

	// 1. Verificar se existe ato com número menor e data/hora estritamente posterior
	var precedingConflict chronoRecord
	precedingQuery := tx.Unscoped().Model(&domain.Document{}).
		Select("documents.order, created_at").
		Where("municipality_id = ? AND type = ?", d.MunicipalityID, d.Type).
		Where("documents.order < ?", d.Order)

	if tx.Dialector.Name() == "postgres" {
		precedingQuery = precedingQuery.Where("date_trunc('minute', created_at) > ?", docTimeTrunc)
	} else {
		precedingQuery = precedingQuery.Where("strftime('%Y-%m-%d %H:%M:00', created_at) > ?", docMinuteStr)
	}

	if d.Type == domain.TypeContract {
		if d.ContractType != nil {
			precedingQuery = precedingQuery.Where("contract_type = ?", *d.ContractType)
		} else {
			precedingQuery = precedingQuery.Where("contract_type IS NULL")
		}
	}
	if year != nil {
		precedingQuery = precedingQuery.Where(fmt.Sprintf("%s = ?", yearExpr), *year)
	}
	if d.ID != uuid.Nil {
		precedingQuery = precedingQuery.Where("id != ?", d.ID)
	}

	if err := precedingQuery.Order("created_at DESC, documents.order DESC").Limit(1).Scan(&precedingConflict).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if precedingConflict.Order > 0 {
		precDate := formatChronoDate(precedingConflict.CreatedAt, recifeLoc)
		return fmt.Errorf("%w: existe %s nº %d com data posterior (%s). Para a data informada (%s), o número deve ser inferior a %d",
			domain.ErrChronologicalOrderInvalid, typeName, precedingConflict.Order, precDate, docDateFormatted, precedingConflict.Order)
	}

	// 2. Verificar se existe ato com número maior e data/hora estritamente anterior
	var succeedingConflict chronoRecord
	succeedingQuery := tx.Unscoped().Model(&domain.Document{}).
		Select("documents.order, created_at").
		Where("municipality_id = ? AND type = ?", d.MunicipalityID, d.Type).
		Where("documents.order > ?", d.Order)

	if tx.Dialector.Name() == "postgres" {
		succeedingQuery = succeedingQuery.Where("date_trunc('minute', created_at) < ?", docTimeTrunc)
	} else {
		succeedingQuery = succeedingQuery.Where("strftime('%Y-%m-%d %H:%M:00', created_at) < ?", docMinuteStr)
	}

	if d.Type == domain.TypeContract {
		if d.ContractType != nil {
			succeedingQuery = succeedingQuery.Where("contract_type = ?", *d.ContractType)
		} else {
			succeedingQuery = succeedingQuery.Where("contract_type IS NULL")
		}
	}
	if year != nil {
		succeedingQuery = succeedingQuery.Where(fmt.Sprintf("%s = ?", yearExpr), *year)
	}
	if d.ID != uuid.Nil {
		succeedingQuery = succeedingQuery.Where("id != ?", d.ID)
	}

	if err := succeedingQuery.Order("created_at ASC, documents.order ASC").Limit(1).Scan(&succeedingConflict).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if succeedingConflict.Order > 0 {
		succDate := formatChronoDate(succeedingConflict.CreatedAt, recifeLoc)
		return fmt.Errorf("%w: existe %s nº %d com data anterior (%s). Para a data informada (%s), o número deve ser superior a %d",
			domain.ErrChronologicalOrderInvalid, typeName, succeedingConflict.Order, succDate, docDateFormatted, succeedingConflict.Order)
	}

	return nil
}

func (r *documentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Document{}, "id = ?", id).Error
}

func (r *documentRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().
		Model(&domain.Document{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *documentRepository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&domain.Document{}, "id = ?", id).Error
}

func (r *documentRepository) GetLastOrder(municipalityID uuid.UUID, docType domain.DocumentType, contractType *domain.ContractType, year *int) (int, error) {
	var lastOrder int
	query := r.db.Unscoped().Model(&domain.Document{}).
		Select("COALESCE(MAX(documents.order), 0)").
		Where("municipality_id = ? AND type = ?", municipalityID, docType)

	if docType == domain.TypeContract && contractType != nil {
		query = query.Where("contract_type = ?", *contractType)
	}

	if year != nil {
		query = query.Where("EXTRACT(YEAR FROM created_at AT TIME ZONE 'America/Recife') = ?", *year)
	}

	err := query.Row().Scan(&lastOrder)
	return lastOrder, err
}

// applyFilters aplica filtros de busca à query.
func applyFilters(query *gorm.DB, filter domain.DocumentFilter) *gorm.DB {
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if len(filter.AllowedTypes) > 0 {
		query = query.Where("type IN ?", filter.AllowedTypes)
	}
	if filter.MunicipalityID != nil {
		query = query.Where("municipality_id = ?", *filter.MunicipalityID)
	}
	if filter.CreatorID != nil {
		query = query.Where("creator_id = ?", *filter.CreatorID)
	}
	if filter.ContractType != nil {
		query = query.Where("contract_type = ?", *filter.ContractType)
	}
	if filter.Year != nil {
		query = query.Where("EXTRACT(YEAR FROM created_at AT TIME ZONE 'America/Recife') = ?", *filter.Year)
	}
	if filter.Search != "" {
		query = query.Where("LOWER(description) LIKE LOWER(?)", "%"+filter.Search+"%")
	}
	return query
}
