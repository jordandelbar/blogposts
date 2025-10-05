package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"personal_website/internal/app/core/domain"
	domain_validation "personal_website/internal/app/core/validation"
	"personal_website/pkg/telemetry"
	"personal_website/pkg/utils"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/time/rate"
)

func (h *Handler) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")

		origin := r.Header.Get("Origin")

		if origin != "" && len(h.config.Cors.TrustedOrigins) > 0 {
			for i := range h.config.Cors.TrustedOrigins {
				if origin == h.config.Cors.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
						w.Header().Set("Access-Control-Allow-Credentials", "true")

						w.WriteHeader(http.StatusOK)
						return
					}

					// Set credentials header for all requests from trusted origins
					w.Header().Set("Access-Control-Allow-Credentials", "true")

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				h.errorResponder.ServerErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) rateLimit(next http.Handler) http.Handler {
	var ipRateLimiter *IPRateLimiter

	if h.config.Limiter.Enabled {
		ipRateLimiter = NewIPRateLimiter(rate.Limit(h.config.Limiter.Rps), h.config.Limiter.Burst)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.config.Limiter.Enabled && ipRateLimiter != nil {
			clientIP := getClientIP(r)
			limiter := ipRateLimiter.getLimiter(clientIP)

			if !limiter.Allow() {
				h.errorResponder.RateLimitExceededResponse(w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) correlationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := uuid.New().String()

		w.Header().Set("X-Correlation-ID", correlationID)

		ctx := context.WithValue(r.Context(), utils.CorrelationIDKey, correlationID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		correlationID := utils.GetCorrelationID(r.Context())

		h.logger.Info("Request started",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"correlation_id", correlationID,
		)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		h.logger.Info("Request completed",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"duration", duration.String(),
			"correlation_id", correlationID,
		)
	})
}

func (h *Handler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		var token string

		// First, try to get token from Authorization header (for API compatibility)
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader != "" {
			headerParts := strings.Split(authorizationHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				h.HandleDomainError(w, r, domain.ErrInvalidAuthToken)
				return
			}
			token = headerParts[1]
		} else {
			// If no Authorization header, try to get token from HTTP-only cookie
			cookie, err := r.Cookie("cms_auth_token")
			if err != nil || cookie.Value == "" {
				// No token found in either header or cookie
				r = h.contextSetAuthenticatedSession(r, domain.AnonymousSession)
				next.ServeHTTP(w, r)
				return
			}
			token = cookie.Value
		}

		validator := domain_validation.NewTokenValidator()
		if validator.ValidateTokenPlaintext(token); !validator.Valid() {
			h.HandleDomainError(w, r, domain.ErrInvalidAuthToken)
			return
		}

		ctx := r.Context()
		session, err := h.datastore.SessionRepo().GetSession(ctx, token, domain.ScopeAuthentication)

		if err != nil {
			switch {
			case errors.Is(err, domain.ErrSessionNotFound), errors.Is(err, domain.ErrSessionExpired):
				h.HandleDomainError(w, r, domain.ErrInvalidAuthToken)
			default:
				h.errorResponder.ServerErrorResponse(w, r, err)
			}
			return
		}

		authenticatedSession := &domain.Session{
			UserID:      session.UserID,
			Email:       session.Email,
			Permissions: session.Permissions,
			Activated:   session.Activated,
		}

		r = h.contextSetAuthenticatedSession(r, authenticatedSession)

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := h.contextGetAuthenticatedSession(r)

		if session.IsAnonymous() {
			h.HandleDomainError(w, r, domain.ErrAuthenticationRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) requireActivatedUser(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := h.contextGetAuthenticatedSession(r)

		if !session.Activated {
			h.HandleDomainError(w, r, domain.ErrInactiveAccount)
			return
		}

		next.ServeHTTP(w, r)
	})

	return h.requireAuthenticatedUser(fn)
}

func (h *Handler) requirePermission(code string, next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := h.contextGetAuthenticatedSession(r)

		if !session.Permissions.Include(code) {
			h.HandleDomainError(w, r, domain.ErrNotPermitted)
			return
		}

		next.ServeHTTP(w, r)
	})

	return h.requireActivatedUser(fn)
}

func (h *Handler) requirePermissionMiddleware(code string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return h.requirePermission(code, next)
	}
}

func (h *Handler) metricsMiddleware(t *telemetry.Telemetry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ctx := r.Context()

			t.RequestsInFlight.Add(ctx, 1)
			defer t.RequestsInFlight.Add(ctx, -1)

			rw := telemetry.NewResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			statusCode := strconv.Itoa(rw.StatusCode)

			labels := []attribute.KeyValue{
				attribute.String("method", r.Method),
				attribute.String("route", r.URL.Path),
				attribute.String("status_code", statusCode),
			}

			t.RequestsTotal.Add(ctx, 1, metric.WithAttributes(labels...))
			t.RequestDuration.Record(ctx, duration, metric.WithAttributes(labels...))
			t.ResponseSize.Record(ctx, rw.Written, metric.WithAttributes(labels...))
		})
	}
}
