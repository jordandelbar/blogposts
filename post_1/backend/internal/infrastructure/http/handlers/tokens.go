package handlers

import (
	"net/http"
	"personal_website/internal/app/core/domain"
	"personal_website/internal/infrastructure/http/dto"
	"personal_website/pkg/utils"
	"strings"
	"time"
)

// CreateAuthenticationToken godoc
// @Summary Create authentication token
// @Description Authenticate user with email and password to get an authentication token
// @Tags authentication
// @Accept json
// @Produce json
// @Param auth body dto.AuthRequest true "Authentication credentials"
// @Success 201 {object} dto.AuthResponse "Authentication token created successfully"
// @Failure 400 {object} string "Invalid request data"
// @Failure 404 {object} string "User not found"
// @Failure 401 {object} string "Invalid credentials"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/auth/token [post]
func (h *Handler) AuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var authRequest dto.AuthRequest

	err := utils.ReadJSON(w, r, &authRequest)
	if err != nil {
		h.errorResponder.BadRequestResponse(w, r, err)
		return
	}

	if !h.validateDTO(w, r, authRequest, "authentication request") {
		return
	}

	ctx := r.Context()
	user, err := h.datastore.UserRepo().GetUserByEmail(ctx, authRequest.Email)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	match, err := user.Password.Matches(authRequest.Password)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
		return
	}

	if !match {
		h.HandleDomainError(w, r, domain.ErrInvalidCredentials)
		return
	}

	// Generate short-lived access token and long-lived refresh token
	accessToken := domain.GenerateAccessToken(user.ID)
	refreshToken := domain.GenerateRefreshToken(user.ID)

	// Get user permissions for the session
	permissions, err := h.datastore.PermissionRepo().GetPermissions(ctx, &user)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
		return
	}

	session := &domain.Session{
		UserID:      user.ID,
		Email:       user.Email,
		Permissions: permissions,
		Activated:   user.Activated,
	}

	// Clean up any existing sessions for this user to ensure fresh state
	err = h.datastore.SessionRepo().DeleteAllSessionsForUser(ctx, user.ID, domain.ScopeAuthentication)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
		return
	}

	err = h.datastore.SessionRepo().DeleteAllSessionsForUser(ctx, user.ID, domain.ScopeRefresh)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
		return
	}

	// Store both access and refresh token sessions
	err = h.datastore.SessionRepo().StoreSession(ctx, accessToken.Plaintext, domain.ScopeAuthentication, session)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	err = h.datastore.SessionRepo().StoreSession(ctx, refreshToken.Plaintext, domain.ScopeRefresh, session)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	// Set HTTP-only cookie with the refresh token
	http.SetCookie(w, &http.Cookie{
		Name:     "cms_refresh_token",
		Value:    refreshToken.Plaintext,
		Expires:  refreshToken.Expiry,
		HttpOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	err = utils.WriteJSON(w, http.StatusCreated, dto.AuthResponse{
		Success:     true,
		AccessToken: accessToken.Plaintext,
		Expiry:      accessToken.Expiry,
	})
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}

// LogoutToken godoc
// @Summary Logout user
// @Description Clear authentication cookie and invalidate session
// @Tags authentication
// @Accept json
// @Produce json
// @Success 200 {object} dto.AuthResponse "Successfully logged out"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/auth/logout [post]
func (h *Handler) LogoutToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get refresh token from cookie to identify the user
	refreshCookie, err := r.Cookie("cms_refresh_token")
	if err == nil && refreshCookie.Value != "" {
		// Get the session to find the user ID
		session, err := h.datastore.SessionRepo().GetSession(ctx, refreshCookie.Value, domain.ScopeRefresh)
		if err == nil {
			// Delete ALL access tokens for this user
			err = h.datastore.SessionRepo().DeleteAllSessionsForUser(ctx, session.UserID, domain.ScopeAuthentication)
			if err != nil {
				h.errorResponder.ServerErrorResponse(w, r, err)
				return
			}

			// Delete ALL refresh tokens for this user
			err = h.datastore.SessionRepo().DeleteAllSessionsForUser(ctx, session.UserID, domain.ScopeRefresh)
			if err != nil {
				h.errorResponder.ServerErrorResponse(w, r, err)
				return
			}
		}
	}

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "cms_refresh_token",
		Value:    "",
		Expires:  time.Unix(0, 0), // Set to past time to delete cookie
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	err = utils.WriteJSON(w, http.StatusOK, dto.AuthResponse{Success: true, AccessToken: "", Expiry: time.Unix(0, 0)})
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}

