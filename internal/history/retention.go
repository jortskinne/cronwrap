package history

import (
	"time"
)

// RetentionPolicy defines rules for purging old history records.
type RetentionPolicy struct {
	// MaxAge is the maximum age of records to keep.
	// Records older than this duration are removed.
	// Zero means no age-based pruning.
	MaxAge time.Duration

	// MaxRecords is the maximum number of records to retain.
	// Zero means no count-based pruning.
	MaxRecords int
}

// Apply removes records from the given slice that violate the policy.
// Age-based pruning is applied first, then count-based pruning.
func (p RetentionPolicy) Apply(records []Record) []Record {
	if p.MaxAge <= 0 && p.MaxRecords <= 0 {
		return records
	}

	result := records

	if p.MaxAge > 0 {
		cutoff := time.Now().Add(-p.MaxAge)
		filtered := result[:0]
		for _, r := range result {
			if r.StartedAt.After(cutoff) {
				filtered = append(filtered, r)
			}
		}
		result = filtered
	}

	if p.MaxRecords > 0 && len(result) > p.MaxRecords {
		result = result[len(result)-p.MaxRecords:]
	}

	return result
}

// ApplyToStore reads all records from the store, applies the retention policy,
// and rewrites the store with the filtered records.
func ApplyToStore(s *Store, policy RetentionPolicy) (int, error) {
	records, err := s.ReadAll()
	if err != nil {
		return 0, err
	}

	origLen := len(records)
	pruned := policy.Apply(records)

	if len(pruned) == origLen {
		return 0, nil
	}

	if err := s.Rewrite(pruned); err != nil {
		return 0, err
	}

	return origLen - len(pruned), nil
}
