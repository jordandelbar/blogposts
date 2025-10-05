package utils

import (
	"context"
	"log/slog"
	"net/http"
)

type ctxKey string

const CorrelationIDKey ctxKey = "correlation_id"

type ErrorResponder struct {
	Logger *slog.Logger
}

func NewErrorResponder(logger *slog.Logger) *ErrorResponder {
	return &ErrorResponder{logger}
}

func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

func (e *ErrorResponder) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	correlationID := GetCorrelationID(r.Context())
	e.Logger.Error(err.Error(), "method", method, "uri", uri, "correlation_id", correlationID)
}

func (e *ErrorResponder) ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := Envelope{"error": message}
	err := WriteJSON(w, status, env)
	if err != nil {
		e.logError(r, err)
	}
}

func (e *ErrorResponder) DtoValidationErrorResponse(w http.ResponseWriter, r *http.Request, errors map[string][]string) {
	env := Envelope{"error": "validation failed", "validation_errors": errors}
	err := WriteJSON(w, http.StatusBadRequest, env)
	if err != nil {
		e.logError(r, err)
	}
}

func (e *ErrorResponder) DomainValidationErrorResponse(w http.ResponseWriter, r *http.Request, errors map[string][]string) {
	env := Envelope{"error": "validation failed", "validation_errors": errors}
	err := WriteJSON(w, http.StatusUnprocessableEntity, env)
	if err != nil {
		e.logError(r, err)
	}
}

func (e *ErrorResponder) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	e.ErrorResponse(w, r, http.StatusInternalServerError, message)
}

func (e *ErrorResponder) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)
	e.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (e *ErrorResponder) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	e.logError(r, err)
	e.ErrorResponse(w, r, http.StatusNotFound, err.Error())
}

func (e *ErrorResponder) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	e.ErrorResponse(w, r, http.StatusTooManyRequests, message)
}

