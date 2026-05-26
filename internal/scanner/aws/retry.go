package aws

import (
	"context"
	"math/rand"
	"time"
)

// retryConfig holds retry parameters for AWS API calls.
type retryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
}

// defaultRetry is the standard retry policy: 3 attempts with 1s/2s/4s backoff.
var defaultRetry = retryConfig{
	MaxAttempts: 3,
	BaseDelay:   time.Second,
}

// withRetry executes fn with exponential backoff + jitter on transient errors.
// Attempts: 1st immediate, 2nd after ~1s, 3rd after ~2s (base doubles each time).
// A random jitter of up to 100ms is added to each delay to avoid thundering-herd.
// The context is checked before every attempt so callers can cancel promptly.
func withRetry(ctx context.Context, cfg retryConfig, fn func() error) error {
	var lastErr error
	delay := cfg.BaseDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Do not sleep after the final attempt.
		if attempt == cfg.MaxAttempts {
			break
		}

		// Exponential backoff with up to 100 ms jitter.
		jitter := time.Duration(rand.Int63n(int64(100 * time.Millisecond)))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay + jitter):
		}

		delay *= 2
	}

	return lastErr
}
