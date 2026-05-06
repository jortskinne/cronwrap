package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPagerDutyNotifier_Trigger(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := NewPagerDutyNotifier("test-key")
	n.eventsURL = ts.URL

	result := JobResult{
		JobName:   "backup",
		Success:   false,
		ExitCode:  1,
		Output:    "disk full",
		StartedAt: time.Now(),
	}

	if err := n.Notify(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.EventAction != "trigger" {
		t.Errorf("expected trigger, got %s", received.EventAction)
	}
	if received.Payload.Severity != "critical" {
		t.Errorf("expected critical severity, got %s", received.Payload.Severity)
	}
	if received.RoutingKey != "test-key" {
		t.Errorf("expected routing key test-key, got %s", received.RoutingKey)
	}
}

func TestPagerDutyNotifier_Resolve(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := NewPagerDutyNotifier("key")
	n.eventsURL = ts.URL

	result := JobResult{JobName: "cleanup", Success: true, StartedAt: time.Now()}
	if err := n.Notify(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.EventAction != "resolve" {
		t.Errorf("expected resolve, got %s", received.EventAction)
	}
	if received.Payload.Severity != "info" {
		t.Errorf("expected info severity, got %s", received.Payload.Severity)
	}
}

func TestPagerDutyNotifier_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewPagerDutyNotifier("key")
	n.eventsURL = ts.URL

	err := n.Notify(JobResult{JobName: "job", StartedAt: time.Now()})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestPagerDutyNotifier_DedupKey(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := NewPagerDutyNotifier("key")
	n.eventsURL = ts.URL

	result := JobResult{JobName: "my-job", StartedAt: time.Now()}
	n.Notify(result)

	if received.DedupKey != "cronwrap-my-job" {
		t.Errorf("expected dedup key cronwrap-my-job, got %s", received.DedupKey)
	}
}
