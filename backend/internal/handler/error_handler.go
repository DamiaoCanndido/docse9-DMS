package handler

import (
	"errors"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// handleDomainError mapeia os erros padrão de domínio para o status HTTP correspondente.
// Retorna true se o erro foi reconhecido e tratado, ou false caso contrário.
func handleDomainError(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}

	switch {
	case errors.Is(err, domain.ErrDocumentNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists) ||
		errors.Is(err, domain.ErrUsernameAlreadyExists) ||
		errors.Is(err, service.ErrMunicipalityNameConflict):
		response.Conflict(c, err.Error())
	case errors.Is(err, domain.ErrInvalidDocumentType) ||
		errors.Is(err, domain.ErrInvalidContractType) ||
		errors.Is(err, domain.ErrContractFieldsRequired) ||
		errors.Is(err, domain.ErrIncorrectCurrentPassword) ||
		errors.Is(err, domain.ErrInvalidUsername) ||
		errors.Is(err, domain.ErrInvalidEmail) ||
		errors.Is(err, service.ErrInvalidUF):
		response.BadRequest(c, err.Error())
	default:
		return false
	}
	return true
}

// handleDocumentError trata erros específicos de operações com documentos.
func handleDocumentError(c *gin.Context, err error) {
	if handleDomainError(c, err) {
		return
	}

	switch {
	case errors.Is(err, service.ErrMunicipalityNotFound) || errors.Is(err, domain.ErrUserNotFound):
		// Foreign keys inválidas na criação de documentos
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c)
	}
}

// handleUserError trata erros específicos de operações com usuários.
func handleUserError(c *gin.Context, err error) {
	if handleDomainError(c, err) {
		return
	}

	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrMunicipalityNotFound):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c)
	}
}

// handleMunicipalityError trata erros específicos de operações com municípios.
func handleMunicipalityError(c *gin.Context, err error) {
	if handleDomainError(c, err) {
		return
	}

	switch {
	case errors.Is(err, service.ErrMunicipalityNotFound):
		response.NotFound(c, err.Error())
	default:
		response.InternalError(c)
	}
}

// handleAuthError trata erros específicos de autenticação.
func handleAuthError(c *gin.Context, err error) {
	if handleDomainError(c, err) {
		return
	}
	response.InternalError(c)
}
