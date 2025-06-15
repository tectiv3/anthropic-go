package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo_Success(t *testing.T) {
	callCount := 0
	f := func() error {
		callCount++
		return nil
	}

	ctx := context.Background()
	err := Do(ctx, f)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected function to be called once, got %d", callCount)
	}
}

func TestDo_NonRecoverableError(t *testing.T) {
	callCount := 0
	testErr := errors.New("non-recoverable error")
	f := func() error {
		callCount++
		return testErr
	}

	ctx := context.Background()
	err := Do(ctx, f)
	if err != testErr {
		t.Errorf("expected %v, got %v", testErr, err)
	}
	if callCount != 1 {
		t.Errorf("expected function to be called once, got %d", callCount)
	}
}

func TestDo_RecoverableErrorEventualSuccess(t *testing.T) {
	callCount := 0
	testErr := NewRecoverableError(errors.New("recoverable error"))
	f := func() error {
		callCount++
		if callCount < 3 {
			return testErr
		}
		return nil
	}

	ctx := context.Background()
	err := Do(ctx, f, WithMaxRetries(5), WithBaseWait(1*time.Millisecond))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected function to be called 3 times, got %d", callCount)
	}
}

func TestDo_RecoverableErrorMaxRetriesExceeded(t *testing.T) {
	callCount := 0
	testErr := NewRecoverableError(errors.New("recoverable error"))
	f := func() error {
		callCount++
		return testErr
	}

	ctx := context.Background()
	err := Do(ctx, f, WithMaxRetries(2), WithBaseWait(1*time.Millisecond))
	if err != testErr {
		t.Errorf("expected %v, got %v", testErr, err)
	}
	if callCount != 3 { // initial attempt + 2 retries
		t.Errorf("expected function to be called 3 times, got %d", callCount)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	callCount := 0
	testErr := NewRecoverableError(errors.New("recoverable error"))
	f := func() error {
		callCount++
		return testErr
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after first attempt
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, f, WithMaxRetries(5), WithBaseWait(10*time.Millisecond))
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	// Should have been called at least once, but not all retries due to cancellation
	if callCount == 0 {
		t.Errorf("expected function to be called at least once, got %d", callCount)
	}
}

func TestDo_WithOptions(t *testing.T) {
	callCount := 0
	testErr := NewRecoverableError(errors.New("recoverable error"))
	f := func() error {
		callCount++
		return testErr
	}

	start := time.Now()
	ctx := context.Background()
	err := Do(ctx, f, WithMaxRetries(1), WithBaseWait(50*time.Millisecond))
	duration := time.Since(start)

	if err != testErr {
		t.Errorf("expected %v, got %v", testErr, err)
	}
	if callCount != 2 { // initial attempt + 1 retry
		t.Errorf("expected function to be called 2 times, got %d", callCount)
	}
	// Should have waited at least the base wait time
	if duration < 50*time.Millisecond {
		t.Errorf("expected to wait at least 50ms, but took %v", duration)
	}
}

func TestWithMaxRetries(t *testing.T) {
	config := &retryConfig{}
	opt := WithMaxRetries(5)
	opt(config)
	if config.MaxRetries != 5 {
		t.Errorf("expected MaxRetries to be 5, got %d", config.MaxRetries)
	}
}

func TestWithBaseWait(t *testing.T) {
	config := &retryConfig{}
	opt := WithBaseWait(10 * time.Second)
	opt(config)
	if config.BaseWait != 10*time.Second {
		t.Errorf("expected BaseWait to be 10s, got %v", config.BaseWait)
	}
}
