package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/cronwrap/internal/notifier"
	"github.com/yourorg/cronwrap/internal/runner"
)

func TestSentryNotifier_Success_NoEvent(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notifier.NewSentryNotifier(ts.URL, "production", "v1.0")
	err := n.Notify(runner.Result{Command: "echo hi", ExitCode: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected no HTTP call for successful job")
	}
}

func TestSentryNotifier_Failure_SendsEvent(t *testing.T) {
	var captured map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notifier.NewSentryNotifier(ts.URL, "staging", "v1.2")
	err := n.Notify(runner.Result{
		Command:  "backup.sh",
		ExitCode: 1,
		Stderr:   "disk full",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured["level"] != "error" {
		t.Errorf("expected level=error, got %v", captured["level"])
	}
	if captured["environment"] != "staging" {
		t.Errorf("expected environment=staging, got %v", captured["environment"])
	}
	if captured["release"] != "v1.2" {
		t.Errorf("expected release=v1.2, got %v", captured["release"])
	}
}

func TestSentryNotifier_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notifier.NewSentryNotifier(ts.URL, "", "")
	err := n.Notify(runner.Result{Command: "fail.sh", ExitCode: 2})
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestSentryNotifier_InvalidURL(t *testing.T) {
	n := notifier.NewSentryNotifier("://bad-url", "", "")
	err := n.Notify(runner.Result{Command: "x", ExitCode: 1})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}
