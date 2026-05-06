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

func TestTeamsNotifier_Success(t *testing.T) {
	var received teamsCard
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewTeamsNotifier(ts.URL)
	err := n.Notify(JobResult{
		JobName:  "backup",
		Success:  true,
		ExitCode: 0,
		Duration: 2 * time.Second,
		Output:   "done",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.ThemeColor != "00FF00" {
		t.Errorf("expected green theme, got %s", received.ThemeColor)
	}
	if !strings.Contains(received.Summary, "backup") {
		t.Errorf("summary missing job name: %s", received.Summary)
	}
}

func TestTeamsNotifier_Failure(t *testing.T) {
	var received teamsCard
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewTeamsNotifier(ts.URL)
	err := n.Notify(JobResult{
		JobName:  "cleanup",
		Success:  false,
		ExitCode: 1,
		Duration: 500 * time.Millisecond,
		Output:   "error occurred",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.ThemeColor != "FF0000" {
		t.Errorf("expected red theme, got %s", received.ThemeColor)
	}
	if len(received.Sections) == 0 {
		t.Fatal("expected at least one section")
	}
	facts := received.Sections[0].Facts
	if len(facts) != 3 {
		t.Errorf("expected 3 facts, got %d", len(facts))
	}
}

func TestTeamsNotifier_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewTeamsNotifier(ts.URL)
	err := n.Notify(JobResult{JobName: "job", Success: false})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status code in error, got: %v", err)
	}
}

func TestTeamsNotifier_InvalidURL(t *testing.T) {
	n := NewTeamsNotifier("http://127.0.0.1:0/no-server")
	err := n.Notify(JobResult{JobName: "job", Success: true})
	if err == nil {
		t.Fatal("expected connection error")
	}
}
