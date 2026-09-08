package adsefid

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		opts       []ClientOption
		wantReject bool
	}{
		{name: "defaults", apiKey: "k"},
		{name: "blank api key", apiKey: "", wantReject: true},
		{name: "blank base url", apiKey: "k", opts: []ClientOption{WithBaseURL("")}, wantReject: true},
		{name: "base url of only slashes", apiKey: "k", opts: []ClientOption{WithBaseURL("///")}, wantReject: true},
		{name: "blank user agent", apiKey: "k", opts: []ClientOption{WithUserAgent("  ")}, wantReject: true},
		{name: "user agent with a newline", apiKey: "k", opts: []ClientOption{WithUserAgent("bad\nagent")}, wantReject: true},
		{name: "user agent with a carriage return", apiKey: "k", opts: []ClientOption{WithUserAgent("bad\ragent")}, wantReject: true},
		{name: "custom user agent", apiKey: "k", opts: []ClientOption{WithUserAgent("my-app/1.0")}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.apiKey, tc.opts...)
			if tc.wantReject {
				var validationErr *ValidationError
				if !asError(err, &validationErr) {
					t.Fatalf("expected a *ValidationError, got %#v", err)
				}
				if client != nil {
					t.Error("a rejected client should be nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			if client.SMS == nil || client.Messenger == nil || client.User == nil {
				t.Error("all three services should be wired up")
			}
		})
	}
}

func TestClientTrimsTrailingSlashFromBaseURL(t *testing.T) {
	client, err := NewClient("k", WithBaseURL("https://api.test/"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client.baseURL != "https://api.test" {
		t.Errorf("baseURL = %q, want no trailing slash", client.baseURL)
	}
}

func TestDefaultUserAgentPrefix(t *testing.T) {
	// Never assert the exact version: it comes from build info and is
	// "0+unknown" when the module is built in place, as it is under `go test`.
	if got := defaultUserAgent(); !strings.HasPrefix(got, "adsefid-go/") {
		t.Errorf("default user agent = %q, want an adsefid-go/ prefix", got)
	}
}

func TestRequestHeaders(t *testing.T) {
	client, requests := newFixtureClient(t, "envelopes/user.get_info.success.json")
	if _, err := client.User.GetInfo(context.Background()); err != nil {
		t.Fatalf("GetInfo: %v", err)
	}

	got := onlyRequest(t, requests)
	if got.Header.Get("X-API-KEY") != testAPIKey {
		t.Errorf("X-API-KEY = %q", got.Header.Get("X-API-KEY"))
	}
	if !strings.HasPrefix(got.Header.Get("User-Agent"), "adsefid-go/") {
		t.Errorf("User-Agent = %q", got.Header.Get("User-Agent"))
	}
	if got.Header.Get("Accept") != "application/json" {
		t.Errorf("Accept = %q", got.Header.Get("Accept"))
	}
	if got.Header.Get("Content-Type") != "" {
		t.Errorf("a GET carries no Content-Type, got %q", got.Header.Get("Content-Type"))
	}
}

func TestWithUserAgentOverridesTheDefault(t *testing.T) {
	var seen string
	srv := newRecordingServer(t, func(r *http.Request) { seen = r.Header.Get("User-Agent") },
		mustFixture(t, "envelopes/user.get_info.success.json"))

	client, err := NewClient(testAPIKey, WithBaseURL(srv), WithUserAgent("my-app/2.1"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := client.User.GetInfo(context.Background()); err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if seen != "my-app/2.1" {
		t.Errorf("User-Agent = %q, want my-app/2.1", seen)
	}
}

// WithHTTPClient takes precedence over WithTimeout, which only configures the
// client this package builds for itself.
func TestWithHTTPClientBeatsWithTimeout(t *testing.T) {
	custom := &http.Client{Timeout: 90 * time.Second}
	client, err := NewClient("k", WithTimeout(1*time.Second), WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client.httpClient != custom {
		t.Error("WithHTTPClient should win")
	}
	if client.httpClient.Timeout != 90*time.Second {
		t.Error("the supplied client's own timeout should be left alone")
	}
}

func TestContextCancellationSurfacesAsTransportError(t *testing.T) {
	client, _ := newFixtureClient(t, "envelopes/user.get_info.success.json")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.User.GetInfo(ctx)
	var transportErr *TransportError
	if !asError(err, &transportErr) {
		t.Fatalf("expected a *TransportError, got %#v", err)
	}
}
