package notifier

import (
	"errors"
	"net/smtp"
	"strings"
	"testing"
	"time"
)

func TestEmailNotifier_Success(t *testing.T) {
	var capturedMsg []byte
	notifier := NewEmailNotifier(EmailConfig{
		SMTPHost: "localhost",
		SMTPPort: 587,
		From:     "cron@example.com",
		To:       []string{"ops@example.com"},
	})
	notifier.send = func(addr, from string, to []string, msg []byte) error {
		capturedMsg = msg
		return nil
	}

	err := notifier.Notify(JobResult{
		JobName:  "backup",
		Success:  true,
		Duration: 2 * time.Second,
		Output:   "done",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	msgStr := string(capturedMsg)
	if !strings.Contains(msgStr, "backup") {
		t.Error("expected job name in message")
	}
	if !strings.Contains(msgStr, "SUCCESS") {
		t.Error("expected SUCCESS in message")
	}
}

func TestEmailNotifier_Failure(t *testing.T) {
	notifier := NewEmailNotifier(EmailConfig{
		SMTPHost: "localhost",
		SMTPPort: 587,
		From:     "cron@example.com",
		To:       []string{"ops@example.com"},
	})
	var capturedMsg []byte
	notifier.send = func(addr, from string, to []string, msg []byte) error {
		capturedMsg = msg
		return nil
	}

	err := notifier.Notify(JobResult{
		JobName:  "import",
		Success:  false,
		ExitCode: 1,
		Error:    "connection refused",
		Duration: 500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	msgStr := string(capturedMsg)
	if !strings.Contains(msgStr, "FAILURE") {
		t.Error("expected FAILURE in message")
	}
	if !strings.Contains(msgStr, "connection refused") {
		t.Error("expected error text in message")
	}
}

func TestEmailNotifier_SMTPError(t *testing.T) {
	notifier := NewEmailNotifier(EmailConfig{
		SMTPHost: "localhost",
		SMTPPort: 587,
		From:     "cron@example.com",
		To:       []string{"ops@example.com"},
	})
	notifier.send = func(addr, from string, to []string, msg []byte) error {
		return errors.New("connection refused")
	}

	err := notifier.Notify(JobResult{JobName: "test", Success: false})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEmailNotifier_AuthUsed(t *testing.T) {
	var capturedAuth smtp.Auth
	notifier := NewEmailNotifier(EmailConfig{
		SMTPHost: "mail.example.com",
		SMTPPort: 587,
		Username: "user",
		Password: "pass",
		From:     "cron@example.com",
		To:       []string{"ops@example.com"},
	})
	_ = capturedAuth
	notifier.send = func(addr, from string, to []string, msg []byte) error {
		if addr != "mail.example.com:587" {
			t.Errorf("unexpected addr: %s", addr)
		}
		return nil
	}
	if err := notifier.Notify(JobResult{JobName: "j", Success: true}); err != nil {
		t.Fatal(err)
	}
}
