package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var idPattern = regexp.MustCompile(`id="([^"]*)"`)

// TestSiteLinksResolve builds the site and checks every link on every page: none may point at a
// Markdown file or a repository path, and every site path and in-page anchor must exist in the build.
func TestSiteLinksResolve(t *testing.T) {
	out := t.TempDir()
	if err := buildSite(filepath.Join("..", ".."), out); err != nil {
		t.Fatal(err)
	}
	ids := map[string]map[string]bool{} // page path ("/developers/") -> heading ids on it
	pages := map[string]string{}
	err := filepath.WalkDir(out, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".html") {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(out, filepath.Dir(p))
		if err != nil {
			return err
		}
		page := "/"
		if rel != "." {
			page = "/" + filepath.ToSlash(rel) + "/"
		}
		pages[page] = string(data)
		ids[page] = map[string]bool{}
		for _, m := range idPattern.FindAllStringSubmatch(string(data), -1) {
			ids[page][m[1]] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for page, html := range pages {
		for _, m := range hrefPattern.FindAllStringSubmatch(html, -1) {
			href := m[1]
			if strings.Contains(href, "://") || strings.HasPrefix(href, "mailto:") {
				continue
			}
			checked++
			target, anchor, _ := strings.Cut(href, "#")
			if strings.Contains(target, ".md") {
				t.Errorf("%s links to Markdown %q", page, href)
				continue
			}
			if target == "" {
				target = page
			}
			if !strings.HasPrefix(target, "/") {
				t.Errorf("%s has a relative link %q", page, href)
				continue
			}
			if strings.HasSuffix(target, "/") {
				if _, ok := pages[target]; !ok {
					t.Errorf("%s links to %q, which the site has no page for", page, href)
					continue
				}
				if anchor != "" && !ids[target][anchor] {
					t.Errorf("%s links to %q, whose page has no id %q", page, href, anchor)
				}
				continue
			}
			if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(target))); err != nil {
				t.Errorf("%s links to %q, which the site doesn't have", page, href)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no links checked")
	}
}

func TestSiteLink(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, c := range []struct{ file, href, want string }{
		{"docs/developer-guide.md", "../GUIDE.md#names-and-aliases", "/guide/#names-and-aliases"},
		{"docs/developer-guide.md", "../spec/v0/SPEC.md#6-setup", "/spec/#6-setup"},
		{"docs/developer-guide.md", "#7-load-a-new-release", "#7-load-a-new-release"},
		{"docs/developer-guide.md", "../vocab/v0", "/vocabulary/"},
		{"docs/developer-guide.md", "../schema/v0/exercise.json", "/schema/v0/exercise.json"},
		{"docs/developer-guide.md", "../catalogue/LICENSE", repoURL + "/blob/main/catalogue/LICENSE"},
		{"docs/developer-guide.md", "../catalogue", repoURL + "/tree/main/catalogue"},
		{"GUIDE.md", "docs/developer-guide.md", "/developers/"},
		{"CHANGELOG.md", "GOVERNANCE.md", repoURL + "/blob/main/GOVERNANCE.md"},
		{"GUIDE.md", "https://wger.de", "https://wger.de"},
	} {
		if got := siteLink(root, c.file, c.href); got != c.want {
			t.Errorf("siteLink(%q, %q) = %q, want %q", c.file, c.href, got, c.want)
		}
	}
}
