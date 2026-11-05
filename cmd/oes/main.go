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
  oes validate [path...]                         check records against the schema and the catalogue rules
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

func validate(paths []string) error {
	if len(paths) == 0 {
		paths = []string{"examples", "catalogue/exercises"}
	}
	v, err := oes.NewValidator(".")
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
		if root == "examples" {
			continue // examples point at catalogue exercises they don't include
		}
		for _, p := range oes.CheckCatalogue(all) {
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
