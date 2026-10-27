# Changelog

## Unreleased

### Breaking
- Monetary response fields (`Cost`, `TotalCost`, `CreditLeft`) are now `json.Number` instead of
  `float64`, preserving the exact decimal text. Use `.String()` or `.Float64()`.
- `UserTemplate.Parameters` now keeps unrecognized parameter types instead of
  dropping them.

### Fixed
- Status lookups send ids joined by literal commas, and a GET whose path and query pass 8,000
  characters returns a `*ValidationError` instead of reaching the gateway's HTML 414.
- `webhooks.Verify` and `VerifyWithKey` reject an empty signing key. Previously an empty secret
  accepted a signature made with an empty key.
- A template number parameter that is NaN or infinite returns a `*ValidationError` instead of an
  encoding failure.

### Added
- `MaxMessengerFileBytes`; `UploadFile` rejects content above 15 MiB before sending the request.
- Response bodies are read through a 16 MiB cap.
- The default HTTP client keeps up to 100 idle connections per host.
- `make lint` runs a pinned golangci-lint with no separate install.