// AuthStatus godoc
// @Summary Check authentication status
// @Description Check if the user is authenticated via HTTP-only cookie
// @Tags authentication
// @Accept json
// @Produce json
// @Success 200 {object} dto.AuthStatusResponse "Authentication status"
// @Failure 401 {object} string "Not authenticated"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/auth/status [get]
func (h *Handler) AuthStatus(w http.ResponseWriter, r *http.Request) {
	var token string

	// First, try to get token from Authorization header
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader != "" {
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) == 2 && headerParts[0] == "Bearer" {
			token = headerParts[1]
		}
	}

	// If no Bearer token, try to get refresh token from cookie for fallback
	if token == "" {
		cookie, err := r.Cookie("cms_refresh_token")
		if err != nil || cookie.Value == "" {
			err = utils.WriteJSON(w, http.StatusOK, dto.AuthStatusResponse{Authenticated: false})
			if err != nil {
				h.errorResponder.ServerErrorResponse(w, r, err)
			}
			return
		}
		token = cookie.Value
	}

	// Check if session exists and is valid
	ctx := r.Context()
	var session *domain.Session
	var err error

	// Try access token first
	if authorizationHeader != "" {
		session, err = h.datastore.SessionRepo().GetSession(ctx, token, domain.ScopeAuthentication)
	} else {
		// Fallback to refresh token scope
		session, err = h.datastore.SessionRepo().GetSession(ctx, token, domain.ScopeRefresh)
	}

	if err != nil {
		err = utils.WriteJSON(w, http.StatusOK, dto.AuthStatusResponse{Authenticated: false})
		if err != nil {
			h.errorResponder.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Set appropriate expiry based on token type
	var expiry time.Time
	if authorizationHeader != "" {
		// Access token - short expiry
		expiry = time.Now().Add(10 * time.Minute)
	} else {
		// Refresh token - longer expiry
		expiry = time.Now().Add(4 * 24 * time.Hour)
	}

	err = utils.WriteJSON(w, http.StatusOK, dto.AuthStatusResponse{
		Authenticated: true,
		Expiry:        &expiry,
		UserEmail:     &session.Email,
	})
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Use refresh token from HTTP-only cookie to get new access token
// @Tags authentication
// @Accept json
// @Produce json
// @Success 200 {object} dto.RefreshTokenResponse "New access token generated"
// @Failure 401 {object} string "Invalid or expired refresh token"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/auth/refresh [post]
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	refreshCookie, err := r.Cookie("cms_refresh_token")
	if err != nil || refreshCookie.Value == "" {
		h.HandleDomainError(w, r, domain.ErrInvalidAuthToken)
		return
	}

	// Validate and get session using refresh token
	ctx := r.Context()
	session, err := h.datastore.SessionRepo().GetSession(ctx, refreshCookie.Value, domain.ScopeRefresh)
	if err != nil {
		h.HandleDomainError(w, r, domain.ErrInvalidAuthToken)
		return
	}

	// Delete all existing access tokens for this user to ensure only one active token
	err = h.datastore.SessionRepo().DeleteAllSessionsForUser(ctx, session.UserID, domain.ScopeAuthentication)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
		return
	}

	// Generate new access token
	accessToken := domain.GenerateAccessToken(session.UserID)

	// Store new access token session
	err = h.datastore.SessionRepo().StoreSession(ctx, accessToken.Plaintext, domain.ScopeAuthentication, session)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, dto.RefreshTokenResponse{
		AccessToken: accessToken.Plaintext,
		Expiry:      accessToken.Expiry,
	})
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}
