package adsefid

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBuildIDsQuery(t *testing.T) {
	tests := []struct {
		name       string
		messageIDs []string
		localIDs   []string
		want       url.Values
		wantEmpty  bool
	}{
		{name: "neither", wantEmpty: true},
		{name: "empty slices", messageIDs: []string{}, localIDs: []string{}, wantEmpty: true},
		{name: "message ids only", messageIDs: []string{"a", "b"}, want: url.Values{"message_ids": {"a,b"}}},
		{name: "local ids only", localIDs: []string{"x"}, want: url.Values{"local_ids": {"x"}}},
		{
			name: "both", messageIDs: []string{"a"}, localIDs: []string{"x", "y"},
			want: url.Values{"message_ids": {"a"}, "local_ids": {"x,y"}},
		},
		{
			name: "values needing escaping", messageIDs: []string{"a b", "c%d"},
			want: url.Values{"message_ids": {"a b,c%d"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildIDsQuery(tc.messageIDs, tc.localIDs)
			if tc.wantEmpty {
				if got != "" {
					t.Errorf("expected no query string at all, got %q", got)
				}
				return
			}
			if !strings.HasPrefix(got, "?") {
				t.Fatalf("a non-empty query must start with '?', got %q", got)
			}
			parsed, err := url.ParseQuery(strings.TrimPrefix(got, "?"))
			if err != nil {
				t.Fatalf("parse %q: %v", got, err)
			}
			if parsed.Encode() != tc.want.Encode() {
				t.Errorf("query = %v, want %v", parsed, tc.want)
			}
		})
	}
}

func TestJoinCSV(t *testing.T) {
	tests := []struct {
		values []string
		want   string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "a,b,c"},
		{[]string{"a,b"}, "a,b"},
	}
	for _, tc := range tests {
		if got := joinCSV(tc.values); got != tc.want {
			t.Errorf("joinCSV(%v) = %q, want %q", tc.values, got, tc.want)
		}
	}
}

func TestEscapeQuotes(t *testing.T) {
	tests := []struct{ in, want string }{
		{"plain.txt", "plain.txt"},
		{`quo"te.txt`, `quo\"te.txt`},
		{`back\slash.txt`, `back\\slash.txt`},
		{`both"\.txt`, `both\"\\.txt`},
	}
	for _, tc := range tests {
		if got := escapeQuotes(tc.in); got != tc.want {
			t.Errorf("escapeQuotes(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMessengerUploadRejectsOversizedFile(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"exactly the limit", MaxMessengerFileBytes, false},
		{"one byte over", MaxMessengerFileBytes + 1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, requests := newTestClient(t, jsonHandler(200, []byte(`{"status":"success","data":{"file_id":"18dc580b-c598-4649-b2f4-dde1e5e35a57"}}`)))
			_, err := client.Messenger.UploadFile(context.Background(), bytes.NewReader(make([]byte, tc.size)), "a.pdf", "application/pdf")
			var verr *ValidationError
			if tc.wantErr {
				if !errors.As(err, &verr) || len(*requests) != 0 {
					t.Fatalf("want a *ValidationError and no request, got %v (%d requests)", err, len(*requests))
				}
				return
			}
			if err != nil {
				t.Fatalf("UploadFile: %v", err)
			}
		})
	}
}

func TestResponseBodyIsBounded(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(bytes.Repeat([]byte(" "), maxResponseBytes+10))
	})
	_, err := client.User.GetInfo(context.Background())
	var terr *TransportError
	if !errors.As(err, &terr) {
		t.Fatalf("want a *TransportError, got %v", err)
	}
}

func TestMoneyKeepsExactDecimalText(t *testing.T) {
	body := []byte(`{"status":"success","data":{"name":"n","company_name":null,"credit_left":1250000.10,"email":null,"phone":null,"account_status":"active"}}`)
	client, _ := newTestClient(t, jsonHandler(200, body))
	info, err := client.User.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.CreditLeft != "1250000.10" {
		t.Errorf("credit_left = %q, want exact wire text", info.CreditLeft)
	}
}

func TestTemplateRejectsNonFiniteNumber(t *testing.T) {
	client, requests := newTestClient(t, jsonHandler(200, []byte(`{}`)))
	_, err := client.SMS.SendTemplate(context.Background(), &SendTemplateSmsRequest{
		TemplateID: "t", Receptor: "09120000000", LineNumber: "3000",
		Parameters: map[string]TemplateParameterValue{"n": NumberParam(math.NaN())},
	})
	var verr *ValidationError
	if !errors.As(err, &verr) || len(*requests) != 0 {
		t.Fatalf("want a *ValidationError and no request, got %v", err)
	}
}

func TestErrorMappingWithoutEnvelope(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantCode  WebServiceResponseCode
		wantName  string
		wantLimit bool
	}{
		{"html 502", 502, "<html>bad gateway</html>", webServiceResponseCodeUnknown, "UNKNOWN_ERROR", false},
		{"bare 429", 429, "", WebServiceResponseCodeRequestLimitReached, "RATE_LIMITED", true},
		{"429 with other code", 429, `{"status":"error","error":{"code":2024,"name":"INVALID_PARAMETER"}}`, 2024, "INVALID_PARAMETER", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := newTestClient(t, jsonHandler(tc.status, []byte(tc.body)))
			_, err := client.User.GetInfo(context.Background())
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("want an *APIError, got %v", err)
			}
			if apiErr.Code != tc.wantCode || apiErr.Name != tc.wantName {
				t.Errorf("code/name = %d/%s", apiErr.Code, apiErr.Name)
			}
			if _, isLimit := IsRateLimitError(err); isLimit != tc.wantLimit {
				t.Errorf("rate limit = %v, want %v", isLimit, tc.wantLimit)
			}
		})
	}
}

func TestStatusLookupURLLength(t *testing.T) {
	ids := func(n int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = fmt.Sprintf("0195ff8c-55aa-7e74-a337-%012d", i)
		}
		return out
	}
	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		{"fits the 8000 character budget", 210, false},
		{"over the budget", 230, true},
		{"documented maximum", 2000, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, requests := newTestClient(t, jsonHandler(200, []byte(`{"status":"success","data":{"receptors":[]}}`)))
			_, err := client.SMS.GetStatus(context.Background(), &GetSmsStatusRequest{MessageIDs: ids(tc.count)})
			var verr *ValidationError
			if tc.wantErr {
				if !errors.As(err, &verr) || len(*requests) != 0 {
					t.Fatalf("want a *ValidationError and no request, got %v (%d requests)", err, len(*requests))
				}
				return
			}
			if err != nil {
				t.Fatalf("GetStatus: %v", err)
			}
			if raw := onlyRequest(t, requests).RawQuery; strings.Contains(raw, "%2C") || !strings.Contains(raw, ",") {
				t.Errorf("ids must be separated by literal commas, got %.80s", raw)
			}
		})
	}
}

func TestBaseURLPathIsKept(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"status":"success","data":{"name":"n","company_name":null,"credit_left":1,"email":null,"phone":null,"account_status":"active"}}`))
	}))
	defer srv.Close()
	client, err := NewClient("k", WithBaseURL(srv.URL+"/webservice/api/"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.User.GetInfo(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/webservice/api/v1/user/info" {
		t.Errorf("path = %q", gotPath)
	}
}
