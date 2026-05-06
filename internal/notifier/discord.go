package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DiscordNotifier sends job result notifications to a Discord webhook.
type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewDiscordNotifier creates a new DiscordNotifier using the given webhook URL.
func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
}

type discordPayload struct {
	Username string         `json:"username"`
	Embeds   []discordEmbed `json:"embeds"`
}

// Notify sends a Discord embed message for the given job result.
func (d *DiscordNotifier) Notify(result JobResult) error {
	color := 0x2ECC71 // green for success
	if !result.Success {
		color = 0xE74C3C // red for failure
	}

	description := fmt.Sprintf("**Job:** `%s`\n**Exit Code:** %d\n**Duration:** %s",
		result.Command,
		result.ExitCode,
		result.Duration.Round(time.Millisecond),
	)
	if result.Output != "" {
		description += fmt.Sprintf("\n**Output:**\n```\n%s\n```", truncate(result.Output, 1800))
	}

	payload := discordPayload{
		Username: "cronwrap",
		Embeds: []discordEmbed{
			{
				Title:       fmt.Sprintf("Cron Job %s", statusText(result.Success)),
				Description: description,
				Color:       color,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}
	return nil
}
