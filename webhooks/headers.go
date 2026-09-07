package webhooks

// Header names carried by every adsefid.com outgoing webhook delivery.
// http.Header.Get canonicalizes header names, so these work regardless of
// wire casing.
const (
	HeaderID        = "X-Atlas-Webhook-Id"
	HeaderSignature = "X-Atlas-Webhook-Signature"
	HeaderTimestamp = "X-Atlas-Webhook-Timestamp"
	HeaderEvent     = "X-Atlas-Webhook-Event"
	HeaderAttempt   = "X-Atlas-Webhook-Attempt"
)
