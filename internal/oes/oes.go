// Package oes validates Open Exercise Schema documents: the JSON Schema, then the
// rules a schema can't state (vocabulary terms, load sense, relation targets).
package oes

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Term is one vocabulary entry.
type Term struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Parent     string `json:"parent,omitempty"`
	Definition string `json:"definition,omitempty"`
}

// Vocabulary is one file in vocab/v0.
type Vocabulary struct {
	Name        string `json:"vocabulary"`
	Description string `json:"description"`
	Terms       []Term `json:"terms"`
}

// Validator checks documents against the schema and vocabularies in a repository root.
type Validator struct {
	schema *jsonschema.Schema
	Vocabs map[string]Vocabulary
}

// NewValidator loads the schema and vocabularies under root.
func NewValidator(root string) (*Validator, error) {
	path := filepath.Join(root, "schema", "v0", "exercise.json")
	c := jsonschema.NewCompiler()
	schema, err := c.Compile(path)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	vocabs, err := LoadVocabularies(root)
	if err != nil {
		return nil, err
	}
	return &Validator{schema: schema, Vocabs: vocabs}, nil
}

// LoadVocabularies reads every file in vocab/v0, keyed by vocabulary name.
func LoadVocabularies(root string) (map[string]Vocabulary, error) {
	paths, err := filepath.Glob(filepath.Join(root, "vocab", "v0", "*.json"))
	if err != nil {
		return nil, fmt.Errorf("list vocabularies: %w", err)
	}
	vocabs := make(map[string]Vocabulary, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read vocabulary: %w", err)
		}
		var v Vocabulary
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, fmt.Errorf("decode %s: %w", p, err)
		}
		vocabs[v.Name] = v
	}
	return vocabs, nil
}

func (v *Validator) hasTerm(vocab, term string) bool {
	if strings.HasPrefix(term, "x-") {
		return true
	}
	return slices.ContainsFunc(v.Vocabs[vocab].Terms, func(t Term) bool { return t.ID == term })
}

// Validate returns every problem in one document.
func (v *Validator) Validate(data []byte) []string {
	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return []string{"invalid JSON: " + err.Error()}
	}
	if err := v.schema.Validate(doc); err != nil {
		var ve *jsonschema.ValidationError
		if errors.As(err, &ve) {
			return []string{ve.Error()}
		}
		return []string{err.Error()}
	}
	var e Exercise
	if err := json.Unmarshal(data, &e); err != nil {
		return []string{"decode: " + err.Error()}
	}
	return append(v.checkTerms(e), checkSense(e, doc)...)
}

func (v *Validator) checkTerms(e Exercise) []string {
	var out []string
	check := func(vocab string, terms ...string) {
		for _, t := range terms {
			if !v.hasTerm(vocab, t) {
				out = append(out, fmt.Sprintf("%q is not in vocabulary %s", t, vocab))
			}
		}
	}
	parents := map[string]string{}
	for _, t := range v.Vocabs["muscles"].Terms {
		parents[t.ID] = t.Parent
	}
	for _, m := range e.Anatomy.Muscles {
		for p := parents[m.Muscle]; p != ""; p = parents[p] {
			if slices.ContainsFunc(e.Anatomy.Muscles, func(o MuscleUse) bool { return o.Muscle == p }) {
				out = append(out, fmt.Sprintf("muscles lists both %s and its parent %s", m.Muscle, p))
			}
		}
	}
	check("modalities", e.Classification.Modalities...)
	check("movement-patterns", e.Classification.MovementPatterns...)
	for _, m := range e.Anatomy.Muscles {
		check("muscles", m.Muscle)
	}
	for _, j := range e.Anatomy.Joints {
		check("joints", j.Joint)
		check("joint-actions", j.Actions...)
	}
	for _, group := range e.Equipment.Required {
		check("equipment", group...)
	}
	check("equipment", e.Equipment.Optional...)
	for _, p := range e.Measurement.Profiles {
		check("metrics", p.Required...)
		check("metrics", p.Optional...)
	}
	return out
}

