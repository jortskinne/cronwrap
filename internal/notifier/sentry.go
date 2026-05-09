package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourorg/cronwrap/internal/runner"
)

// SentryNotifier sends error events to a Sentry DSN-compatible endpoint.
type SentryNotifier struct {
	dsn        string
	environment string
	release    string
	client     *http.Client
}

type sentryEvent struct {
	EventID     string            `json:"event_id"`
	Timestamp   string            `json:"timestamp"`
	Level       string            `json:"level"`
	Logger      string            `json:"logger"`
	Message     string            `json:"message"`
	Environment string            `json:"environment,omitempty"`
	Release     string            `json:"release,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`
}

// NewSentryNotifier creates a Sentry notifier that reports failures as error events.
func NewSentryNotifier(dsn, environment, release string) *SentryNotifier {
	return &SentryNotifier{
		dsn:         dsn,
		environment: environment,
		release:     release,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a Sentry event for failed job runs.
func (s *SentryNotifier) Notify(result runner.Result) error {
	if result.ExitCode == 0 {
		return nil
	}

	level := "error"
	msg := fmt.Sprintf("cronwrap: job %q failed (exit %d)", result.Command, result.ExitCode)

	event := sentryEvent{
		EventID:     fmt.Sprintf("%x", time.Now().UnixNano()),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Level:       level,
		Logger:      "cronwrap",
		Message:     msg,
		Environment: s.environment,
		Release:     s.release,
		Extra: map[string]string{
			"command": result.Command,
			"stdout":  truncate(result.Stdout, 1024),
			"stderr":  truncate(result.Stderr, 1024),
			"exit_code": fmt.Sprintf("%d", result.ExitCode),
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("sentry: marshal event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.dsn, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sentry: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sentry: send event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("sentry: unexpected status %d", resp.StatusCode)
	}
	return nil
}
