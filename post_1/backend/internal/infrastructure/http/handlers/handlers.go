package handlers

import (
	"log/slog"
	"personal_website/config"
	"personal_website/internal/app/core/ports"
	"personal_website/pkg/telemetry"
	"personal_website/pkg/utils"
)

type Handler struct {
	config         *config.AppConfig
	logger         *slog.Logger
	datastore      ports.Datastore
	emailService   ports.EmailService
	resumeService  ports.ResumeService
	userService    ports.UserService
	errorResponder *utils.ErrorResponder
	telemetry      *telemetry.Telemetry
}

func NewHandler(
	cfg *config.AppConfig,
	logger *slog.Logger,
	datastore ports.Datastore,
	emailService ports.EmailService,
	resumeService ports.ResumeService,
	userService ports.UserService,
	errorResponder *utils.ErrorResponder,
	telemetry *telemetry.Telemetry,
) *Handler {
	return &Handler{
		config:         cfg,
		logger:         logger,
		datastore:      datastore,
		emailService:   emailService,
		resumeService:  resumeService,
		userService:    userService,
		errorResponder: errorResponder,
		telemetry:      telemetry,
	}
}
