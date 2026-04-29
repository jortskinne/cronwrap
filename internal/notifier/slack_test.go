package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlackNotifier_Success(t *testing.T) {
	var received slackPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewSlackNotifier(server.URL, "#alerts", "cronwrap")
	notif := &Notification{
		JobName:   "backup",
		Success:   true,
		ExitCode:  0,
		Output:    "backup complete",
		Duration:  2 * time.Second,
		StartedAt: time.Now(),
	}

	if err := n.Notify(notif); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received.Channel != "#alerts" {
		t.Errorf("expected channel #alerts, got %q", received.Channel)
	}
	if len(received.Attachments) == 0 || received.Attachments[0].Color != "good" {
		t.Errorf("expected green attachment for success")
	}
}

func TestSlackNotifier_Failure(t *testing.T) {
	var received slackPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received) //nolint:errcheck
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewSlackNotifier(server.URL, "", "")
	notif := &Notification{
		JobName:  "deploy",
		Success:  false,
		ExitCode: 1,
		Output:   "error: connection refused",
		Duration: 500 * time.Millisecond,
		StartedAt: time.Now(),
	}

	if err := n.Notify(notif); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received.Attachments) == 0 || received.Attachments[0].Color != "danger" {
		t.Errorf("expected red attachment for failure")
	}
}

func TestSlackNotifier_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := NewSlackNotifier(server.URL, "#ops", "cronwrap")
	notif := &Notification{JobName: "test", StartedAt: time.Now()}

	if err := n.Notify(notif); err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestTruncate(t *testing.T) {
	long := string(make([]byte, 1500))
	result := truncate(long, 1000)
	if len(result) <= 1000 {
		t.Logf("truncated to %d chars (with suffix)", len(result))
	}
	short := "hello"
	if truncate(short, 100) != short {
		t.Errorf("short string should not be truncated")
	}
}
