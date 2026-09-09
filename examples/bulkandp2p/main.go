// Command bulkandp2p demonstrates the two multi-receptor SMS endpoints and,
// more importantly, how to read a partial success.
//
// Bulk and P2P sends answer HTTP 200 even when some receptors failed. The
// per-item Status is the raw WebServiceCode; MessageStatus() and ErrorCode()
// split it into the typed enum for its range. Always inspect the items; a nil
// error does not mean every message went out.
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
// report prints one item. ErrorCode() is set for a rejected item and
// MessageStatus() for an accepted one; neither is set for a code this SDK
// does not know yet, so the raw Status is printed in that case.
func report(receptor string, localID *string, status int, messageID *string) {
	label := "-"
	if localID != nil {
		label = *localID
	}

	if code, ok := adsefid.ErrorCodeOf(status); ok {
		fmt.Printf("  %-12s (%s) FAILED with code %d (%s)\n", receptor, label, status, code)
		return
	}

	id := "-"
	if messageID != nil {
		id = *messageID
	}
	if messageStatus, ok := adsefid.MessageStatusOf(status); ok {
		fmt.Printf("  %-12s (%s) accepted as %s: %s\n", receptor, label, id, messageStatus)
		return
	}
	fmt.Printf("  %-12s (%s) reported an unknown status code %d\n", receptor, label, status)
}

func strPtr(s string) *string { return &s }
