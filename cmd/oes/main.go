// Command oes validates Open Exercise Schema documents and builds the catalogue
// and website.
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/openexercise/open-exercise/internal/oes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: oes validate [path...] | oes site [outdir]")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "validate":
		err = validate(os.Args[2:])
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
