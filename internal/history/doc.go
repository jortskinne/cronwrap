// Package history implements run history tracking for cronwrap jobs.
//
// Records are stored in a newline-delimited JSON (JSONL) file, with one
// JSON object per line. Each record captures the job name, command,
// start time, duration, exit code, and captured output.
//
// Example usage:
//
//	store, err := history.NewStore("/var/lib/cronwrap/history.jsonl")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = store.Append(history.Record{
//		JobName:  "nightly-backup",
//		Command:  "./backup.sh",
//		Success:  true,
//		ExitCode: 0,
//	})
//
// Records can be retrieved with ReadAll for reporting or auditing.
package history
