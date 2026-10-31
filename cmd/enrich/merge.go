package main

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/openexercise/open-exercise/internal/oes"
)

const enrichedOn = "2026-10-31"

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	return strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(name), "-"), "-")
}

// group is one catalogue record: the source records whose canonical names agree.
type group struct {
	id      string
	answer  answer
	members []wger // lowest wger ID first; the first one's answer is the record's
}

func loadAnswer(id int) (answer, error) {
	var a answer
	data, err := os.ReadFile(enrichPath(id))
	if err != nil {
		return a, err
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return a, fmt.Errorf("decode %d: %w", id, err)
	}
	return a, nil
}

// buildGroups merges records with the same canonical name and sets aside non-exercises.
func buildGroups(exercises []wger) (groups []*group, rejected []wger, err error) {
	byID := map[string]*group{}
	slices.SortFunc(exercises, func(a, b wger) int { return cmp.Compare(a.ID, b.ID) })
	for _, e := range exercises {
		a, err := loadAnswer(e.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("answer for %d (%s): %w", e.ID, e.Name, err)
		}
		if !a.IsExercise {
			rejected = append(rejected, e)
			continue
		}
		id := "oec:" + slugify(a.Name)
		g, ok := byID[id]
		if !ok {
			g = &group{id: id, answer: a}
			byID[id] = g
			groups = append(groups, g)
		}
		g.members = append(g.members, e)
	}
	slices.SortFunc(groups, func(a, b *group) int { return cmp.Compare(a.id, b.id) })
	return groups, rejected, nil
}

type relationAnswer struct {
	Relations []oes.Relation `json:"relations"`
}

func relationPath(id string) string {
	return filepath.Join(relationDir, strings.TrimPrefix(id, "oec:")+".json")
}

var relationKinds = []string{"variation-of", "regression-of", "progression-of", "substitute-for", "antagonist-of", "part-of"}

// relate asks, per exercise, how it connects to others with an overlapping movement
// pattern or prime mover. It offers candidates, so a target always exists.
func relate(ctx context.Context) error {
	exercises, err := loadSnapshot()
	if err != nil {
		return err
	}
	groups, _, err := buildGroups(exercises)
	if err != nil {
		return err
	}
	client, err := newClient(ctx)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	if err := os.MkdirAll(relationDir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	const system = `You connect one exercise to others in a catalogue. Answer with JSON: {"relations":[{"type":..., "target":...}]}. Read each as: this exercise <type> target.
variation-of: the same basic movement changed in one facet; target is the plainer form.
regression-of: this is easier than the target and leads toward it. progression-of: this is harder than the target and builds on it.
substitute-for: a sensible swap for the target when the target can't be done; same pattern and main muscles.
antagonist-of: works the opposing action, as a push does a pull. part-of: this is a component of the target, as a clean is of a clean and jerk.
Use only the candidate IDs given. Give only clear relations, at most eight, none is fine. Prefer precision: leave out anything a coach would argue with.`
	var todo []*group
	for _, g := range groups {
		if _, err := os.Stat(relationPath(g.id)); err != nil {
			todo = append(todo, g)
		}
	}
	fmt.Printf("%d to do\n", len(todo))
	var failed atomic.Int64
	jobs := make(chan *group)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for g := range jobs {
				raw, err := ask(ctx, client, system, relationPrompt(g, groups))
				if err == nil {
					err = saveRelations(g, groups, raw)
				}
				if err != nil {
					failed.Add(1)
					fmt.Printf("FAILED %s: %v\n", g.id, err)
				}
			}
		})
	}
	for _, g := range todo {
		jobs <- g
	}
	close(jobs)
	wg.Wait()
	fmt.Printf("%d failed\n", failed.Load())
	return nil
}

func primeMovers(a answer) []string {
	var out []string
	for _, m := range a.Muscles {
		if m.Role == "prime-mover" {
			out = append(out, m.Muscle)
		}
	}
	return out
}

func overlaps(a, b []string) bool {
	return slices.ContainsFunc(a, func(s string) bool { return slices.Contains(b, s) })
}

func relationPrompt(g *group, groups []*group) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Exercise: %s (%s). %s\nCandidates:\n", g.answer.Name, g.id, g.answer.Summary)
	for _, o := range groups {
		if o != g && (overlaps(g.answer.Classification.MovementPatterns, o.answer.Classification.MovementPatterns) && overlaps(primeMovers(g.answer), primeMovers(o.answer))) {
			fmt.Fprintf(&b, "%s: %s\n", o.id, o.answer.Name)
		}
	}
	return b.String()
}

// saveRelations keeps only allowed kinds and existing targets, so a bad answer loses
// a link rather than failing the record.
func saveRelations(g *group, groups []*group, raw []byte) error {
	var r relationAnswer
	if err := json.Unmarshal(raw, &r); err != nil {
		return fmt.Errorf("decode relations: %w", err)
	}
	kept := []oes.Relation{}
	for _, rel := range r.Relations {
		exists := slices.ContainsFunc(groups, func(o *group) bool { return o.id == rel.Target })
		if exists && rel.Target != g.id && slices.Contains(relationKinds, rel.Type) {
			kept = append(kept, oes.Relation{Type: rel.Type, Target: rel.Target})
		}
	}
	data, err := json.Marshal(kept)
	if err != nil {
		return err
	}
	return os.WriteFile(relationPath(g.id), data, 0o644)
}

var aliasKinds = []string{"common", "abbreviation", "regional", "brand", "misspelling"}

var spdx = map[string]string{"CC-BY-SA 4": "CC-BY-SA-4.0", "CC-BY-SA 3": "CC-BY-SA-3.0", "CC0": "CC0-1.0"}

