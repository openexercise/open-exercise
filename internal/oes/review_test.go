package oes

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestMarkReviewed(t *testing.T) {
	raw, err := os.ReadFile("../../catalogue/exercises/barbell-bench-press.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := MarkReviewed(raw, CommunityReviewed, "ana", "2026-11-05", false); err == nil || !strings.Contains(err.Error(), "/summary") {
		t.Errorf("a record with unchecked values was marked reviewed: %v", err)
	}
	for _, bad := range []struct{ level, by, at string }{
		{"machine-generated", "ana", "2026-11-05"}, {CommunityReviewed, " ", "2026-11-05"}, {CommunityReviewed, "ana", "5 Nov"},
	} {
		if _, err := MarkReviewed(raw, bad.level, bad.by, bad.at, true); err == nil {
			t.Errorf("%+v was accepted", bad)
		}
	}
	out, err := MarkReviewed(raw, ExpertReviewed, "ana", "2026-11-05", true)
	if err != nil {
		t.Fatal(err)
	}
	var before, after map[string]json.RawMessage
	_ = json.Unmarshal(raw, &before)
	_ = json.Unmarshal(out, &after)
	for k := range before {
		if k != "provenance" && string(before[k]) != string(after[k]) {
			t.Errorf("%s changed", k)
		}
	}
	var e Exercise
	if err := json.Unmarshal(out, &e); err != nil {
		t.Fatal(err)
	}
	if r := e.Provenance.Review; r == nil || r.Status != ExpertReviewed || r.By != "ana" || r.At != "2026-11-05" || r.Model != "" {
		t.Errorf("review = %+v", r)
	}
	if len(e.Provenance.Generated) != 0 || len(e.Provenance.Sources) == 0 || e.Provenance.License == "" {
		t.Errorf("provenance = %+v", e.Provenance)
	}
	if problems := checkSense(e, decode(t, out)); len(problems) > 0 {
		t.Errorf("a reviewed record has problems: %v", problems)
	}
}

func TestReviewedRecordsAreChecked(t *testing.T) {
	raw, _ := os.ReadFile("../../catalogue/exercises/barbell-bench-press.json")
	var e Exercise
	_ = json.Unmarshal(raw, &e)
	e.Provenance.Review = &Review{Status: CommunityReviewed}
	problems := strings.Join(checkSense(e, decode(t, raw)), "; ")
	if !strings.Contains(problems, "still lists unchecked values") || !strings.Contains(problems, "needs by and at") {
		t.Errorf("problems = %s", problems)
	}
}

func TestReplacedByMustNotLoop(t *testing.T) {
	all := []Exercise{{ID: "oec:a", ReplacedBy: "oec:b"}, {ID: "oec:b", ReplacedBy: "oec:c"}, {ID: "oec:c", ReplacedBy: "oec:a"}, {ID: "oec:d", ReplacedBy: "oec:a"}}
	if got := strings.Join(CheckCatalogue(all), "; "); !strings.Contains(got, "loops") {
		t.Errorf("a loop wasn't caught: %s", got)
	}
	ok := []Exercise{{ID: "oec:a", ReplacedBy: "oec:b"}, {ID: "oec:b", ReplacedBy: "oec:c"}, {ID: "oec:c"}}
	if got := CheckCatalogue(ok); len(got) > 0 {
		t.Errorf("a chain was refused: %v", got)
	}
}

func decode(t *testing.T, raw []byte) any {
	t.Helper()
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}
