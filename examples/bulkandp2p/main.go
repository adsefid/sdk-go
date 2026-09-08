// Command bulkandp2p demonstrates the two multi-receptor SMS endpoints and,
// more importantly, how to read a partial success.
//
// Bulk and P2P sends answer HTTP 200 even when some receptors failed. The
// per-item Status is a plain int because a failed item carries a
// WebServiceResponseCode (2000+) where a successful one carries a
// WebServiceMessageStatus (1000-1999). Always inspect the items; a nil error
// does not mean every message went out.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	adsefid "github.com/adsefid/sdk-go"
)

func main() {
	apiKey := os.Getenv("ADSEFID_API_KEY")
	lineNumber := os.Getenv("ADSEFID_LINE_NUMBER")
	if apiKey == "" || lineNumber == "" {
		log.Fatal("set ADSEFID_API_KEY and ADSEFID_LINE_NUMBER")
	}

	client, err := adsefid.NewClient(apiKey, adsefid.WithTimeout(30*time.Second))
	if err != nil {
		log.Fatalf("failed to construct client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	sendBulk(ctx, client, lineNumber)
	sendP2P(ctx, client, lineNumber)
}

// sendBulk sends one identical message to many receptors.
func sendBulk(ctx context.Context, client *adsefid.Client, lineNumber string) {
	resp, err := client.SMS.SendBulk(ctx, &adsefid.SendBulkSmsRequest{
		LineNumber: lineNumber,
		Message:    "Scheduled maintenance tonight from 01:00 to 03:00.",
		Receptors: []adsefid.BulkSmsReceptor{
			// local_id is your own idempotency handle: it comes back on the
			// response and on the status webhook, so you can match a delivery
			// report to your own record without storing our message IDs.
			{Receptor: "09120000000", LocalID: strPtr("maint-1")},
			{Receptor: "09120000001", LocalID: strPtr("maint-2")},
		},
	})
	if err != nil {
		log.Fatalf("bulk send failed outright: %v", err)
	}

	fmt.Printf("\nbulk group %s: %d receptors, total cost %g\n", resp.GroupID, resp.TotalCount, resp.TotalCost)
	for _, item := range resp.Receptors {
		report(item.Receptor, item.LocalID, item.Status, item.MessageID)
	}
	fmt.Printf("  status histogram: %v\n", resp.Counts)
}

// sendP2P sends a different message to each receptor in one request.
func sendP2P(ctx context.Context, client *adsefid.Client, lineNumber string) {
	resp, err := client.SMS.SendP2P(ctx, &adsefid.SendP2PSmsRequest{
		LineNumber: lineNumber,
		Messages: []adsefid.P2PSmsMessage{
			{Receptor: "09120000000", Message: "Hi Ali, your order #1001 shipped.", LocalID: strPtr("ship-1001")},
			{Receptor: "09120000001", Message: "Hi Reza, your order #1002 shipped.", LocalID: strPtr("ship-1002")},
		},
	})
	if err != nil {
		log.Fatalf("p2p send failed outright: %v", err)
	}

	fmt.Printf("\np2p group %s: total cost %g\n", resp.GroupID, resp.TotalCost)
	for _, item := range resp.Messages {
		report(item.Receptor, item.LocalID, item.Status, item.MessageID)
	}
}

// report prints one per-item result. A status below 2000 is a message status;
// 2000 and above is an error code explaining why that one receptor failed.
func report(receptor string, localID *string, status int, messageID *string) {
	label := "-"
	if localID != nil {
		label = *localID
	}

	if status >= 2000 {
		fmt.Printf("  %-12s (%s) FAILED with code %d (%s)\n",
			receptor, label, status, adsefid.WebServiceResponseCode(status))
		return
	}

	id := "-"
	if messageID != nil {
		id = *messageID
	}
	fmt.Printf("  %-12s (%s) accepted as %s: %s\n",
		receptor, label, id, adsefid.WebServiceMessageStatus(status))
}

func strPtr(s string) *string { return &s }
