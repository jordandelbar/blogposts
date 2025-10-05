package handlers

import (
	"context"
	"net/http"
	"personal_website/internal/app/core/domain"
)

type contextKey string

const authenticatedUserContextKey = contextKey("authenticatedUser")

func (h *Handler) contextSetAuthenticatedSession(r *http.Request, user *domain.Session) *http.Request {
	ctx := context.WithValue(r.Context(), authenticatedUserContextKey, user)
	return r.WithContext(ctx)
}

func (h *Handler) contextGetAuthenticatedSession(r *http.Request) *domain.Session {
	user, ok := r.Context().Value(authenticatedUserContextKey).(*domain.Session)
	if !ok {
		panic("missing authenticated user value in request context")
	}

	return user
}
