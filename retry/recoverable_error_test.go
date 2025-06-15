package retry

import (
	"errors"
	"testing"
)

func TestIsRecoverable_Nil(t *testing.T) {
	if IsRecoverable(nil) {
		t.Error("expected IsRecoverable(nil) to be false")
	}
}

func TestIsRecoverable_NonRecoverableError(t *testing.T) {
	err := errors.New("regular error")
	if IsRecoverable(err) {
		t.Error("expected regular error to not be recoverable")
	}
}

func TestIsRecoverable_RecoverableError(t *testing.T) {
	err := NewRecoverableError(errors.New("test error"))
	if !IsRecoverable(err) {
		t.Error("expected recoverable error to be recoverable")
	}
}

func TestNewRecoverableError(t *testing.T) {
	originalErr := errors.New("original error")
	recoverableErr := NewRecoverableError(originalErr)

	if !recoverableErr.IsRecoverable() {
		t.Error("expected IsRecoverable() to return true")
	}

	if recoverableErr.Error() != "original error" {
		t.Errorf("expected error message 'original error', got '%s'", recoverableErr.Error())
	}

	if recoverableErr.Unwrap() != originalErr {
		t.Error("expected Unwrap() to return original error")
	}
}

func TestRecoverableError_ErrorInterface(t *testing.T) {
	originalErr := errors.New("test error")
	recoverableErr := NewRecoverableError(originalErr)

	// Test that it implements error interface
	var err error = recoverableErr
	if err.Error() != "test error" {
		t.Errorf("expected error message 'test error', got '%s'", err.Error())
	}
}

func TestRecoverableError_Unwrap(t *testing.T) {
	originalErr := errors.New("wrapped error")
	recoverableErr := NewRecoverableError(originalErr)

	unwrapped := errors.Unwrap(recoverableErr)
	if unwrapped != originalErr {
		t.Error("expected Unwrap to return the original error")
	}
}

func TestIsRecoverable_WrappedRecoverableError(t *testing.T) {
	originalErr := errors.New("base error")
	recoverableErr := NewRecoverableError(originalErr)
	wrappedErr := errors.New("wrapped: " + recoverableErr.Error())

	// This should not be recoverable because it's wrapped in a regular error
	if IsRecoverable(wrappedErr) {
		t.Error("expected wrapped recoverable error to not be recoverable")
	}
}

type customRecoverableError struct {
	msg string
}

func (e *customRecoverableError) Error() string {
	return e.msg
}

func (e *customRecoverableError) IsRecoverable() bool {
	return true
}

func TestIsRecoverable_CustomRecoverableError(t *testing.T) {
	err := &customRecoverableError{msg: "custom recoverable error"}
	if !IsRecoverable(err) {
		t.Error("expected custom recoverable error to be recoverable")
	}
}

type customNonRecoverableError struct {
	msg string
}

func (e *customNonRecoverableError) Error() string {
	return e.msg
}

func (e *customNonRecoverableError) IsRecoverable() bool {
	return false
}

func TestIsRecoverable_CustomNonRecoverableError(t *testing.T) {
	err := &customNonRecoverableError{msg: "custom non-recoverable error"}
	if IsRecoverable(err) {
		t.Error("expected custom non-recoverable error to not be recoverable")
	}
}
