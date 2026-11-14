// Command enrich turns the wger snapshot into Open Exercise records with Gemini.
// Each model answer is kept in sources/enrichment/<wger id>.json, so a run resumes and
// merging never calls the model again.
//
//	go run ./cmd/enrich generate [-limit N] [-ids 9,31]
//	go run ./cmd/enrich relate
//	go run ./cmd/enrich merge [-out DIR] [-force]
//
// merge rebuilt the catalogue from the snapshot once. Records have been edited by hand since, so it
// refuses to overwrite catalogue/exercises when that would change or remove any record: write to
// -out DIR to compare instead, or pass -force to throw the edits away.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/genai"

	"github.com/openexercise/open-exercise/internal/oes"
)

const (
	model       = "gemini-3.8-flash"
	location    = "global"
	workers     = 12
	enrichDir   = "sources/enrichment"
	relationDir = "sources/relations"
)

type wger struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Category       string   `json:"category"`
	Muscles        []string `json:"muscles"`
	Secondary      []string `json:"muscles_secondary"`
	Equipment      []string `json:"equipment"`
	License        string   `json:"license"`
	Author         string   `json:"license_author"`
	WgerAliases    []string `json:"wger_aliases"`
	LegacyID       string   `json:"legacy_id"`
	CuratedAliases []string `json:"curated_aliases"`
}

func main() {
	if len(os.Args) < 2 {
		fatal(errors.New("usage: enrich generate|relate|merge"))
	}
	ctx := context.Background()
	var err error
	switch os.Args[1] {
	case "generate":
		err = generate(ctx, os.Args[2:])
	case "relate":
		err = relate(ctx)
	case "merge":
		err = merge(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func loadSnapshot() ([]wger, error) {
	data, err := os.ReadFile("sources/wger/exercises.json")
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}
	var snap struct {
		Exercises []wger `json:"exercises"`
	}
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	return snap.Exercises, nil
}

func newClient(ctx context.Context) (*genai.Client, error) {
	project := os.Getenv("OES_GCP_PROJECT")
	if project == "" {
		return nil, errors.New("OES_GCP_PROJECT is required")
	}
	return genai.NewClient(ctx, &genai.ClientConfig{Backend: genai.BackendVertexAI, Project: project, Location: location})
}

// ask sends no response schema: constraint decoding times out at this size. The shape
// is in the prompt and the validator enforces it.
func ask(ctx context.Context, c *genai.Client, system, prompt string) ([]byte, error) {
	var last error
	for attempt := range 5 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*attempt) * 3 * time.Second)
		}
		callCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
		resp, err := c.Models.GenerateContent(callCtx, model, genai.Text(prompt), &genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(system, genai.RoleUser),
			ResponseMIMEType:  "application/json",
			Temperature:       genai.Ptr[float32](0),
		})
		cancel()
		if err != nil {
			last = err
			fmt.Fprintf(os.Stderr, "attempt %d: %v\n", attempt+1, err)
			continue
		}
		text := resp.Text()
		if json.Valid([]byte(text)) {
			return []byte(text), nil
		}
		last = errors.New("model returned invalid JSON")
	}
	return nil, last
}

func generate(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	limit := fs.Int("limit", 0, "stop after this many exercises")
	idList := fs.String("ids", "", "only these wger IDs, comma separated")
	if err := fs.Parse(args); err != nil {
		return err
	}
	exercises, err := loadSnapshot()
	if err != nil {
		return err
	}
	v, err := oes.NewValidator(".")
	if err != nil {
		return err
	}
	client, err := newClient(ctx)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	if err := os.MkdirAll(enrichDir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	var todo []wger
	for _, e := range exercises {
		if *idList != "" && !slices.Contains(strings.Split(*idList, ","), strconv.Itoa(e.ID)) {
			continue
		}
		if _, err := os.Stat(enrichPath(e.ID)); err == nil {
			continue
		}
		todo = append(todo, e)
	}
	if *limit > 0 && len(todo) > *limit {
		todo = todo[:*limit]
	}
	fmt.Printf("%d to do\n", len(todo))
	system := systemPrompt(v.Vocabs)
	var done, failed atomic.Int64
	jobs := make(chan wger)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for e := range jobs {
				if err := enrichOne(ctx, client, v, system, e); err != nil {
					failed.Add(1)
					fmt.Printf("FAILED %d %s: %v\n", e.ID, e.Name, err)
					continue
				}
				if n := done.Add(1); n%50 == 0 {
					fmt.Printf("%d done\n", n)
				}
			}
		})
	}
	for _, e := range todo {
		jobs <- e
	}
	close(jobs)
	wg.Wait()
	fmt.Printf("%d done, %d failed\n", done.Load(), failed.Load())
	return nil
}

func enrichPath(id int) string { return filepath.Join(enrichDir, strconv.Itoa(id)+".json") }

func enrichOne(ctx context.Context, c *genai.Client, v *oes.Validator, system string, e wger) error {
	input, err := json.Marshal(map[string]any{
		"name": e.Name, "description": e.Description, "category": e.Category, "primary_muscles": e.Muscles,
		"secondary_muscles": e.Secondary, "equipment": e.Equipment, "aliases": slices.Concat(e.WgerAliases, e.CuratedAliases),
	})
	if err != nil {
		return err
	}
	prompt := "Source record:\n" + string(input)
	var problems []string
	for range 3 {
		full := prompt
		if len(problems) > 0 {
			full += "\n\nYour previous answer was rejected for: " + strings.Join(problems, "; ") + ". Answer again, fixing these."
		}
		raw, err := ask(ctx, c, system, full)
		if err != nil {
			return err
		}
		var a answer
		if err := json.Unmarshal(raw, &a); err != nil {
			problems = []string{"wrong shape: " + err.Error() + " (optional_equipment is a list of strings; each common_faults item is an object with fault and correction)"}
			continue
		}
		if !a.IsExercise {
			return os.WriteFile(enrichPath(e.ID), raw, 0o644)
		}
		doc, err := json.Marshal(a.record("oec:check", nil))
		if err != nil {
			return err
		}
		problems = v.Validate(doc)
		if len(problems) == 0 {
			return os.WriteFile(enrichPath(e.ID), raw, 0o644)
		}
	}
	return fmt.Errorf("rejected: %s", strings.Join(problems, "; "))
}
