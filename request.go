package adsefid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const (
	// MaxMessengerFileBytes is the largest file content the Messenger upload
	// endpoint accepts: 15 MiB, inclusive. Multipart framing is not counted.
	MaxMessengerFileBytes = 15 * 1024 * 1024

	maxResponseBytes = 16 * 1024 * 1024

	// maxRequestTargetLength is the longest path-and-query the SDK sends. The
	// service sits behind a proxy that rejects request lines over 8 KiB with an
	// HTML 414, so a GET with a very long ID list is refused locally with a
	// *ValidationError instead.
	maxRequestTargetLength = 8000
)

func uniqueIDs(values []string) []string {
	seen := make(map[string]bool, len(values))
	ids := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			ids = append(ids, value)
		}
	}
	return ids
}

// buildIDsQuery builds the "?message_ids=...&local_ids=..." query string
// shared by the get-status and similar lookup endpoints, CSV-joining each
// slice and omitting empty ones entirely. It returns "" when both slices are
// empty.
func buildIDsQuery(messageIDs, localIDs []string) string {
	parts := make([]string, 0, 2)
	if csv := joinEscaped(messageIDs); csv != "" {
		parts = append(parts, "message_ids="+csv)
	}
	if csv := joinEscaped(localIDs); csv != "" {
		parts = append(parts, "local_ids="+csv)
	}
	if len(parts) == 0 {
		return ""
	}
	return "?" + strings.Join(parts, "&")
}

// joinEscaped query-escapes each value and joins them with a literal comma.
// The service reads the comma as the list separator, and leaving it unescaped
// keeps the URL about a third shorter than %2C would. Duplicates are dropped
// because the service collapses them anyway.
func joinEscaped(values []string) string {
	seen := make(map[string]struct{}, len(values))
	escaped := make([]string, 0, len(values))
	for _, value := range values {
		if _, dup := seen[value]; dup {
			continue
		}
		seen[value] = struct{}{}
		escaped = append(escaped, url.QueryEscape(value))
	}
	return strings.Join(escaped, ",")
}

func joinCSV(values []string) string {
	return strings.Join(values, ",")
}

// doGet performs a single-attempt GET request against path (which may already
// contain a query string) and decodes the envelope into T.
func doGet[T any](ctx context.Context, c *Client, path string) (T, error) {
	return doRequest[T](ctx, c, http.MethodGet, path, "", nil)
}

// doJSON performs a single-attempt request with a JSON-encoded body
// and decodes the envelope into T.
func doJSON[T any](ctx context.Context, c *Client, method, path string, body any) (T, error) {
	var zero T
	encoded, err := json.Marshal(body)
	if err != nil {
		return zero, &TransportError{Message: "failed to encode request body", Err: err}
	}
	return doRequest[T](ctx, c, method, path, "application/json", bytes.NewReader(encoded))
}

// doPostMultipart performs a single-attempt multipart/form-data POST request
// uploading r under the "file" form field, and decodes the envelope into T.
func doPostMultipart[T any](ctx context.Context, c *Client, path, fileName, contentType string, r io.Reader) (T, error) {
	var zero T

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	partHeader := make(map[string][]string)
	partHeader["Content-Disposition"] = []string{
		fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeQuotes(fileName)),
	}
	if contentType != "" {
		partHeader["Content-Type"] = []string{contentType}
	}

	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return zero, &TransportError{Message: "failed to build multipart request body", Err: err}
	}
	written, err := io.Copy(part, io.LimitReader(r, MaxMessengerFileBytes+1))
	if err != nil {
		return zero, &TransportError{Message: "failed to read the file to upload", Err: err}
	}
	if written > MaxMessengerFileBytes {
		return zero, &ValidationError{Field: "file", Message: "exceeds the maximum allowed size of 15 MiB"}
	}
	if err := writer.Close(); err != nil {
		return zero, &TransportError{Message: "failed to finalize multipart request body", Err: err}
	}

	return doRequest[T](ctx, c, http.MethodPost, path, writer.FormDataContentType(), &buf)
}

func escapeQuotes(s string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s)
}

// doRequest is the single choke point for every HTTP call this SDK makes. It
// never retries: exactly one HTTP request is sent per call.
func doRequest[T any](ctx context.Context, c *Client, method, path, contentType string, body io.Reader) (T, error) {
	var zero T

	if method == http.MethodGet {
		if length := len(c.basePath) + len(path); length > maxRequestTargetLength {
			return zero, &ValidationError{
				Message: fmt.Sprintf("the request URL is %d characters long, over the %d the service accepts; send fewer IDs per call", length, maxRequestTargetLength),
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return zero, &TransportError{Message: "failed to build the HTTP request", Err: err}
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, ctxErr) {
			return zero, &TransportError{Message: "request canceled or timed out", Err: err}
		}
		return zero, &TransportError{Message: "a transport-level error occurred while calling the adsefid.com API", Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return zero, &TransportError{Message: "failed to read the response body", Err: err}
	}
	if len(respBody) > maxResponseBytes {
		return zero, &TransportError{Message: "the response body exceeds the maximum supported size"}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var envelope struct {
			Status string          `json:"status"`
			Data   json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(respBody, &envelope); err != nil {
			return zero, &TransportError{Message: "failed to decode a successful adsefid.com API response", Err: err}
		}
		if envelope.Status == "error" {
			return zero, mapErrorResponse(resp.StatusCode, respBody)
		}
		if envelope.Status != "success" || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
			return zero, &TransportError{Message: "the adsefid.com API returned a malformed success envelope"}
		}
		var data T
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return zero, &TransportError{Message: "failed to decode a successful adsefid.com API response", Err: err}
		}
		return data, nil
	}

	return zero, mapErrorResponse(resp.StatusCode, respBody)
}

func mapErrorResponse(httpStatusCode int, respBody []byte) error {
	var envelope errorEnvelope
	if err := json.Unmarshal(respBody, &envelope); err != nil || envelope.Error == nil {
		apiErr := &APIError{
			Code:           webServiceResponseCodeUnknown,
			Name:           "UNKNOWN_ERROR",
			HTTPStatusCode: httpStatusCode,
		}
		if httpStatusCode == http.StatusTooManyRequests {
			apiErr.Code = WebServiceResponseCodeRequestLimitReached
			apiErr.Name = "RATE_LIMITED"
			return &RateLimitError{APIError: apiErr}
		}
		return apiErr
	}

	code := WebServiceResponseCode(envelope.Error.Code)
	apiErr := &APIError{
		Code:           code,
		Name:           envelope.Error.Name,
		HTTPStatusCode: httpStatusCode,
		Details:        envelope.Error.Details,
	}

	if code == WebServiceResponseCodeMessageLimitReached || code == WebServiceResponseCodeRequestLimitReached || httpStatusCode == http.StatusTooManyRequests {
		return &RateLimitError{APIError: apiErr}
	}

	return apiErr
}
