package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetrier_Do_Success(t *testing.T) {
	retrier := New(3, 100*time.Millisecond, 1*time.Second)

	callCount := 0
	err := retrier.Do(context.Background(), "test_operation", func() error {
		callCount++
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetrier_Do_SuccessAfterRetries(t *testing.T) {
	retrier := New(3, 100*time.Millisecond, 1*time.Second)

	callCount := 0
	err := retrier.Do(context.Background(), "test_operation", func() error {
		callCount++
		if callCount < 3 {
			return &pq.Error{Code: "08006"} // connection_failure - retryable
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 3, callCount)
}

func TestRetrier_Do_MaxRetriesExceeded(t *testing.T) {
	retrier := New(2, 50*time.Millisecond, 500*time.Millisecond)

	callCount := 0
	err := retrier.Do(context.Background(), "test_operation", func() error {
		callCount++
		return &pq.Error{Code: "08006"} // Always fail with retryable error
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries (2) exceeded")
	assert.Equal(t, 3, callCount) // Initial attempt + 2 retries
}

func TestRetrier_Do_NonRetryableError(t *testing.T) {
	retrier := New(3, 100*time.Millisecond, 1*time.Second)

	callCount := 0
	nonRetryableErr := &pq.Error{Code: "42501"} // insufficient_privilege - not retryable

	err := retrier.Do(context.Background(), "test_operation", func() error {
		callCount++
		return nonRetryableErr
	})

	require.Error(t, err)
	assert.Equal(t, nonRetryableErr, err)
	assert.Equal(t, 1, callCount) // Should not retry
}

func TestRetrier_Do_ContextCancellation(t *testing.T) {
	retrier := New(5, 200*time.Millisecond, 2*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	callCount := 0
	err := retrier.Do(ctx, "test_operation", func() error {
		callCount++
		return &pq.Error{Code: "08006"} // Always fail with retryable error
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled during retry")
	assert.True(t, callCount >= 1 && callCount <= 3) // Should be cancelled during retries
}

func TestRetrier_DoWithTimeout(t *testing.T) {
	retrier := New(5, 100*time.Millisecond, 1*time.Second)

	callCount := 0
	err := retrier.DoWithTimeout(context.Background(), 250*time.Millisecond, "test_operation", func() error {
		callCount++
		return &pq.Error{Code: "08006"} // Always fail with retryable error
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled during retry")
	assert.True(t, callCount >= 1 && callCount <= 3) // Should timeout during retries
}

func TestRetrier_calculateDelay(t *testing.T) {
	retrier := New(5, 100*time.Millisecond, 2*time.Second)

	tests := []struct {
		attempt        int
		expectedMin    time.Duration
		expectedMax    time.Duration
		shouldBeCapped bool
	}{
		{1, 100 * time.Millisecond, 100 * time.Millisecond, false},
		{2, 200 * time.Millisecond, 200 * time.Millisecond, false},
		{3, 400 * time.Millisecond, 400 * time.Millisecond, false},
		{4, 800 * time.Millisecond, 800 * time.Millisecond, false},
		{5, 1600 * time.Millisecond, 1600 * time.Millisecond, false},
		{6, 2 * time.Second, 2 * time.Second, true},  // Should be capped
		{10, 2 * time.Second, 2 * time.Second, true}, // Should be capped
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.attempt)), func(t *testing.T) {
			delay := retrier.calculateDelay(tt.attempt)

			if tt.shouldBeCapped {
				assert.Equal(t, retrier.maxDelay, delay)
			} else {
				assert.Equal(t, tt.expectedMin, delay)
			}
		})
	}
}

func TestRetrier_isRetryableError(t *testing.T) {
	retrier := New(3, 100*time.Millisecond, 1*time.Second)

	tests := []struct {
		name        string
		err         error
		shouldRetry bool
	}{
		{
			name:        "PostgreSQL connection failure",
			err:         &pq.Error{Code: "08006"},
			shouldRetry: true,
		},
		{
			name:        "PostgreSQL too many connections",
			err:         &pq.Error{Code: "53300"},
			shouldRetry: true,
		},
		{
			name:        "PostgreSQL authentication error",
			err:         &pq.Error{Code: "28000"},
			shouldRetry: false,
		},
		{
			name:        "Generic connection refused error",
			err:         errors.New("connection refused"),
			shouldRetry: true,
		},
		{
			name:        "Generic timeout error",
			err:         errors.New("i/o timeout"),
			shouldRetry: true,
		},
		{
			name:        "Non-retryable generic error",
			err:         errors.New("invalid syntax"),
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := retrier.isRetryableError(tt.err)
			assert.Equal(t, tt.shouldRetry, result)
		})
	}
}
