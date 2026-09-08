## Scope

This repository is the Go client SDK for the adsefid.com SMS Web Service API (module:
`github.com/adsefid/sdk-go`), independently versioned and published with its own `go.mod`.
Equivalent SDKs exist for the same API in sibling repositories (`sdk-dotnet`, `sdk-js`, `sdk-php`,
`sdk-python`); a behavior change here should generally be considered for parity there.

## Source of truth

The API surface (endpoints, field names, types, validation rules, enums, example payloads,
webhook behavior) is defined by the published adsefid.com SMS Web Service API documentation. This
SDK is verified against doc version v1.11.0. Re-read the relevant documentation before changing
any endpoint, request/response model, or enum. The SDK follows independent Semantic Versioning
from repository tags; never copy the API-document version into a tag. Record both versions in the
README. A future v2 must add `/v2` to the module and import paths.

A small number of facts below are empirically observed behaviors of the live API that are easy to
get wrong from a literal reading of the documentation's prose or pseudo-code. Trust these notes
over an ambiguous doc reading:

- Webhook signatures are plain Base64, not hex-then-Base64. The signature is HMAC-SHA256 over the
  literal string "{timestamp}.{raw_body}", and the raw digest bytes are Base64-encoded directly —
  there is no intermediate hex-encoding step, even though a literal reading of some spec
  pseudo-code can suggest one. See `webhooks/verify.go`.
- `TemplateParameterType` has an undocumented third value in the wild. The documented, supported
  public set is {string, number}. The live API has been observed to also emit a value for some
  templates that isn't documented; this SDK intentionally models only the two documented values —
  do not add support for it without first confirming it against current, documented API behavior.
  See `enums.go`.
- `error.details` shape varies per endpoint and is intentionally untyped (`json.RawMessage`). It
  may be a validation map, a bulk/P2P per-item list, a cancel-specific map, or absent entirely —
  never give it a strong type; decode it defensively per endpoint if you need it.

## Deliberate deviations from sibling SDKs (documented, not accidental)

- **JSON via stdlib `encoding/json` (reflection-based), not a zero-reflection approach.** The
  other SDKs' "no reflection" preference targets .NET/AOT-trimming concerns specifically; forcing
  a zero-reflection JSON approach in Go would mean hand-rolling every struct's (un)marshaling or
  pulling in a codegen dependency that ~no real-world Go SDK uses. Exported struct fields +
  `json:"snake_case"` tags is the idiomatic, minimal-dependency Go choice.
- **IDs are plain `string`, not a dedicated UUID type.** Go's stdlib has no UUID type, and adding
  a third-party one would violate this module's zero-dependency policy.
- **Enums are permissively-typed by construction.** An unrecognized wire integer just becomes that
  typed `WebServiceResponseCode`/`WebServiceMessageStatus` value — Go's named-integer-type system
  makes this automatic, unlike the explicit "parse permissively" workarounds the other SDKs need.

## Architecture map

- `client.go` — `Client`, functional-options construction, holds `SMS`/`Messenger`/`User` service handles.
- `request.go` — single-attempt HTTP transport, envelope decoding, error mapping. No retry logic anywhere, by design.
- `validate.go` — client-side pre-flight checks (local_id format, length/count limits) that fail fast before any network call.
- `errors.go` — the full error type hierarchy.
- `enums.go` — every documented enum as typed Go constants, plus `TemplateParameterValue`. That
  type stores a number as a `json.Number` so exact wire text survives a round trip; a `number`
  template parameter may legitimately travel as a JSON *string*, which is how leading zeros
  (`"001234"`) and exact decimals (`"1.50"`) reach the service intact. Do not "simplify" it back to
  a `float64`.
- `sms.go` / `messenger.go` / `user.go` — one `*Service` type per resource area.
- `webhooks/` — signature verification and typed webhook event payloads. The endpoint secret is
  the Base64 encoding of 32 random bytes and the service signs with the **decoded bytes**, so
  `Verify` Base64-decodes before keying the HMAC; `VerifyWithKey` takes raw key bytes. Keying the
  HMAC with the UTF-8 bytes of the Base64 string does not verify against the live service.
  This package imports the root package but is never imported by it. `headers.go` holds the `Header*` constants and `events.go` the `EventType*` constants — use instead of typing header/type strings.

## Hard rules

- Every change ships with tests. Suites live beside the code they cover, in `package adsefid`
  (and `package webhooks`), stdlib `testing` only — no testify, no anything: `go.mod` stays
  dependency-free. Prefer one table-driven test over many near-identical functions. A new endpoint
  needs rows in the request-building, response-parsing and validation tables at minimum.
- `testdata/` holds golden fixtures that are **byte-identical** to the same tree in the sibling SDK
  repositories, so all five agree on the wire. Never edit a fixture in isolation: change it in all
  five and regenerate every `CHECKSUMS.txt`, or `TestFixturesIntegrity` fails.
- No retry logic anywhere in this SDK — every request is a single attempt.
- No third-party dependencies — stdlib only. Think hard before adding one; the answer is almost always no.
- No magic literals — every documented enum value is a named Go constant, never a bare int/string at a call site.
- Return errors, never panic, for any expected failure condition.
- Keep this SDK's public surface structurally parallel to the sibling SDKs' — translated idiomatically to Go, not mechanically transliterated.

## Development

Run `make lint`/`make fmt`/`make build`/`make test` before finishing any change. Prerequisites:
`gofmt`/`go vet` ship with the Go toolchain; `golangci-lint` needs a one-time
`brew install golangci-lint` (see README's Development section).

Note: this repository's `.golangci.yml` is written in golangci-lint's v2 configuration schema
(`version: "2"`), since that is the version available in this environment. It enables the same
linter set as originally specified: errcheck, govet, staticcheck, unused, ineffassign, gofmt,
goimports, revive.
