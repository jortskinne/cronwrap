package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends alerts to PagerDuty via the Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	httpClient     *http.Client
	eventsURL      string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	DedupKey    string    `json:"dedup_key,omitempty"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary   string `json:"summary"`
	Source    string `json:"source"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
	CustomDetails map[string]string `json:"custom_details,omitempty"`
}

// NewPagerDutyNotifier creates a PagerDutyNotifier with the given integration key.
func NewPagerDutyNotifier(integrationKey string) *PagerDutyNotifier {
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		eventsURL:      pagerDutyEventsURL,
	}
}

// Notify sends a PagerDuty event for failed jobs; resolves on success.
func (p *PagerDutyNotifier) Notify(result JobResult) error {
	action := "resolve"
	if !result.Success {
		action = "trigger"
	}

	details := map[string]string{
		"exit_code": fmt.Sprintf("%d", result.ExitCode),
		"output":    truncate(result.Output, 512),
	}

	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: action,
		DedupKey:    fmt.Sprintf("cronwrap-%s", result.JobName),
		Payload: pdDetails{
			Summary:       fmt.Sprintf("cronwrap: job %q %s", result.JobName, statusText(result.Success)),
			Source:        "cronwrap",
			Severity:      pdSeverity(result.Success),
			Timestamp:     result.StartedAt.UTC().Format(time.RFC3339),
			CustomDetails: details,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	resp, err := p.httpClient.Post(p.eventsURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func pdSeverity(success bool) string {
	if success {
		return "info"
	}
	return "critical"
}
