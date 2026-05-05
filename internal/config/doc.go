// Package config provides loading, parsing, and validation of cronwrap
// configuration files.
//
// Configuration is read from a YAML file whose path is supplied at runtime.
// Selected fields can be overridden by environment variables so that secrets
// (e.g. Slack webhook URLs, SMTP passwords) are never stored on disk.
//
// Supported environment overrides:
//
//	CRONWRAP_SLACK_WEBHOOK  – overrides slack.webhook_url
//	CRONWRAP_EMAIL_PASSWORD – overrides email.password
//	CRONWRAP_HISTORY_PATH   – overrides history.path
//
// Example usage:
//
//	cfg, err := config.Load("/etc/cronwrap/config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
package config
