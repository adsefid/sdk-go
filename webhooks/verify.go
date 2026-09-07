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
// respectively. secret is the webhook endpoint's signing secret.
//
// The signature is HMAC-SHA256 over the literal string "{timestamp}.{rawBody}",
// keyed with secret, with the raw digest bytes Base64-encoded directly (no
// intermediate hex-encoding step) and prefixed with "v1=". Verify strips that
// prefix and compares using a constant-time comparison.
//
// It returns a *adsefid.WebhookVerificationError if the signature is invalid,
// the timestamp is missing, malformed, or too old (5 minutes by default; see
// WithMaxAge), or the body cannot be parsed as a recognized event.
func Verify(rawBody []byte, signatureHeader, timestampHeader, secret string, opts ...VerifyOption) (WebhookEvent, error) {
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

	if !signatureValid(timestampSeconds, rawBody, secret, providedSignature) {
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
// (timestampSeconds, rawBody) under secret, and compares it against
// providedSignature in constant time.
func signatureValid(timestampSeconds int64, rawBody []byte, secret, providedSignature string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
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
