package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const opsgenieAlertsURL = "https://api.opsgenie.com/v2/alerts"

// OpsGenieNotifier sends alerts to OpsGenie when a job fails.
type OpsGenieNotifier struct {
	apiKey   string
	team     string
	baseURL  string
	client   *http.Client
}

type opsgeniePayload struct {
	Message     string            `json:"message"`
	Alias       string            `json:"alias"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Tags        []string          `json:"tags,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
}

// NewOpsGenieNotifier creates a notifier that sends alerts to OpsGenie.
// apiKey is the OpsGenie API key, team is an optional team tag.
func NewOpsGenieNotifier(apiKey, team string) *OpsGenieNotifier {
	return &OpsGenieNotifier{
		apiKey:  apiKey,
		team:    team,
		baseURL: opsgenieAlertsURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify implements the Notifier interface.
func (o *OpsGenieNotifier) Notify(result JobResult) error {
	if result.Success {
		return nil
	}

	tags := []string{"cronwrap"}
	if o.team != "" {
		tags = append(tags, o.team)
	}

	payload := opsgeniePayload{
		Message:     fmt.Sprintf("cronwrap: job failed — %s", result.JobName),
		Alias:       fmt.Sprintf("cronwrap-%s", result.JobName),
		Description: truncate(result.Output, 1000),
		Priority:    "P2",
		Tags:        tags,
		Details: map[string]string{
			"exit_code": fmt.Sprintf("%d", result.ExitCode),
			"duration":  result.Duration.String(),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, o.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status %d", resp.StatusCode)
	}
	return nil
}
