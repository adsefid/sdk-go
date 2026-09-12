package adsefid

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestMessengerRequestBuilding(t *testing.T) {
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
			fixture: "envelopes/messenger.send_single.success.json",
			call: func(c *Client) error {
				_, err := c.Messenger.SendSingle(context.Background(), &SendSingleMessengerRequest{
					Message: "hi", Receptor: "98912xxxxxxx", Profile: "profile-1",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/messenger/single",
			assertBody: func(t *testing.T, body map[string]any) {
				assertNoKeys(t, body, "hide", "file_id", "send_time", "local_id")
			},
		},
		{
			name:    "send bulk",
			fixture: "envelopes/messenger.send_bulk.partial_success.json",
			call: func(c *Client) error {
				_, err := c.Messenger.SendBulk(context.Background(), &SendBulkMessengerRequest{
					Receptors: []BulkMessengerReceptor{{Receptor: "a"}, {Receptor: "", LocalID: ptr("-bad")}}, Message: "m", Profile: "p",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/messenger/bulk",
			assertBody: func(t *testing.T, body map[string]any) {
				if receptors, ok := body["receptors"].([]any); !ok || len(receptors) != 2 {
					t.Errorf("expected 2 receptors, got %v", body["receptors"])
				}
			},
		},
		{
			name:    "send p2p",
			fixture: "envelopes/messenger.send_p2p.partial_success.json",
			call: func(c *Client) error {
				_, err := c.Messenger.SendP2P(context.Background(), &SendP2PMessengerRequest{
					Receptors: []P2PMessengerReceptor{{Receptor: "a", Message: "m"}, {Receptor: "", Message: ""}}, Profile: "p",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/messenger/p2p",
			assertBody: func(t *testing.T, body map[string]any) {
				if receptors, ok := body["receptors"].([]any); !ok || len(receptors) != 2 {
					t.Errorf("expected 2 receptors, got %v", body["receptors"])
				}
			},
		},
		{
			name:    "cancel",
			fixture: "envelopes/messenger.cancel.success.json",
			call: func(c *Client) error {
				_, err := c.Messenger.Cancel(context.Background(), &CancelMessengerRequest{MessageIDs: []string{"m1"}})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/messenger/cancel",
		},
		{
			name:    "send template",
			fixture: "envelopes/messenger.send_template.success.json",
			call: func(c *Client) error {
				_, err := c.Messenger.SendTemplate(context.Background(), &SendTemplateMessengerRequest{
					TemplateID: "invoice_notice",
					Parameters: map[string]TemplateParameterValue{"invoice": StringParam("001234")},
					Receptor:   "98912xxxxxxx",
					Profile:    "p",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/messenger/template",
			assertBody: func(t *testing.T, body map[string]any) {
				params, _ := body["parameters"].(map[string]any)
				if params["invoice"] != "001234" {
					t.Errorf("leading zeros lost: %v", params["invoice"])
				}
			},
		},
		{
			name:    "get status",
			fixture: "envelopes/messenger.get_status.success.json",
			call: func(c *Client) error {
				_, err := c.Messenger.GetStatus(context.Background(), &GetMessengerStatusRequest{MessageIDs: []string{"m1"}, LocalIDs: []string{"l1"}})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/messenger/status",
			wantQuery:  url.Values{"message_ids": {"m1"}, "local_ids": {"l1"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, requests := newFixtureClient(t, tc.fixture)
			if err := tc.call(client); err != nil {
				t.Fatalf("call: %v", err)
			}
			got := onlyRequest(t, requests)
			if got.Method != tc.wantMethod || got.Path != tc.wantPath {
				t.Errorf("%s %s, want %s %s", got.Method, got.Path, tc.wantMethod, tc.wantPath)
			}
			if tc.wantQuery != nil {
				gotQuery, _ := url.ParseQuery(got.RawQuery)
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

// TestMessengerUploadFileMultipart parses the recorded body with a real
// multipart reader — the boundary is random, so a raw string comparison would
// be meaningless.
func TestMessengerUploadFileMultipart(t *testing.T) {
	tests := []struct {
		name            string
		fileName        string
		contentType     string
		wantFileName    string
		wantContentType string
	}{
		{"plain", "invoice.pdf", "application/pdf", "invoice.pdf", "application/pdf"},
		{"quotes in filename", `we"ird\name.txt`, "text/plain", `we"ird\name.txt`, "text/plain"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, requests := newFixtureClient(t, "envelopes/messenger.upload_file.success.json")
			payload := "file-contents-here"
			got, err := client.Messenger.UploadFile(context.Background(), strings.NewReader(payload), tc.fileName, tc.contentType)
			if err != nil {
				t.Fatalf("UploadFile: %v", err)
			}
			if got.FileID == "" {
				t.Error("expected a file_id")
			}

			req := onlyRequest(t, requests)
			mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
			if err != nil {
				t.Fatalf("parse Content-Type %q: %v", req.Header.Get("Content-Type"), err)
			}
			if mediaType != "multipart/form-data" {
				t.Fatalf("media type = %s", mediaType)
			}

			reader := multipart.NewReader(strings.NewReader(string(req.Body)), params["boundary"])
			part, err := reader.NextPart()
			if err != nil {
				t.Fatalf("read first part: %v", err)
			}
			if part.FormName() != "file" {
				t.Errorf("form field = %q, want %q", part.FormName(), "file")
			}
			if part.FileName() != tc.wantFileName {
				t.Errorf("filename = %q, want %q", part.FileName(), tc.wantFileName)
			}
			if got := part.Header.Get("Content-Type"); got != tc.wantContentType {
				t.Errorf("part Content-Type = %q, want %q", got, tc.wantContentType)
			}
			content, err := io.ReadAll(part)
			if err != nil {
				t.Fatalf("read part body: %v", err)
			}
			if string(content) != payload {
				t.Errorf("part body = %q, want %q", content, payload)
			}
			if _, err := reader.NextPart(); err == nil {
				t.Error("expected exactly one part")
			}
		})
	}
}

func TestMessengerValidationRejectsBeforeSending(t *testing.T) {
	tests := []struct {
		name string
		call func(*Client) error
	}{
		{"empty message", func(c *Client) error {
			_, err := c.Messenger.SendSingle(context.Background(), &SendSingleMessengerRequest{Receptor: "a", Profile: "p"})
			return err
		}},
		{"empty profile", func(c *Client) error {
			_, err := c.Messenger.SendSingle(context.Background(), &SendSingleMessengerRequest{Receptor: "a", Message: "m"})
			return err
		}},
		{"message over 4000 units", func(c *Client) error {
			_, err := c.Messenger.SendSingle(context.Background(), &SendSingleMessengerRequest{
				Receptor: "a", Profile: "p", Message: repeatRune('a', maxMessengerMessageLength+1),
			})
			return err
		}},
		{"upload with empty filename", func(c *Client) error {
			_, err := c.Messenger.UploadFile(context.Background(), strings.NewReader("x"), "", "text/plain")
			return err
		}},
		{"upload with empty content type", func(c *Client) error {
			_, err := c.Messenger.UploadFile(context.Background(), strings.NewReader("x"), "note.txt", "")
			return err
		}},
		{"cancel with neither id list", func(c *Client) error {
			_, err := c.Messenger.Cancel(context.Background(), &CancelMessengerRequest{})
			return err
		}},
		{"get status with neither id list", func(c *Client) error {
			_, err := c.Messenger.GetStatus(context.Background(), &GetMessengerStatusRequest{})
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, requests := newFixtureClient(t, "envelopes/messenger.send_single.success.json")
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

// TestMessengerPartialSuccessIsNotAnError mirrors the SMS case: a messenger
// bulk or P2P send answers HTTP 200 even when individual receptors failed.
func TestMessengerPartialSuccessIsNotAnError(t *testing.T) {
	t.Run("bulk", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/messenger.send_bulk.partial_success.json")
		got, err := client.Messenger.SendBulk(context.Background(), &SendBulkMessengerRequest{
			Receptors: []BulkMessengerReceptor{{Receptor: "a"}, {Receptor: "b"}},
			Message:   "m",
			Profile:   "p",
		})
		if err != nil {
			t.Fatalf("a partial success must not be an error, got %v", err)
		}
		if len(got.Receptors) != 2 {
			t.Fatalf("expected 2 receptor results, got %d", len(got.Receptors))
		}
		if got.Receptors[0].Status != 1000 || got.Receptors[1].Status != 2025 {
			t.Errorf("statuses = %d/%d, want 1000/2025", got.Receptors[0].Status, got.Receptors[1].Status)
		}
		if got.Receptors[1].MessageID != nil {
			t.Error("a failed receptor should have a null message_id")
		}
		if got.Counts["2025"] != 1 {
			t.Errorf("counts = %v", got.Counts)
		}
	})

	t.Run("p2p", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/messenger.send_p2p.partial_success.json")
		got, err := client.Messenger.SendP2P(context.Background(), &SendP2PMessengerRequest{
			Receptors: []P2PMessengerReceptor{{Receptor: "a", Message: "m"}},
			Profile:   "p",
		})
		if err != nil {
			t.Fatalf("a partial success must not be an error, got %v", err)
		}
		if len(got.Receptors) != 2 || got.Receptors[1].Status != 2014 {
			t.Errorf("unexpected receptors %+v", got.Receptors)
		}
	})
}
