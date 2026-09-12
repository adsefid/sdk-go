# adsefid.com SMS Web Service — Go SDK

[![CI](https://github.com/adsefid/sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/adsefid/sdk-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/adsefid/sdk-go.svg)](https://pkg.go.dev/github.com/adsefid/sdk-go)

A Go client SDK for the [adsefid.com](https://adsefid.com) SMS Web Service REST API: send single,
bulk, P2P, and templated SMS and Messenger messages; check delivery status; cancel scheduled
messages; fetch inbound SMS; look up account info, lines, profiles, and templates; and verify
inbound webhook deliveries.

Equivalent SDKs for the same API exist for .NET, JavaScript, PHP, and Python. This SDK's public
surface is structurally parallel to those, translated idiomatically to Go — see
[AGENTS.md](AGENTS.md) for the specific, deliberate deviations.

## Requirements

Go 1.22 or later. Zero third-party dependencies — this module only uses the Go standard library.

## Install

```sh
go get github.com/adsefid/sdk-go
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	adsefid "github.com/adsefid/sdk-go"
)

func main() {
	client, err := adsefid.NewClient(os.Getenv("ADSEFID_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.SMS.SendSingle(context.Background(), &adsefid.SendSingleSmsRequest{
		Receptor:   "09120000000",
		LineNumber: "3000000000",
		Message:    "Hello from adsefid.com!",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("sent message %s, status=%s\n", resp.MessageID, resp.Status)
}
```

See [`examples/quickstart`](examples/quickstart) for a fuller version with error triage, and
[`examples/webhookserver`](examples/webhookserver) for a working webhook receiver.

## Authentication

Every request is authenticated with a single `X-API-KEY` header, set from the API key you pass to
`NewClient`. Keep it out of source control — the examples read it from an environment variable:

```sh
export ADSEFID_API_KEY="your-api-key"
```

```go
client, err := adsefid.NewClient(os.Getenv("ADSEFID_API_KEY"))
```

`NewClient` returns a `*adsefid.ValidationError` if the key is blank.

## Configuration

`NewClient` takes functional options:

```go
client, err := adsefid.NewClient(
	apiKey,
	adsefid.WithBaseURL("https://api.adsefid.com"), // default; override for testing
	adsefid.WithTimeout(15 * time.Second),          // only used when WithHTTPClient is not set
	adsefid.WithUserAgent("my-service/1.0.0"),      // defaults to adsefid-go/<SDK_VERSION>
)
```

- **`WithBaseURL(string)`** — overrides the default API base URL (`https://api.adsefid.com`).
- **`WithHTTPClient(*http.Client)`** — supplies your own `*http.Client`, e.g. to install a custom
  `http.RoundTripper` for logging or a corporate proxy. Takes precedence over `WithTimeout` if
  both are set.
- **`WithTimeout(time.Duration)`** — sets the timeout of the default `*http.Client` this package
  builds internally. Has no effect if `WithHTTPClient` is also supplied.
- **`WithUserAgent(string)`** — replaces the default `adsefid-go/<SDK_VERSION>` User-Agent value.

Monetary response fields (`Cost`, `TotalCost`, and `CreditLeft`) use `float64` and may contain fractional values.

## Resource reference

| Resource | SDK method | HTTP endpoint | Notes |
|---|---|---|---|
| `client.SMS` | `SendSingle(ctx, *SendSingleSmsRequest)` | `POST /v1/sms/single` | |
| `client.SMS` | `SendBulk(ctx, *SendBulkSmsRequest)` | `POST /v1/sms/bulk` | Partial success is a normal typed return |
| `client.SMS` | `SendP2P(ctx, *SendP2PSmsRequest)` | `POST /v1/sms/p2p` | Partial success is a normal typed return |
| `client.SMS` | `SendTemplate(ctx, *SendTemplateSmsRequest)` | `POST /v1/sms/template` | |
| `client.SMS` | `GetStatus(ctx, *GetSmsStatusRequest)` | `GET /v1/sms/status` | |
| `client.SMS` | `Cancel(ctx, *CancelSmsRequest)` | `POST /v1/sms/cancel` | |
| `client.SMS` | `GetReceived(ctx, *GetReceivedSmsRequest)` | `GET /v1/sms/receive` | |
| `client.Messenger` | `SendSingle(ctx, *SendSingleMessengerRequest)` | `POST /v1/messenger/single` | |
| `client.Messenger` | `SendBulk(ctx, *SendBulkMessengerRequest)` | `POST /v1/messenger/bulk` | Partial success is a normal typed return |
| `client.Messenger` | `SendP2P(ctx, *SendP2PMessengerRequest)` | `POST /v1/messenger/p2p` | Partial success is a normal typed return |
| `client.Messenger` | `UploadFile(ctx, io.Reader, fileName, contentType string)` | `POST /v1/messenger/file` | Caller owns the reader |
| `client.Messenger` | `Cancel(ctx, *CancelMessengerRequest)` | `POST /v1/messenger/cancel` | |
| `client.Messenger` | `SendTemplate(ctx, *SendTemplateMessengerRequest)` | `POST /v1/messenger/template` | |
| `client.Messenger` | `GetStatus(ctx, *GetMessengerStatusRequest)` | `GET /v1/messenger/status` | |
| `client.User` | `GetInfo(ctx)` | `GET /v1/user/info` | |
| `client.User` | `GetLines(ctx)` | `GET /v1/user/lines` | |
| `client.User` | `GetProfiles(ctx)` | `GET /v1/user/profiles` | |
| `client.User` | `GetTemplates(ctx, *GetUserTemplatesRequest)` | `GET /v1/user/templates` | `nil` lists the first page in every state |

Every method takes a `context.Context` as its first argument and returns `(T, error)` or
`(*T, error)` — there are no panics for expected failure conditions.
Bulk/P2P item values are sent unchanged so the API can accept or reject them independently; only
request-level fields are prevalidated.

## Error handling

This SDK never panics for expected failures. It defines four error types, all implementing
`error`, and you distinguish them with `errors.As`:

- **`*adsefid.ValidationError`** — a client-side pre-flight check failed (bad input shape, a
  message too long, a malformed `local_id`, etc.). No network call was made.
- **`*adsefid.APIError`** — the API returned a non-success envelope, or a non-2xx HTTP status.
  Carries `Code adsefid.WebServiceResponseCode`, `Name string`, `HTTPStatusCode int`, and optional
  typed `Details *adsefid.APIErrorDetails`.
- **`*adsefid.RateLimitError`** — embeds `*adsefid.APIError` for the rate-limit-shaped failures
  (`WebServiceResponseCode` 2035 `MessageLimitReached`, 2036 `RequestLimitReached`, or a bare HTTP
  429). It unwraps to the embedded `*APIError`, so `errors.As` also matches that.
  `adsefid.IsRateLimitError(err)` is a convenience wrapper around the same check.
- **`*adsefid.TransportError`** — a network-level failure: connection error, timeout, or a
  response body that could not be decoded. Wraps the underlying error (`Unwrap()`).

A worked example, checking the more specific `*RateLimitError` before the `*APIError` it embeds:

```go
resp, err := client.SMS.SendSingle(ctx, req)
if err != nil {
	var valErr *adsefid.ValidationError
	if errors.As(err, &valErr) {
		log.Fatalf("bad request: %s", valErr)
	}

	var rlErr *adsefid.RateLimitError
	if errors.As(err, &rlErr) {
		log.Fatalf("rate limited (%s)", rlErr.Name)
	}

	var apiErr *adsefid.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("API error %s (code %d, HTTP %d)", apiErr.Name, int(apiErr.Code), apiErr.HTTPStatusCode)
	}

	var transportErr *adsefid.TransportError
	if errors.As(err, &transportErr) {
		log.Fatalf("transport error: %v", transportErr)
	}

	log.Fatalf("unexpected error: %v", err)
}
```

`APIError.Details` contains optional `Errors map[string]APIFieldError` for field errors (and rejected
cancel IDs), plus optional `Items []APIItemError` for indexed bulk/P2P errors. Named integer codes
preserve unknown future values.

### Rate limits

Codes `2035`, `2036`, and bare HTTP `429` responses surface as `*adsefid.RateLimitError`.

## Enums

All documented enums are typed Go constants implementing `fmt.Stringer`. Unrecognized wire values
degrade gracefully (e.g. `WebServiceResponseCode(9999)`) rather than failing to parse:

- **`LineSelector`** (int, 0-5): `LineSelectorPromotionalSendBased`, `LineSelectorPromotionalDeliverBased`,
  `LineSelectorBulkServiceSendBased`, `LineSelectorBulkServiceDeliverBased`,
  `LineSelectorCustomerClubServiceSendBased`, `LineSelectorCustomerClubServiceDeliverBased`.
- **`WebServiceMessageStatus`** (int, 1000-1999): the status of a single message — e.g.
  `WebServiceMessageStatusDelivered`, `WebServiceMessageStatusBlacklisted`,
  `WebServiceMessageStatusUnknown`. See `enums.go` for the full set of 18 values.
- **`WebServiceResponseCode`** (int, 2000-2047): the API's error codes — e.g.
  `WebServiceResponseCodeInvalidAPIKey`, `WebServiceResponseCodeNotEnoughCredit`,
  `WebServiceResponseCodeMessageLimitReached`, `WebServiceResponseCodeInvalidMessageIDs`, and
  `WebServiceResponseCodeFileTooLarge`. See `enums.go` for the full set of 48 values.
- **`TemplateState`** (string): `TemplateStatePendingApproval`, `TemplateStateApproved`,
  `TemplateStateRejected`.
- **`TemplateParameterType`** (string): `TemplateParameterTypeString`, `TemplateParameterTypeNumber`.
  The documented public set is exactly these two; the live API has been observed to also emit an
  undocumented third value for some templates, which this SDK intentionally does not model (see
  AGENTS.md).

Bulk and P2P per-item results (`BulkSmsReceptorResult.Status`, `P2PSmsMessageResult.Status`,
`BulkMessengerReceptorResult.Status`, `P2PMessengerReceptorResult.Status`) are a plain `int`, not a
typed enum, because a failed item's `status` can carry either a `WebServiceMessageStatus` (1000+)
or a `WebServiceResponseCode` (2000+). Each item's `MessageStatus()` and `ErrorCode()` methods split
that value into the typed enum for its range (`ok` is false for the other range and for a code this
SDK does not know yet), and `adsefid.MessageStatusOf`/`adsefid.ErrorCodeOf` do the same for any raw
value:

```go
for _, item := range resp.Receptors {
	if code, ok := item.ErrorCode(); ok {
		fmt.Printf("%s: rejected (%s)\n", item.Receptor, code)
		continue
	}
	status, _ := item.MessageStatus()
	fmt.Printf("%s: %s\n", item.Receptor, status)
}
```

Top-level statuses (single send, get-status, cancel) are always in the 1000-1999 range and use the
typed `WebServiceMessageStatus`.

`TemplateParameterValue` models a template parameter's value, which the API accepts and returns as
either a JSON string or a JSON number:

```go
params := map[string]adsefid.TemplateParameterValue{
	"name":  adsefid.StringParam("Ali"),
	"count": adsefid.IntParam(3),
	"price": adsefid.NumberParam(19.99),
}

if s, ok := params["name"].StringValue(); ok {
	fmt.Println(s)
}
```

### Numbers, leading zeros and decimals

A parameter the template declares as `number` may be sent **either** as a JSON number or as a JSON
string. The service substitutes a numeric string verbatim, so a string is the only way to keep a
value's exact digits:

```go
params := map[string]adsefid.TemplateParameterValue{
	"invoice": adsefid.StringParam("001234"), // renders as 001234, not 1234
	"amount":  adsefid.StringParam("1.50"),   // renders as 1.50, not 1.5
	"count":   adsefid.IntParam(2),           // an ordinary integer
	"rate":    adsefid.NumberParam(19.99),    // a decimal, where float rounding is acceptable
}
```

Use `StringParam` whenever the rendered text must match the digits you supplied — invoice numbers,
account numbers, zero-padded codes, and money amounts with a fixed number of decimal places.
`NumberParam` takes a `float64` and therefore cannot represent every decimal exactly.

On the way back, `RawNumber()` returns a number as its exact wire text, while `NumberValue()`
converts to `float64`:

```go
value := resp.Parameters["amount"]
if raw, ok := value.RawNumber(); ok {
	fmt.Println(raw) // "1.50" — the trailing zero survives
}
```

See [`examples/templates`](examples/templates) for a runnable version.

## File upload example

```go
f, err := os.Open("banner.png")
if err != nil {
	log.Fatal(err)
}
defer f.Close()

uploaded, err := client.Messenger.UploadFile(ctx, f, "banner.png", "image/png")
if err != nil {
	log.Fatal(err)
}

_, err = client.Messenger.SendSingle(ctx, &adsefid.SendSingleMessengerRequest{
	Receptor: "09120000000",
	Profile:  profileID,
	Message:  "Check out our new banner!",
	FileID:   &uploaded.FileID,
})
```

The service enforces its documented MIME allowlist and 15 MB limit. Oversized uploads return
`WebServiceResponseCodeFileTooLarge` (2047, HTTP 413).

## Webhooks

The `webhooks` subpackage verifies and parses inbound webhook deliveries. It has no dependency on
`*adsefid.Client` — a webhook receiver is typically a separate process from whatever sends
messages.

Every webhook delivery carries the following headers. Use the exported constants so the wire names
remain exact:

- `webhooks.HeaderID` (`X-Atlas-Webhook-Id`) — the delivery's unique ID.
- `webhooks.HeaderSignature` (`X-Atlas-Webhook-Signature`) — `"v1=" + base64(hmac_sha256("{timestamp}.{raw_body}", secret))`.
- `webhooks.HeaderTimestamp` (`X-Atlas-Webhook-Timestamp`) — the Unix timestamp (seconds) the delivery was signed at.
- `webhooks.HeaderEvent` (`X-Atlas-Webhook-Event`) — the event type (also present in the body as `"type"`).
- `webhooks.HeaderAttempt` (`X-Atlas-Webhook-Attempt`) — the 1-based delivery attempt number (also present in the body).

`webhooks.Verify` needs only the raw body plus the Signature and Timestamp headers — the event id,
type, and attempt are also present in (and covered by the signature of) the JSON body itself.

### The signing secret is Base64

Your endpoint's signing secret is shown in the adsefid.com panel as the Base64 encoding of 32
random bytes, and the service signs with **those raw bytes** — not with the text of the Base64
string. `webhooks.Verify` takes the secret exactly as the panel shows it and decodes it for you; a
secret that is not valid Base64 is rejected with a `*adsefid.WebhookVerificationError`.

If you already hold the decoded key, use `webhooks.VerifyWithKey(rawBody, signature, timestamp,
key, opts...)` instead and skip the decoding step.

You configure, per webhook endpoint, which event types it receives (in your adsefid.com panel) —
an endpoint subscribed only to `webhooks.EventTypeReceive` will never see a `*StatusEvent` arrive,
so don't assume every deployment gets all three; handle whichever ones you've subscribed to.

```go
package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/adsefid/sdk-go/webhooks"
)

func main() {
	secret := os.Getenv("ADSEFID_WEBHOOK_SECRET")

	http.HandleFunc("/webhooks/adsefid", func(w http.ResponseWriter, r *http.Request) {
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		signature := r.Header.Get(webhooks.HeaderSignature)
		timestamp := r.Header.Get(webhooks.HeaderTimestamp)

		event, err := webhooks.Verify(rawBody, signature, timestamp, secret)
		if err != nil {
			http.Error(w, "invalid webhook", http.StatusUnauthorized)
			return
		}

		switch e := event.(type) {
		case *webhooks.ReceiveEvent:
			for _, item := range e.Data {
				log.Printf("received SMS on %s from %s: %q", item.LineNumber, item.Sender, item.Message)
			}
		case *webhooks.StatusEvent:
			for _, item := range e.Data {
				log.Printf("SMS %s -> %s", item.ID, item.StatusDelivery)
			}
		case *webhooks.MessengerStatusEvent:
			for _, item := range e.Data {
				log.Printf("Messenger %s -> %s", item.ID, item.StatusDelivery)
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

See [`examples/webhookserver`](examples/webhookserver) for the complete, runnable version.

`webhooks.Verify` accepts a `webhooks.WithMaxAge(time.Duration)` option to override the default
5-minute staleness window, and returns a `*adsefid.WebhookVerificationError` for any verification
or parsing failure (bad signature, stale timestamp, malformed or unrecognized payload).

## Development

```sh
make deps    # go mod download
make fmt     # gofmt + goimports
make lint    # go vet + golangci-lint
make build   # go build ./...
make test    # go test ./...
```

`gofmt`, `go vet` and `go test` ship with the Go toolchain. `golangci-lint` needs a one-time
install:

```sh
brew install golangci-lint
```

The test suite uses only the standard library (`testing` plus `net/http/httptest`), so `go.mod`
stays dependency-free. Golden fixtures live in `testdata/` and are byte-identical to the same tree
in the sibling SDK repositories; `TestFixturesIntegrity` verifies them against `CHECKSUMS.txt`.

### Examples

Each directory under `examples/` is a standalone `package main`. Run one with `go run`:

```sh
export ADSEFID_API_KEY=...
export ADSEFID_LINE_NUMBER=983000XXX

go run ./examples/account          # account info, lines, profiles, templates; client configuration
go run ./examples/quickstart       # send one SMS, with full error triage
go run ./examples/bulkandp2p       # bulk + P2P sends, and reading a partial success
go run ./examples/templates        # list templates and send one, incl. exact numeric values
go run ./examples/statusandcancel  # delivery status, cancelling, inbound messages
go run ./examples/messenger        # upload an attachment and send it via a messenger profile
go run ./examples/webhookserver    # verify and dispatch inbound webhooks
```

`examples/account` sends nothing, so it is the safest one to try first.

## Versioning

This SDK follows Semantic Versioning independently of the API documentation.

- SDK version: **`0.5.0`** (repository tag `v0.5.0`)
- Verified API documentation: **`v1.13.0`**

SDK releases use `v<SDK_VERSION>` tags. The two version numbers move independently. A future major
version `v2` must also change the module path to `github.com/adsefid/sdk-go/v2`.

## License

MIT — see [LICENSE](LICENSE).
