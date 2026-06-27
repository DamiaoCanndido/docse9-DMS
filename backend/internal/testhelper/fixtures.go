package testhelper

import (
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/google/uuid"
)

// ─── UUIDs fixos para testes determinísticos ──────────────────────────────────

var (
	MunPassagemID  = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	MunPatosID     = uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")
	NonExistentID  = uuid.MustParse("999e8400-e29b-41d4-a716-446655440099")
)

// ─── Factories ────────────────────────────────────────────────────────────────

func MakePassagem() domain.Municipality {
	return domain.Municipality{
		ID:        MunPassagemID,
		Name:      "Passagem",
		UF:        "PB",
		ImageURL:  "https://example.com/passagem.png",
		CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}

func MakePatos() domain.Municipality {
	return domain.Municipality{
		ID:        MunPatosID,
		Name:      "Patos",
		UF:        "PB",
		CreatedAt: time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC),
	}
}

func MakeCreateInput(name, uf, imageURL string) domain.CreateMunicipalityInput {
	return domain.CreateMunicipalityInput{
		Name:     name,
		UF:       uf,
		ImageURL: imageURL,
	}
}

// StringPtr converte string para *string (útil para UpdateMunicipalityInput).
func StringPtr(s string) *string { return &s }

func MakeUserAdmin(munID uuid.UUID) domain.User {
	return domain.User{
		ID:             uuid.MustParse("770e8400-e29b-41d4-a716-446655440002"),
		Username:       "admin_user",
		Email:          "admin@example.com",
		Password:       "hashed_password_123",
		Role:           domain.RoleAdmin,
		MunicipalityID: munID,
	}
}

func MakeUserCommon(munID uuid.UUID) domain.User {
	return domain.User{
		ID:             uuid.MustParse("880e8400-e29b-41d4-a716-446655440003"),
		Username:       "common_user",
		Email:          "common@example.com",
		Password:       "hashed_password_456",
		Role:           domain.RoleCommon,
		MunicipalityID: munID,
	}
}

