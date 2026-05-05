package history

import (
	"testing"
	"time"
)

func baseTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestSummarize_Empty(t *testing.T) {
	s := Summarize(nil)
	if s.Total != 0 {
		t.Errorf("expected Total=0, got %d", s.Total)
	}
}

func TestSummarize_AllSuccess(t *testing.T) {
	records := []Record{
		{ExitCode: 0, Duration: 2 * time.Second, StartedAt: baseTime()},
		{ExitCode: 0, Duration: 4 * time.Second, StartedAt: baseTime().Add(time.Hour)},
	}
	s := Summarize(records)
	if s.Total != 2 {
		t.Errorf("expected Total=2, got %d", s.Total)
	}
	if s.Successes != 2 {
		t.Errorf("expected Successes=2, got %d", s.Successes)
	}
	if s.Failures != 0 {
		t.Errorf("expected Failures=0, got %d", s.Failures)
	}
	if s.LastStatus != "success" {
		t.Errorf("expected LastStatus=success, got %s", s.LastStatus)
	}
	if s.AvgRuntime != 3*time.Second {
		t.Errorf("expected AvgRuntime=3s, got %v", s.AvgRuntime)
	}
}

func TestSummarize_MixedResults(t *testing.T) {
	records := []Record{
		{ExitCode: 0, Duration: 1 * time.Second, StartedAt: baseTime()},
		{ExitCode: 1, Duration: 3 * time.Second, StartedAt: baseTime().Add(time.Hour)},
	}
	s := Summarize(records)
	if s.Successes != 1 || s.Failures != 1 {
		t.Errorf("expected 1 success and 1 failure, got %d/%d", s.Successes, s.Failures)
	}
	if s.LastStatus != "failure" {
		t.Errorf("expected LastStatus=failure, got %s", s.LastStatus)
	}
	if !s.LastRun.Equal(baseTime().Add(time.Hour)) {
		t.Errorf("unexpected LastRun: %v", s.LastRun)
	}
}

func TestSummarize_SingleRecord(t *testing.T) {
	records := []Record{
		{ExitCode: 2, Duration: 5 * time.Second, StartedAt: baseTime()},
	}
	s := Summarize(records)
	if s.Total != 1 {
		t.Errorf("expected Total=1, got %d", s.Total)
	}
	if s.AvgRuntime != 5*time.Second {
		t.Errorf("expected AvgRuntime=5s, got %v", s.AvgRuntime)
	}
}
