package oes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validDoc() map[string]any {
	return map[string]any{
		"id": "oec:test-press", "name": "Test Press",
		"measurement": map[string]any{
			"profiles": []any{map[string]any{"required": []any{"reps", "load"}}},
			"load":     map[string]any{"kind": "external", "convention": "total", "implements": 1},
		},
	}
}

func TestValidate(t *testing.T) {
	v, err := NewValidator("../..")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		change func(d map[string]any)
		want   string
	}{
		{"valid", func(map[string]any) {}, ""},
		{"bad id", func(d map[string]any) { d["id"] = "Bench Press" }, "/id"},
		{"unknown field", func(d map[string]any) { d["nope"] = 1 }, "nope"},
		{"extension allowed", func(d map[string]any) { d["x-acme"] = 1 }, ""},
		{"load without meaning", func(d map[string]any) {
			delete(d["measurement"].(map[string]any), "load")
		}, "must say what the number means"},
		{"bodyweight with load", func(d map[string]any) {
			d["measurement"].(map[string]any)["load"] = map[string]any{"kind": "bodyweight"}
		}, "takes no number"},
		{"per-implement needs a count", func(d map[string]any) {
			d["measurement"].(map[string]any)["load"] = map[string]any{"kind": "external", "convention": "per-implement", "implements": 1}
		}, "implements of 2"},
		{"unknown muscle", func(d map[string]any) {
			d["anatomy"] = map[string]any{"muscles": []any{map[string]any{"muscle": "pecs", "role": "prime-mover"}}}
		}, "vocabulary muscles"},
		{"private term allowed", func(d map[string]any) {
			d["anatomy"] = map[string]any{"muscles": []any{map[string]any{"muscle": "x-pecs", "role": "prime-mover"}}}
		}, ""},
		{"muscle and its parent", func(d map[string]any) {
			d["anatomy"] = map[string]any{"muscles": []any{
				map[string]any{"muscle": "biceps-long-head", "role": "prime-mover"},
				map[string]any{"muscle": "biceps-brachii", "role": "prime-mover"},
			}}
		}, "and its parent"},
		{"e1rm needs external load", func(d map[string]any) {
			m := d["measurement"].(map[string]any)
			m["load"] = map[string]any{"kind": "bodyweight-plus-external", "convention": "total"}
			m["records"] = map[string]any{"e1rm": true}
		}, "e1rm"},
		{"replaced needs target", func(d map[string]any) { d["status"] = "replaced" }, "replaced_by"},
		{"generated pointer must exist", func(d map[string]any) {
			d["provenance"] = map[string]any{"generated": []any{"/summary"}}
		}, "doesn't exist"},
		{"self relation", func(d map[string]any) {
			d["relations"] = []any{map[string]any{"type": "same-as", "target": "oec:test-press"}}
		}, "itself"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := validDoc()
			tt.change(d)
			data, err := json.Marshal(d)
			if err != nil {
				t.Fatal(err)
			}
			got := strings.Join(v.Validate(data), "; ")
			if tt.want == "" && got != "" {
				t.Fatalf("want valid, got %s", got)
			}
			if !strings.Contains(got, tt.want) {
				t.Fatalf("want %q in %q", tt.want, got)
			}
		})
	}
}

func TestExamplesAreValid(t *testing.T) {
	v, err := NewValidator("../..")
	if err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob("../../examples/*.json")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no examples: %v", err)
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if problems := v.Validate(data); len(problems) > 0 {
			t.Errorf("%s: %v", p, problems)
		}
	}
}

func TestCheckCatalogue(t *testing.T) {
	all := []Exercise{
		{ID: "oec:a", PreviousIDs: []string{"oec:b"}, Relations: []Relation{{Type: "same-as", Target: "oec:gone"}}},
		{ID: "oec:b"},
		{ID: "oec:b"},
	}
	got := strings.Join(CheckCatalogue(all), "; ")
	for _, want := range []string{"duplicate id oec:b", "current id", "oec:gone"} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q in %q", want, got)
		}
	}
}

func TestVocabularyParentsExist(t *testing.T) {
	vocabs, err := LoadVocabularies("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(vocabs) != 7 {
		t.Fatalf("got %d vocabularies", len(vocabs))
	}
	for _, v := range vocabs {
		ids := map[string]bool{}
		for _, term := range v.Terms {
			ids[term.ID] = true
		}
		for _, term := range v.Terms {
			if term.Parent != "" && !ids[term.Parent] {
				t.Errorf("%s: %s has unknown parent %s", v.Name, term.ID, term.Parent)
			}
		}
	}
}

// CheckAgainst checks some records with the catalogue they belong with: targets may be catalogue
// records, a checked record stands in for the catalogue's with its ID, and only problems with the
// checked records are reported.
func TestCheckAgainst(t *testing.T) {
	known := []Exercise{
		{ID: "oec:a"},
		{ID: "oec:b", PreviousIDs: []string{"oec:old-b"}},
		{ID: "oec:broken", Relations: []Relation{{Type: "variation-of", Target: "oec:missing"}}},
	}
	tests := []struct {
		name    string
		checked []Exercise
		want    []string
	}{
		{"own record relating to the catalogue", []Exercise{{ID: "acme:x", Relations: []Relation{{Type: "variation-of", Target: "oec:a"}}}}, nil},
		{"a catalogue record checked alone, edited", []Exercise{{ID: "oec:b", PreviousIDs: []string{"oec:old-b"}, ReplacedBy: "oec:a"}}, nil},
		{"a missing target", []Exercise{{ID: "acme:x", Relations: []Relation{{Type: "variation-of", Target: "oec:nope"}}}},
			[]string{"acme:x: relation target oec:nope doesn't exist"}},
		{"a missing replaced_by", []Exercise{{ID: "acme:x", ReplacedBy: "oec:nope"}}, []string{"acme:x: replaced_by oec:nope doesn't exist"}},
		{"a previous id that's a catalogue id", []Exercise{{ID: "acme:x", PreviousIDs: []string{"oec:a"}}},
			[]string{"acme:x lists previous id oec:a, which is a current id"}},
		{"a previous id the catalogue claims", []Exercise{{ID: "acme:x", PreviousIDs: []string{"oec:old-b"}}},
			[]string{"previous id oec:old-b is claimed by oec:b and acme:x"}},
		{"an id the catalogue lists as previous", []Exercise{{ID: "oec:old-b"}}, []string{"oec:old-b is a previous id of oec:b"}},
		{"a loop through the catalogue", []Exercise{{ID: "oec:a", ReplacedBy: "oec:c"}, {ID: "oec:c", ReplacedBy: "oec:a"}},
			[]string{"oec:a: replaced_by loops back through oec:a", "oec:c: replaced_by loops back through oec:c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckAgainst(tt.checked, known)
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
