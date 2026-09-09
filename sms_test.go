package adsefid

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

// TestSMSRequestBuilding pins the method, path and body shape of every SMS
// endpoint. The body assertions deliberately check that omitted optionals are
// absent rather than null — that is the `omitempty` contract callers rely on.
func TestSMSRequestBuilding(t *testing.T) {
	sendTime := time.Date(2026, 4, 4, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		fixture    string
		call       func(*Client) error
		wantMethod string
		wantPath   string
		wantQuery  url.Values
		assertBody func(*testing.T, map[string]any)
	}{
		{
			name:    "send single",
			fixture: "envelopes/sms.send_single.success.json",
			call: func(c *Client) error {
				_, err := c.SMS.SendSingle(context.Background(), &SendSingleSmsRequest{
					Receptor:   "98912xxxxxxx",
					LineNumber: "3000xxxx",
					Message:    "hello",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/sms/single",
			assertBody: func(t *testing.T, body map[string]any) {
				if body["receptor"] != "98912xxxxxxx" || body["message"] != "hello" {
					t.Errorf("unexpected body %v", body)
				}
				assertNoKeys(t, body, "send_time", "local_id", "hide", "line_selector")
			},
		},
		{
			name:    "send single with every optional",
			fixture: "envelopes/sms.send_single.success.json",
			call: func(c *Client) error {
				_, err := c.SMS.SendSingle(context.Background(), &SendSingleSmsRequest{
					Receptor:     "98912xxxxxxx",
					LineNumber:   "3000xxxx",
					Message:      "hello",
					SendTime:     &sendTime,
					LocalID:      ptr("order-1"),
					Hide:         ptr(true),
					LineSelector: ptr(LineSelectorBulkServiceSendBased),
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/sms/single",
			assertBody: func(t *testing.T, body map[string]any) {
				for _, key := range []string{"send_time", "local_id", "hide", "line_selector"} {
					if _, present := body[key]; !present {
						t.Errorf("body should carry %q", key)
					}
				}
				if body["local_id"] != "order-1" || body["hide"] != true {
					t.Errorf("unexpected body %v", body)
				}
			},
		},
		{
			name:    "send bulk",
			fixture: "envelopes/sms.send_bulk.partial_success.json",
			call: func(c *Client) error {
				_, err := c.SMS.SendBulk(context.Background(), &SendBulkSmsRequest{
					Receptors:  []BulkSmsReceptor{{Receptor: "98912xxxxxxx"}, {Receptor: "98993xxxxxxx"}},
					Message:    "bulk",
					LineNumber: "3000xxxx",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/sms/bulk",
			assertBody: func(t *testing.T, body map[string]any) {
				receptors, ok := body["receptors"].([]any)
				if !ok || len(receptors) != 2 {
					t.Fatalf("expected 2 receptors, got %v", body["receptors"])
				}
				first, _ := receptors[0].(map[string]any)
				assertNoKeys(t, first, "local_id", "hide")
			},
		},
		{
			name:    "send p2p",
			fixture: "envelopes/sms.send_p2p.partial_success.json",
			call: func(c *Client) error {
				_, err := c.SMS.SendP2P(context.Background(), &SendP2PSmsRequest{
					Messages:   []P2PSmsMessage{{Receptor: "98912xxxxxxx", Message: "a"}},
					LineNumber: "3000xxxx",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/sms/p2p",
			assertBody: func(t *testing.T, body map[string]any) {
				if _, ok := body["messages"].([]any); !ok {
					t.Errorf("expected a messages array, got %v", body["messages"])
				}
			},
		},
		{
			name:    "send template",
			fixture: "envelopes/sms.send_template.success.json",
			call: func(c *Client) error {
				_, err := c.SMS.SendTemplate(context.Background(), &SendTemplateSmsRequest{
					TemplateID: "otp_login",
					Parameters: map[string]TemplateParameterValue{
						"code":    StringParam("459122"),
						"invoice": StringParam("001234"),
						"minutes": IntParam(2),
					},
					Receptor:   "98912xxxxxxx",
					LineNumber: "3000xxxx",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/sms/template",
			assertBody: func(t *testing.T, body map[string]any) {
				params, ok := body["parameters"].(map[string]any)
				if !ok {
					t.Fatalf("expected a parameters object, got %v", body["parameters"])
				}
				// A number-typed parameter sent as a string keeps its leading
				// zeros; the server substitutes such a value verbatim.
				if params["invoice"] != "001234" {
					t.Errorf("leading zeros lost: %v", params["invoice"])
				}
				if params["minutes"] != float64(2) {
					t.Errorf("expected minutes to travel as a JSON number, got %T %v", params["minutes"], params["minutes"])
				}
			},
		},
		{
			name:    "get status",
			fixture: "envelopes/sms.get_status.success.json",
			call: func(c *Client) error {
				_, err := c.SMS.GetStatus(context.Background(), &GetSmsStatusRequest{MessageIDs: []string{"m1", "m2"}, LocalIDs: []string{"l1"}})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/sms/status",
			wantQuery:  url.Values{"message_ids": {"m1,m2"}, "local_ids": {"l1"}},
		},
		{
			name:    "cancel",
			fixture: "envelopes/sms.cancel.success.json",
			call: func(c *Client) error {
				_, err := c.SMS.Cancel(context.Background(), &CancelSmsRequest{MessageIDs: []string{"m1"}})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/sms/cancel",
			assertBody: func(t *testing.T, body map[string]any) {
				assertNoKeys(t, body, "local_ids")
			},
		},
		{
			name:    "get received",
			fixture: "envelopes/sms.get_received.success.json",
			call: func(c *Client) error {
				_, err := c.SMS.GetReceived(context.Background(), &GetReceivedSmsRequest{LineNumber: "3000xxxx", Count: ptr(10)})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/sms/receive",
			wantQuery:  url.Values{"line_number": {"3000xxxx"}, "count": {"10"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, requests := newFixtureClient(t, tc.fixture)
			if err := tc.call(client); err != nil {
				t.Fatalf("call: %v", err)
			}

			got := onlyRequest(t, requests)
			if got.Method != tc.wantMethod {
				t.Errorf("method = %s, want %s", got.Method, tc.wantMethod)
			}
			if got.Path != tc.wantPath {
				t.Errorf("path = %s, want %s", got.Path, tc.wantPath)
			}
			if got.Header.Get("X-API-KEY") != testAPIKey {
				t.Errorf("X-API-KEY = %q", got.Header.Get("X-API-KEY"))
			}
			if got.Header.Get("Accept") != "application/json" {
				t.Errorf("Accept = %q", got.Header.Get("Accept"))
			}
			if tc.wantQuery != nil {
				// Compare parsed values: Go sorts query keys, other SDKs do not.
				gotQuery, err := url.ParseQuery(got.RawQuery)
				if err != nil {
					t.Fatalf("parse query %q: %v", got.RawQuery, err)
				}
				if gotQuery.Encode() != tc.wantQuery.Encode() {
					t.Errorf("query = %v, want %v", gotQuery, tc.wantQuery)
				}
			}
			if tc.assertBody != nil {
				tc.assertBody(t, decodeBody(t, got))
			}
		})
	}
}

// TestSMSResponseParsing feeds each endpoint the API documentation's own
// example body and checks the typed result.
func TestSMSResponseParsing(t *testing.T) {
	t.Run("send single", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.send_single.success.json")
		got, err := client.SMS.SendSingle(context.Background(), &SendSingleSmsRequest{
			Receptor: "98912xxxxxxx", LineNumber: "3000xxxx", Message: "hi",
		})
		if err != nil {
			t.Fatalf("SendSingle: %v", err)
		}
		if got.Status != WebServiceMessageStatusScheduled {
			t.Errorf("status = %v, want scheduled(1000)", got.Status)
		}
		if got.SegmentCount != 1 || got.Cost != 120 {
			t.Errorf("segment/cost = %d/%v", got.SegmentCount, got.Cost)
		}
		if got.SendTime == nil {
			t.Error("send_time should have parsed")
		}
	})

	t.Run("send template round-trips its parameters", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.send_template.success.json")
		got, err := client.SMS.SendTemplate(context.Background(), &SendTemplateSmsRequest{
			TemplateID: "otp_login",
			Parameters: map[string]TemplateParameterValue{"code": StringParam("459122")},
			Receptor:   "98912xxxxxxx", LineNumber: "3000xxxx",
		})
		if err != nil {
			t.Fatalf("SendTemplate: %v", err)
		}
		code, ok := got.Parameters["code"].StringValue()
		if !ok || code != "459122" {
			t.Errorf("code parameter = %q (string=%v)", code, ok)
		}
		if !got.Parameters["minutes"].IsNumber() {
			t.Error("minutes should have come back as a number")
		}
		if got.ExpiryDate == nil {
			t.Error("expiry_date should have parsed")
		}
	})

	t.Run("get status", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.get_status.success.json")
		got, err := client.SMS.GetStatus(context.Background(), &GetSmsStatusRequest{MessageIDs: []string{"m1"}})
		if err != nil {
			t.Fatalf("GetStatus: %v", err)
		}
		if len(got.Receptors) == 0 {
			t.Fatal("expected at least one receptor")
		}
	})

	t.Run("get received", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.get_received.success.json")
		got, err := client.SMS.GetReceived(context.Background(), &GetReceivedSmsRequest{LineNumber: "3000xxxx"})
		if err != nil {
			t.Fatalf("GetReceived: %v", err)
		}
		if len(got.Messages) == 0 {
			t.Fatal("expected at least one message")
		}
		if got.Messages[0].ReceiveDate.IsZero() {
			t.Error("receive_date should have parsed")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.cancel.success.json")
		got, err := client.SMS.Cancel(context.Background(), &CancelSmsRequest{LocalIDs: []string{"l1"}})
		if err != nil {
			t.Fatalf("Cancel: %v", err)
		}
		if len(got.CancelledMessages)+len(got.FailedToCancel) == 0 {
			t.Error("expected some cancel results")
		}
	})
}

// TestPartialSuccessIsNotAnError locks in the contract that a bulk or P2P send
// returning HTTP 200 with per-item failure codes is a normal typed response,
// never an error.
func TestPartialSuccessIsNotAnError(t *testing.T) {
	t.Run("bulk", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.send_bulk.partial_success.json")
		got, err := client.SMS.SendBulk(context.Background(), &SendBulkSmsRequest{
			Receptors:  []BulkSmsReceptor{{Receptor: "a"}, {Receptor: "b"}},
			Message:    "m",
			LineNumber: "3000xxxx",
		})
		if err != nil {
			t.Fatalf("a partial success must not be an error, got %v", err)
		}
		if len(got.Receptors) != 2 {
			t.Fatalf("expected 2 receptor results, got %d", len(got.Receptors))
		}
		if got.Receptors[0].Status != 1000 {
			t.Errorf("first receptor status = %d, want 1000", got.Receptors[0].Status)
		}
		// 2025 RECEPTOR_BLACKLISTED is a WebServiceResponseCode, not a message
		// status — which is exactly why Status is a plain int here.
		if got.Receptors[1].Status != 2025 {
			t.Errorf("second receptor status = %d, want 2025", got.Receptors[1].Status)
		}
		if got.Receptors[1].MessageID != nil {
			t.Error("a failed receptor should have a null message_id")
		}
		// The typed views split the WebServiceCode by range, so a caller never compares raw ints.
		if status, ok := got.Receptors[0].MessageStatus(); !ok || status != WebServiceMessageStatusScheduled {
			t.Errorf("first receptor MessageStatus() = %v, %v; want Scheduled, true", status, ok)
		}
		if _, ok := got.Receptors[0].ErrorCode(); ok {
			t.Error("an accepted receptor must not report an ErrorCode")
		}
		if _, ok := got.Receptors[1].MessageStatus(); ok {
			t.Error("a rejected receptor must not report a MessageStatus")
		}
		if code, ok := got.Receptors[1].ErrorCode(); !ok || code != WebServiceResponseCodeReceptorBlacklisted {
			t.Errorf("second receptor ErrorCode() = %v, %v; want ReceptorBlacklisted, true", code, ok)
		}
		if got.Counts["2025"] != 1 || got.TotalCount != 2 {
			t.Errorf("counts = %v, total = %d", got.Counts, got.TotalCount)
		}
	})

	t.Run("p2p", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.send_p2p.partial_success.json")
		got, err := client.SMS.SendP2P(context.Background(), &SendP2PSmsRequest{
			Messages:   []P2PSmsMessage{{Receptor: "a", Message: "x"}},
			LineNumber: "3000xxxx",
		})
		if err != nil {
			t.Fatalf("a partial success must not be an error, got %v", err)
		}
		if code, ok := got.Messages[1].ErrorCode(); !ok || code != WebServiceResponseCodeInvalidReceptor {
			t.Errorf("second message ErrorCode() = %v, %v; want InvalidReceptor, true", code, ok)
		}
		if len(got.Messages) != 2 || got.Messages[1].Status != 2014 {
			t.Errorf("unexpected messages %+v", got.Messages)
		}
	})
}

// TestSMSValidationRejectsBeforeSending asserts every pre-flight check fires
// without a request ever reaching the network.
func TestSMSValidationRejectsBeforeSending(t *testing.T) {
	tests := []struct {
		name string
		call func(*Client) error
	}{
		{"nil request", func(c *Client) error {
			_, err := c.SMS.SendSingle(context.Background(), nil)
			return err
		}},
		{"empty receptor", func(c *Client) error {
			_, err := c.SMS.SendSingle(context.Background(), &SendSingleSmsRequest{LineNumber: "3000", Message: "m"})
			return err
		}},
		{"empty line number", func(c *Client) error {
			_, err := c.SMS.SendSingle(context.Background(), &SendSingleSmsRequest{Receptor: "98912", Message: "m"})
			return err
		}},
		{"empty message", func(c *Client) error {
			_, err := c.SMS.SendSingle(context.Background(), &SendSingleSmsRequest{Receptor: "98912", LineNumber: "3000"})
			return err
		}},
		{"message over 900 units", func(c *Client) error {
			_, err := c.SMS.SendSingle(context.Background(), &SendSingleSmsRequest{
				Receptor: "98912", LineNumber: "3000", Message: repeatRune('a', maxSmsMessageLength+1),
			})
			return err
		}},
		{"invalid local id", func(c *Client) error {
			_, err := c.SMS.SendSingle(context.Background(), &SendSingleSmsRequest{
				Receptor: "98912", LineNumber: "3000", Message: "m", LocalID: ptr("-bad"),
			})
			return err
		}},
		{"empty receptors slice", func(c *Client) error {
			_, err := c.SMS.SendBulk(context.Background(), &SendBulkSmsRequest{
				Receptors: nil, Message: "m", LineNumber: "3000",
			})
			return err
		}},
		{"empty messages slice", func(c *Client) error {
			_, err := c.SMS.SendP2P(context.Background(), &SendP2PSmsRequest{Messages: nil, LineNumber: "3000"})
			return err
		}},
		{"get status with neither id list", func(c *Client) error {
			_, err := c.SMS.GetStatus(context.Background(), &GetSmsStatusRequest{})
			return err
		}},
		{"cancel with neither id list", func(c *Client) error {
			_, err := c.SMS.Cancel(context.Background(), &CancelSmsRequest{})
			return err
		}},
		{"received count over 499", func(c *Client) error {
			_, err := c.SMS.GetReceived(context.Background(), &GetReceivedSmsRequest{LineNumber: "3000", Count: ptr(maxReceivedCount + 1)})
			return err
		}},
		{"received count of zero", func(c *Client) error {
			// The service requires 1..499, so 0 is rejected here rather than
			// being sent and bounced back as an API error.
			_, err := c.SMS.GetReceived(context.Background(), &GetReceivedSmsRequest{LineNumber: "3000", Count: ptr(0)})
			return err
		}},
		{"negative received count", func(c *Client) error {
			_, err := c.SMS.GetReceived(context.Background(), &GetReceivedSmsRequest{LineNumber: "3000", Count: ptr(-1)})
			return err
		}},
		{"received with empty line number", func(c *Client) error {
			_, err := c.SMS.GetReceived(context.Background(), &GetReceivedSmsRequest{})
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, requests := newFixtureClient(t, "envelopes/sms.send_single.success.json")
			err := tc.call(client)
			var validationErr *ValidationError
			if !asError(err, &validationErr) {
				t.Fatalf("expected a *ValidationError, got %#v", err)
			}
			if len(*requests) != 0 {
				t.Errorf("validation must reject before any HTTP call, but %d were made", len(*requests))
			}
		})
	}
}

// TestGetStatusCombinedIDLimit checks the 2000-ID ceiling counts distinct
// values, so duplicates do not push a valid call over the limit.
func TestGetStatusCombinedIDLimit(t *testing.T) {
	atLimit := make([]string, maxCombinedIDsLookup)
	for i := range atLimit {
		atLimit[i] = "id-" + itoa(i)
	}

	client, requests := newFixtureClient(t, "envelopes/sms.get_status.success.json")
	if _, err := client.SMS.GetStatus(context.Background(), &GetSmsStatusRequest{MessageIDs: atLimit}); err != nil {
		t.Fatalf("exactly 2000 distinct ids should be accepted: %v", err)
	}
	if len(*requests) != 1 {
		t.Fatalf("expected the call to reach the network")
	}

	overLimit := append(append([]string{}, atLimit...), "one-too-many")
	client2, requests2 := newFixtureClient(t, "envelopes/sms.get_status.success.json")
	if _, err := client2.SMS.GetStatus(context.Background(), &GetSmsStatusRequest{MessageIDs: overLimit}); err == nil {
		t.Error("2001 distinct ids should be rejected")
	}
	if len(*requests2) != 0 {
		t.Error("the over-limit call must not reach the network")
	}

	duplicates := make([]string, maxCombinedIDsLookup+50)
	for i := range duplicates {
		duplicates[i] = "same-id"
	}
	client3, _ := newFixtureClient(t, "envelopes/sms.get_status.success.json")
	if _, err := client3.SMS.GetStatus(context.Background(), &GetSmsStatusRequest{MessageIDs: duplicates}); err != nil {
		t.Errorf("duplicates collapse to one distinct id and should be accepted: %v", err)
	}
}

// TestGetReceivedCountBoundaries pins the inclusive 1..499 range the service
// enforces.
func TestGetReceivedCountBoundaries(t *testing.T) {
	for _, count := range []int{minReceivedCount, 250, maxReceivedCount} {
		client, requests := newFixtureClient(t, "envelopes/sms.get_received.success.json")
		if _, err := client.SMS.GetReceived(context.Background(), &GetReceivedSmsRequest{LineNumber: "3000xxxx", Count: ptr(count)}); err != nil {
			t.Errorf("count=%d should be accepted: %v", count, err)
		}
		if len(*requests) != 1 {
			t.Errorf("count=%d should have reached the network", count)
		}
	}
}
