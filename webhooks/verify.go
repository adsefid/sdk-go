package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	adsefid "github.com/adsefid/sdk-go"
)

const (
	signaturePrefix = "v1="
	defaultMaxAge   = 5 * time.Minute
)

// VerifyOption configures a call to Verify.
type VerifyOption func(*verifyConfig)

type verifyConfig struct {
	maxAge time.Duration
}

// WithMaxAge overrides the default maximum age (5 minutes) a webhook
// timestamp may have before Verify rejects it as stale.
func WithMaxAge(maxAge time.Duration) VerifyOption {
	return func(c *verifyConfig) {
		c.maxAge = maxAge
	}
}

// rawEnvelope is the wire shape of every webhook delivery body.
type rawEnvelope struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Attempt    int             `json:"attempt"`
	Version    string          `json:"version"`
	Data       json.RawMessage `json:"data"`
}

// Verify authenticates and parses one inbound webhook delivery.
//
// rawBody must be the exact, unmodified request body bytes (verifying against
// a body that has been re-serialized, pretty-printed, or otherwise mutated
// will fail). signatureHeader and timestampHeader are the values of the
// X-Atlas-Webhook-Signature and X-Atlas-Webhook-Timestamp request headers,
// respectively.
//
// secret is the webhook endpoint's signing secret exactly as shown in your
// adsefid.com panel: the Base64 encoding of 32 random bytes. Verify decodes it
// to those raw bytes and uses them as the HMAC key, which is what the server
// signs with. A secret that is not valid Base64 is rejected. If you hold the
// key as raw bytes already, use VerifyWithKey instead.
//
// The signature is HMAC-SHA256 over the literal string "{timestamp}.{rawBody}",
// with the raw digest bytes Base64-encoded directly (no intermediate
// hex-encoding step) and prefixed with "v1=". Verify strips that prefix and
// compares using a constant-time comparison.
//
// It returns a *adsefid.WebhookVerificationError if the secret is not valid
// Base64, the signature is invalid, the timestamp is missing, malformed, or too
// old (5 minutes by default; see WithMaxAge), or the body cannot be parsed as a
// recognized event.
func Verify(rawBody []byte, signatureHeader, timestampHeader, secret string, opts ...VerifyOption) (WebhookEvent, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(secret))
	if err != nil {
		return nil, &adsefid.WebhookVerificationError{Message: "webhook secret is not valid Base64; use the secret exactly as shown in your adsefid.com panel"}
	}
	return VerifyWithKey(rawBody, signatureHeader, timestampHeader, key, opts...)
}

// VerifyWithKey is Verify with the signing key supplied as raw bytes rather
// than as the Base64 string from the adsefid.com panel. Use it when you store
// the decoded key yourself, for example in a secret manager.
func VerifyWithKey(rawBody []byte, signatureHeader, timestampHeader string, key []byte, opts ...VerifyOption) (WebhookEvent, error) {
	cfg := &verifyConfig{maxAge: defaultMaxAge}
	for _, opt := range opts {
		opt(cfg)
	}

	if !strings.HasPrefix(signatureHeader, signaturePrefix) {
		return nil, &adsefid.WebhookVerificationError{Message: "signature header is missing the 'v1=' prefix"}
	}
	providedSignature := signatureHeader[len(signaturePrefix):]

	timestampSeconds, err := strconv.ParseInt(strings.TrimSpace(timestampHeader), 10, 64)
	if err != nil {
		return nil, &adsefid.WebhookVerificationError{Message: "timestamp header is not a valid unix timestamp"}
	}

	if !signatureValid(timestampSeconds, rawBody, key, providedSignature) {
		return nil, &adsefid.WebhookVerificationError{Message: "signature mismatch"}
	}

	age := time.Since(time.Unix(timestampSeconds, 0))
	if age < 0 {
		age = -age
	}
	if age > cfg.maxAge {
		return nil, &adsefid.WebhookVerificationError{Message: "webhook timestamp is older than the allowed max age"}
	}

	var envelope rawEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return nil, &adsefid.WebhookVerificationError{Message: "webhook body could not be parsed as JSON"}
	}

	meta := eventMeta{
		id:         envelope.ID,
		eventType:  EventType(envelope.Type),
		occurredAt: envelope.OccurredAt,
		attempt:    envelope.Attempt,
		version:    envelope.Version,
	}

	switch meta.eventType {
	case EventTypeReceive:
		var data []ReceivedMessageItem
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return nil, &adsefid.WebhookVerificationError{Message: "webhook payload 'data' field could not be parsed for a receive event"}
		}
		return &ReceiveEvent{eventMeta: meta, Data: data}, nil

	case EventTypeStatus:
		var data []StatusUpdateItem
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return nil, &adsefid.WebhookVerificationError{Message: "webhook payload 'data' field could not be parsed for a status event"}
		}
		return &StatusEvent{eventMeta: meta, Data: data}, nil

	case EventTypeMessengerStatus:
		var data []StatusUpdateItem
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return nil, &adsefid.WebhookVerificationError{Message: "webhook payload 'data' field could not be parsed for a messenger.status event"}
		}
		return &MessengerStatusEvent{eventMeta: meta, Data: data}, nil

	default:
		return nil, &adsefid.WebhookVerificationError{Message: fmt.Sprintf("unsupported webhook event type %q", envelope.Type)}
	}
}

// signatureValid recomputes the expected HMAC-SHA256 signature for
// (timestampSeconds, rawBody) under key, and compares it against
// providedSignature in constant time.
func signatureValid(timestampSeconds int64, rawBody []byte, key []byte, providedSignature string) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(strconv.FormatInt(timestampSeconds, 10)))
	mac.Write([]byte("."))
	mac.Write(rawBody)
	expectedDigest := mac.Sum(nil)

	providedDigest, err := base64.StdEncoding.DecodeString(providedSignature)
	if err != nil {
		return false
	}

	return hmac.Equal(providedDigest, expectedDigest)
}
