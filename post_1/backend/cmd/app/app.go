package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"personal_website/config"
	"personal_website/internal/app/core/ports"
	"personal_website/internal/app/core/services/mailer"
	"personal_website/internal/app/core/services/registration"
	"personal_website/internal/infrastructure/adapters/email_sender"
	datastore_adapter "personal_website/internal/infrastructure/adapters/repository/datastore"
	postgres_adapter "personal_website/internal/infrastructure/adapters/repository/postgres"
	valkey_adapter "personal_website/internal/infrastructure/adapters/repository/valkey"
	"personal_website/internal/infrastructure/adapters/resume"
	"personal_website/internal/infrastructure/http"
	"personal_website/pkg/telemetry"
	"personal_website/pkg/utils"
	"syscall"

	_ "github.com/lib/pq"
)

func StartApp(cfg *config.Config) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("Initializing telemetry...")
	telemetryInstance, err := telemetry.NewTelemetry(logger)
	if err != nil {
		logger.Error("Failed to initialize telemetry", "error", err)
		os.Exit(1)
	}
	defer telemetryInstance.Shutdown(context.Background())

	logger.Info("Initializing database connection...")
	pgDatabase, err := postgres_adapter.NewDatabase(&cfg.Postgres)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer pgDatabase.Close()

	logger.Info("Initializing valkey connection...")
	vkDatabase, err := valkey_adapter.NewDatabase(&cfg.Valkey)
	if err != nil {
		logger.Error("Failed to initialize valkey", "error", err)
		os.Exit(1)
	}
	defer vkDatabase.Close()

	resumeService, err := resume.NewService(&cfg.Minio)
	if err != nil {
		logger.Error("Failed to initialize resume service", "error", err)
		os.Exit(1)
	}

	datastore := datastore_adapter.NewDatastore(pgDatabase, vkDatabase)

	server, err := NewServer(ServerDeps{
		Logger:        logger,
		Config:        cfg,
		Datastore:     datastore,
		EmailSender:   email_sender.NewEmailSender(),
		ResumeService: resumeService,
		Telemetry:     telemetryInstance,
	})
	if err != nil {
		logger.Error("Error when initializing server", "error", err.Error())
		os.Exit(1)
	}

	// Initialize metrics server
	metricsServer := http.NewMetricsServer(logger, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start main server
	go func() {
		logger.Info("Starting main server...")
		if err := server.Serve(ctx); err != nil {
			logger.Error("Main server stopped with error", "error", err)
		}
	}()

	// Start metrics server
	go func() {
		logger.Info("Starting metrics server...")
		if err := metricsServer.Serve(ctx); err != nil {
			logger.Error("Metrics server stopped with error", "error", err)
		}
	}()

	<-ctx.Done()

	logger.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	// Shutdown both servers
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Main server shutdown failed", "error", err)
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Metrics server shutdown failed", "error", err)
	}

	logger.Info("Servers shut down successfully.")
}

type ServerDeps struct {
	Logger        *slog.Logger
	Config        *config.Config
	Datastore     ports.Datastore
	EmailSender   ports.EmailSender
	ResumeService ports.ResumeService
	Telemetry     *telemetry.Telemetry
}

func NewServer(deps ServerDeps) (*http.Server, error) {
	errorReponder := utils.NewErrorResponder(deps.Logger)

	emailService, err := mailer.NewService(&deps.Config.SMTP, deps.EmailSender, deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("error when initializing email service: %w", err)
	}

	userService := registration.NewUserService(emailService, deps.Datastore)

	server := http.NewServer(
		deps.Logger,
		deps.Config,
		deps.Datastore,
		deps.ResumeService,
		emailService,
		userService,
		errorReponder,
		deps.Telemetry,
	)
	return server, nil
}
