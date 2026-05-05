// Package history provides persistent run-history storage and reporting for
// cronwrap.
//
// Records are appended to a newline-delimited JSON file on disk.  The package
// exposes helpers to:
//
//   - Append a new Record after each job execution (Store.Append).
//   - Read all stored records back (Store.ReadAll).
//   - Prune the file to a maximum number of entries (Prune).
//   - Compute aggregate statistics over a slice of records (Summarize).
//   - Render a human-readable tabular report to any io.Writer (PrintReport).
package history
