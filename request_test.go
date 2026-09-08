package adsefid

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildIDsQuery(t *testing.T) {
	tests := []struct {
		name       string
		messageIDs []string
		localIDs   []string
		want       url.Values
		wantEmpty  bool
	}{
		{name: "neither", wantEmpty: true},
		{name: "empty slices", messageIDs: []string{}, localIDs: []string{}, wantEmpty: true},
		{name: "message ids only", messageIDs: []string{"a", "b"}, want: url.Values{"message_ids": {"a,b"}}},
		{name: "local ids only", localIDs: []string{"x"}, want: url.Values{"local_ids": {"x"}}},
		{
			name: "both", messageIDs: []string{"a"}, localIDs: []string{"x", "y"},
			want: url.Values{"message_ids": {"a"}, "local_ids": {"x,y"}},
		},
		{
			name: "values needing escaping", messageIDs: []string{"a b", "c%d"},
			want: url.Values{"message_ids": {"a b,c%d"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildIDsQuery(tc.messageIDs, tc.localIDs)
			if tc.wantEmpty {
				if got != "" {
					t.Errorf("expected no query string at all, got %q", got)
				}
				return
			}
			if !strings.HasPrefix(got, "?") {
				t.Fatalf("a non-empty query must start with '?', got %q", got)
			}
			parsed, err := url.ParseQuery(strings.TrimPrefix(got, "?"))
			if err != nil {
				t.Fatalf("parse %q: %v", got, err)
			}
			if parsed.Encode() != tc.want.Encode() {
				t.Errorf("query = %v, want %v", parsed, tc.want)
			}
		})
	}
}

func TestJoinCSV(t *testing.T) {
	tests := []struct {
		values []string
		want   string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "a,b,c"},
		{[]string{"a,b"}, "a,b"},
	}
	for _, tc := range tests {
		if got := joinCSV(tc.values); got != tc.want {
			t.Errorf("joinCSV(%v) = %q, want %q", tc.values, got, tc.want)
		}
	}
}

func TestEscapeQuotes(t *testing.T) {
	tests := []struct{ in, want string }{
		{"plain.txt", "plain.txt"},
		{`quo"te.txt`, `quo\"te.txt`},
		{`back\slash.txt`, `back\\slash.txt`},
		{`both"\.txt`, `both\"\\.txt`},
	}
	for _, tc := range tests {
		if got := escapeQuotes(tc.in); got != tc.want {
			t.Errorf("escapeQuotes(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
