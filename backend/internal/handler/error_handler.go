package handler

import (
	"errors"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// formatBindingError converte erros de validação do Gin/Validator para mensagens amigáveis em português.
func formatBindingError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			switch fe.Field() {
			case "Description":
				switch fe.Tag() {
				case "required":
					return "A descrição do documento é obrigatória"
				case "min":
					return "A descrição deve conter no mínimo 3 caracteres"
				}
			case "Type":
				return "Tipo de documento inválido"
			case "Duration":
				return "Duração do contrato é obrigatória e deve ser maior que zero"
			case "ContractType":
				return "Tipo de contrato inválido"
			case "Value":
				return "Valor do contrato é obrigatório e deve ser maior que zero"
			case "StartIn":
				return "Data de início do contrato é obrigatória"
			}
		}
	}
	return err.Error()
}

// handleDomainError mapeia os erros padrão de domínio para o status HTTP correspondente.
// Retorna true se o erro foi reconhecido e tratado, ou false caso contrário.
func handleDomainError(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}

	switch {
	case errors.Is(err, domain.ErrDocumentNotFound) ||
		errors.Is(err, domain.ErrSequenceOffsetNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, domain.ErrManualOrderForbidden) ||
		errors.Is(err, domain.ErrSequenceAccessForbidden):
		response.Forbidden(c, err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists) ||
		errors.Is(err, domain.ErrUsernameAlreadyExists) ||
		errors.Is(err, domain.ErrOrderAlreadyExists) ||
		errors.Is(err, domain.ErrChronologicalOrderInvalid) ||
		errors.Is(err, service.ErrMunicipalityNameConflict):
		response.Conflict(c, err.Error())
	case errors.Is(err, domain.ErrInvalidDocumentType) ||
		errors.Is(err, domain.ErrInvalidContractType) ||
		errors.Is(err, domain.ErrContractFieldsRequired) ||
		errors.Is(err, domain.ErrIncorrectCurrentPassword) ||
		errors.Is(err, domain.ErrInvalidUsername) ||
		errors.Is(err, domain.ErrInvalidEmail) ||
		errors.Is(err, domain.ErrInvalidInitialOrder) ||
		errors.Is(err, domain.ErrInvalidSequenceYear) ||
		errors.Is(err, domain.ErrInvalidSequenceContractType) ||
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
	case errors.Is(err, domain.ErrFileNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrInvalidContentType) || errors.Is(err, domain.ErrFileTooLarge) || errors.Is(err, domain.ErrInvalidFileKey):
		response.BadRequest(c, err.Error())
	case errors.Is(err, domain.ErrPDFMissingOCR):
		response.Error(c, 422, err.Error()) // 422 Unprocessable Entity
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
