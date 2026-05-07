package notifier

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/cronwrap/internal/runner"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // normal operation
	CircuitOpen                         // failing, requests blocked
	CircuitHalfOpen                     // probing if service recovered
)

// CircuitBreaker wraps a Notifier and stops calling it after consecutive failures.
type CircuitBreaker struct {
	mu           sync.Mutex
	inner        Notifier
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	state        CircuitState
	openedAt     time.Time
}

// NewCircuitBreaker returns a Notifier that opens after maxFailures consecutive
// errors and resets after resetTimeout.
func NewCircuitBreaker(inner Notifier, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	if maxFailures < 1 {
		maxFailures = 3
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreaker{
		inner:        inner,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Notify forwards the result to the inner notifier unless the circuit is open.
func (cb *CircuitBreaker) Notify(result runner.Result) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitOpen:
		if time.Since(cb.openedAt) >= cb.resetTimeout {
			cb.state = CircuitHalfOpen
		} else {
			return fmt.Errorf("circuit breaker open: notifier skipped")
		}
	case CircuitHalfOpen:
		// allow one probe through
	}

	err := cb.inner.Notify(result)
	if err != nil {
		cb.failures++
		if cb.failures >= cb.maxFailures || cb.state == CircuitHalfOpen {
			cb.state = CircuitOpen
			cb.openedAt = time.Now()
		}
		return err
	}

	// success — reset
	cb.failures = 0
	cb.state = CircuitClosed
	return nil
}

// State returns the current circuit state (safe for inspection in tests).
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
