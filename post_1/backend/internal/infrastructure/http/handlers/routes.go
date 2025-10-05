package handlers

import (
	"net/http"

	_ "personal_website/internal/infrastructure/http/docs"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(h.recoverPanic)
	r.Use(h.correlationID)
	r.Use(h.requestLogger)
	r.Use(h.enableCORS)
	r.Use(h.rateLimit)
	r.Use(h.metricsMiddleware(h.telemetry))

	r.NotFound(h.NotFoundResponse)
	r.MethodNotAllowed(h.MethodNotAllowedResponse)

	r.Get("/health", h.HealthcheckHandler)
	r.Get("/docs/*", httpSwagger.WrapHandler)

	r.Route("/v1", func(r chi.Router) {
		h.registerV1Routes(r)
	})

	return r
}

func (h *Handler) registerV1Routes(r chi.Router) {
	// Public endpoints
	r.Post("/contact", h.ContactHandler)
	r.Get("/resume", h.ResumeHandler)

	// Public article endpoints
	h.registerPublicArticleRoutes(r)

	// User registration and authentication
	h.registerAuthRoutes(r)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(h.authenticate)
		h.registerProtectedArticleRoutes(r)
		h.registerProtectedUserRoutes(r)
	})
}

func (h *Handler) registerPublicArticleRoutes(r chi.Router) {
	r.Get("/articles", h.ListArticles)
	r.Get("/articles/slug/{slug}", h.GetArticleBySlug)
}

func (h *Handler) registerProtectedArticleRoutes(r chi.Router) {
	r.With(h.requirePermissionMiddleware("articles:write")).Post("/articles", h.CreateArticle)
	r.With(h.requirePermissionMiddleware("articles:read")).Get("/articles/all", h.ListAllArticles)
	r.With(h.requirePermissionMiddleware("articles:read")).Get("/articles/id/preview/{id}", h.GetArticleById)
	r.With(h.requirePermissionMiddleware("articles:read")).Get("/articles/id/edit/{id}", h.GetArticleForEdit)
	r.With(h.requirePermissionMiddleware("articles:write")).Put("/articles/id/{id}", h.UpdateArticle)
	r.With(h.requirePermissionMiddleware("articles:write")).Patch("/articles/id/{id}/publish", h.PublishArticle)
	r.With(h.requirePermissionMiddleware("articles:write")).Patch("/articles/id/{id}/unpublish", h.UnpublishArticle)
	r.With(h.requirePermissionMiddleware("articles:write")).Delete("/articles/id/{id}", h.SoftDeleteArticle)
	r.With(h.requirePermissionMiddleware("articles:write")).Delete("/articles/id/{id}/permanent", h.DeleteArticle)
	r.With(h.requirePermissionMiddleware("articles:write")).Post("/articles/id/{id}/restore", h.RestoreArticle)
	r.With(h.requirePermissionMiddleware("articles:read")).Get("/articles/trash", h.ListDeletedArticles)
}

func (h *Handler) registerAuthRoutes(r chi.Router) {
	// User registration and activation
	r.Post("/users", h.RegisterUser)
	r.Patch("/users/activate", h.ActivateUser)

	// Authentication token creation, refresh, logout, and status
	r.Post("/auth/login", h.AuthenticationToken)
	r.Post("/auth/refresh", h.RefreshToken)
	r.Post("/auth/logout", h.LogoutToken)
	r.Get("/auth/status", h.AuthStatus)
}

func (h *Handler) registerProtectedUserRoutes(r chi.Router) {
	// User account management
	r.Patch("/users/deactivate", h.DeactivateUser)
	r.Delete("/users", h.DeleteUser)
}
