package notifier

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// EmailConfig holds configuration for the email notifier.
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	To       []string
}

// EmailNotifier sends job result notifications via email.
type EmailNotifier struct {
	cfg  EmailConfig
	send func(addr, from string, to []string, msg []byte) error
}

// NewEmailNotifier creates an EmailNotifier with the given configuration.
func NewEmailNotifier(cfg EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		cfg:  cfg,
		send: smtp.SendMail,
	}
}

// Notify sends an email notification for the given job result.
func (e *EmailNotifier) Notify(result JobResult) error {
	subject := fmt.Sprintf("[cronwrap] %s — %s", statusText(result.Success), result.JobName)
	body := buildEmailBody(result)

	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.SMTPHost)

	msg := buildRawMessage(e.cfg.From, e.cfg.To, subject, body)
	return e.send(addr, auth, e.cfg.From, e.cfg.To, msg)
}

func buildEmailBody(r JobResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Job:      %s\n", r.JobName))
	sb.WriteString(fmt.Sprintf("Status:   %s\n", statusText(r.Success)))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", r.Duration.Round(time.Millisecond)))
	if r.ExitCode != 0 {
		sb.WriteString(fmt.Sprintf("Exit Code: %d\n", r.ExitCode))
	}
	if r.Output != "" {
		sb.WriteString("\nOutput:\n")
		sb.WriteString(truncate(r.Output, 4096))
	}
	if r.Error != "" {
		sb.WriteString("\nError:\n")
		sb.WriteString(truncate(r.Error, 2048))
	}
	return sb.String()
}

func buildRawMessage(from string, to []string, subject, body string) []byte {
	header := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n",
		from, strings.Join(to, ", "), subject,
	)
	return []byte(header + body)
}
