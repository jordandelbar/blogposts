package retry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"time"

	"github.com/lib/pq"
)

type Retrier struct {
	maxRetries   int
	initialDelay time.Duration
	maxDelay     time.Duration
	logger       *slog.Logger
}

func New(maxRetries int, initialDelay, maxDelay time.Duration) *Retrier {
	return &Retrier{
		maxRetries:   maxRetries,
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
	}
}

type RetryableFunc func() error

func (r *Retrier) Do(ctx context.Context, operation string, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			delay := r.calculateDelay(attempt)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
			}
		}

		if err := fn(); err != nil {
			lastErr = err

			if !r.isRetryableError(err) {
				return err
			}

			if attempt == r.maxRetries {
				return fmt.Errorf("max retries (%d) exceeded for %s: %w", r.maxRetries, operation, err)
			}

			continue
		}

		if attempt > 0 {
		}
		return nil
	}

	return lastErr
}

func (r *Retrier) DoWithTimeout(ctx context.Context, timeout time.Duration, operation string, fn RetryableFunc) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.Do(timeoutCtx, operation, fn)
}

func (r *Retrier) calculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(r.initialDelay) * math.Pow(2, float64(attempt-1)))

	if delay > r.maxDelay {
		delay = r.maxDelay
	}

	return delay
}

func (r *Retrier) isRetryableError(err error) bool {
	// Check for network timeout errors
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// Check for network operation errors
	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return true
	}

	// PostgreSQL-specific errors
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "53300", // too_many_connections
			"53400", // configuration_limit_exceeded
			"08000", // connection_exception
			"08003", // connection_does_not_exist
			"08006", // connection_failure
			"08001", // sqlclient_unable_to_establish_sqlconnection
			"08004": // sqlserver_rejected_establishment_of_sqlconnection
			return true
		default:
			// Authentication, permission, and syntax errors are not retryable
			return false
		}
	}

	errorMsg := err.Error()
	retryableMessages := []string{
		"connection refused",
		"connection reset by peer",
		"no such host",
		"network is unreachable",
		"i/o timeout",
		"connection timeout",
		"temporary failure",
		"server is not ready",
	}

	for _, msg := range retryableMessages {
		if contains(errorMsg, msg) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
