// Package notifier provides integrations for sending cron job result
// notifications to various external services.
//
// Supported notifiers:
//   - Slack (incoming webhook)
//   - Email (SMTP)
//   - Generic Webhook (JSON POST)
//   - PagerDuty (Events API v2)
//   - OpsGenie (Alerts API)
//   - Microsoft Teams (incoming webhook)
//   - Discord (incoming webhook)
//
// All notifiers implement the Notifier interface defined in notifier.go.
// Use NewMulti to fan out notifications to multiple destinations.
//
// Example:
//
//	n := notifier.NewMulti(
//		notifier.NewSlackNotifier(webhookURL),
//		notifier.NewDiscordNotifier(discordURL),
//	)
//	n.Notify(result)
package notifier
