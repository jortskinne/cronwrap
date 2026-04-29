// Package notifier provides notification backends for reporting
// cron job failures and results to external services.
package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackNotifier sends job result notifications to a Slack webhook.
type SlackNotifier struct {
	WebhookURL string
	Channel    string
	Username   string
	client     *http.Client
}

// slackPayload represents the JSON body sent to Slack's incoming webhook API.
type slackPayload struct {
	Channel  string            `json:"channel,omitempty"`
	Username string            `json:"username,omitempty"`
	Text     string            `json:"text"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

type slackAttachment struct {
	Color  string `json:"color"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Footer string `json:"footer"`
}

// NewSlackNotifier creates a SlackNotifier with the given webhook URL.
func NewSlackNotifier(webhookURL, channel, username string) *SlackNotifier {
	return &SlackNotifier{
		WebhookURL: webhookURL,
		Channel:    channel,
		Username:   username,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a notification about a job result to Slack.
func (s *SlackNotifier) Notify(n *Notification) error {
	color := "good"
	if !n.Success {
		color = "danger"
	}

	payload := slackPayload{
		Channel:  s.Channel,
		Username: s.Username,
		Text:     fmt.Sprintf("cronwrap: job *%s* %s", n.JobName, statusText(n.Success)),
		Attachments: []slackAttachment{
			{
				Color: color,
				Title: fmt.Sprintf("Exit code: %d | Duration: %s", n.ExitCode, n.Duration.Round(time.Millisecond)),
				Text:  truncate(n.Output, 1000),
				Footer: fmt.Sprintf("ran at %s", n.StartedAt.Format(time.RFC3339)),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func statusText(success bool) string {
	if success {
		return "succeeded ✅"
	}
	return "failed ❌"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "... (truncated)"
}
