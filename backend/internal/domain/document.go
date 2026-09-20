package domain

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Erros de domínio da entidade Document.
var (
	ErrDocumentNotFound       = errors.New("documento não encontrado")
	ErrInvalidDocumentType    = errors.New("tipo de documento inválido")
	ErrInvalidContractType    = errors.New("tipo de contrato inválido")
	ErrContractFieldsRequired = errors.New("campos de contrato (duration, type, value, startIn) são obrigatórios para contratos")
	ErrFileNotFound           = errors.New("documento não possui anexo")
	ErrInvalidContentType     = errors.New("somente arquivos no formato PDF são permitidos")
	ErrFileTooLarge           = errors.New("arquivo excede o tamanho máximo permitido de 25MB")
	ErrPDFMissingOCR          = errors.New("o arquivo PDF não possui camada de texto pesquisável (OCR)")
	ErrInvalidFileKey         = errors.New("fileKey inválido ou não pertence a este documento")
)

// DocumentType representa o tipo de documento.
type DocumentType string

const (
	TypeNotice    DocumentType = "NOTICE"
	TypeDecree    DocumentType = "DECREE"
	TypeOrdinance DocumentType = "ORDINANCE"
	TypeLaw       DocumentType = "LAW"
	TypeContract  DocumentType = "CONTRACT"
)

// ContractType representa o tipo de contrato.
type ContractType string

const (
	ContractPublicInterest ContractType = "publicinterest"
	ContractBidding        ContractType = "bidding"
	ContractService        ContractType = "service"
)

// Document representa um documento (metadados de arquivo) no sistema.
type Document struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Type           DocumentType   `gorm:"type:varchar(50);not null;index"                json:"type"`
	Order          int            `gorm:"type:integer;not null"                          json:"order"`
	Description    string         `gorm:"type:text;not null"                              json:"description"`
	FileKey        string         `gorm:"type:varchar(255);not null;default:''"          json:"fileKey"`
	CreatorID      uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"creatorId"`
	CreatedBy      User           `gorm:"foreignKey:CreatorID"                           json:"createdBy,omitempty"`
	MunicipalityID uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"municipalityId"`
	Municipality   Municipality   `gorm:"foreignKey:MunicipalityID"                      json:"municipality,omitempty"`

	// Campos específicos para Contratos (Type == "CONTRACT")
	Duration     *int          `gorm:"type:integer"                                   json:"duration,omitempty"`
	ContractType *ContractType `gorm:"type:varchar(50)"                                 json:"contractType,omitempty"`
	Value        *float64      `gorm:"type:numeric(15,2)"                             json:"value,omitempty"`
	StartIn      *time.Time    `gorm:"type:timestamptz"                               json:"startIn,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime;type:timestamptz"                json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;type:timestamptz"                json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"` // soft-delete
}

// ──────────────────────────────────────────────
// DTOs
// ──────────────────────────────────────────────

type DocumentFilter struct {
	Type              *DocumentType  `json:"type"           form:"type"`
	AllowedTypes      []DocumentType `json:"-"`
	MunicipalityIDRaw string         `json:"-"              form:"municipalityId"`
	MunicipalityID    *uuid.UUID     `json:"municipalityId" form:"-"`
	CreatorIDRaw      string         `json:"-"              form:"creatorId"`
	CreatorID         *uuid.UUID     `json:"creatorId"      form:"-"`
	Search            string         `json:"search"         form:"search"` // Busca textual na descrição
	Year              *int           `json:"year"           form:"year"`   // Ano do documento ou contrato
	ContractType      *ContractType  `json:"contractType"   form:"contractType"`
}

func (f *DocumentFilter) Process() error {
	if f.MunicipalityIDRaw != "" && f.MunicipalityID == nil {
		raw := strings.Trim(f.MunicipalityIDRaw, " \"[]")
		if raw != "" {
			parsed, err := uuid.Parse(raw)
			if err != nil {
				return err
			}
			f.MunicipalityID = &parsed
		}
	}
	if f.CreatorIDRaw != "" && f.CreatorID == nil {
		raw := strings.Trim(f.CreatorIDRaw, " \"[]")
		if raw != "" {
			parsed, err := uuid.Parse(raw)
			if err != nil {
				return err
			}
			f.CreatorID = &parsed
		}
	}
	return nil
}

