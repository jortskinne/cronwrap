// Package history provides persistent run history tracking for cronwrap.
//
// Records are stored as newline-delimited JSON in a configurable file path.
// Each record captures the job label, exit status, output, and timing
// information for later review or reporting.
//
// Example usage:
//
//	store, err := history.NewStore("/var/log/cronwrap/history.jsonl")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := store.Append(record); err != nil {
//		log.Println("failed to save history:", err)
//	}
//	records, err := store.ReadAll()
package history
