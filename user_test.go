package adsefid

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

func TestUserEndpoints(t *testing.T) {
	t.Run("get info", func(t *testing.T) {
		client, requests := newFixtureClient(t, "envelopes/user.get_info.success.json")
		got, err := client.User.GetInfo(context.Background())
		if err != nil {
			t.Fatalf("GetInfo: %v", err)
		}
		req := onlyRequest(t, requests)
		if req.Method != http.MethodGet || req.Path != "/v1/user/info" {
			t.Errorf("%s %s", req.Method, req.Path)
		}
		if got == nil {
			t.Fatal("expected user info")
		}
	})

	t.Run("get lines", func(t *testing.T) {
		client, requests := newFixtureClient(t, "envelopes/user.get_lines.success.json")
		got, err := client.User.GetLines(context.Background())
		if err != nil {
			t.Fatalf("GetLines: %v", err)
		}
		if onlyRequest(t, requests).Path != "/v1/user/lines" {
			t.Error("wrong path")
		}
		if len(got) == 0 {
			t.Fatal("expected at least one line")
		}
	})

	t.Run("get profiles", func(t *testing.T) {
		client, requests := newFixtureClient(t, "envelopes/user.get_profiles.success.json")
		got, err := client.User.GetProfiles(context.Background())
		if err != nil {
			t.Fatalf("GetProfiles: %v", err)
		}
		if onlyRequest(t, requests).Path != "/v1/user/profiles" {
			t.Error("wrong path")
		}
		if len(got) == 0 {
			t.Fatal("expected at least one profile")
		}
	})
}

func TestGetTemplates(t *testing.T) {
	t.Run("parses the documented shape", func(t *testing.T) {
		client, requests := newFixtureClient(t, "envelopes/user.get_templates.success.json")
		state := TemplateStateApproved
		got, err := client.User.GetTemplates(context.Background(), &GetUserTemplatesRequest{State: &state, Skip: ptr(0), Take: ptr(50)})
		if err != nil {
			t.Fatalf("GetTemplates: %v", err)
		}

		req := onlyRequest(t, requests)
		gotQuery, _ := url.ParseQuery(req.RawQuery)
		want := url.Values{"state": {"approved"}, "skip": {"0"}, "take": {"50"}}
		if gotQuery.Encode() != want.Encode() {
			t.Errorf("query = %v, want %v", gotQuery, want)
		}

		if got.Total != 1 || len(got.Items) != 1 {
			t.Fatalf("total/items = %d/%d", got.Total, len(got.Items))
		}
		item := got.Items[0]
		if item.State != TemplateStateApproved {
			t.Errorf("state = %v", item.State)
		}
		if item.Parameters["OTPCode"] != TemplateParameterTypeString {
			t.Errorf("OTPCode type = %v", item.Parameters["OTPCode"])
		}
		if item.Parameters["amount"] != TemplateParameterTypeNumber {
			t.Errorf("amount type = %v", item.Parameters["amount"])
		}
		if item.Description != nil {
			t.Errorf("description should be nil, got %v", *item.Description)
		}
		if item.CreatedAt.IsZero() || item.UpdatedAt.IsZero() {
			t.Error("timestamps should have parsed")
		}
	})

	// The live service emits an undocumented parameter type the SDK does not
	// model. Dropping the entry keeps the typed map honest instead of surfacing
	// a value callers cannot switch on.
	t.Run("drops undocumented parameter types", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/user.get_templates.unknown_type.json")
		got, err := client.User.GetTemplates(context.Background(), nil)
		if err != nil {
			t.Fatalf("GetTemplates: %v", err)
		}
		params := got.Items[0].Parameters
		if _, present := params["link"]; present {
			t.Error(`the undocumented "url" parameter type should have been dropped`)
		}
		if len(params) != 2 {
			t.Errorf("expected the two documented parameters, got %v", params)
		}
	})

	t.Run("omits absent query parameters entirely", func(t *testing.T) {
		client, requests := newFixtureClient(t, "envelopes/user.get_templates.success.json")
		if _, err := client.User.GetTemplates(context.Background(), nil); err != nil {
			t.Fatalf("GetTemplates: %v", err)
		}
		if got := onlyRequest(t, requests).RawQuery; got != "" {
			t.Errorf("expected no query string, got %q", got)
		}
	})

	t.Run("rejects out-of-range paging", func(t *testing.T) {
		for _, tc := range []struct {
			name       string
			skip, take *int
		}{
			{"take zero", nil, ptr(0)},
			{"take over 100", nil, ptr(maxTemplatesTake + 1)},
			{"negative skip", ptr(-1), nil},
		} {
			t.Run(tc.name, func(t *testing.T) {
				client, requests := newFixtureClient(t, "envelopes/user.get_templates.success.json")
				_, err := client.User.GetTemplates(context.Background(), &GetUserTemplatesRequest{Skip: tc.skip, Take: tc.take})
				var validationErr *ValidationError
				if !asError(err, &validationErr) {
					t.Fatalf("expected a *ValidationError, got %#v", err)
				}
				if len(*requests) != 0 {
					t.Error("validation must reject before any HTTP call")
				}
			})
		}
	})

	t.Run("accepts the take boundaries", func(t *testing.T) {
		for _, take := range []int{1, maxTemplatesTake} {
			client, _ := newFixtureClient(t, "envelopes/user.get_templates.success.json")
			if _, err := client.User.GetTemplates(context.Background(), &GetUserTemplatesRequest{Take: ptr(take)}); err != nil {
				t.Errorf("take=%d should be accepted: %v", take, err)
			}
		}
	})
}
