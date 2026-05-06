package notifier

import (
	"errors"
	"testing"
)

// stubNotifier records every Notify call so tests can inspect them.
type stubNotifier struct {
	calls []JobResult
	err   error
}

func (s *stubNotifier) Notify(r JobResult) error {
	s.calls = append(s.calls, r)
	return s.err
}

func result(code int) JobResult {
	return JobResult{ExitCode: code, Command: "echo hi"}
}

func TestFilteredNotifier_FailurePolicy_Failure(t *testing.T) {
	stub := &stubNotifier{}
	f := NewFilteredNotifier(stub, NotifyOnFailure)

	if err := f.Notify(result(1)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(stub.calls))
	}
}

func TestFilteredNotifier_FailurePolicy_Success(t *testing.T) {
	stub := &stubNotifier{}
	f := NewFilteredNotifier(stub, NotifyOnFailure)

	if err := f.Notify(result(0)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.calls) != 0 {
		t.Fatalf("expected 0 calls, got %d", len(stub.calls))
	}
}

func TestFilteredNotifier_AlwaysPolicy(t *testing.T) {
	stub := &stubNotifier{}
	f := NewFilteredNotifier(stub, NotifyOnAlways)

	_ = f.Notify(result(0))
	_ = f.Notify(result(1))

	if len(stub.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(stub.calls))
	}
}

func TestFilteredNotifier_SuccessPolicy_Success(t *testing.T) {
	stub := &stubNotifier{}
	f := NewFilteredNotifier(stub, NotifyOnSuccess)

	_ = f.Notify(result(0))
	if len(stub.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(stub.calls))
	}
}

func TestFilteredNotifier_SuccessPolicy_Failure(t *testing.T) {
	stub := &stubNotifier{}
	f := NewFilteredNotifier(stub, NotifyOnSuccess)

	_ = f.Notify(result(2))
	if len(stub.calls) != 0 {
		t.Fatalf("expected 0 calls, got %d", len(stub.calls))
	}
}

func TestFilteredNotifier_PropagatesError(t *testing.T) {
	want := errors.New("send failed")
	stub := &stubNotifier{err: want}
	f := NewFilteredNotifier(stub, NotifyOnAlways)

	if err := f.Notify(result(1)); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestParseNotifyOn(t *testing.T) {
	cases := []struct {
		input string
		want  NotifyOn
	}{
		{"failure", NotifyOnFailure},
		{"always", NotifyOnAlways},
		{"success", NotifyOnSuccess},
		{"", NotifyOnFailure},
		{"unknown", NotifyOnFailure},
	}
	for _, c := range cases {
		if got := ParseNotifyOn(c.input); got != c.want {
			t.Errorf("ParseNotifyOn(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
