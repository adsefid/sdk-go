// Command account reads the account-level endpoints and shows how to configure
// the client beyond the defaults.
//
// Nothing here sends a message, so it is the safest example to run first
// against a real API key.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	adsefid "github.com/adsefid/sdk-go"
)

func main() {
	apiKey := os.Getenv("ADSEFID_API_KEY")
	if apiKey == "" {
		log.Fatal("set ADSEFID_API_KEY")
	}

	// The SDK never retries a request. If you want retries, connection
	// pooling, tracing or a proxy, configure your own *http.Client and hand it
	// over — it takes precedence over WithTimeout.
	httpClient := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	client, err := adsefid.NewClient(
		apiKey,
		adsefid.WithHTTPClient(httpClient),
		// WithBaseURL is normally only needed to point at a mock server.
		adsefid.WithBaseURL(baseURLOrDefault()),
		// Identify your own application in the User-Agent; the SDK's default
		// is "adsefid-go/<version>".
		adsefid.WithUserAgent("my-billing-service/1.4 (+https://example.com)"),
	)
	if err != nil {
		log.Fatalf("failed to construct client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.User.GetInfo(ctx)
	if err != nil {
		log.Fatalf("get info: %v", err)
	}
	fmt.Printf("account %s (%s)\n  credit left: %g\n", info.Name, info.AccountStatus, info.CreditLeft)
	if info.Email != nil {
		fmt.Printf("  email: %s\n", *info.Email)
	}

	lines, err := client.User.GetLines(ctx)
	if err != nil {
		log.Fatalf("get lines: %v", err)
	}
	fmt.Printf("\n%d SMS line(s):\n", len(lines))
	for _, line := range lines {
		state := "disabled"
		if line.Enabled {
			state = "enabled"
		}
		fmt.Printf("  %-14s %-24s %-8s default selector: %s\n",
			line.LineNumber, line.LineName, state, line.LineSelector)
	}

	profiles, err := client.User.GetProfiles(ctx)
	if err != nil {
		log.Fatalf("get profiles: %v", err)
	}
	fmt.Printf("\n%d messenger profile(s):\n", len(profiles))
	for _, profile := range profiles {
		fmt.Printf("  %-40s %-20s %s\n", profile.ID, profile.Name, profile.Messenger)
	}

	// Templates are paged: take is capped at 100.
	page, err := client.User.GetTemplates(ctx, nil, intPtr(0), intPtr(100))
	if err != nil {
		log.Fatalf("get templates: %v", err)
	}
	fmt.Printf("\n%d template(s) (showing %d):\n", page.Total, len(page.Items))
	for _, item := range page.Items {
		fmt.Printf("  %-24s %-16s %d parameter(s)\n", item.TemplateID, item.State, len(item.Parameters))
	}
}

func baseURLOrDefault() string {
	if custom := os.Getenv("ADSEFID_BASE_URL"); custom != "" {
		return custom
	}
	return adsefid.DefaultBaseURL
}

func intPtr(i int) *int { return &i }
