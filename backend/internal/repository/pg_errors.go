package repository

import (
	"errors"
	"strings"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgErrUniqueViolation = "23505"

// translatePgError converte erros do PostgreSQL em erros de domínio.
// Tenta via pgconn.PgError primeiro; usa string matching como fallback
// para garantir compatibilidade independente da versão do driver.
func translatePgError(err error) error {
	if err == nil {
		return nil
	}

	if isUniqueViolation(err) {
		return domain.ErrNameAlreadyExists
	}

	return err
}

func isUniqueViolation(err error) bool {
	// Tentativa 1: pgx v5 nativo
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgErrUniqueViolation
	}

	// Tentativa 2: fallback por string (cobre lib/pq e wrappers do GORM)
	msg := err.Error()
	return strings.Contains(msg, "23505") ||
		strings.Contains(msg, "duplicate key value violates unique constraint")
}
