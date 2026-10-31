# adsefid.com SMS Web Service: Go SDK

[![CI](https://github.com/adsefid/sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/adsefid/sdk-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/adsefid/sdk-go.svg)](https://pkg.go.dev/github.com/adsefid/sdk-go)

A Go client for the [adsefid.com](https://adsefid.com) SMS Web Service API. It sends SMS and
Messenger messages (single, bulk, P2P and template), checks delivery status, cancels unsent
messages, reads inbound SMS, looks up account data, and verifies signed webhooks.

The SDK uses only the standard library and needs Go 1.22 or later. It is verified against API
documentation **v1.15.0**. Sibling SDKs exist for .NET, JavaScript, PHP and Python with the same
shape.

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

	selector := adsefid.LineSelectorBulkServiceSendBased
	resp, err := client.SMS.SendSingle(context.Background(), &adsefid.SendSingleSmsRequest{
		Receptor:   "989120000000",
		LineNumber: "983000",
		LineSelector: &selector,
		Message:    "Hello from adsefid.com!",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("sent %s, status=%s, cost=%s\n", resp.MessageID, resp.Status, resp.Cost)
}
```

The SDK never reads environment variables. Pass the API key yourself. It goes out as the
`X-API-KEY` header on every request.

**Line selector.** An account can own several lines with the same number, each with its own
selector (see `client.User.GetLines`). When it does, `LineSelector` is required on every SMS send;
without it the service answers `WebServiceResponseCodeInvalidLineSelector`.

## Client options

```go
client, err := adsefid.NewClient(
	apiKey,
	adsefid.WithBaseURL("https://api.adsefid.com"), // default
	adsefid.WithTimeout(15*time.Second),            // ignored if WithHTTPClient is set
	adsefid.WithUserAgent("my-service/1.0.0"),      // default: adsefid-go/<SDK_VERSION>
	adsefid.WithHTTPClient(myHTTPClient),           // custom transport, proxy, logging
)
```

`NewClient` returns a `*ValidationError` for a blank key, a base URL that is not an absolute
http(s) URL, or a bad user agent. The default `*http.Client` keeps up to 100 idle connections per
host, so concurrent callers reuse TLS connections.

Every method takes a `context.Context` first. Use it for cancellation and per-call deadlines.

## Methods

| Service | Method | Endpoint |
|---|---|---|
| `client.SMS` | `SendSingle`, `SendBulk`, `SendP2P`, `SendTemplate` | `POST /v1/sms/{single,bulk,p2p,template}` |
| `client.SMS` | `GetStatus`, `Cancel`, `GetReceived` | `QUERY /v1/sms/status`, `POST /v1/sms/cancel`, `GET /v1/sms/receive` |
| `client.Messenger` | `SendSingle`, `SendBulk`, `SendP2P`, `SendTemplate` | `POST /v1/messenger/{single,bulk,p2p,template}` |
| `client.Messenger` | `UploadFile(ctx, io.Reader, fileName, contentType)` | `POST /v1/messenger/file` |
| `client.Messenger` | `GetStatus`, `Cancel` | `QUERY /v1/messenger/status`, `POST /v1/messenger/cancel` |
| `client.User` | `GetInfo`, `GetLines`, `GetProfiles`, `GetTemplates` | `GET /v1/user/{info,lines,profiles,templates}` |

Every method except `GetInfo`, `GetLines`, `GetProfiles` and `UploadFile` takes one
`*XxxRequest` struct, including the GET lookups. `GetTemplates(ctx, nil)` lists the first page in
every state. Cancel accepts messages that are awaiting approval, queued or scheduled, with no
refund, and at most 2,000 distinct IDs across both lists. Status lookups use `QUERY` with JSON arrays, supporting up to 2,000 distinct IDs in one request.
Unmatched cancellation IDs appear in `failed_to_cancel` with error codes 2046 or 2029; a missing
local ID match has a null message ID. Check each item's status or error code.

## Bulk and P2P: partial success is not an error

A bulk or P2P call returns HTTP 200 even when some items are rejected. Each item's `Status` holds
either a message status (1000 to 1999) or an error code (2000 and up). `MessageStatus()` and
`ErrorCode()` split it into the typed enum for its range:

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

Item values reach the API unchanged, so valid items still go out. If every item is rejected, the
call fails with an `*APIError` whose `Details.Items` lists one entry per rejected index. Only
request-level fields are checked locally.

## Template parameters

A parameter the template declares as `number` can be sent as a JSON number or a JSON string.
The service substitutes a numeric string as written, so a string is the only way to keep exact
digits:

```go
params := map[string]adsefid.TemplateParameterValue{
	"name":    adsefid.StringParam("Ali"),
	"invoice": adsefid.StringParam("001234"), // renders 001234, not 1234
	"amount":  adsefid.StringParam("1.50"),   // renders 1.50, not 1.5
	"count":   adsefid.IntParam(2),
	"rate":    adsefid.NumberParam(19.99), // a float64, so rounding is possible
}
```

`RawNumber()` returns a number's exact wire text (`"1.50"` keeps its trailing zero);
`NumberValue()` converts to `float64`. `NumberParam` rejects NaN and infinities with a
`*ValidationError`.

Money fields (`Cost`, `TotalCost`, `CreditLeft`) are `json.Number`, so you keep the exact decimal
the service sent. Call `.String()`, or `.Float64()` when rounding is fine.

## File upload

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

File content is limited to 15 MiB inclusive (`adsefid.MaxMessengerFileBytes`). A larger reader
fails with a `*ValidationError` before any request is sent. The service also enforces a MIME
allowlist and answers `WebServiceResponseCodeFileTooLarge` (2047) for anything it still rejects
as too large. The SDK reads your reader but does not close it.

## Errors

Every failure is a returned `error`. The SDK never panics for an expected failure. Tell the types
apart with `errors.As`, checking `*RateLimitError` before the `*APIError` it embeds:

| Type | Meaning |
|---|---|
| `*ValidationError` | A local check failed. No request was sent. |
| `*APIError` | The API returned an error envelope, or a non-2xx status without one. Has `Code`, `Name`, `HTTPStatusCode` and optional typed `Details`. |
| `*RateLimitError` | Embeds `*APIError`. Codes 2035 and 2036, or any HTTP 429. `adsefid.IsRateLimitError(err)` is a shortcut. |
| `*TransportError` | Connection failure, timeout, or a response body that could not be decoded. Wraps the cause. |

```go
resp, err := client.SMS.SendSingle(ctx, req)
if err != nil {
	var rl *adsefid.RateLimitError
	var api *adsefid.APIError
	switch {
	case errors.As(err, &rl):
		log.Printf("rate limited (%s)", rl.Name)
	case errors.As(err, &api):
		log.Printf("API error %s (code %d, HTTP %d)", api.Name, int(api.Code), api.HTTPStatusCode)
	default:
		log.Printf("failed: %v", err)
	}
	return
}
_ = resp
```

A non-2xx response with no error envelope is an `*APIError` with code `-1` and name
`UNKNOWN_ERROR`. A bare 429 is code 2036 with name `RATE_LIMITED`. The SDK sends each request
once and never retries. Add backoff in your own `http.RoundTripper` if you need it.

`APIError.Details.Errors` maps field names to `{Code, Name}`.
`Details.Items` carries indexed bulk/P2P errors.

## Enums and unknown values

`LineSelector`, `WebServiceMessageStatus`, `WebServiceResponseCode`, `TemplateState` and
`TemplateParameterType` are typed constants that implement `fmt.Stringer`. A value the service
adds later is kept as that type's raw value (for example `WebServiceResponseCode(9999)`) instead
of failing to parse. See `enums.go` for the full lists.

## Webhooks

The `webhooks` package verifies the signature, checks the timestamp, and parses the payload into
a typed event. It does not depend on `*adsefid.Client`, so a receiver can be a separate process.

The signing secret appears in the panel Base64-encoded, and the service signs with the decoded
bytes. `webhooks.Verify` takes the secret exactly as the panel shows it. If you store the decoded
key, call `webhooks.VerifyWithKey`. An empty secret or key is always rejected.

```go
http.HandleFunc("/webhooks/adsefid", func(w http.ResponseWriter, r *http.Request) {
	rawBody, err := io.ReadAll(r.Body) // verify the exact bytes, before any JSON parsing
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	event, err := webhooks.Verify(
		rawBody,
		r.Header.Get(webhooks.HeaderSignature),
		r.Header.Get(webhooks.HeaderTimestamp),
		os.Getenv("ADSEFID_WEBHOOK_SECRET"),
	)
	if err != nil {
		http.Error(w, "invalid webhook", http.StatusUnauthorized) // do not say why
		return
	}

	switch e := event.(type) {
	case *webhooks.ReceiveEvent:
		for _, m := range e.Data {
			log.Printf("SMS on %s from %s: %q", m.LineNumber, m.Sender, m.Message)
		}
	case *webhooks.StatusEvent:
		for _, s := range e.Data {
			log.Printf("SMS %s -> %s", s.ID, s.StatusDelivery)
		}
	case *webhooks.MessengerStatusEvent:
		for _, s := range e.Data {
			log.Printf("Messenger %s -> %s", s.ID, s.StatusDelivery)
		}
	}
	w.WriteHeader(http.StatusOK)
})
```

Header and event-type constants (`webhooks.HeaderSignature`, `webhooks.EventTypeReceive`, and so
on) hold the exact wire names. `webhooks.WithMaxAge(d)` changes the default five-minute freshness
window. Each endpoint only receives the event types you subscribed it to in the panel. Deliveries
can repeat, so make your handler idempotent and return a 2xx quickly.

## Using this SDK with coding agents

[AGENTS.md](AGENTS.md) holds the repository rules. The points below prevent most wrong guesses:

- Pass the API key and webhook secret explicitly. The SDK reads no environment variables.
- Treat a bulk or P2P response as data, not an error. Check each item's status.
- Send exact numeric text (`"001234"`, `"1.50"`) as `StringParam`.
- Hand `webhooks.Verify` the raw request body, never re-serialized JSON.
- There are no retries and no automatic batching. Split large requests yourself.
- Check errors with `errors.As`, not string matching. Branch on `Code`, which is stable.

## Examples and contributing

Runnable programs are in [`examples/`](examples): `account`, `quickstart`, `bulkandp2p`,
`templates`, `statusandcancel`, `messenger` and `webhookserver`. `account` sends nothing, so try it
first.

```sh
export ADSEFID_API_KEY=...
export ADSEFID_LINE_NUMBER=983000
go run ./examples/quickstart
```

To work on the SDK: `make lint`, `make test` and `make build`. `make lint` runs a pinned
golangci-lint, so nothing needs installing. See [AGENTS.md](AGENTS.md).

## Versioning and license

Semantic Versioning, independent of the API documentation. Current SDK version: **1.1.0**
(tag `v1.1.0`). See [CHANGELOG.md](CHANGELOG.md). A future v2 changes the module path to
`github.com/adsefid/sdk-go/v2`. MIT licensed, see [LICENSE](LICENSE).
