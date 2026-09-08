package adsefid

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

// TestErrorMapping walks the full HTTP status × envelope matrix. Note the
// 2036-at-HTTP-200 row: the response envelope's own status wins over the HTTP
// status code, which is the API's documented behaviour.
func TestErrorMapping(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		httpStatus    int
		wantRateLimit bool
		wantCode      WebServiceResponseCode
		wantName      string
		wantDetails   bool
	}{
		{
			name:       "unauthorized",
			fixture:    "errors/error.invalid_api_key.json",
			httpStatus: http.StatusUnauthorized,
			wantCode:   WebServiceResponseCodeUnauthorized,
			wantName:   "UNAUTHORIZED",
		},
		{
			name:          "message rate limit",
			fixture:       "errors/error.rate_limit_message.json",
			httpStatus:    http.StatusTooManyRequests,
			wantRateLimit: true,
			wantCode:      WebServiceResponseCodeMessageLimitReached,
			wantName:      "MESSAGE_LIMIT_REACHED",
		},
		{
			name:          "request rate limit inside a 200 envelope",
			fixture:       "errors/error.rate_limit_request.json",
			httpStatus:    http.StatusOK,
			wantRateLimit: true,
			wantCode:      WebServiceResponseCodeRequestLimitReached,
			wantName:      "REQUEST_LIMIT_REACHED",
		},
		{
			name:        "invalid parameter keeps its details",
			fixture:     "errors/error.invalid_parameter.json",
			httpStatus:  http.StatusBadRequest,
			wantCode:    WebServiceResponseCodeInvalidParameter,
			wantName:    "INVALID_PARAMETER",
			wantDetails: true,
		},
		{
			name:       "an unmapped code is carried through, not rejected",
			fixture:    "errors/error.unknown_code.json",
			httpStatus: http.StatusBadRequest,
			wantCode:   WebServiceResponseCode(2999),
			wantName:   "SOME_FUTURE_SERVER_ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := mustFixture(t, tc.fixture)
			client, _ := newTestClient(t, jsonHandler(tc.httpStatus, body))

			_, err := client.User.GetInfo(context.Background())

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected an *APIError, got %#v", err)
			}
			if apiErr.Code != tc.wantCode {
				t.Errorf("code = %d, want %d", apiErr.Code, tc.wantCode)
			}
			if apiErr.Name != tc.wantName {
				t.Errorf("name = %q, want %q", apiErr.Name, tc.wantName)
			}
			if apiErr.HTTPStatusCode != tc.httpStatus {
				t.Errorf("http status = %d, want %d", apiErr.HTTPStatusCode, tc.httpStatus)
			}

			var rateLimitErr *RateLimitError
			gotRateLimit := errors.As(err, &rateLimitErr)
			if gotRateLimit != tc.wantRateLimit {
				t.Errorf("rate limit = %v, want %v", gotRateLimit, tc.wantRateLimit)
			}
			if _, ok := IsRateLimitError(err); ok != tc.wantRateLimit {
				t.Errorf("IsRateLimitError = %v, want %v", ok, tc.wantRateLimit)
			}

			if tc.wantDetails && len(apiErr.Details) == 0 {
				t.Error("expected details to survive")
			}
		})
	}
}

func TestErrorMappingForUndecodableBodies(t *testing.T) {
	body := mustFixture(t, "errors/error.not_json.txt")

	t.Run("500 with an HTML body", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(http.StatusInternalServerError, body))
		_, err := client.User.GetInfo(context.Background())

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected an *APIError, got %#v", err)
		}
		if apiErr.Name != "UNKNOWN_ERROR" || apiErr.HTTPStatusCode != http.StatusInternalServerError {
			t.Errorf("got %+v", apiErr)
		}
		var rateLimitErr *RateLimitError
		if errors.As(err, &rateLimitErr) {
			t.Error("a 500 is not a rate limit")
		}
	})

	t.Run("bare 429 with no envelope is still a rate limit", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(http.StatusTooManyRequests, body))
		_, err := client.User.GetInfo(context.Background())

		var rateLimitErr *RateLimitError
		if !errors.As(err, &rateLimitErr) {
			t.Fatalf("expected a *RateLimitError, got %#v", err)
		}
		if rateLimitErr.Code != WebServiceResponseCodeRequestLimitReached {
			t.Errorf("code = %d", rateLimitErr.Code)
		}
	})
}

