package handlers

import (
	"fmt"
	"net/http"
	"personal_website/internal/infrastructure/dto_validation"
	"strconv"
)

func (h *Handler) validateDTO(w http.ResponseWriter, r *http.Request, dto any, context string) bool {
	validator := dto_validation.NewDtoValidator()
	if !validator.ValidateStruct(dto).Valid() {
		h.logger.Warn("Validation failed", "context", context, "errors", validator.Errors)
		h.errorResponder.DtoValidationErrorResponse(w, r, validator.Errors)
		return false
	}
	return true
}

func (h *Handler) extractIDParam(w http.ResponseWriter, r *http.Request) (int32, bool) {
	idStr := r.PathValue("id")
	if idStr == "" {
		h.errorResponder.BadRequestResponse(w, r, fmt.Errorf("id parameter is required"))
		return 0, false
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponder.BadRequestResponse(w, r, fmt.Errorf("invalid id parameter"))
		return 0, false
	}

	return int32(id), true
}

func (h *Handler) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error":"not found"}`))
}

func (h *Handler) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"error":"method not allowed"}`))
}
