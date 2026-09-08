package adsefid

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const testAPIKey = "test-api-key"

// recorded is one request as the test server saw it.
type recorded struct {
	Method   string
	Path     string
	RawQuery string
	Header   http.Header
	Body     []byte
}

// jsonHandler answers every request with a fixed status and body.
func jsonHandler(status int, body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}
}

// newTestClient starts an httptest server running h, points a Client at it and
// returns both the client and a pointer to the slice of requests the server
// observed. The server is closed when the test ends.
func newTestClient(t *testing.T, h http.HandlerFunc) (*Client, *[]recorded) {
	t.Helper()

	var requests []recorded
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests = append(requests, recorded{
			Method:   r.Method,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
			Header:   r.Header.Clone(),
			Body:     body,
		})
		h(w, r)
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(testAPIKey, WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, &requests
}

// newFixtureClient is newTestClient answering every request with the named
// fixture at HTTP 200.
func newFixtureClient(t *testing.T, fixture string) (*Client, *[]recorded) {
	t.Helper()
	return newTestClient(t, jsonHandler(http.StatusOK, mustFixture(t, fixture)))
}

func mustFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", filepath.FromSlash(name)))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return body
}

// onlyRequest asserts exactly one request reached the server and returns it.
func onlyRequest(t *testing.T, requests *[]recorded) recorded {
	t.Helper()
	if len(*requests) != 1 {
		t.Fatalf("expected exactly 1 request, got %d", len(*requests))
	}
	return (*requests)[0]
}

// decodeBody parses a recorded JSON request body into a generic map.
func decodeBody(t *testing.T, r recorded) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatalf("decode request body %q: %v", r.Body, err)
	}
	return body
}

// assertNoKeys fails if the request body carries any of the named keys. This is
// the "an omitted optional is absent, not null" contract.
func assertNoKeys(t *testing.T, body map[string]any, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if _, present := body[key]; present {
			t.Errorf("body should not carry %q, got %v", key, body[key])
		}
	}
}

// errRoundTripper fails every request at the transport layer.
type errRoundTripper struct{ err error }

func (rt errRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, rt.err
}

// repeatRune builds a string of n copies of r, for the length-limit tests.
func repeatRune(r rune, n int) string {
	return strings.Repeat(string(r), n)
}

// asError is errors.As with a friendlier name for the table-driven tests.
func asError[T error](err error, target *T) bool {
	return errors.As(err, target)
}

func itoa(i int) string { return strconv.Itoa(i) }

// newRecordingServer starts a server that calls inspect on every request and
// then answers with body at HTTP 200. It returns the server's base URL.
func newRecordingServer(t *testing.T, inspect func(*http.Request), body []byte) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inspect(r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}