type CreateDocumentInput struct {
	Type           DocumentType  `json:"type"           binding:"required,oneof=NOTICE DECREE ORDINANCE LAW CONTRACT"`
	Description    string        `json:"description"    binding:"required,min=3"`
	CreatorID      uuid.UUID     `json:"creatorId"`
	MunicipalityID uuid.UUID     `json:"municipalityId"`

	// Campos exclusivos para Contrato (devem ser validados no Service/Handler)
	Duration     *int          `json:"duration"     binding:"omitempty,gt=0"`
	ContractType *ContractType `json:"contractType" binding:"omitempty,oneof=publicinterest bidding service"`
	Value        *float64      `json:"value"        binding:"omitempty,gt=0"`
	StartIn      *time.Time    `json:"startIn"      binding:"omitempty"`
}

type UpdateDocumentInput struct {
	Description *string        `json:"description" binding:"omitempty,min=3"`
	CreatedAt   *time.Time     `json:"createdAt"   binding:"omitempty"`

	// Campos exclusivos para Contrato
	Duration     *int          `json:"duration"     binding:"omitempty,gt=0"`
	ContractType *ContractType `json:"contractType" binding:"omitempty,oneof=publicinterest bidding service"`
	Value        *float64      `json:"value"        binding:"omitempty,gt=0"`
	StartIn      *time.Time    `json:"startIn"      binding:"omitempty"`
}

type UploadURLInput struct {
	FileName    string `json:"fileName"    binding:"required"`
	FileSize    int64  `json:"fileSize"    binding:"required,gt=0,lte=26214400"` // máx 25MB
	ContentType string `json:"contentType" binding:"required,eq=application/pdf"`
}

type UploadURLResponse struct {
	UploadURL        string `json:"uploadUrl"`
	FileKey          string `json:"fileKey"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

type ConfirmUploadInput struct {
	FileKey string `json:"fileKey" binding:"required"`
}

type FileURLResponse struct {
	URL              string `json:"url"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

// ──────────────────────────────────────────────
// Repository interface (porta de saída)
// ──────────────────────────────────────────────

type DocumentRepository interface {
	Create(d *Document) error
	CreateWithNextOrder(d *Document, year *int) error
	FindAll(filter DocumentFilter, page, pageSize int) ([]Document, int64, error)
	FindDeleted(filter DocumentFilter, page, pageSize int) ([]Document, int64, error)
	FindByID(id uuid.UUID) (*Document, error)
	FindByIDUnscoped(id uuid.UUID) (*Document, error)
	Update(d *Document) error
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) error
	HardDelete(id uuid.UUID) error

	// GetLastOrder retorna o maior order para um tipo de documento em um município.
	// Se year for nil, ignora o ano (usado para LAWS).
	GetLastOrder(municipalityID uuid.UUID, docType DocumentType, contractType *ContractType, year *int) (int, error)
}

// ──────────────────────────────────────────────
// Service interface (porta de entrada)
// ──────────────────────────────────────────────

type DocumentService interface {
	Create(input CreateDocumentInput) (*Document, error)
	GetAll(filter DocumentFilter, page, pageSize int) ([]Document, int64, error)
	GetDeleted(filter DocumentFilter, page, pageSize int) ([]Document, int64, error)
	GetByID(id uuid.UUID) (*Document, error)
	GetByIDUnscoped(id uuid.UUID) (*Document, error)
	Update(id uuid.UUID, input UpdateDocumentInput) (*Document, error)
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) (*Document, error)
	HardDelete(id uuid.UUID) error

	// Gestão de Anexos e Storage (Cloudflare R2 / S3)
	GenerateUploadURL(ctx context.Context, docID uuid.UUID, input UploadURLInput) (*UploadURLResponse, error)
	ConfirmUpload(ctx context.Context, docID uuid.UUID, input ConfirmUploadInput) (*Document, error)
	GenerateFileURL(ctx context.Context, docID uuid.UUID, download bool) (*FileURLResponse, error)
}
