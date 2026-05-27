package aws

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), retryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond}, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_SuccessOnThirdAttempt(t *testing.T) {
	calls := 0
	sentinel := errors.New("transient")
	err := withRetry(context.Background(), retryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond}, func() error {
		calls++
		if calls < 3 {
			return sentinel
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success on third attempt, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_ExhaustsAllAttempts(t *testing.T) {
	sentinel := errors.New("permanent error")
	calls := 0
	err := withRetry(context.Background(), retryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond}, func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected exactly 3 calls, got %d", calls)
	}
}

func TestWithRetry_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	calls := 0
	err := withRetry(ctx, retryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond}, func() error {
		calls++
		return nil
	})
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if calls != 0 {
		t.Errorf("expected 0 calls with cancelled context, got %d", calls)
	}
}