func checkSense(e Exercise, doc any) []string {
	var out []string
	usesLoad := false
	for _, p := range e.Measurement.Profiles {
		usesLoad = usesLoad || slices.Contains(p.Required, "load") || slices.Contains(p.Optional, "load")
	}
	if usesLoad && e.Measurement.Load == nil {
		out = append(out, "a profile records load, so measurement.load must say what the number means")
	}
	if l := e.Measurement.Load; l != nil {
		if l.Kind == "none" || l.Kind == "bodyweight" {
			if usesLoad {
				out = append(out, fmt.Sprintf("load kind %s takes no number, but a profile records load", l.Kind))
			}
		} else if l.Convention == "" {
			out = append(out, fmt.Sprintf("load kind %s needs a convention", l.Kind))
		}
		if l.Kind == "external" && (l.Convention == "" || l.Convention == "not-applicable") {
			out = append(out, "external load needs a real convention")
		}
		if l.Convention == "per-implement" && l.Implements < 2 {
			out = append(out, "per-implement load needs implements of 2 or more")
		}
	}
	if e.Measurement.Records.E1RM && (!usesLoad || e.Measurement.Load == nil || e.Measurement.Load.Kind != "external") {
		out = append(out, "e1rm needs an external load")
	}
	seen := map[string]bool{}
	for _, m := range e.Anatomy.Muscles {
		if seen[m.Muscle] {
			out = append(out, fmt.Sprintf("muscle %s is listed twice", m.Muscle))
		}
		seen[m.Muscle] = true
	}
	for _, r := range e.Relations {
		if r.Target == e.ID {
			out = append(out, "a relation points at the exercise itself")
		}
	}
	if slices.Contains(e.PreviousIDs, e.ID) {
		out = append(out, "previous_ids contains the current id")
	}
	for _, p := range e.Provenance.Generated {
		if !pointerResolves(doc, p) {
			out = append(out, fmt.Sprintf("provenance.generated names %q, which doesn't exist", p))
		}
	}
	// A machine-verified record names its pipeline, day and at least two models (SPEC §13).
	if r := e.Provenance.Review; r != nil && r.Status == MachineVerified {
		if r.By == "" || r.At == "" {
			out = append(out, "review.status machine-verified needs by and at")
		}
		if distinct(r.Models) < 2 {
			out = append(out, "review.status machine-verified needs at least two different models")
		}
	}
	// A person who reviewed a record checked every value and says who and when (SPEC §13).
	if r := e.Provenance.Review; isReviewed(r) {
		if len(e.Provenance.Generated) > 0 {
			out = append(out, fmt.Sprintf("review.status is %s, but provenance.generated still lists unchecked values", r.Status))
		}
		if r.By == "" || r.At == "" {
			out = append(out, fmt.Sprintf("review.status %s needs by and at", r.Status))
		}
	}
	return out
}

func pointerResolves(doc any, pointer string) bool {
	cur := doc
	for _, part := range strings.Split(pointer, "/")[1:] {
		part = strings.NewReplacer("~1", "/", "~0", "~").Replace(part)
		switch node := cur.(type) {
		case map[string]any:
			next, ok := node[part]
			if !ok {
				return false
			}
			cur = next
		case []any:
			var i int
			if _, err := fmt.Sscanf(part, "%d", &i); err != nil || i < 0 || i >= len(node) {
				return false
			}
			cur = node[i]
		default:
			return false
		}
	}
	return true
}

// CheckCatalogue checks rules across records: unique IDs, and relation targets and
// replaced_by that exist, with no replaced_by loop. An ID may not be another record's previous ID.
func CheckCatalogue(all []Exercise) []string {
	return CheckAgainst(all, nil)
}

// CheckAgainst checks the same rules for some records alongside a catalogue they belong with:
// a relation target or replaced_by may be a catalogue record, and previous IDs may not clash with
// one. A checked record replaces the catalogue's record with the same ID, so a file from the
// catalogue can be checked on its own after an edit. Only problems with the checked records are
// reported.
func CheckAgainst(checked, known []Exercise) []string {
	var out []string
	ids := make(map[string]bool, len(checked))
	for _, e := range checked {
		if ids[e.ID] {
			out = append(out, "duplicate id "+e.ID)
		}
		ids[e.ID] = true
	}
	// Every record, the checked ones in place of the catalogue's with the same ID.
	byID := make(map[string]Exercise, len(known)+len(checked))
	for _, e := range known {
		byID[e.ID] = e
	}
	for _, e := range checked {
		byID[e.ID] = e
	}
	claimed := map[string][]string{} // previous ID -> the records listing it
	for _, id := range slices.Sorted(maps.Keys(byID)) {
		for _, p := range byID[id].PreviousIDs {
			claimed[p] = append(claimed[p], id)
		}
	}
	for _, e := range checked {
		for _, p := range e.PreviousIDs {
			if _, current := byID[p]; current && p != e.ID {
				out = append(out, fmt.Sprintf("%s lists previous id %s, which is a current id", e.ID, p))
			}
			for _, other := range claimed[p] {
				// A pair of checked records is reported once.
				if other != e.ID && (!ids[other] || other < e.ID) {
					out = append(out, fmt.Sprintf("previous id %s is claimed by %s and %s", p, other, e.ID))
					break
				}
			}
		}
		for _, other := range claimed[e.ID] {
			if other != e.ID && !ids[other] {
				out = append(out, fmt.Sprintf("%s is a previous id of %s", e.ID, other))
			}
		}
		for _, r := range e.Relations {
			if _, ok := byID[r.Target]; !ok {
				out = append(out, fmt.Sprintf("%s: relation target %s doesn't exist", e.ID, r.Target))
			}
		}
		if _, ok := byID[e.ReplacedBy]; e.ReplacedBy != "" && !ok {
			out = append(out, fmt.Sprintf("%s: replaced_by %s doesn't exist", e.ID, e.ReplacedBy))
		}
	}
	// A replaced_by chain ends: a consumer follows it to the record to use.
	for _, e := range checked {
		seen := map[string]bool{e.ID: true}
		for at := byID[e.ID].ReplacedBy; at != ""; at = byID[at].ReplacedBy {
			if seen[at] {
				out = append(out, fmt.Sprintf("%s: replaced_by loops back through %s", e.ID, at))
				break
			}
			seen[at] = true
		}
	}
	return out
}
