package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TeamsNotifier sends job result notifications to a Microsoft Teams channel
// via an Incoming Webhook URL.
type TeamsNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewTeamsNotifier creates a TeamsNotifier that posts to the given webhook URL.
func NewTeamsNotifier(webhookURL string) *TeamsNotifier {
	return &TeamsNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

type teamsCard struct {
	Type       string         `json:"@type"`
	Context    string         `json:"@context"`
	ThemeColor string         `json:"themeColor"`
	Summary    string         `json:"summary"`
	Sections   []teamsSection `json:"sections"`
}

type teamsSection struct {
	ActivityTitle    string      `json:"activityTitle"`
	ActivitySubtitle string      `json:"activitySubtitle"`
	Facts            []teamsFact `json:"facts"`
	Markdown         bool        `json:"markdown"`
}

type teamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Notify sends a Teams message card for the given job result.
func (t *TeamsNotifier) Notify(result JobResult) error {
	color := "00FF00"
	if !result.Success {
		color = "FF0000"
	}

	card := teamsCard{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: color,
		Summary:    fmt.Sprintf("cronwrap: %s", result.JobName),
		Sections: []teamsSection{
			{
				ActivityTitle:    fmt.Sprintf("Job: %s", result.JobName),
				ActivitySubtitle: statusText(result.Success),
				Facts: []teamsFact{
					{Name: "Exit Code", Value: fmt.Sprintf("%d", result.ExitCode)},
					{Name: "Duration", Value: result.Duration.String()},
					{Name: "Output", Value: truncate(result.Output, 500)},
				},
				Markdown: false,
			},
		},
	}

	body, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("teams: marshal card: %w", err)
	}

	resp, err := t.client.Post(t.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("teams: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams: unexpected status %d", resp.StatusCode)
	}
	return nil
}
