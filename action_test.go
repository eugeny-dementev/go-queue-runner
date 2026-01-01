package queuerunner

import (
	"testing"
	"time"
)

func TestWithErrorHandlerHandlesError(t *testing.T) {
	handled := false
	action := WithErrorHandler(func(_ *Context) error {
		return ErrInvalidScope
	}, func(_ error, _ *Context) {
		handled = true
	})

	if err := action(&Context{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handled {
		t.Fatal("expected error handler to be invoked")
	}
}

func TestWithDelaySleepsAfterAction(t *testing.T) {
	action := WithDelay(func(_ *Context) error {
		return nil
	}, 5*time.Millisecond)

	start := time.Now()
	if err := action(&Context{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) < 5*time.Millisecond {
		t.Fatalf("expected delay to be applied")
	}
}

func TestWithLockInvalidScopeReturnsError(t *testing.T) {
	action := WithLock(" ", func(_ *Context) error { return nil })

	if err := action(&Context{}); err == nil {
		t.Fatal("expected error for invalid lock scope")
	}
}
