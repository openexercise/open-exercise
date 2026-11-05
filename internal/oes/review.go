package oes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Review levels a person can give a record (SPEC §13).
const (
	CommunityReviewed = "community-reviewed"
	ExpertReviewed    = "expert-reviewed"
)

// MarkReviewed records that a person checked a record: provenance.review becomes level, by and at,
// and provenance.generated goes, since no value is unchecked any more. A record that still lists
// generated values is refused unless checked says the reviewer checked them all. Only the
// provenance changes; every other byte of the record stays as it was.
func MarkReviewed(raw []byte, level, by, at string, checked bool) ([]byte, error) {
	if level != CommunityReviewed && level != ExpertReviewed {
		return nil, fmt.Errorf("level %q: use %s or %s", level, CommunityReviewed, ExpertReviewed)
	}
	if strings.TrimSpace(by) == "" {
		return nil, errors.New("say who reviewed it (by)")
	}
	if _, err := time.Parse(time.DateOnly, at); err != nil {
		return nil, fmt.Errorf("at %q: use YYYY-MM-DD", at)
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return nil, fmt.Errorf("read the record: %w", err)
	}
	provenance := map[string]json.RawMessage{}
	if p, ok := record["provenance"]; ok {
		if err := json.Unmarshal(p, &provenance); err != nil {
			return nil, fmt.Errorf("read provenance: %w", err)
		}
	}
	var generated []string
	if g, ok := provenance["generated"]; ok {
		if err := json.Unmarshal(g, &generated); err != nil {
			return nil, fmt.Errorf("read provenance.generated: %w", err)
		}
	}
	if len(generated) > 0 && !checked {
		return nil, fmt.Errorf("provenance.generated still lists %s: check them, remove each you checked, or say you checked them all", strings.Join(generated, ", "))
	}
	review, err := json.Marshal(Review{Status: level, At: at, By: by})
	if err != nil {
		return nil, err
	}
	provenance["review"] = review
	delete(provenance, "generated")
	if record["provenance"], err = json.Marshal(provenance); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(record); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// isReviewed says a person checked the record.
func isReviewed(r *Review) bool {
	return r != nil && (r.Status == CommunityReviewed || r.Status == ExpertReviewed)
}
