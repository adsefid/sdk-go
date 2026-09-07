// Command webhookserver runs a minimal HTTP server that receives, verifies,
// and dispatches adsefid.com webhook deliveries.
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
	if secret == "" {
		log.Fatal("ADSEFID_WEBHOOK_SECRET environment variable is not set")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhooks/adsefid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// The Id, Event, and Attempt headers are also sent on every
		// delivery and are useful for logging or idempotency keys; Verify
		// itself only needs the Signature and Timestamp headers, since the
		// event id/type/attempt are also present (and covered by the
		// signature) in the body.
		deliveryID := r.Header.Get(webhooks.HeaderID)
		signature := r.Header.Get(webhooks.HeaderSignature)
		timestamp := r.Header.Get(webhooks.HeaderTimestamp)

		event, err := webhooks.Verify(rawBody, signature, timestamp, secret)
		if err != nil {
			log.Printf("webhook verification failed (delivery %s): %v", deliveryID, err)
			http.Error(w, "invalid webhook", http.StatusUnauthorized)
			return
		}

		// Only event types this webhook endpoint is subscribed to (in your adsefid.com panel)
		// ever arrive here — an endpoint subscribed to just "receive" never sees a *StatusEvent.
		switch e := event.(type) {
		case *webhooks.ReceiveEvent:
			for _, item := range e.Data {
				log.Printf("received SMS on %s from %s: %q", item.LineNumber, item.Sender, item.Message)
			}
		case *webhooks.StatusEvent:
			for _, item := range e.Data {
				log.Printf("SMS %s status update: %s", item.ID, item.StatusDelivery)
			}
		case *webhooks.MessengerStatusEvent:
			for _, item := range e.Data {
				log.Printf("Messenger %s status update: %s", item.ID, item.StatusDelivery)
			}
		default:
			log.Printf("unhandled webhook event type: %s", event.Type())
		}

		w.WriteHeader(http.StatusOK)
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