// TestMalformedSuccessEnvelopes: a 2xx whose body is not a usable success
// envelope is a transport problem, not an API error.
func TestMalformedSuccessEnvelopes(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"no data field", `{"status":"success"}`},
		{"null data", `{"status":"success","data":null}`},
		{"empty body", ``},
		{"not json at all", `<html>nope</html>`},
		{"unknown status word", `{"status":"partial","data":{}}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := newTestClient(t, jsonHandler(http.StatusOK, []byte(tc.body)))
			_, err := client.User.GetInfo(context.Background())

			var transportErr *TransportError
			if !errors.As(err, &transportErr) {
				t.Fatalf("expected a *TransportError, got %#v", err)
			}
		})
	}
}

func TestTransportFailureIsWrapped(t *testing.T) {
	underlying := errors.New("dial tcp: connection refused")
	client, err := NewClient(testAPIKey,
		WithBaseURL("https://api.test"),
		WithHTTPClient(&http.Client{Transport: errRoundTripper{err: underlying}}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.User.GetInfo(context.Background())

	var transportErr *TransportError
	if !errors.As(err, &transportErr) {
		t.Fatalf("expected a *TransportError, got %#v", err)
	}
	if !errors.Is(err, underlying) {
		t.Error("the underlying transport error should stay unwrappable")
	}
}

// TestErrorHierarchy pins the relationship the READMEs promise: a rate limit
// is an API error, so a caller matching *APIError first would swallow it.
func TestErrorHierarchy(t *testing.T) {
	apiErr := &APIError{Code: WebServiceResponseCodeRequestLimitReached, Name: "REQUEST_LIMIT_REACHED", HTTPStatusCode: 429}
	rateLimitErr := &RateLimitError{APIError: apiErr}

	var asAPI *APIError
	if !errors.As(error(rateLimitErr), &asAPI) {
		t.Error("a *RateLimitError should unwrap to an *APIError")
	}
	if asAPI != apiErr {
		t.Error("unwrapping should yield the same *APIError")
	}
	if errors.Unwrap(rateLimitErr) != apiErr {
		t.Error("Unwrap should return the wrapped *APIError")
	}
}

// TestErrorDetailsShapes: Details is deliberately left as raw JSON because the
// service uses a different shape per endpoint. Decode each real shape to prove
// none of them is lost or coerced on the way through.
func TestErrorDetailsShapes(t *testing.T) {
	t.Run("validation errors are a field to message map under \"errors\"", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(
			http.StatusBadRequest, mustFixture(t, "errors/error.invalid_parameter.json")))

		_, err := client.User.GetTemplates(context.Background(), nil, nil, nil)

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected an *APIError, got %#v", err)
		}

		var details struct {
			Errors map[string]string `json:"errors"`
		}
		if err := json.Unmarshal(apiErr.Details, &details); err != nil {
			t.Fatalf("decode details: %v", err)
		}
		// Values are plain strings, one per field — not arrays.
		if got := details.Errors["take"]; got != "invalid value for take" {
			t.Errorf("errors[take] = %q", got)
		}
		if got := details.Errors["state"]; got != "invalid value for state" {
			t.Errorf("errors[state] = %q", got)
		}
	})

	t.Run("single-send errors are a flat field to message map", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(
			http.StatusBadRequest, mustFixture(t, "errors/error.details_single.json")))

		_, err := client.User.GetInfo(context.Background())

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected an *APIError, got %#v", err)
		}

		var details map[string]string
		if err := json.Unmarshal(apiErr.Details, &details); err != nil {
			t.Fatalf("decode details: %v", err)
		}
		if got := details["receptor"]; got != "invalid value for receptor" {
			t.Errorf("details[receptor] = %q", got)
		}
	})

	t.Run("bulk errors carry per-item errors keyed by index", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(
			http.StatusBadRequest, mustFixture(t, "errors/error.details_bulk.json")))

		_, err := client.User.GetInfo(context.Background())

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected an *APIError, got %#v", err)
		}

		var details struct {
			Errors   map[string]string `json:"errors"`
			Messages []struct {
				Index  int               `json:"index"`
				Errors map[string]string `json:"errors"`
			} `json:"messages"`
		}
		if err := json.Unmarshal(apiErr.Details, &details); err != nil {
			t.Fatalf("decode details: %v", err)
		}
		if details.Errors["line_number"] == "" {
			t.Error("request-level errors should survive")
		}
		if len(details.Messages) != 2 {
			t.Fatalf("expected 2 per-item errors, got %d", len(details.Messages))
		}
		// The index says which item of YOUR array failed; it is not positional
		// in this list, so a gap (0 then 2) is normal.
		if details.Messages[1].Index != 2 {
			t.Errorf("second item index = %d, want 2", details.Messages[1].Index)
		}
		if details.Messages[1].Errors["local_id"] == "" {
			t.Error("per-item errors should survive")
		}
	})

	t.Run("cancel errors are the one shape whose values are arrays", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(
			http.StatusBadRequest, mustFixture(t, "errors/error.details_cancel.json")))

		_, err := client.SMS.Cancel(context.Background(), &CancelSmsRequest{LocalIDs: []string{"a"}})

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected an *APIError, got %#v", err)
		}

		var details map[string][]string
		if err := json.Unmarshal(apiErr.Details, &details); err != nil {
			t.Fatalf("decode details: %v", err)
		}
		if len(details["local_ids"]) != 2 {
			t.Errorf("details[local_ids] = %v, want 2 entries", details["local_ids"])
		}
	})
}
