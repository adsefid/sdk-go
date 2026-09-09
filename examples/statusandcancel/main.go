// Command statusandcancel demonstrates the SMS lookup endpoints: checking
// delivery status by ID, cancelling scheduled messages, and reading inbound
// messages.
//
// Status and cancel both accept message IDs (ours) and local IDs (yours) in
// the same call. Their combined distinct count may not exceed 2000; the SDK
// checks that before making the request.
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

	// Schedule a message far enough ahead that there is something to cancel.
	sendAt := time.Now().Add(2 * time.Hour)
	sent, err := client.SMS.SendSingle(ctx, &adsefid.SendSingleSmsRequest{
		Receptor:   "09120000000",
		LineNumber: lineNumber,
		Message:    "This one is scheduled, and about to be cancelled.",
		SendTime:   &sendAt,
		LocalID:    strPtr("demo-cancel-1"),
	})
	if err != nil {
		log.Fatalf("send: %v", err)
	}
	fmt.Printf("scheduled %s for %s\n", sent.MessageID, sendAt.Format(time.RFC3339))

	// Look it up by our ID and by your own local_id in one call.
	status, err := client.SMS.GetStatus(ctx, &adsefid.GetSmsStatusRequest{MessageIDs: []string{sent.MessageID}, LocalIDs: []string{"demo-cancel-1"}})
	if err != nil {
		log.Fatalf("get status: %v", err)
	}
	fmt.Printf("\nstatus for %d message(s):\n", len(status.Receptors))
	for _, item := range status.Receptors {
		delivered := "not yet"
		if item.DeliveryTime != nil {
			delivered = item.DeliveryTime.Format(time.RFC3339)
		}
		fmt.Printf("  %s -> %s (delivered: %s)\n", item.MessageID, item.Status, delivered)
	}

	// Cancelling reports each message separately: a message already sent
	// cannot be recalled and comes back under FailedToCancel.
	cancelled, err := client.SMS.Cancel(ctx, &adsefid.CancelSmsRequest{
		MessageIDs: []string{sent.MessageID},
	})
	if err != nil {
		log.Fatalf("cancel: %v", err)
	}
	fmt.Printf("\ncancelled %d, failed to cancel %d\n",
		len(cancelled.CancelledMessages), len(cancelled.FailedToCancel))
	for _, item := range cancelled.FailedToCancel {
		fmt.Printf("  %s could not be cancelled: %s\n", item.MessageID, item.Status)
	}

	// Inbound messages. count is capped at 499; since filters by arrival time.
	since := time.Now().Add(-24 * time.Hour)
	count := 50
	received, err := client.SMS.GetReceived(ctx, &adsefid.GetReceivedSmsRequest{LineNumber: lineNumber, Count: &count, Since: &since})
	if err != nil {
		log.Fatalf("get received: %v", err)
	}
	fmt.Printf("\n%d inbound message(s) in the last 24h:\n", len(received.Messages))
	for _, item := range received.Messages {
		fmt.Printf("  %s from %s at %s: %s\n",
			item.LineNumber, item.Sender, item.ReceiveDate.Format(time.RFC3339), item.Message)
	}
}

func strPtr(s string) *string { return &s }
