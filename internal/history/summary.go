package history

import "time"

// Summary holds aggregated statistics derived from a slice of Records.
type Summary struct {
	Total      int
	Successes  int
	Failures   int
	LastRun    time.Time
	LastStatus string
	AvgRuntime time.Duration
}

// Summarize computes a Summary from the provided records.
// Records are assumed to be in append order (oldest first).
func Summarize(records []Record) Summary {
	if len(records) == 0 {
		return Summary{}
	}

	s := Summary{
		Total: len(records),
	}

	var totalDuration time.Duration

	for _, r := range records {
		if r.ExitCode == 0 {
			s.Successes++
		} else {
			s.Failures++
		}
		totalDuration += r.Duration
	}

	last := records[len(records)-1]
	s.LastRun = last.StartedAt
	if last.ExitCode == 0 {
		s.LastStatus = "success"
	} else {
		s.LastStatus = "failure"
	}

	s.AvgRuntime = totalDuration / time.Duration(len(records))

	return s
}
