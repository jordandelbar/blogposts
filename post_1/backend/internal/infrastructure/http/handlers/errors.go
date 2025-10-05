package handlers

import (
	"errors"
	"net/http"
	"personal_website/internal/app/core/domain"
)

func (h *Handler) mapDomainErrorToHttp(err domain.DomainError) int {
	switch err.Type {
	case domain.ErrorTypeValidation:
		return http.StatusUnprocessableEntity
	case domain.ErrorTypeNotFound:
		return http.StatusNotFound
	case domain.ErrorTypeConflict:
		return http.StatusConflict
	case domain.ErrorTypeAuth:
		return http.StatusUnauthorized
	case domain.ErrorTypeInternal:
		h.logger.Error("Internal domain error", "code", err.Code, "msg", err.Message, "underlying", err.Underlying)
		return http.StatusInternalServerError
	default:
		h.logger.Error("BUG: unmapped domain error type",
			"type", string(err.Type),
			"code", err.Code,
			"msg", err.Message,
			"underlying", err.Underlying,
		)
		return http.StatusInternalServerError
	}
}

func (h *Handler) RespondError(w http.ResponseWriter, r *http.Request, err domain.DomainError) {
	status := h.mapDomainErrorToHttp(err)

	clientMsg := "internal server error"

	if status >= 400 && status < 500 {
		clientMsg = err.Message
	}

	h.errorResponder.ErrorResponse(w, r, status, clientMsg)
}

func (h *Handler) HandleDomainError(w http.ResponseWriter, r *http.Request, err error) {
	var domainErr domain.DomainError
	if errors.As(err, &domainErr) {
		h.RespondError(w, r, domainErr)
	} else {
		h.RespondError(w, r, domain.NewInternalError(err))
	}
}
