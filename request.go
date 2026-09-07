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

// buildIDsQuery builds the "?message_ids=...&local_ids=..." query string
// shared by the get-status and similar lookup endpoints, CSV-joining each
// slice and omitting empty ones entirely. It returns "" when both slices are
// empty.
func buildIDsQuery(messageIDs, localIDs []string) string {
	values := url.Values{}
	if csv := joinCSV(messageIDs); csv != "" {
		values.Set("message_ids", csv)
	}
	if csv := joinCSV(localIDs); csv != "" {
		values.Set("local_ids", csv)
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func joinCSV(values []string) string {
	return strings.Join(values, ",")
}

// doGet performs a single-attempt GET request against path (which may already
// contain a query string) and decodes the envelope into T.
func doGet[T any](ctx context.Context, c *Client, path string) (T, error) {
	return doRequest[T](ctx, c, http.MethodGet, path, "", nil)
}

// doPostJSON performs a single-attempt POST request with a JSON-encoded body
// and decodes the envelope into T.
func doPostJSON[T any](ctx context.Context, c *Client, path string, body any) (T, error) {
	var zero T
	encoded, err := json.Marshal(body)
	if err != nil {
		return zero, &TransportError{Message: "failed to encode request body", Err: err}
	}
	return doRequest[T](ctx, c, http.MethodPost, path, "application/json", bytes.NewReader(encoded))
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
	if _, err := io.Copy(part, r); err != nil {
		return zero, &TransportError{Message: "failed to read the file to upload", Err: err}
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

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return zero, &TransportError{Message: "failed to build the HTTP request", Err: err}
	}
	req.Header.Set("X-API-KEY", c.apiKey)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, &TransportError{Message: "failed to read the response body", Err: err}
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
