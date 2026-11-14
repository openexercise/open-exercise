// Command oes validates Open Exercise Schema documents and builds the catalogue
// and website.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openexercise/open-exercise/internal/oes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "validate":
		err = validate(os.Args[2:])
	case "review":
		err = review(os.Args[2:])
	case "site":
		out := "dist"
		if len(os.Args) > 2 {
			out = os.Args[2]
		}
		err = buildSite(".", out)
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const usage = `usage:
  oes validate [-catalogue PATH] [path...]       check records against the schema and the catalogue rules;
                                                 records outside the catalogue are checked with it
  oes review --by NAME [--level L] [--at DATE] [--checked] <id>...
                                                 mark catalogue records as reviewed by a person
  oes site [outdir]                              build the website (dist/ by default)`

// review marks catalogue records as checked by a person (CONTRIBUTING.md, Reviewing a record).
func review(args []string) error {
	flags := flag.NewFlagSet("review", flag.ContinueOnError)
	by := flags.String("by", "", "who reviewed them: a name or GitHub handle (required)")
	level := flags.String("level", oes.CommunityReviewed, "community-reviewed, or expert-reviewed for a qualified coach or clinician")
	at := flags.String("at", time.Now().Format(time.DateOnly), "the day of the review, YYYY-MM-DD")
	checked := flags.Bool("checked", false, "you checked every value provenance.generated still lists")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "usage: oes review --by NAME [--level L] [--at DATE] [--checked] <id>...\n\n"+
			"Sets provenance.review and clears provenance.generated on each record (oec:slug or slug).\n"+
			"A record that still lists generated values is refused unless --checked says you checked them all.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() == 0 {
		flags.Usage()
		return errors.New("name at least one record")
	}
	var failed []string
	for _, id := range flags.Args() {
		path := filepath.Join("catalogue", "exercises", strings.TrimPrefix(id, "oec:")+".json")
		raw, err := os.ReadFile(path)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", id, err))
			continue
		}
		out, err := oes.MarkReviewed(raw, *level, *by, *at, *checked)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", id, err))
			continue
		}
		if err := os.WriteFile(path, out, 0o644); err != nil {
			return err
		}
		fmt.Printf("%s: %s by %s on %s\n", id, *level, *by, *at)
	}
	if len(failed) > 0 {
		return errors.New(strings.Join(failed, "\n"))
	}
	return nil
}

const catalogueDir = "catalogue/exercises"

func validate(args []string) error {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	against := flags.String("catalogue", "", "the catalogue other records are checked with: a directory of records or a release's catalogue.json (default catalogue/exercises, when there is one)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	paths := flags.Args()
	if len(paths) == 0 {
		paths = []string{"examples", catalogueDir}
	}
	v, err := oes.NewValidator(".")
	if err != nil {
		return err
	}
	known, err := readCatalogue(*against)
	if err != nil {
		return err
	}
	var problems, count int
	for _, root := range paths {
		var all []oes.Exercise
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			count++
			for _, p := range v.Validate(data) {
				problems++
				fmt.Printf("%s: %s\n", path, p)
			}
			var e oes.Exercise
			if err := json.Unmarshal(data, &e); err == nil {
				all = append(all, e)
			}
			return nil
		})
		if err != nil {
			return err
		}
		var found []string
		if isCatalogue(root) && *against == "" {
			found = oes.CheckCatalogue(all) // the whole catalogue, checked by itself
		} else {
			found = oes.CheckAgainst(all, known)
		}
		for _, p := range found {
			problems++
			fmt.Printf("%s: %s\n", root, p)
		}
	}
	fmt.Printf("%d documents, %d problems\n", count, problems)
	if problems > 0 {
		return fmt.Errorf("validation failed")
	}
	return nil
}

// isCatalogue says a path is the catalogue directory itself.
func isCatalogue(path string) bool {
	return filepath.Clean(path) == filepath.Clean(catalogueDir)
}

// readCatalogue reads the records other files are checked with: a directory of records, a
// release's catalogue.json (an array of records), or by default this checkout's catalogue. No
// catalogue is no records: relations then can't resolve, and validate says so.
func readCatalogue(path string) ([]oes.Exercise, error) {
	if path == "" {
		if _, err := os.Stat(catalogueDir); err != nil {
			return nil, nil
		}
		path = catalogueDir
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("catalogue: %w", err)
	}
	if !info.IsDir() {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("catalogue: %w", err)
		}
		var all []oes.Exercise
		if err := json.Unmarshal(data, &all); err != nil {
			return nil, fmt.Errorf("catalogue %s: want an array of records: %w", path, err)
		}
		return all, nil
	}
	var all []oes.Exercise
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".json") {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		var e oes.Exercise
		if err := json.Unmarshal(data, &e); err != nil {
			return fmt.Errorf("catalogue %s: %w", p, err)
		}
		all = append(all, e)
		return nil
	})
	return all, err
}
