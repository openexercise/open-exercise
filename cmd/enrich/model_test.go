package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

func TestUnavailable(t *testing.T) {
	status := func(code int) error {
		return fmt.Errorf("claude: %w", &anthropic.Error{StatusCode: code, Request: &http.Request{}, Response: &http.Response{StatusCode: code}})
	}
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"rate limit", status(429), true},
		{"server error", status(500), true},
		{"overloaded", status(529), true},
		{"bad request", status(400), false},
		{"no credential", status(401), false},
		{"forbidden", status(403), false},
		{"not found", status(404), false},
		{"timeout", fmt.Errorf("claude: %w", context.DeadlineExceeded), true},
		{"transport", errors.New("claude: connection reset"), true},
	}
	for _, c := range cases {
		if got := unavailable(c.err); got != c.want {
			t.Errorf("%s: unavailable = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestJSONBody(t *testing.T) {
	cases := map[string]string{
		`{"a":1}`:                 `{"a":1}`,
		"  {\"a\":1}\n":           `{"a":1}`,
		"```json\n{\"a\":1}\n```": `{"a":1}`,
		"```\n{\"a\":1}\n```\n":   `{"a":1}`,
		"not json":                "not json",
	}
	for in, want := range cases {
		if got := string(jsonBody(in)); got != want {
			t.Errorf("jsonBody(%q) = %q, want %q", in, got, want)
		}
	}
}
