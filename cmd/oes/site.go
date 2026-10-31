package main

import (
	"bytes"
	"cmp"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"github.com/openexercise/open-exercise/internal/oes"
)

//go:embed templates/*.html templates/style.css templates/search.js
var templateFS embed.FS

type page struct {
	Title       string
	Description string
	Path        string
	Exercises   []oes.Exercise
	Exercise    *oes.Exercise
	Vocabs      []oes.Vocabulary
	Labels      map[string]string
	Names       map[string]string
	HTML        template.HTML
	Raw         string
	Patterns    []oes.Term
	Muscles     []oes.Term
	Equipment   []oes.Term
	Count       int
}

// indexEntry is one row of the search index, with parents included so filtering by
// "chest" finds a record that names only the sternal head.
type indexEntry struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Summary   string   `json:"summary,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
	Patterns  []string `json:"patterns,omitempty"`
	Muscles   []string `json:"muscles,omitempty"`
	Equipment []string `json:"equipment,omitempty"`
}

func buildSite(root, out string) error {
	vocabs, err := oes.LoadVocabularies(root)
	if err != nil {
		return err
	}
	exercises, raw, err := loadCatalogue(filepath.Join(root, "catalogue", "exercises"))
	if err != nil {
		return err
	}
	labels := map[string]string{}
	var vocabList []oes.Vocabulary
	for _, v := range vocabs {
		vocabList = append(vocabList, v)
		for _, t := range v.Terms {
			labels[t.ID] = t.Label
		}
	}
	slices.SortFunc(vocabList, func(a, b oes.Vocabulary) int { return cmp.Compare(a.Name, b.Name) })
	names := map[string]string{}
	for _, e := range exercises {
		names[e.ID] = e.Name
	}
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"label": func(id string) string {
			if l, ok := labels[id]; ok {
				return l
			}
			return id
		},
		"exerciseName": func(id string) string { return names[id] },
		"slug":         func(id string) string { return strings.TrimPrefix(id, "oec:") },
		"words":        func(s string) string { return strings.ReplaceAll(s, "-", " ") },
		"join":         strings.Join,
		"isGenerated": func(e *oes.Exercise) bool {
			return e.Provenance.Review != nil && e.Provenance.Review.Status == "machine-generated"
		},
	}).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}
	if err := os.RemoveAll(out); err != nil {
		return fmt.Errorf("clear output: %w", err)
	}
	render := func(path, name string, p page) error {
		p.Path = path
		dest := filepath.Join(out, path, "index.html")
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		var b bytes.Buffer
		if err := tmpl.ExecuteTemplate(&b, name, p); err != nil {
			return fmt.Errorf("render %s: %w", path, err)
		}
		return os.WriteFile(dest, b.Bytes(), 0o644)
	}

	active := slices.DeleteFunc(slices.Clone(exercises), func(e oes.Exercise) bool { return e.Status == "deprecated" })
	if err := render("", "home.html", page{Title: "Open Exercise", Description: "An open standard for describing exercises, and a public catalogue in that format.", Count: len(active)}); err != nil {
		return err
	}
	spec, err := renderMarkdown(filepath.Join(root, "spec", "v0", "SPEC.md"))
	if err != nil {
		return err
	}
	if err := render("spec", "doc.html", page{Title: "Specification", Description: "The Open Exercise Schema, version 0 draft.", HTML: spec}); err != nil {
		return err
	}
	if err := render("vocabulary", "vocabulary.html", page{Title: "Vocabularies", Description: "The controlled terms the schema uses.", Vocabs: vocabList}); err != nil {
		return err
	}
	if err := render("catalogue", "catalogue.html", page{
		Title: "Catalogue", Description: "Search the Open Exercise Catalogue.", Exercises: active, Count: len(active),
		Patterns: vocabs["movement-patterns"].Terms, Muscles: vocabs["muscles"].Terms, Equipment: vocabs["equipment"].Terms,
	}); err != nil {
		return err
	}
	for _, e := range exercises {
		rec := e
		desc := e.Summary
		if desc == "" {
			desc = e.Name + " in the Open Exercise Catalogue."
		}
		if err := render("exercises/"+strings.TrimPrefix(e.ID, "oec:"), "exercise.html", page{Title: e.Name, Description: desc, Exercise: &rec, Raw: raw[e.ID]}); err != nil {
			return err
		}
	}
	if err := writeData(root, out, exercises, raw, active, vocabs); err != nil {
		return err
	}
	fmt.Printf("site: %d pages for %d exercises in %s\n", len(exercises)+4, len(active), out)
	return nil
}

func loadCatalogue(dir string) ([]oes.Exercise, map[string]string, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, nil, err
	}
	exercises := make([]oes.Exercise, 0, len(paths))
	raw := make(map[string]string, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, err
		}
		var e oes.Exercise
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, nil, fmt.Errorf("decode %s: %w", p, err)
		}
		exercises = append(exercises, e)
		raw[e.ID] = string(data)
	}
	slices.SortFunc(exercises, func(a, b oes.Exercise) int { return cmp.Compare(a.Name, b.Name) })
	return exercises, raw, nil
}

func renderMarkdown(path string) (template.HTML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := goldmark.New(goldmark.WithExtensions(extension.Table)).Convert(data, &b); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return template.HTML(b.String()), nil // our own file in this repository
}

func ancestors(terms []oes.Term, id string) []string {
	parent := map[string]string{}
	for _, t := range terms {
		parent[t.ID] = t.Parent
	}
	out := []string{}
	for id != "" {
		out = append(out, id)
		id = parent[id]
	}
	return out
}

func writeData(root, out string, exercises []oes.Exercise, raw map[string]string, active []oes.Exercise, vocabs map[string]oes.Vocabulary) error {
	index := make([]indexEntry, 0, len(active))
	for _, e := range active {
		row := indexEntry{ID: e.ID, Name: e.Name, Summary: e.Summary, Patterns: e.Classification.MovementPatterns}
		for _, a := range e.Aliases {
			row.Aliases = append(row.Aliases, a.Text)
		}
		for _, m := range e.Anatomy.Muscles {
			for _, id := range ancestors(vocabs["muscles"].Terms, m.Muscle) {
				if !slices.Contains(row.Muscles, id) {
					row.Muscles = append(row.Muscles, id)
				}
			}
		}
		for _, group := range e.Equipment.Required {
			for _, item := range group {
				for _, id := range ancestors(vocabs["equipment"].Terms, item) {
					if !slices.Contains(row.Equipment, id) {
						row.Equipment = append(row.Equipment, id)
					}
				}
			}
		}
		index = append(index, row)
	}
	if err := writeJSON(filepath.Join(out, "catalogue", "index.json"), index); err != nil {
		return err
	}
	all := make([]json.RawMessage, 0, len(exercises))
	for _, e := range exercises {
		all = append(all, json.RawMessage(raw[e.ID]))
		if err := writeFile(filepath.Join(out, "exercises", strings.TrimPrefix(e.ID, "oec:")+".json"), []byte(raw[e.ID])); err != nil {
			return err
		}
	}
	if err := writeJSON(filepath.Join(out, "catalogue.json"), all); err != nil {
		return err
	}
	for _, dir := range []string{"schema", "vocab"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			return writeFile(filepath.Join(out, rel), data)
		})
		if err != nil {
			return fmt.Errorf("copy %s: %w", dir, err)
		}
	}
	for _, name := range []string{"style.css", "search.js"} {
		data, err := templateFS.ReadFile("templates/" + name)
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(out, name), data); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return writeFile(path, data)
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
