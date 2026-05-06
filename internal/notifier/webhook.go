package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookNotifier sends job results to a generic HTTP webhook endpoint.
type WebhookNotifier struct {
	URL    string
	client *http.Client
}

// webhookPayload is the JSON body sent to the webhook endpoint.
type webhookPayload struct {
	Job       string    `json:"job"`
	Success   bool      `json:"success"`
	ExitCode  int       `json:"exit_code"`
	Duration  float64   `json:"duration_seconds"`
	StartedAt time.Time `json:"started_at"`
	Output    string    `json:"output,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// NewWebhookNotifier creates a WebhookNotifier that posts to the given URL.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL: url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends the JobResult as a JSON POST request to the configured webhook URL.
func (w *WebhookNotifier) Notify(result JobResult) error {
	output := truncate(result.Output, 1000)

	payload := webhookPayload{
		Job:       result.Job,
		Success:   result.Success,
		ExitCode:  result.ExitCode,
		Duration:  result.Duration.Seconds(),
		StartedAt: result.StartedAt,
		Output:    output,
	}
	if !result.Success {
		payload.Error = result.ErrorMessage
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
