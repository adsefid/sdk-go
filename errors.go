package adsefid

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ValidationError reports a client-side pre-flight validation failure. It is
// returned before any network request is made.
type ValidationError struct {
	// Field is the name of the offending request field, when known.
	Field string
	// Message describes what is wrong with Field.
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// APIError reports a non-success response envelope, or a non-2xx HTTP status
// that could not be decoded as one, returned by the adsefid.com API.
type APIError struct {
	// Code is the API's error code. It is -1 when the HTTP status was non-2xx
	// but the response body could not be decoded as an error envelope.
	Code WebServiceResponseCode
	// Name is the SCREAMING_SNAKE error name reported by the API (e.g.
	// "INVALID_API_KEY"), or "UNKNOWN_ERROR" when Code is -1.
	Name string
	// HTTPStatusCode is the HTTP status code of the response.
	HTTPStatusCode int
	// Details carries the endpoint-specific error details payload, if any.
	// Its shape varies by endpoint and is intentionally left untyped; decode
	// it defensively per endpoint if you need it.
	Details json.RawMessage
}

func (e *APIError) Error() string {
	return fmt.Sprintf("adsefid: API error %s (%d), HTTP status %d", e.Name, int(e.Code), e.HTTPStatusCode)
}

// RateLimitError reports an APIError whose code indicates the caller has been
// rate-limited (WebServiceResponseCode 2035 MessageLimitReached, 2036
// RequestLimitReached, or a bare HTTP 429 with no decodable error envelope).
type RateLimitError struct {
	*APIError
}

func (e *RateLimitError) Unwrap() error { return e.APIError }

// IsRateLimitError reports whether err is, or wraps, a *RateLimitError, and
// returns it if so. It is a thin convenience wrapper around errors.As.
func IsRateLimitError(err error) (*RateLimitError, bool) {
	var rlErr *RateLimitError
	if errors.As(err, &rlErr) {
		return rlErr, true
	}
	return nil, false
}

// TransportError reports a network-level failure (connection error, timeout,
// response body/JSON that could not be decoded) that occurred while talking
// to the adsefid.com API.
type TransportError struct {
	Message string
	Err     error
}

func (e *TransportError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("adsefid: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("adsefid: %s", e.Message)
}

func (e *TransportError) Unwrap() error { return e.Err }

// WebhookVerificationError reports a failure to verify or parse an inbound
// webhook delivery: a bad signature, a stale timestamp, or a malformed or
// unrecognized payload.
type WebhookVerificationError struct {
	Message string
}

func (e *WebhookVerificationError) Error() string {
	return fmt.Sprintf("adsefid: webhook verification failed: %s", e.Message)
}
