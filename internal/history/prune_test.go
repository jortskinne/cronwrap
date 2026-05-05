package history

import (
	"os"
	"testing"
	"time"
)

func TestPrune_ReducesToMax(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/history.jsonl"

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	base := time.Now()
	for i := 0; i < 10; i++ {
		r := Record{
			Label:     "job",
			StartedAt: base.Add(time.Duration(i) * time.Minute),
			Success:   true,
		}
		if err := store.Append(r); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	if err := store.Prune(5); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	records, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 5 {
		t.Errorf("expected 5 records after prune, got %d", len(records))
	}
	// Newest 5 should be kept (indices 5-9).
	for _, r := range records {
		if r.StartedAt.Before(base.Add(5 * time.Minute)) {
			t.Errorf("old record not pruned: %v", r.StartedAt)
		}
	}
}

func TestPrune_NoOpWhenUnderLimit(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir + "/h.jsonl")

	for i := 0; i < 3; i++ {
		_ = store.Append(Record{Label: "job", StartedAt: time.Now()})
	}

	if err := store.Prune(10); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	records, _ := store.ReadAll()
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}

func TestPrune_ZeroIsNoOp(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir + "/h.jsonl")
	_ = store.Append(Record{Label: "job", StartedAt: time.Now()})

	if err := store.Prune(0); err != nil {
		t.Fatalf("Prune(0) should be no-op, got: %v", err)
	}
}

func TestPrune_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/h.jsonl"
	os.WriteFile(path, []byte{}, 0644)
	store, _ := NewStore(path)

	if err := store.Prune(5); err != nil {
		t.Fatalf("Prune on empty file: %v", err)
	}
}
