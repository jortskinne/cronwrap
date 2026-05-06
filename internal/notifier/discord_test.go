package notifier

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDiscordNotifier_Success(t *testing.T) {
	var received discordPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := NewDiscordNotifier(ts.URL)
	err := n.Notify(JobResult{
		Command:  "backup.sh",
		Success:  true,
		ExitCode: 0,
		Duration: 2 * time.Second,
		Output:   "done",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(received.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(received.Embeds))
	}
	if received.Embeds[0].Color != 0x2ECC71 {
		t.Errorf("expected green color for success, got %d", received.Embeds[0].Color)
	}
	if !strings.Contains(received.Embeds[0].Description, "backup.sh") {
		t.Errorf("expected command in description")
	}
}

func TestDiscordNotifier_Failure(t *testing.T) {
	var received discordPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := NewDiscordNotifier(ts.URL)
	err := n.Notify(JobResult{
		Command:  "sync.sh",
		Success:  false,
		ExitCode: 1,
		Duration: 500 * time.Millisecond,
		Output:   "error occurred",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Embeds[0].Color != 0xE74C3C {
		t.Errorf("expected red color for failure, got %d", received.Embeds[0].Color)
	}
	if !strings.Contains(received.Embeds[0].Title, "Failure") {
		t.Errorf("expected Failure in title, got %q", received.Embeds[0].Title)
	}
}

func TestDiscordNotifier_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewDiscordNotifier(ts.URL)
	err := n.Notify(JobResult{Command: "job", Success: false})
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status 500 in error, got %v", err)
	}
}

func TestDiscordNotifier_InvalidURL(t *testing.T) {
	n := NewDiscordNotifier("http://127.0.0.1:0/invalid")
	err := n.Notify(JobResult{Command: "job", Success: true})
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestDiscordNotifier_LongOutputTruncated(t *testing.T) {
	var received discordPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := NewDiscordNotifier(ts.URL)
	longOutput := strings.Repeat("x", 3000)
	err := n.Notify(JobResult{Command: "job", Success: true, Output: longOutput})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received.Embeds[0].Description) > 2200 {
		t.Errorf("description too long, output was not truncated")
	}
}
