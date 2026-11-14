package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"maps"
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

const catalogueDir = "catalogue/exercises"

// merge rebuilds every record from the snapshot, the model's answers and the relations. It works
// out each file first and writes only when that loses nothing: a record changed or removed since the
// last merge is a person's edit (a review, a correction), and merging would discard it.
func merge(args []string) error {
	flags := flag.NewFlagSet("merge", flag.ContinueOnError)
	out := flags.String("out", "", "write the records and REPORT.md under this directory instead of the catalogue, to compare")
	force := flags.Bool("force", false, "overwrite catalogue/exercises even though that discards edits made since the last merge")
	if err := flags.Parse(args); err != nil {
		return err
	}
	exercises, err := loadSnapshot()
	if err != nil {
		return err
	}
	files, report, groups, rejected, err := build(exercises)
	if err != nil {
		return err
	}
	dir, reportPath := catalogueDir, "catalogue/REPORT.md"
	if *out != "" {
		dir, reportPath = filepath.Join(*out, "exercises"), filepath.Join(*out, "REPORT.md")
	} else if !*force {
		lost, err := discarded(catalogueDir, files)
		if err != nil {
			return err
		}
		if len(lost) > 0 {
			return fmt.Errorf("merge would discard edits to %d record(s) in %s (%s); nothing was written.\n"+
				"Records are edited by hand now, so merge no longer owns them. To see what it would make,\n"+
				"run it with -out DIR and diff. To overwrite anyway, pass -force", len(lost), catalogueDir, sample(lost))
		}
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("clear %s: %w", dir, err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	for _, name := range slices.Sorted(maps.Keys(files)) {
		if err := os.WriteFile(filepath.Join(dir, name), files[name], 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	if err := os.WriteFile(reportPath, report, 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	fmt.Printf("%d source records, %d exercises, %d set aside, written to %s\n", len(exercises), groups, rejected, dir)
	return nil
}

// build works out every record file, by name, and the report, without writing anything.
func build(exercises []wger) (files map[string][]byte, report []byte, groups, rejected int, err error) {
	gs, rej, err := buildGroups(exercises)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	vocabs, err := oes.LoadVocabularies(".")
	if err != nil {
		return nil, nil, 0, 0, err
	}
	muscleParents := map[string]string{}
	for _, t := range vocabs["muscles"].Terms {
		muscleParents[t.ID] = t.Parent
	}
	current := map[string]bool{}
	for _, g := range gs {
		current[g.id] = true
	}
	files = map[string][]byte{}
	var b strings.Builder
	b.WriteString("# Enrichment report\n\nWhat merging did to the wger snapshot. Every line here needs a person's eye.\n\n## Merged because they got the same canonical name\n\n")
	for _, g := range gs {
		name, data, err := recordFile(g, current, muscleParents)
		if err != nil {
			return nil, nil, 0, 0, err
		}
		files[name] = data
		if len(g.members) > 1 {
			names := []string{}
			for _, m := range g.members {
				names = append(names, m.Name)
			}
			fmt.Fprintf(&b, "- `%s`: %s\n", g.id, strings.Join(names, "; "))
		}
	}
	b.WriteString("\n## Set aside as not being exercises\n\n")
	for _, e := range rej {
		fmt.Fprintf(&b, "- %s (wger %d, was `%s`)\n", e.Name, e.ID, e.LegacyID)
		name, data, err := rejectedFile(e)
		if err != nil {
			return nil, nil, 0, 0, err
		}
		files[name] = data
	}
	return files, []byte(b.String()), len(gs), len(rej), nil
}

// discarded lists the records in dir that writing files would change or remove.
func discarded(dir string, files map[string][]byte) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}
	var lost []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		have, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		if want, ok := files[e.Name()]; !ok || !bytes.Equal(have, want) {
			lost = append(lost, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return lost, nil
}

// sample names a few records, for a message.
func sample(ids []string) string {
	if len(ids) <= 3 {
		return strings.Join(ids, ", ")
	}
	return fmt.Sprintf("%s and %d more", strings.Join(ids[:3], ", "), len(ids)-3)
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

func recordFile(g *group, current map[string]bool, muscleParents map[string]string) (string, []byte, error) {
	var relations []oes.Relation
	if data, err := os.ReadFile(relationPath(g.id)); err == nil {
		if err := json.Unmarshal(data, &relations); err != nil {
			return "", nil, fmt.Errorf("decode relations of %s: %w", g.id, err)
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
	return encodeDoc(g.id, doc)
}

// rejectedFile keeps a deprecated record, so an old ID still resolves.
func rejectedFile(e wger) (string, []byte, error) {
	doc := map[string]any{
		"$schema": "https://openexercise.org/schema/v0/exercise.json",
		"id":      e.LegacyID, "name": e.Name, "status": "deprecated",
		"measurement":  map[string]any{"profiles": []oes.Profile{{Required: []string{"reps"}}}},
		"external_ids": map[string]string{"wger": strconv.Itoa(e.ID)},
		"provenance":   provenanceFor([]wger{e}),
	}
	return encodeDoc(e.LegacyID, doc)
}

// encodeDoc is a record's file name and contents.
func encodeDoc(id string, doc map[string]any) (string, []byte, error) {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("encode %s: %w", id, err)
	}
	return strings.TrimPrefix(id, "oec:") + ".json", append(data, '\n'), nil
}
