package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	adsefid "github.com/adsefid/sdk-go"
)

type signatureVector struct {
	Secret            string `json:"secret"`
	Timestamp         string `json:"timestamp"`
	BodyFile          string `json:"body_file"`
	Signature         string `json:"signature"`
	TamperedSignature string `json:"tampered_signature"`
}

func mustFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", filepath.FromSlash(name)))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return body
}

func loadVector(t *testing.T) (signatureVector, []byte) {
	t.Helper()
	var vector signatureVector
	if err := json.Unmarshal(mustFixture(t, "webhooks/signature_vector.json"), &vector); err != nil {
		t.Fatalf("decode signature_vector.json: %v", err)
	}
	return vector, mustFixture(t, vector.BodyFile)
}

// sign reproduces exactly what the server does: HMAC-SHA256 over
// "{timestamp}.{rawBody}" keyed with the raw secret bytes, Base64-encoded.
func sign(t *testing.T, secretBase64, timestamp string, rawBody []byte) string {
	t.Helper()
	key, err := base64.StdEncoding.DecodeString(secretBase64)
	if err != nil {
		t.Fatalf("decode secret: %v", err)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(timestamp + "."))
	mac.Write(rawBody)
	return "v1=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func nowTimestamp() string { return strconv.FormatInt(time.Now().Unix(), 10) }

// TestVerifyGoldenVector is the cross-SDK vector. It is the check that proves
// the HMAC key is the Base64-DECODED secret bytes, not the UTF-8 bytes of the
// Base64 string the panel displays. The vector's timestamp is fixed, so the
// max age has to be wide open.
func TestVerifyGoldenVector(t *testing.T) {
	vector, body := loadVector(t)

	event, err := Verify(body, vector.Signature, vector.Timestamp, vector.Secret, WithMaxAge(100*365*24*time.Hour))
	if err != nil {
		t.Fatalf("the golden vector must verify: %v", err)
	}

	receive, ok := event.(*ReceiveEvent)
	if !ok {
		t.Fatalf("expected a *ReceiveEvent, got %T", event)
	}
	if receive.Type() != EventTypeReceive {
		t.Errorf("type = %q", receive.Type())
	}
	if receive.ID() == "" || receive.Attempt() != 1 || receive.Version() != "1" {
		t.Errorf("metadata = %q/%d/%q", receive.ID(), receive.Attempt(), receive.Version())
	}
	if receive.OccurredAt().IsZero() {
		t.Error("occurred_at should have parsed")
	}
	if len(receive.Data) != 1 {
		t.Fatalf("expected 1 data item, got %d", len(receive.Data))
	}
	if receive.Data[0].Sender == "" || receive.Data[0].Message == "" {
		t.Errorf("data item = %+v", receive.Data[0])
	}
}

// A UTF-8-keyed HMAC (what this SDK used to compute) must NOT verify. This is
// the regression guard for the fix.
func TestVerifyRejectsAUTF8KeyedSignature(t *testing.T) {
	vector, body := loadVector(t)

	mac := hmac.New(sha256.New, []byte(vector.Secret))
	mac.Write([]byte(vector.Timestamp + "."))
	mac.Write(body)
	wrongSignature := "v1=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if wrongSignature == vector.Signature {
		t.Fatal("the test vector is degenerate: both keyings produced the same signature")
	}
	_, err := Verify(body, wrongSignature, vector.Timestamp, vector.Secret, WithMaxAge(100*365*24*time.Hour))
	if err == nil {
		t.Fatal("a signature keyed with the UTF-8 bytes of the Base64 secret must be rejected")
	}
}

func TestVerifyWithKeyAcceptsRawKeyBytes(t *testing.T) {
	vector, body := loadVector(t)
	key, err := base64.StdEncoding.DecodeString(vector.Secret)
	if err != nil {
		t.Fatalf("decode secret: %v", err)
	}

	if _, err := VerifyWithKey(body, vector.Signature, vector.Timestamp, key, WithMaxAge(100*365*24*time.Hour)); err != nil {
		t.Fatalf("VerifyWithKey: %v", err)
	}
}

