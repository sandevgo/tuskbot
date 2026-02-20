package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry_SuccessOnFirstTry(t *testing.T) {
	ctx := context.Background()
	retrier := NewDefaultRetrier()

	counter := 0
	var resultValue int
	operation := func() error {
		counter++
		resultValue = 42
		return nil
	}

	err := retrier.Do(ctx, operation)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resultValue != 42 {
		t.Errorf("expected 42, got %d", resultValue)
	}
	if counter != 1 {
		t.Errorf("expected 1 attempt, got %d", counter)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	retrier := NewDefaultRetrier()

	counter := 0
	var resultValue int
	operation := func() error {
		counter++
		if counter < 2 {
			return errors.New("temporary error")
		}
		resultValue = 42
		return nil
	}

	err := retrier.Do(ctx, operation)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resultValue != 42 {
		t.Errorf("expected 42, got %d", resultValue)
	}
	if counter != 2 {
		t.Errorf("expected 2 attempts, got %d", counter)
	}
}

func TestRetry_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	config := NewDefaultConfig()
	config.MaxRetries = 2
	retrier := NewRetrier(config)

	expectedErr := errors.New("permanent error")
	counter := 0
	operation := func() error {
		counter++
		return expectedErr
	}

	err := retrier.Do(ctx, operation)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) { // Use errors.Is for checking wrapped errors, though here it's direct.
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if counter != 3 { // Initial try + 2 retries
		t.Errorf("expected 3 attempts, got %d", counter)
	}
}

func TestRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	retrier := NewDefaultRetrier()

	operation := func() error {
		cancel() // Cancel the context during the operation
		return errors.New("operation error after cancel")
	}

	err := retrier.Do(ctx, operation)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRetry_BackoffAndJitter(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		MaxRetries:    2,
		BackoffFactor: 2.0,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		Jitter:        50 * time.Millisecond,
	}
	retrier := NewRetrier(config)

	start := time.Now()
	counter := 0
	operation := func() error {
		counter++
		return errors.New("error")
	}

	_ = retrier.Do(ctx, operation)
	elapsed := time.Since(start)

	// Total attempts = MaxRetries + 1 = 3
	// Delays happen *before* retries. So 2 delays.
	// 1st delay: InitialDelay + jitter_1 (where jitter_1 is up to config.Jitter)
	//   delay = 100ms. nextDelay = (100ms + jitter_1)
	// 2nd delay: (InitialDelay * BackoffFactor) + jitter_2
	//   delay = 100ms * 2.0 = 200ms. nextDelay = (200ms + jitter_2)

	// Minimum total delay: (100ms + 0ms) + (200ms + 0ms) = 300ms
	minExpectedDelay := config.InitialDelay + time.Duration(float64(config.InitialDelay)*config.BackoffFactor)
	// Maximum total delay: (100ms + 50ms) + (200ms + 50ms) = 150ms + 250ms = 400ms
	maxExpectedDelay := (config.InitialDelay + config.Jitter) + (time.Duration(float64(config.InitialDelay)*config.BackoffFactor) + config.Jitter)
	if maxExpectedDelay > (config.MaxDelay + config.Jitter + config.MaxDelay + config.Jitter) { // Ensure MaxDelay logic is sound
		// This calculation needs to be careful if MaxDelay is hit.
		// Delay 1: min(100, MaxDelay) + jitter. Let MaxDelay = 1s. So 100ms + jitter.
		// Delay 2: min(200, MaxDelay) + jitter. So 200ms + jitter.
		// This is correct for the given values.
	}

	// The test measures total elapsed time, which includes the execution of the operation itself.
	// However, the operation is very fast (just increments a counter and returns an error).
	// So, elapsed time will be dominated by the delays.

	if elapsed < minExpectedDelay || elapsed > maxExpectedDelay {
		t.Errorf("expected total delay to be roughly between %v and %v, got %v", minExpectedDelay, maxExpectedDelay, elapsed)
	}
	if counter != 3 { // Initial try + 2 retries
		t.Errorf("expected 3 attempts, got %d", counter)
	}
}
