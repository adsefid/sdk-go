package adsefid

import (
	"encoding/json"
	"testing"
)

type localIDCase struct {
	Value string `json:"value"`
	Valid bool   `json:"valid"`
	Why   string `json:"why"`
}

// TestLocalIDPattern runs the golden table shared byte-for-byte with the
// sibling SDK repositories, so all five agree on what a local_id may be.
func TestLocalIDPattern(t *testing.T) {
	var cases []localIDCase
	if err := json.Unmarshal(mustFixture(t, "validation/local_ids.json"), &cases); err != nil {
		t.Fatalf("decode local_ids.json: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("the golden table is empty")
	}

	for _, tc := range cases {
		t.Run(tc.Why, func(t *testing.T) {
			err := validateLocalID(&tc.Value, "local_id")
			if tc.Valid && err != nil {
				t.Errorf("%q should be valid (%s), got %v", tc.Value, tc.Why, err)
			}
			if !tc.Valid && err == nil {
				t.Errorf("%q should be rejected (%s)", tc.Value, tc.Why)
			}
		})
	}
}

// A nil or empty local_id means "not supplied" and is skipped rather than run
// through the pattern; the field is optional on every endpoint that takes it.
func TestLocalIDOptionality(t *testing.T) {
	if err := validateLocalID(nil, "local_id"); err != nil {
		t.Errorf("a nil local_id is not supplied, so it should pass: %v", err)
	}
	empty := ""
	if err := validateLocalID(&empty, "local_id"); err != nil {
		t.Errorf("an empty local_id is treated as not supplied: %v", err)
	}
}

// TestMaxLengthCountsUTF16CodeUnits: the server enforces its length limits in
// UTF-16 code units, so a character outside the Basic Multilingual Plane costs
// two. Counting runes here would accept a message the server rejects.
func TestMaxLengthCountsUTF16CodeUnits(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		maxLength  int
		wantReject bool
	}{
		{"ascii at the limit", repeatRune('a', 900), 900, false},
		{"ascii one over", repeatRune('a', 901), 900, true},
		{"persian at the limit", repeatRune('س', 900), 900, false},
		{"persian one over", repeatRune('س', 901), 900, true},
		{"450 emoji is exactly 900 code units", repeatRune('😀', 450), 900, false},
		{"451 emoji is 902 code units", repeatRune('😀', 451), 900, true},
		{"900 emoji is 1800 code units", repeatRune('😀', 900), 900, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := requireMaxLength(tc.value, tc.maxLength, "message")
			if tc.wantReject && err == nil {
				t.Errorf("expected %q (%d units) to be rejected", tc.name, utf16Length(tc.value))
			}
			if !tc.wantReject && err != nil {
				t.Errorf("expected %q (%d units) to be accepted: %v", tc.name, utf16Length(tc.value), err)
			}
		})
	}
}

func TestUTF16Length(t *testing.T) {
	tests := []struct {
		value string
		want  int
	}{
		{"", 0},
		{"abc", 3},
		{"سلام", 4},
		{"😀", 2},
		{"a😀b", 4},
		{"é", 1},
	}
	for _, tc := range tests {
		if got := utf16Length(tc.value); got != tc.want {
			t.Errorf("utf16Length(%q) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestValidationHelpers(t *testing.T) {
	t.Run("requireNonEmpty", func(t *testing.T) {
		if requireNonEmpty("", "field") == nil {
			t.Error("an empty string should be rejected")
		}
		if err := requireNonEmpty("x", "field"); err != nil {
			t.Errorf("a non-empty string should pass: %v", err)
		}
	})

	t.Run("requireNonEmptySlice", func(t *testing.T) {
		if requireNonEmptySlice([]string(nil), "field") == nil {
			t.Error("a nil slice should be rejected")
		}
		if requireNonEmptySlice([]string{}, "field") == nil {
			t.Error("an empty slice should be rejected")
		}
		if err := requireNonEmptySlice([]string{"a"}, "field"); err != nil {
			t.Errorf("a populated slice should pass: %v", err)
		}
	})

	t.Run("requireInRange is inclusive at both ends", func(t *testing.T) {
		for _, value := range []int{0, 250, 499} {
			if err := requireInRange(value, 0, 499, "count"); err != nil {
				t.Errorf("%d should be in range: %v", value, err)
			}
		}
		for _, value := range []int{-1, 500} {
			if requireInRange(value, 0, 499, "count") == nil {
				t.Errorf("%d should be out of range", value)
			}
		}
	})

	t.Run("requireAtLeastOne", func(t *testing.T) {
		if requireAtLeastOne(false, false, "need one") == nil {
			t.Error("neither provided should be rejected")
		}
		for _, tc := range [][2]bool{{true, false}, {false, true}, {true, true}} {
			if err := requireAtLeastOne(tc[0], tc[1], "need one"); err != nil {
				t.Errorf("%v should pass: %v", tc, err)
			}
		}
	})

	t.Run("requireCombinedCountAtMost", func(t *testing.T) {
		if err := requireCombinedCountAtMost(1000, 1000, 2000, "too many"); err != nil {
			t.Errorf("exactly the limit should pass: %v", err)
		}
		if requireCombinedCountAtMost(1000, 1001, 2000, "too many") == nil {
			t.Error("one over the limit should be rejected")
		}
	})

	t.Run("distinctCount", func(t *testing.T) {
		if got := distinctCount(nil); got != 0 {
			t.Errorf("distinctCount(nil) = %d", got)
		}
		if got := distinctCount([]string{"a", "b", "a", "a"}); got != 2 {
			t.Errorf("distinctCount = %d, want 2", got)
		}
	})

	t.Run("requireRequest", func(t *testing.T) {
		if requireRequest[SendSingleSmsRequest](nil) == nil {
			t.Error("a nil request should be rejected")
		}
		if err := requireRequest(&SendSingleSmsRequest{}); err != nil {
			t.Errorf("a non-nil request should pass: %v", err)
		}
	})
}

func TestValidationErrorMessage(t *testing.T) {
	withField := &ValidationError{Field: "receptor", Message: "is required"}
	if got := withField.Error(); got != "receptor: is required" {
		t.Errorf("Error() = %q", got)
	}
	withoutField := &ValidationError{Message: "at least one id is required"}
	if got := withoutField.Error(); got != "at least one id is required" {
		t.Errorf("Error() = %q", got)
	}
}
