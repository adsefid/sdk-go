// Package webhooks verifies and parses inbound webhook deliveries from the
// adsefid.com SMS Web Service. It has no dependency on adsefid.Client — a
// webhook receiver is typically a separate process (an HTTP server) from
// whatever sends messages.
package webhooks

import (
	"time"

	adsefid "github.com/adsefid/sdk-go"
)

// EventType identifies the kind of a webhook event, mirroring the API's
// "type" field.
type EventType string

const (
	// EventTypeReceive is fired when an inbound SMS message arrives on one
	// of the account's lines.
	EventTypeReceive EventType = "receive"
	// EventTypeStatus is fired when an SMS message's delivery status
	// changes.
	EventTypeStatus EventType = "status"
	// EventTypeMessengerStatus is fired when a Messenger message's delivery
	// status changes.
	EventTypeMessengerStatus EventType = "messenger.status"
)

// WebhookEvent is implemented by every event type this package can parse:
// *ReceiveEvent, *StatusEvent, and *MessengerStatusEvent. Use a type switch
// on the value returned by Verify to access event-specific data.
type WebhookEvent interface {
	// ID is the unique identifier of this webhook delivery.
	ID() string
	// Type is the event's kind.
	Type() EventType
	// OccurredAt is when the underlying event happened, as reported by the
	// API (not when this delivery attempt was made).
	OccurredAt() time.Time
	// Attempt is the 1-based delivery attempt number for this webhook call.
	Attempt() int
	// Version is the webhook payload schema version.
	Version() string

	// webhookEvent is an unexported marker method that seals this interface
	// to the types defined in this package.
	webhookEvent()
}

// eventMeta holds the fields common to every webhook event and implements
// the metadata methods of WebhookEvent. It is embedded in each concrete
// event type.
type eventMeta struct {
	id         string
	eventType  EventType
	occurredAt time.Time
	attempt    int
	version    string
}

func (m eventMeta) ID() string            { return m.id }
func (m eventMeta) Type() EventType       { return m.eventType }
func (m eventMeta) OccurredAt() time.Time { return m.occurredAt }
func (m eventMeta) Attempt() int          { return m.attempt }
func (m eventMeta) Version() string       { return m.version }
func (m eventMeta) webhookEvent()         {}

// ReceivedMessageItem is one inbound message carried by a ReceiveEvent.
type ReceivedMessageItem struct {
	ID          int64     `json:"id"`
	LineNumber  string    `json:"line_number"`
	Sender      string    `json:"sender"`
	Message     string    `json:"message"`
	ReceiveDate time.Time `json:"receive_date"`
}

// ReceiveEvent is fired when one or more inbound SMS messages arrive.
type ReceiveEvent struct {
	eventMeta
	Data []ReceivedMessageItem
}

// StatusUpdateItem is one message's status change, carried by a StatusEvent
// or MessengerStatusEvent.
type StatusUpdateItem struct {
	ID             string                          `json:"id"`
	LocalID        *string                         `json:"local_id"`
	StatusDelivery adsefid.WebServiceMessageStatus `json:"status_delivery"`
	DeliveryTime   *time.Time                      `json:"delivery_time"`
}

// StatusEvent is fired when one or more SMS messages' delivery status
// changes.
type StatusEvent struct {
	eventMeta
	Data []StatusUpdateItem
}

// MessengerStatusEvent is fired when one or more Messenger messages'
// delivery status changes.
type MessengerStatusEvent struct {
	eventMeta
	Data []StatusUpdateItem
}
