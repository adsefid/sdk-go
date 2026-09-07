// Package adsefid is a client SDK for the adsefid.com SMS Web Service REST API.
//
// It provides SMS, Messenger, and User resources through a single [Client], and a
// sibling [github.com/adsefid/sdk-go/webhooks] package for verifying and parsing
// inbound webhook deliveries.
//
// # Getting started
//
//	client, err := adsefid.NewClient(os.Getenv("ADSEFID_API_KEY"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := client.SMS.SendSingle(ctx, &adsefid.SendSingleSmsRequest{
//		Receptor:   "09120000000",
//		LineNumber: "3000...",
//		Message:    "hello",
//	})
//
// # Errors
//
// Every method returns a plain Go error. Client-side pre-flight failures (bad
// input, never touching the network) come back as a [*ValidationError]. Network
// or timeout problems come back as a [*TransportError]. A non-success API
// response comes back as an [*APIError], or as a [*RateLimitError] (which wraps
// an *APIError) for rate-limit-shaped failures. Use errors.As to distinguish
// them; see the package README for a full worked example.
//
// # Deliberate deviations from the sibling SDKs
//
// Equivalent SDKs exist for this same API in other languages (.NET, JS, PHP,
// Python). This package intentionally departs from their approach in a few
// places where the idiomatic Go choice differs:
//
//   - JSON is handled via the standard library's reflection-based
//     encoding/json, with `json:"snake_case"` struct tags. The other SDKs
//     favor a "zero reflection" approach for AOT-trimming reasons that are
//     specific to their runtimes (chiefly .NET Native AOT); that concern
//     doesn't apply to Go, where encoding/json is the idiomatic default and
//     hand-rolling (un)marshaling for every type would be unusual and costly
//     for no real benefit.
//   - Identifiers (group_id, message_id, profile, file_id, and webhook item
//     ids) are plain strings, not a dedicated UUID type. The standard library
//     has no UUID type, and importing one (e.g. google/uuid) would violate
//     this module's zero-dependency policy.
//   - Enums are permissively-typed by construction: an unrecognized integer
//     or string value from the wire simply becomes that named Go type's
//     value (e.g. a WebServiceResponseCode of 9999), rather than failing to
//     parse. Go's named-integer/string type system makes this automatic,
//     unlike the explicit "parse permissively, keep the raw value alongside"
//     workarounds needed in more strictly-enumerated languages.
//
// # Known doc-vs-reality subtleties
//
// See the doc comments on [TemplateParameterType] and on the bulk/P2P
// per-item result types (their Status fields are a plain int, not a typed
// enum) for two places where this SDK intentionally departs from a literal
// reading of the published API documentation to match observed live-API
// behavior. Both are also called out in AGENTS.md.
package adsefid
