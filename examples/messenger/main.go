// Command messenger demonstrates the Messenger resource end to end: uploading
// an attachment, sending it, and checking delivery status.
//
// Messenger sends go through a "profile" configured in your adsefid.com panel
// rather than an SMS line, and allow a longer message body (4000 characters
// against SMS's 900). File upload is the only multipart endpoint in the API.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	adsefid "github.com/adsefid/sdk-go"
)

func main() {
	apiKey := os.Getenv("ADSEFID_API_KEY")
	if apiKey == "" {
		log.Fatal("set ADSEFID_API_KEY")
	}

	client, err := adsefid.NewClient(apiKey, adsefid.WithTimeout(60*time.Second))
	if err != nil {
		log.Fatalf("failed to construct client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	profile := firstProfile(ctx, client)
	fileID := uploadAttachment(ctx, client)

	sent, err := client.Messenger.SendSingle(ctx, &adsefid.SendSingleMessengerRequest{
		Message:  "Your statement is attached.",
		Receptor: "09120000000",
		Profile:  profile,
		FileID:   &fileID,
		LocalID:  strPtr("statement-2026-09"),
	})
	if err != nil {
		log.Fatalf("send: %v", err)
	}
	fmt.Printf("\nsent %s via %s: %s (cost %g)\n", sent.MessageID, sent.Messenger, sent.Status, sent.Cost)

	status, err := client.Messenger.GetStatus(ctx, []string{sent.MessageID}, nil)
	if err != nil {
		log.Fatalf("get status: %v", err)
	}
	for _, item := range status.Receptors {
		fmt.Printf("  %s -> %s\n", item.MessageID, item.Status)
	}
}

func firstProfile(ctx context.Context, client *adsefid.Client) string {
	profiles, err := client.User.GetProfiles(ctx)
	if err != nil {
		log.Fatalf("list profiles: %v", err)
	}
	if len(profiles) == 0 {
		log.Fatal("no messenger profiles on this account — add one in the adsefid.com panel first")
	}
	fmt.Printf("%d messenger profile(s):\n", len(profiles))
	for _, profile := range profiles {
		fmt.Printf("  %-40s %-20s %s\n", profile.ID, profile.Name, profile.Messenger)
	}
	return profiles[0].ID
}

// uploadAttachment sends any io.Reader. Here it is an in-memory string; in a
// real program it is usually an *os.File, which is also an io.Reader — the SDK
// streams it rather than buffering the whole file.
func uploadAttachment(ctx context.Context, client *adsefid.Client) string {
	content := strings.NewReader("Statement for September 2026\nTotal: 1,250,000 IRR\n")

	uploaded, err := client.Messenger.UploadFile(ctx, content, "statement.txt", "text/plain")
	if err != nil {
		log.Fatalf("upload: %v", err)
	}
	fmt.Printf("\nuploaded attachment as file_id %s\n", uploaded.FileID)
	return uploaded.FileID
}

func strPtr(s string) *string { return &s }
