package handlers

import (
	"net/http"
	"personal_website/internal/infrastructure/http/dto"
	"personal_website/internal/infrastructure/http/mappers"
	"personal_website/pkg/utils"
)

// RegisterUser godoc
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UserRequest true "User registration data"
// @Success 202 "User registered successfully, activation required"
// @Failure 400 {object} string "Invalid request data or validation error"
// @Failure 409 {object} string ""
// @Failure 500 {object} string "Internal server error"
// @Router /v1/users [post]
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var userRequest dto.UserRequest

	err := utils.ReadJSON(w, r, &userRequest)
	if err != nil {
		h.errorResponder.BadRequestResponse(w, r, err)
		return
	}

	if !h.validateDTO(w, r, userRequest, "user registration") {
		return
	}

	user, err := mappers.UserRequestToDomain(userRequest)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
		return
	}

	ctx := r.Context()
	err = h.userService.RegisterUser(ctx, user, h.config.ActivationUrl)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// ActivateUser godoc
// @Summary Activate user account
// @Description Activate a user account using an activation token
// @Tags users
// @Accept json
// @Produce json
// @Param token query string true "Activation token"
// @Success 200 "User account activated successfully"
// @Failure 400 {object} string "Invalid or expired activation token"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/users/activate [patch]
func (h *Handler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	var activationToken dto.ActivationToken

	err := utils.ReadJSON(w, r, &activationToken)
	if err != nil {
		h.errorResponder.BadRequestResponse(w, r, err)
		return
	}

	if !h.validateDTO(w, r, activationToken, "user activation") {
		return
	}

	ctx := r.Context()
	user, err := h.userService.ActivateUser(ctx, activationToken.TokenPlaintext)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	userResponse := mappers.UserToResponse(user)

	err = utils.WriteJSON(w, http.StatusOK, userResponse)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}

// DeactivateUser godoc
// @Summary Deactivate user account
// @Description Deactivate the authenticated user's account
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 "User account deactivated successfully"
// @Failure 401 {object} string "Unauthorized - authentication required"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/users/deactivate [patch]
func (h *Handler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	session := h.contextGetAuthenticatedSession(r)
	if session == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.datastore.UserRepo().DeactivateUser(r.Context(), session.UserID)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteUser godoc
// @Summary Delete user account
// @Description Permanently delete the authenticated user's account
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 "User account deleted successfully"
// @Failure 401 {object} string "Unauthorized - authentication required"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/users [delete]
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	session := h.contextGetAuthenticatedSession(r)
	if session == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.datastore.UserRepo().DeleteUser(r.Context(), session.UserID)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
