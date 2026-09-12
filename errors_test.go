package adsefid

import (
	"context"
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

			if tc.wantDetails && apiErr.Details == nil {
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

// TestErrorDetailsShapes covers the structured details shared by every
// endpoint family.
func TestErrorDetailsShapes(t *testing.T) {
	t.Run("validation field errors", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(
			http.StatusBadRequest, mustFixture(t, "errors/error.invalid_parameter.json")))

		_, err := client.User.GetTemplates(context.Background(), nil)

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected an *APIError, got %#v", err)
		}

		if apiErr.Details == nil {
			t.Fatal("expected details")
		}
		if got := apiErr.Details.Errors["take"]; got.Code != WebServiceResponseCodeInvalidParameter || got.Name != "INVALID_PARAMETER" {
			t.Errorf("errors[take] = %+v", got)
		}
	})

	t.Run("single-send field errors", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(
			http.StatusBadRequest, mustFixture(t, "errors/error.details_single.json")))

		_, err := client.User.GetInfo(context.Background())

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected an *APIError, got %#v", err)
		}

		if apiErr.Details == nil {
			t.Fatal("expected details")
		}
		if got := apiErr.Details.Errors["receptor"]; got.Code != WebServiceResponseCodeInvalidReceptor {
			t.Errorf("errors[receptor] = %+v", got)
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

		if apiErr.Details == nil {
			t.Fatal("expected details")
		}
		if len(apiErr.Details.Items) != 2 {
			t.Fatalf("expected 2 per-item errors, got %d", len(apiErr.Details.Items))
		}
		// The index says which item of YOUR array failed; it is not positional
		// in this list, so a gap (0 then 2) is normal.
		if apiErr.Details.Items[1].Index != 2 {
			t.Errorf("second item index = %d, want 2", apiErr.Details.Items[1].Index)
		}
		if got := apiErr.Details.Items[1].Errors["local_id"]; got.Code != WebServiceResponseCodeDuplicateLocalID {
			t.Errorf("local_id error = %+v", got)
		}
	})

	t.Run("cancel errors keyed by rejected id", func(t *testing.T) {
		client, _ := newTestClient(t, jsonHandler(
			http.StatusBadRequest, mustFixture(t, "errors/error.details_cancel.json")))

		_, err := client.SMS.Cancel(context.Background(), &CancelSmsRequest{LocalIDs: []string{"a"}})

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected an *APIError, got %#v", err)
		}

		if apiErr.Details == nil || len(apiErr.Details.Errors) != 2 {
			t.Fatalf("details = %+v, want 2 errors", apiErr.Details)
		}
		if got := apiErr.Details.Errors["order-10001"]; got.Code != WebServiceResponseCodeInvalidLocalIDs {
			t.Errorf("errors[order-10001] = %+v", got)
		}
	})

	t.Run("unknown nested code keeps its integer", func(t *testing.T) {
		body := []byte(`{"status":"error","error":{"code":2024,"name":"INVALID_PARAMETER","details":{"errors":{"future":{"code":2999,"name":"FUTURE_CODE"}}}}}`)
		client, _ := newTestClient(t, jsonHandler(http.StatusBadRequest, body))

		_, err := client.User.GetInfo(context.Background())
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Details == nil {
			t.Fatalf("expected typed API details, got %#v", err)
		}
		if got := apiErr.Details.Errors["future"].Code; got != WebServiceResponseCode(2999) {
			t.Errorf("code = %d, want 2999", got)
		}
	})
}
