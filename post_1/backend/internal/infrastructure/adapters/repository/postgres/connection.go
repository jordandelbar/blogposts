package postgres_adapter

import (
	"context"
	"database/sql"
	"fmt"
	"personal_website/config"
	"personal_website/pkg/retry"
	"time"
)

func NewConnection(cfg *config.PostgresConfig) (*sql.DB, error) {
	retrier := retry.New(cfg.MaxRetries, cfg.RetryDelay, cfg.MaxRetryDelay)

	var db *sql.DB

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	err := retrier.Do(ctx, "database_connection", func() error {
		conn, err := sql.Open("postgres", cfg.DSN())
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pingCancel()

		if err := conn.PingContext(pingCtx); err != nil {
			conn.Close()
			return fmt.Errorf("failed to ping database: %w", err)
		}

		db = conn
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}
