package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// discarded finds every record merging would change or remove, so merge refuses to overwrite
// hand edits; records it would write unchanged, and new ones, lose nothing.
func TestDiscardedFindsEditedAndRemovedRecords(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"same.json":    "{}\n",
		"edited.json":  "{\"name\": \"by hand\"}\n",
		"removed.json": "{}\n",
		"notes.txt":    "not a record",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string][]byte{
		"same.json":   []byte("{}\n"),
		"edited.json": []byte("{\"name\": \"generated\"}\n"),
		"new.json":    []byte("{}\n"),
	}
	lost, err := discarded(dir, files)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(lost)
	if want := []string{"edited", "removed"}; !slices.Equal(lost, want) {
		t.Errorf("discarded = %v, want %v", lost, want)
	}
	if lost, err := discarded(filepath.Join(dir, "missing"), files); err != nil || len(lost) != 0 {
		t.Errorf("no catalogue yet: %v %v, want nothing lost", lost, err)
	}
}

func TestSample(t *testing.T) {
	if got := sample([]string{"a", "b"}); got != "a, b" {
		t.Errorf("sample = %q", got)
	}
	if got := sample([]string{"a", "b", "c", "d", "e"}); got != "a, b, c and 2 more" {
		t.Errorf("sample = %q", got)
	}
}