func mergeAliases(g *group) []oes.Alias {
	seen := map[string]bool{strings.ToLower(g.answer.Name): true}
	var out []oes.Alias
	add := func(a oes.Alias) {
		key := strings.ToLower(strings.TrimSpace(a.Text))
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		a.Text = strings.TrimSpace(a.Text)
		out = append(out, a)
	}
	for _, a := range g.answer.Aliases {
		if !slices.Contains(aliasKinds, a.Kind) {
			a.Kind = "common"
		}
		add(a)
	}
	for _, m := range g.members {
		for _, s := range slices.Concat(m.CuratedAliases, m.WgerAliases, []string{m.Name}) {
			add(oes.Alias{Text: s, Kind: "common"})
		}
	}
	return out
}

func merge() error {
	exercises, err := loadSnapshot()
	if err != nil {
		return err
	}
	groups, rejected, err := buildGroups(exercises)
	if err != nil {
		return err
	}
	vocabs, err := oes.LoadVocabularies(".")
	if err != nil {
		return err
	}
	muscleParents := map[string]string{}
	for _, t := range vocabs["muscles"].Terms {
		muscleParents[t.ID] = t.Parent
	}
	current := map[string]bool{}
	for _, g := range groups {
		current[g.id] = true
	}
	dir := "catalogue/exercises"
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("clear catalogue: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create catalogue: %w", err)
	}
	var report strings.Builder
	report.WriteString("# Enrichment report\n\nWhat merging did to the wger snapshot. Every line here needs a person's eye.\n\n## Merged because they got the same canonical name\n\n")
	for _, g := range groups {
		if err := writeRecord(dir, g, current, muscleParents); err != nil {
			return err
		}
		if len(g.members) > 1 {
			names := []string{}
			for _, m := range g.members {
				names = append(names, m.Name)
			}
			fmt.Fprintf(&report, "- `%s`: %s\n", g.id, strings.Join(names, "; "))
		}
	}
	report.WriteString("\n## Set aside as not being exercises\n\n")
	for _, e := range rejected {
		fmt.Fprintf(&report, "- %s (wger %d, was `%s`)\n", e.Name, e.ID, e.LegacyID)
		if err := writeRejected(dir, e); err != nil {
			return err
		}
	}
	if err := os.WriteFile("catalogue/REPORT.md", []byte(report.String()), 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	fmt.Printf("%d source records, %d exercises, %d set aside\n", len(exercises), len(groups), len(rejected))
	return nil
}

func provenanceFor(members []wger) map[string]any {
	sources := []oes.Source{}
	for _, m := range members {
		sources = append(sources, oes.Source{
			Title: "wger: " + m.Name, URL: fmt.Sprintf("https://wger.de/en/exercise/%d/view-base", m.ID),
			License: spdx[m.License], Author: m.Author,
		})
	}
	return map[string]any{"license": "CC-BY-SA-4.0", "sources": sources}
}

// dropParents removes a muscle when one of its descendants is also listed, as the spec requires.
func dropParents(muscles []oes.MuscleUse, parents map[string]string) []oes.MuscleUse {
	covered := map[string]bool{}
	for _, m := range muscles {
		for p := parents[m.Muscle]; p != ""; p = parents[p] {
			covered[p] = true
		}
	}
	return slices.DeleteFunc(slices.Clone(muscles), func(m oes.MuscleUse) bool { return covered[m.Muscle] })
}

func writeRecord(dir string, g *group, current map[string]bool, muscleParents map[string]string) error {
	var relations []oes.Relation
	if data, err := os.ReadFile(relationPath(g.id)); err == nil {
		if err := json.Unmarshal(data, &relations); err != nil {
			return fmt.Errorf("decode relations of %s: %w", g.id, err)
		}
	}
	g.answer.Muscles = dropParents(g.answer.Muscles, muscleParents)
	doc := g.answer.record(g.id, relations)
	var previous []string
	ids := []string{}
	for _, m := range g.members {
		ids = append(ids, strconv.Itoa(m.ID))
		// A legacy ID that is now another exercise's current ID can't resolve here.
		if m.LegacyID != g.id && !current[m.LegacyID] {
			previous = append(previous, m.LegacyID)
		}
	}
	if len(previous) > 0 {
		doc["previous_ids"] = previous
	}
	if aliases := mergeAliases(g); len(aliases) > 0 {
		doc["aliases"] = aliases
	}
	doc["external_ids"] = map[string]string{"wger": strings.Join(ids, ",")}
	generated := []string{"/name", "/summary", "/classification", "/anatomy", "/equipment", "/measurement", "/instructions"}
	if _, ok := doc["setup"]; ok {
		generated = append(generated, "/setup")
	}
	if len(relations) > 0 {
		generated = append(generated, "/relations")
	}
	prov := provenanceFor(g.members)
	prov["review"] = oes.Review{Status: "machine-generated", Model: model, At: enrichedOn}
	prov["generated"] = generated
	doc["provenance"] = prov
	return writeDoc(dir, g.id, doc)
}

// writeRejected keeps a deprecated record, so an old ID still resolves.
func writeRejected(dir string, e wger) error {
	doc := map[string]any{
		"$schema": "https://openexercise.org/schema/v0/exercise.json",
		"id":      e.LegacyID, "name": e.Name, "status": "deprecated",
		"measurement":  map[string]any{"profiles": []oes.Profile{{Required: []string{"reps"}}}},
		"external_ids": map[string]string{"wger": strconv.Itoa(e.ID)},
		"provenance":   provenanceFor([]wger{e}),
	}
	return writeDoc(dir, e.LegacyID, doc)
}

func writeDoc(dir, id string, doc map[string]any) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", id, err)
	}
	path := filepath.Join(dir, strings.TrimPrefix(id, "oec:")+".json")
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
