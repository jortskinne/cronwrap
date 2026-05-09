package notifier

import (
	"fmt"
	"net/http"
	"time"

	"github.com/example/cronwrap/internal/runner"
)

// HealthCheckNotifier pings a URL (e.g. healthchecks.io or Uptime Kuma)
// after each job run. On success it appends "/success"; on failure "/fail".
// A zero-value Duration disables the HTTP timeout guard.
type HealthCheckNotifier struct {
	baseURL string
	client  *http.Client
}

// NewHealthCheckNotifier creates a notifier that reports job outcomes to a
// ping-based health-check service.
func NewHealthCheckNotifier(baseURL string, timeout time.Duration) *HealthCheckNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &HealthCheckNotifier{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

// Notify sends a ping to the health-check endpoint.
func (h *HealthCheckNotifier) Notify(result runner.Result) error {
	if h.baseURL == "" {
		return nil
	}

	suffix := "/success"
	if !result.Success {
		suffix = "/fail"
	}

	url := h.baseURL + suffix
	resp, err := h.client.Get(url) //nolint:noctx
	if err != nil {
		return fmt.Errorf("healthcheck ping failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("healthcheck ping returned HTTP %d", resp.StatusCode)
	}
	return nil
}
