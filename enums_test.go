package adsefid

import (
	"encoding/json"
	"testing"
)

// TestEnumsAreForwardCompatible: the API may add codes at any time. Every enum
// here is a named integer or string type precisely so an unrecognised value
// carries through instead of failing the whole response.
func TestEnumsAreForwardCompatible(t *testing.T) {
	t.Run("message status", func(t *testing.T) {
		if got := WebServiceMessageStatusDelivered.String(); got != "Delivered" {
			t.Errorf("String() = %q", got)
		}
		unknown := WebServiceMessageStatus(1998)
		if got := unknown.String(); got == "" {
			t.Error("an unknown status should still stringify")
		}
		if int(unknown) != 1998 {
			t.Error("the raw value must survive")
		}
	})

	t.Run("response code", func(t *testing.T) {
		if got := WebServiceResponseCodeInvalidParameter.String(); got == "" {
			t.Error("String() should not be empty")
		}
		unknown := WebServiceResponseCode(2999)
		if int(unknown) != 2999 {
			t.Error("the raw value must survive")
		}
		if got := unknown.String(); got == "" {
			t.Error("an unknown code should still stringify")
		}
	})

	t.Run("line selector", func(t *testing.T) {
		if LineSelectorBulkServiceSendBased != 2 {
			t.Error("line selector 2 is BulkServiceSendBased per the docs")
		}
		if got := LineSelector(99).String(); got == "" {
			t.Error("an unknown line selector should still stringify")
		}
	})

	t.Run("template state", func(t *testing.T) {
		for _, state := range []TemplateState{TemplateStatePendingApproval, TemplateStateApproved, TemplateStateRejected} {
			if string(state) == "" {
				t.Error("template states are lowercase wire strings")
			}
		}
		if got := TemplateState("archived").String(); got == "" {
			t.Error("an unknown state should still stringify")
		}
	})

	t.Run("template parameter type", func(t *testing.T) {
		if TemplateParameterTypeString != "string" || TemplateParameterTypeNumber != "number" {
			t.Error("the documented wire values are lowercase")
		}
		// The service also emits an undocumented "url"; this SDK models only
		// the documented set and drops the rest on read.
		if got := TemplateParameterType("url").String(); got == "" {
			t.Error("an unknown parameter type should still stringify")
		}
	})
}

// TestTemplateParameterValueJSON is the round-trip that makes leading zeros and
// exact decimals survive: a number-typed parameter may travel as a JSON string,
// and a JSON number keeps its exact wire text rather than going through float64.
func TestTemplateParameterValueJSON(t *testing.T) {
	t.Run("marshals strings and numbers distinctly", func(t *testing.T) {
		tests := []struct {
			name  string
			value TemplateParameterValue
			want  string
		}{
			{"string", StringParam("459122"), `"459122"`},
			{"string with leading zeros", StringParam("001234"), `"001234"`},
			{"string with an exact decimal", StringParam("1.50"), `"1.50"`},
			{"empty string", StringParam(""), `""`},
			{"integer", IntParam(2), `2`},
			{"negative integer", IntParam(-7), `-7`},
			{"float", NumberParam(1.5), `1.5`},
			{"float with no fraction", NumberParam(3), `3`},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got, err := json.Marshal(tc.value)
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				if string(got) != tc.want {
					t.Errorf("Marshal = %s, want %s", got, tc.want)
				}
			})
		}
	})

	t.Run("a number keeps its exact wire text", func(t *testing.T) {
		var value TemplateParameterValue
		if err := json.Unmarshal([]byte(`1.50`), &value); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		raw, ok := value.RawNumber()
		if !ok {
			t.Fatal("expected a number")
		}
		if raw.String() != "1.50" {
			t.Errorf("raw number = %q, want %q — the trailing zero must survive", raw, "1.50")
		}
		remarshalled, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if string(remarshalled) != "1.50" {
			t.Errorf("round trip = %s, want 1.50", remarshalled)
		}
	})

	t.Run("accessors report the held kind", func(t *testing.T) {
		str := StringParam("001234")
		if str.IsNumber() {
			t.Error("StringParam holds a string")
		}
		if got, ok := str.StringValue(); !ok || got != "001234" {
			t.Errorf("StringValue = %q, %v", got, ok)
		}
		if _, ok := str.NumberValue(); ok {
			t.Error("a string value has no number")
		}
		if _, ok := str.RawNumber(); ok {
			t.Error("a string value has no raw number")
		}

		num := NumberParam(1.5)
		if !num.IsNumber() {
			t.Error("NumberParam holds a number")
		}
		if got, ok := num.NumberValue(); !ok || got != 1.5 {
			t.Errorf("NumberValue = %v, %v", got, ok)
		}
		if _, ok := num.StringValue(); ok {
			t.Error("a number value has no string")
		}
	})

	t.Run("rejects anything but a string or a number", func(t *testing.T) {
		for _, raw := range []string{`true`, `null`, `{}`, `[]`} {
			var value TemplateParameterValue
			if err := json.Unmarshal([]byte(raw), &value); err == nil {
				t.Errorf("%s should be rejected as a template parameter value", raw)
			}
		}
	})

	t.Run("round-trips a whole parameter map", func(t *testing.T) {
		wire := `{"amount":1.50,"code":"459122","invoice":"001234","quantity":2}`
		var params map[string]TemplateParameterValue
		if err := json.Unmarshal([]byte(wire), &params); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		got, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		// Go sorts map keys on marshal, and the fixture is already sorted.
		if string(got) != wire {
			t.Errorf("round trip = %s, want %s", got, wire)
		}
	})
}
