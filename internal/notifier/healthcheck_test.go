package notifier

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/cronwrap/internal/runner"
)

func successHCResult() runner.Result {
	return runner.Result{Command: "echo hi", Success: true, ExitCode: 0}
}

func failHCResult() runner.Result {
	return runner.Result{Command: "false", Success: false, ExitCode: 1}
}

func TestHealthCheckNotifier_Success(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewHealthCheckNotifier(srv.URL, 5*time.Second)
	if err := n.Notify(successHCResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/success") {
		t.Errorf("expected /success ping, got %q", gotPath)
	}
}

func TestHealthCheckNotifier_Failure(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewHealthCheckNotifier(srv.URL, 5*time.Second)
	if err := n.Notify(failHCResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/fail") {
		t.Errorf("expected /fail ping, got %q", gotPath)
	}
}

func TestHealthCheckNotifier_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := NewHealthCheckNotifier(srv.URL, 5*time.Second)
	err := n.Notify(successHCResult())
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status code in error, got: %v", err)
	}
}

func TestHealthCheckNotifier_EmptyBaseURL(t *testing.T) {
	n := NewHealthCheckNotifier("", 5*time.Second)
	if err := n.Notify(successHCResult()); err != nil {
		t.Fatalf("expected no-op for empty URL, got: %v", err)
	}
}

func TestHealthCheckNotifier_InvalidURL(t *testing.T) {
	n := NewHealthCheckNotifier("http://127.0.0.1:0", time.Second)
	err := n.Notify(successHCResult())
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
