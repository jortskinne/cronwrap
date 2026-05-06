package notifier

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookNotifier_Success(t *testing.T) {
	var received webhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewWebhookNotifier(server.URL)
	err := n.Notify(JobResult{
		Job:      "test-job",
		Success:  true,
		ExitCode: 0,
		Duration: 2 * time.Second,
		Output:   "all good",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received.Job != "test-job" {
		t.Errorf("expected job 'test-job', got %q", received.Job)
	}
	if !received.Success {
		t.Error("expected success=true")
	}
	if received.DurationSeconds := received.Duration; received.DurationSeconds != 2.0 {
		t.Errorf("expected duration 2.0, got %f", received.DurationSeconds)
	}
}

func TestWebhookNotifier_Failure(t *testing.T) {
	var received webhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewWebhookNotifier(server.URL)
	err := n.Notify(JobResult{
		Job:          "fail-job",
		Success:      false,
		ExitCode:     1,
		ErrorMessage: "something broke",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Error != "something broke" {
		t.Errorf("expected error message 'something broke', got %q", received.Error)
	}
	if received.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", received.ExitCode)
	}
}

func TestWebhookNotifier_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := NewWebhookNotifier(server.URL)
	err := n.Notify(JobResult{Job: "x", Success: true})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestWebhookNotifier_InvalidURL(t *testing.T) {
	n := NewWebhookNotifier("http://127.0.0.1:0/no-server")
	err := n.Notify(JobResult{Job: "x", Success: true})
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}
