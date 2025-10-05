package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"personal_website/config"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer struct {
	server *http.Server
	logger *slog.Logger
	config *config.Config
}

func NewMetricsServer(logger *slog.Logger, cfg *config.Config) *MetricsServer {
	r := chi.NewRouter()

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Health check for the metrics server
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.MetricsPort),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	return &MetricsServer{
		server: srv,
		logger: logger,
		config: cfg,
	}
}

func (s *MetricsServer) Serve(ctx context.Context) error {
	s.logger.Info("Starting metrics server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *MetricsServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down metrics server")
	return s.server.Shutdown(ctx)
}