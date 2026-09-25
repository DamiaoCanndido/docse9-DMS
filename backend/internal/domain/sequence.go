package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Erros de domínio relacionados a sequências e numeração.
var (
	ErrOrderAlreadyExists          = errors.New("o número já está cadastrado para este tipo e ano")
	ErrInvalidInitialOrder         = errors.New("initial_order deve ser maior ou igual a 1")
	ErrSequenceOffsetNotFound      = errors.New("marco inicial não encontrado")
	ErrInvalidSequenceYear         = errors.New("o ano é obrigatório para este tipo de documento")
	ErrInvalidSequenceContractType = errors.New("tipo de contrato é obrigatório para contratos")
	ErrManualOrderForbidden        = errors.New("apenas moderadores podem definir número manual de documento")
	ErrSequenceAccessForbidden     = errors.New("acesso restrito a moderadores do município")
	ErrChronologicalOrderInvalid   = errors.New("inconsistência cronológica de numeração na série documental")
)

// SequenceOffset representa a configuração de marco inicial numérico para um tipo de documento/ano.
type SequenceOffset struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MunicipalityID uuid.UUID     `gorm:"type:uuid;not null;index"                       json:"municipalityId"`
	Municipality   Municipality  `gorm:"foreignKey:MunicipalityID"                      json:"municipality,omitempty"`
	Type           DocumentType  `gorm:"type:varchar(50);not null"                      json:"type"`
	ContractType   *ContractType `gorm:"type:varchar(50)"                               json:"contractType,omitempty"`
	Year           *int          `gorm:"type:integer"                                   json:"year,omitempty"`
	InitialOrder   int           `gorm:"type:integer;not null"                          json:"initialOrder"`
	CreatedAt      time.Time     `gorm:"autoCreateTime;type:timestamptz"                json:"createdAt"`
	UpdatedAt      time.Time     `gorm:"autoUpdateTime;type:timestamptz"                json:"updatedAt"`
}

// TableName especifica o nome da tabela no PostgreSQL.
func (SequenceOffset) TableName() string {
	return "sequence_offsets"
}

// SequenceItemResponse representa o estado de uma série documental:
// marco inicial configurado, último número emitido e próximo a ser gerado.
type SequenceItemResponse struct {
	ID           *uuid.UUID    `json:"id,omitempty"`
	Type         DocumentType  `json:"type"`
	ContractType *ContractType `json:"contractType,omitempty"`
	Year         *int          `json:"year,omitempty"`
	InitialOrder int           `json:"initialOrder"`
	CurrentOrder int           `json:"currentOrder"`
	NextOrder    int           `json:"nextOrder"`
}

// SetSequenceOffsetInput DTO para configurar ou atualizar o marco inicial.
type SetSequenceOffsetInput struct {
	Type         DocumentType  `json:"type"         binding:"required,oneof=NOTICE DECREE ORDINANCE LAW CONTRACT"`
	ContractType *ContractType `json:"contractType" binding:"omitempty,oneof=publicinterest bidding service"`
	Year         *int          `json:"year"         binding:"omitempty"`
	InitialOrder int           `json:"initialOrder" binding:"required,gte=1"`
}

// SequenceRepository porta de saída para persistência de sequence_offsets.
type SequenceRepository interface {
	GetByMunicipality(municipalityID uuid.UUID, year *int) ([]SequenceOffset, error)
	FindOffset(municipalityID uuid.UUID, docType DocumentType, contractType *ContractType, year *int) (*SequenceOffset, error)
	Upsert(offset *SequenceOffset) error
}

// SequenceService porta de entrada para regras de negócio de sequências.
type SequenceService interface {
	GetSequences(municipalityID uuid.UUID, year *int) ([]SequenceItemResponse, error)
	SetOffset(municipalityID uuid.UUID, input SetSequenceOffsetInput) (*SequenceOffset, error)
}