func TestVerifyParsesEachEventType(t *testing.T) {
	vector, _ := loadVector(t)
	timestamp := nowTimestamp()

	tests := []struct {
		name    string
		fixture string
		assert  func(*testing.T, WebhookEvent)
	}{
		{
			name:    "receive",
			fixture: "webhooks/receive.body.json",
			assert: func(t *testing.T, event WebhookEvent) {
				got, ok := event.(*ReceiveEvent)
				if !ok {
					t.Fatalf("expected *ReceiveEvent, got %T", event)
				}
				if got.Data[0].LineNumber == "" {
					t.Error("line_number should have parsed")
				}
				if got.Data[0].ReceiveDate.IsZero() {
					t.Error("receive_date should have parsed")
				}
			},
		},
		{
			name:    "status",
			fixture: "webhooks/status.body.json",
			assert: func(t *testing.T, event WebhookEvent) {
				got, ok := event.(*StatusEvent)
				if !ok {
					t.Fatalf("expected *StatusEvent, got %T", event)
				}
				if got.Data[0].StatusDelivery != adsefid.WebServiceMessageStatusDelivered {
					t.Errorf("status_delivery = %v", got.Data[0].StatusDelivery)
				}
				if got.Data[0].LocalID == nil {
					t.Error("local_id should have parsed")
				}
			},
		},
		{
			name:    "messenger status",
			fixture: "webhooks/messenger_status.body.json",
			assert: func(t *testing.T, event WebhookEvent) {
				if _, ok := event.(*MessengerStatusEvent); !ok {
					t.Fatalf("expected *MessengerStatusEvent, got %T", event)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := mustFixture(t, tc.fixture)
			signature := sign(t, vector.Secret, timestamp, body)

			event, err := Verify(body, signature, timestamp, vector.Secret)
			if err != nil {
				t.Fatalf("Verify: %v", err)
			}
			tc.assert(t, event)
		})
	}
}

func TestVerifyRejections(t *testing.T) {
	vector, body := loadVector(t)
	freshTimestamp := nowTimestamp()
	freshSignature := sign(t, vector.Secret, freshTimestamp, body)

	staleTimestamp := strconv.FormatInt(time.Now().Add(-6*time.Minute).Unix(), 10)
	futureTimestamp := strconv.FormatInt(time.Now().Add(6*time.Minute).Unix(), 10)

	tests := []struct {
		name      string
		body      []byte
		signature string
		timestamp string
		secret    string
	}{
		{"tampered signature", body, vector.TamperedSignature, vector.Timestamp, vector.Secret},
		{"wrong secret", body, freshSignature, freshTimestamp, base64.StdEncoding.EncodeToString([]byte("a-different-32-byte-secret-value"))},
		{"missing v1= prefix", body, strings.TrimPrefix(freshSignature, "v1="), freshTimestamp, vector.Secret},
		{"signature is not base64", body, "v1=not-base-64-!!", freshTimestamp, vector.Secret},
		{"secret is not base64", body, freshSignature, freshTimestamp, "not base64 !!"},
		{"stale timestamp", body, sign(t, vector.Secret, staleTimestamp, body), staleTimestamp, vector.Secret},
		{"future timestamp", body, sign(t, vector.Secret, futureTimestamp, body), futureTimestamp, vector.Secret},
		{"non-numeric timestamp", body, freshSignature, "not-a-number", vector.Secret},
		{"empty timestamp", body, freshSignature, "", vector.Secret},
		{"tampered body", append(body, ' '), freshSignature, freshTimestamp, vector.Secret},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Assert the error type only, never the message: the order in which
			// the five SDKs run these checks differs, so the specific failure
			// reported for a doubly-invalid request is not portable.
			_, err := Verify(tc.body, tc.signature, tc.timestamp, tc.secret)
			var verificationErr *adsefid.WebhookVerificationError
			if !errors.As(err, &verificationErr) {
				t.Fatalf("expected a *WebhookVerificationError, got %#v", err)
			}
		})
	}
}

func TestVerifyRejectsUnparseablePayloads(t *testing.T) {
	vector, _ := loadVector(t)
	timestamp := nowTimestamp()

	tests := []struct {
		name string
		body []byte
	}{
		{"unknown event type", mustFixture(t, "webhooks/unknown_type.body.json")},
		{"not json", []byte(`<html>nope</html>`)},
		{"json but not an object", []byte(`[1,2,3]`)},
		{"data is the wrong shape", []byte(`{"id":"a","type":"receive","occurred_at":"2026-04-04T12:00:00+03:30","attempt":1,"version":"1","data":"nope"}`)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			signature := sign(t, vector.Secret, timestamp, tc.body)
			_, err := Verify(tc.body, signature, timestamp, vector.Secret)
			var verificationErr *adsefid.WebhookVerificationError
			if !errors.As(err, &verificationErr) {
				t.Fatalf("expected a *WebhookVerificationError, got %#v", err)
			}
		})
	}
}

// This SDK checks the signature before it checks freshness, so a request that
// fails both reports the signature. The five SDKs agree on this order.
func TestVerifyChecksSignatureBeforeFreshness(t *testing.T) {
	vector, body := loadVector(t)
	staleTimestamp := strconv.FormatInt(time.Now().Add(-6*time.Minute).Unix(), 10)

	_, err := Verify(body, vector.TamperedSignature, staleTimestamp, vector.Secret)
	if err == nil {
		t.Fatal("expected a rejection")
	}
	if !strings.Contains(err.Error(), "signature") {
		t.Errorf("expected the signature failure to be reported first, got %q", err)
	}
}

func TestWithMaxAge(t *testing.T) {
	vector, body := loadVector(t)
	timestamp := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	signature := sign(t, vector.Secret, timestamp, body)

	if _, err := Verify(body, signature, timestamp, vector.Secret); err == nil {
		t.Error("10 minutes old should exceed the 5-minute default")
	}
	if _, err := Verify(body, signature, timestamp, vector.Secret, WithMaxAge(15*time.Minute)); err != nil {
		t.Errorf("15 minutes of slack should accept it: %v", err)
	}
}

func TestHeaderNamesAreTheWireProtocol(t *testing.T) {
	// These are the literal header names the service sends. Renaming them
	// breaks every deployed receiver.
	want := map[string]string{
		HeaderID:        "X-Atlas-Webhook-Id",
		HeaderSignature: "X-Atlas-Webhook-Signature",
		HeaderTimestamp: "X-Atlas-Webhook-Timestamp",
		HeaderEvent:     "X-Atlas-Webhook-Event",
		HeaderAttempt:   "X-Atlas-Webhook-Attempt",
	}
	for got, expected := range want {
		if got != expected {
			t.Errorf("header constant = %q, want %q", got, expected)
		}
	}
}

func TestFixturesIntegrity(t *testing.T) {
	manifest, err := os.ReadFile(filepath.Join("testdata", "CHECKSUMS.txt"))
	if err != nil {
		t.Fatalf("read CHECKSUMS.txt: %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(manifest)), "\n") {
		want, name, ok := strings.Cut(line, "  ")
		if !ok {
			t.Fatalf("malformed manifest line %q", line)
		}
		sum := sha256.Sum256(mustFixture(t, name))
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Errorf("fixture %s changed: manifest has %s, file hashes to %s", name, want, got)
		}
	}
}
