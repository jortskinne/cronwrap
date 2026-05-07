package notifier

import (
	"errors"
	"testing"
	"time"

	"github.com/fatih/cronwrap/internal/runner"
)

type stubNotifier struct {
	calls int
	err   error
}

func (s *stubNotifier) Notify(_ runner.Result) error {
	s.calls++
	return s.err
}

var okResult = runner.Result{ExitCode: 0}
var errResult = runner.Result{ExitCode: 1}

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	stub := &stubNotifier{}
	cb := NewCircuitBreaker(stub, 3, time.Minute)

	if err := cb.Notify(okResult); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected Closed, got %d", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	stub := &stubNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(stub, 3, time.Minute)

	for i := 0; i < 3; i++ {
		_ = cb.Notify(errResult)
	}
	if cb.State() != CircuitOpen {
		t.Errorf("expected Open, got %d", cb.State())
	}
}

func TestCircuitBreaker_BlocksWhenOpen(t *testing.T) {
	stub := &stubNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(stub, 2, time.Minute)

	_ = cb.Notify(errResult)
	_ = cb.Notify(errResult)

	stub.err = nil // fix the inner notifier
	err := cb.Notify(okResult)
	if err == nil {
		t.Fatal("expected circuit-open error, got nil")
	}
	if stub.calls != 2 {
		t.Errorf("expected inner called 2 times, got %d", stub.calls)
	}
}

func TestCircuitBreaker_HalfOpenProbe_Success(t *testing.T) {
	stub := &stubNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(stub, 2, 0) // zero resets to 30s default, force via field
	cb.resetTimeout = time.Millisecond

	_ = cb.Notify(errResult)
	_ = cb.Notify(errResult) // opens circuit

	time.Sleep(5 * time.Millisecond)
	stub.err = nil

	err := cb.Notify(okResult)
	if err != nil {
		t.Fatalf("probe should succeed: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected Closed after successful probe, got %d", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenProbe_Failure(t *testing.T) {
	stub := &stubNotifier{err: errors.New("still broken")}
	cb := NewCircuitBreaker(stub, 2, time.Millisecond)

	_ = cb.Notify(errResult)
	_ = cb.Notify(errResult)
	time.Sleep(5 * time.Millisecond)

	err := cb.Notify(errResult)
	if err == nil {
		t.Fatal("expected error on failed probe")
	}
	if cb.State() != CircuitOpen {
		t.Errorf("expected Open after failed probe, got %d", cb.State())
	}
}

func TestCircuitBreaker_DefaultsApplied(t *testing.T) {
	cb := NewCircuitBreaker(&stubNotifier{}, 0, 0)
	if cb.maxFailures != 3 {
		t.Errorf("expected default maxFailures=3, got %d", cb.maxFailures)
	}
	if cb.resetTimeout != 30*time.Second {
		t.Errorf("expected default resetTimeout=30s, got %v", cb.resetTimeout)
	}
}
