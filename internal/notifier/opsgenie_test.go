package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpsGenieNotifier_Success(t *testing.T) {
	result := JobResult{
		JobName: "backup",
		Success: true,
		ExitCode: 0,
		Duration: 2 * time.Second,
		Output:   "done",
	}
	n := NewOpsGenieNotifier("key123", "ops-team")
	if err := n.Notify(result); err != nil {
		t.Fatalf("expected no error on success, got: %v", err)
	}
}

func TestOpsGenieNotifier_Failure(t *testing.T) {
	var received opsgeniePayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "GenieKey testkey" {
			t.Errorf("missing or wrong Authorization header")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := NewOpsGenieNotifier("testkey", "infra")
	n.baseURL = ts.URL

	result := JobResult{
		JobName:  "db-backup",
		Success:  false,
		ExitCode: 1,
		Duration: 5 * time.Second,
		Output:   "error: disk full",
	}
	if err := n.Notify(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Alias != "cronwrap-db-backup" {
		t.Errorf("alias = %q, want cronwrap-db-backup", received.Alias)
	}
	if received.Priority != "P2" {
		t.Errorf("priority = %q, want P2", received.Priority)
	}
	found := false
	for _, tag := range received.Tags {
		if tag == "infra" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected tag 'infra' in %v", received.Tags)
	}
}

func TestOpsGenieNotifier_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	n := NewOpsGenieNotifier("badkey", "")
	n.baseURL = ts.URL

	result := JobResult{JobName: "job", Success: false, ExitCode: 2, Duration: time.Second}
	if err := n.Notify(result); err == nil {
		t.Fatal("expected error on HTTP 401")
	}
}

func TestOpsGenieNotifier_NoTeamTag(t *testing.T) {
	var received opsgeniePayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := NewOpsGenieNotifier("key", "")
	n.baseURL = ts.URL

	n.Notify(JobResult{JobName: "job", Success: false, ExitCode: 1, Duration: time.Second})

	if len(received.Tags) != 1 || received.Tags[0] != "cronwrap" {
		t.Errorf("expected only 'cronwrap' tag, got %v", received.Tags)
	}
}
