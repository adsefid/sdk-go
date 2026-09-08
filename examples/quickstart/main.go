// Command quickstart demonstrates constructing an adsefid.Client, sending a
// single SMS message, and triaging the possible error types.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	adsefid "github.com/adsefid/sdk-go"
)

func main() {
	apiKey := os.Getenv("ADSEFID_API_KEY")
	if apiKey == "" {
		log.Fatal("ADSEFID_API_KEY environment variable is not set")
	}

	client, err := adsefid.NewClient(
		apiKey,
		adsefid.WithTimeout(15*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to construct client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.SMS.SendSingle(ctx, &adsefid.SendSingleSmsRequest{
		Receptor:   "09120000000",
		LineNumber: "3000000000",
		Message:    "Hello from the adsefid.com Go SDK!",
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf(
		"sent message %s to %s: status=%s, cost=%g, segments=%d\n",
		resp.MessageID, resp.Receptor, resp.Status, resp.Cost, resp.SegmentCount,
	)
}

// handleError shows the full errors.As triage recommended in the README: a
// *RateLimitError is checked before the more general *APIError it embeds,
// since errors.As matches the first assignable type it finds and a
// RateLimitError would otherwise also satisfy the APIError check.
func handleError(err error) {
	var valErr *adsefid.ValidationError
	if errors.As(err, &valErr) {
		log.Fatalf("request failed local validation: %s", valErr.Error())
	}

	var rlErr *adsefid.RateLimitError
	if errors.As(err, &rlErr) {
		log.Fatalf("rate limited (%s, HTTP %d) — this SDK never retries automatically; back off and try again later", rlErr.Name, rlErr.HTTPStatusCode)
	}

	var apiErr *adsefid.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("API returned an error: %s (code=%d, HTTP status=%d)", apiErr.Name, int(apiErr.Code), apiErr.HTTPStatusCode)
	}

	var transportErr *adsefid.TransportError
	if errors.As(err, &transportErr) {
		log.Fatalf("transport error talking to the adsefid.com API: %v", transportErr)
	}

	log.Fatalf("unexpected error: %v", err)
}
