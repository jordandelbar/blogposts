package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"personal_website/config"
	"personal_website/internal/app/core/ports"
	"personal_website/internal/infrastructure/http/handlers"
	"personal_website/pkg/telemetry"
	"personal_website/pkg/utils"

	_ "github.com/lib/pq"
)

type Server struct {
	server         *http.Server
	handler        *handlers.Handler
	logger         *slog.Logger
	config         *config.Config
	datastore      ports.Datastore
	errorResponder *utils.ErrorResponder
}

func NewServer(
	logger *slog.Logger,
	cfg *config.Config,
	datastore ports.Datastore,
	resumeService ports.ResumeService,
	emailService ports.EmailService,
	userService ports.UserService,
	errorResponder *utils.ErrorResponder,
	telemetryInstance *telemetry.Telemetry,
) *Server {
	handler := handlers.NewHandler(
		&cfg.App,
		logger,
		datastore,
		emailService,
		resumeService,
		userService,
		errorResponder,
		telemetryInstance,
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      handler.Routes(),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	return &Server{
		server:         srv,
		handler:        handler,
		logger:         logger,
		datastore:      datastore,
		config:         cfg,
		errorResponder: errorResponder,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	s.logger.Info("Starting api server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}