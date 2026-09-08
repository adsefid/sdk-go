package adsefid

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// SMSService groups the SMS endpoints of the adsefid.com API. Obtain one via
// Client.SMS rather than constructing it directly.
type SMSService struct {
	client *Client
}

// SendSingleSmsRequest is the request body for SMSService.SendSingle.
type SendSingleSmsRequest struct {
	Receptor     string        `json:"receptor"`
	LineNumber   string        `json:"line_number"`
	LineSelector *LineSelector `json:"line_selector,omitempty"`
	Message      string        `json:"message"`
	SendTime     *time.Time    `json:"send_time,omitempty"`
	LocalID      *string       `json:"local_id,omitempty"`
	Hide         *bool         `json:"hide,omitempty"`
}

// SendSingleSmsResponse is the response body for SMSService.SendSingle.
type SendSingleSmsResponse struct {
	GroupID      string                  `json:"group_id"`
	LocalID      *string                 `json:"local_id"`
	Status       WebServiceMessageStatus `json:"status"`
	LineNumber   string                  `json:"line_number"`
	LineSelector LineSelector            `json:"line_selector"`
	Cost         float64                 `json:"cost"`
	Receptor     string                  `json:"receptor"`
	SendTime     *time.Time              `json:"send_time"`
	MessageID    string                  `json:"message_id"`
	SegmentCount int                     `json:"segment_count"`
	Hide         bool                    `json:"hide"`
}

// BulkSmsReceptor is one entry of SendBulkSmsRequest.Receptors.
type BulkSmsReceptor struct {
	Receptor string  `json:"receptor"`
	LocalID  *string `json:"local_id,omitempty"`
	Hide     *bool   `json:"hide,omitempty"`
}

// SendBulkSmsRequest is the request body for SMSService.SendBulk.
type SendBulkSmsRequest struct {
	Receptors    []BulkSmsReceptor `json:"receptors"`
	Message      string            `json:"message"`
	SendTime     *time.Time        `json:"send_time,omitempty"`
	LineNumber   string            `json:"line_number"`
	LineSelector *LineSelector     `json:"line_selector,omitempty"`
}

// BulkSmsReceptorResult is one entry of SendBulkSmsResponse.Receptors.
//
// Status is a plain int, not a typed WebServiceMessageStatus, because a
// failed per-receptor result can carry a WebServiceResponseCode (2000+)
// instead of a WebServiceMessageStatus (1000-1999). Compare it against both
// enums' underlying int values as needed.
type BulkSmsReceptorResult struct {
	MessageID *string `json:"message_id"`
	Receptor  string  `json:"receptor"`
	LocalID   *string `json:"local_id"`
	Status    int     `json:"status"`
	Hide      bool    `json:"hide"`
	Cost      float64 `json:"cost"`
}

// SendBulkSmsResponse is the response body for SMSService.SendBulk.
type SendBulkSmsResponse struct {
	GroupID      string                  `json:"group_id"`
	Receptors    []BulkSmsReceptorResult `json:"receptors"`
	Message      string                  `json:"message"`
	SegmentCount int                     `json:"segment_count"`
	SendTime     *time.Time              `json:"send_time"`
	LineNumber   string                  `json:"line_number"`
	LineSelector LineSelector            `json:"line_selector"`
	Counts       map[string]int          `json:"counts"`
	TotalCount   int                     `json:"total_count"`
	TotalCost    float64                 `json:"total_cost"`
}

// P2PSmsMessage is one entry of SendP2PSmsRequest.Messages.
type P2PSmsMessage struct {
	Receptor string  `json:"receptor"`
	Message  string  `json:"message"`
	LocalID  *string `json:"local_id,omitempty"`
	Hide     *bool   `json:"hide,omitempty"`
}

// SendP2PSmsRequest is the request body for SMSService.SendP2P.
type SendP2PSmsRequest struct {
	Messages     []P2PSmsMessage `json:"messages"`
	SendTime     *time.Time      `json:"send_time,omitempty"`
	LineNumber   string          `json:"line_number"`
	LineSelector *LineSelector   `json:"line_selector,omitempty"`
}

// P2PSmsMessageResult is one entry of SendP2PSmsResponse.Messages.
//
// Status is a plain int; see BulkSmsReceptorResult.Status for why.
type P2PSmsMessageResult struct {
	MessageID    *string `json:"message_id"`
	Receptor     string  `json:"receptor"`
	Status       int     `json:"status"`
	LocalID      *string `json:"local_id"`
	Message      string  `json:"message"`
	Hide         bool    `json:"hide"`
	SegmentCount int     `json:"segment_count"`
	Cost         float64 `json:"cost"`
}

// SendP2PSmsResponse is the response body for SMSService.SendP2P.
type SendP2PSmsResponse struct {
	GroupID      string                `json:"group_id"`
	Messages     []P2PSmsMessageResult `json:"messages"`
	SendTime     *time.Time            `json:"send_time"`
	LineNumber   string                `json:"line_number"`
	LineSelector LineSelector          `json:"line_selector"`
	TotalCost    float64               `json:"total_cost"`
	Counts       map[string]int        `json:"counts"`
}

// SendTemplateSmsRequest is the request body for SMSService.SendTemplate.
type SendTemplateSmsRequest struct {
	TemplateID   string                            `json:"template_id"`
	Parameters   map[string]TemplateParameterValue `json:"parameters"`
	Receptor     string                            `json:"receptor"`
	LocalID      *string                           `json:"local_id,omitempty"`
	LineNumber   string                            `json:"line_number"`
	LineSelector *LineSelector                     `json:"line_selector,omitempty"`
	ExpiryDate   *time.Time                        `json:"expiry_date,omitempty"`
}

// SendTemplateSmsResponse is the response body for SMSService.SendTemplate.
type SendTemplateSmsResponse struct {
	GroupID      string                            `json:"group_id"`
	MessageID    string                            `json:"message_id"`
	Status       WebServiceMessageStatus           `json:"status"`
	LocalID      *string                           `json:"local_id"`
	LineNumber   string                            `json:"line_number"`
	TemplateID   string                            `json:"template_id"`
	SendTime     *time.Time                        `json:"send_time"`
	ExpiryDate   *time.Time                        `json:"expiry_date"`
	LineSelector LineSelector                      `json:"line_selector"`
	Cost         float64                           `json:"cost"`
	Receptor     string                            `json:"receptor"`
	Message      string                            `json:"message"`
	SegmentCount int                               `json:"segment_count"`
	Parameters   map[string]TemplateParameterValue `json:"parameters"`
}

// SmsStatusItem is one entry of GetSmsStatusResponse.Receptors.
type SmsStatusItem struct {
	MessageID    string                  `json:"message_id"`
	LocalID      *string                 `json:"local_id"`
	Status       WebServiceMessageStatus `json:"status"`
	Receptor     string                  `json:"receptor"`
	SendTime     *time.Time              `json:"send_time"`
	DeliveryTime *time.Time              `json:"delivery_time"`
}

// GetSmsStatusResponse is the response body for SMSService.GetStatus.
type GetSmsStatusResponse struct {
	Receptors []SmsStatusItem `json:"receptors"`
}

// CancelSmsRequest is the request body for SMSService.Cancel. At least one of
// MessageIDs or LocalIDs must be non-empty.
type CancelSmsRequest struct {
	MessageIDs []string `json:"message_ids,omitempty"`
	LocalIDs   []string `json:"local_ids,omitempty"`
}

// CancelledSmsItem is one entry of CancelSmsResponse.CancelledMessages or
// CancelSmsResponse.FailedToCancel.
type CancelledSmsItem struct {
	MessageID string                  `json:"message_id"`
	LocalID   *string                 `json:"local_id"`
	Status    WebServiceMessageStatus `json:"status"`
}

// CancelSmsResponse is the response body for SMSService.Cancel.
type CancelSmsResponse struct {
	CancelledMessages []CancelledSmsItem `json:"cancelled_messages"`
	FailedToCancel    []CancelledSmsItem `json:"failed_to_cancel"`
}

// ReceivedSmsMessage is one entry of GetReceivedSmsResponse.Messages.
type ReceivedSmsMessage struct {
	Message     string    `json:"message"`
	LineNumber  string    `json:"line_number"`
	ReceiveDate time.Time `json:"receive_date"`
	Sender      string    `json:"sender"`
}

// GetReceivedSmsResponse is the response body for SMSService.GetReceived.
type GetReceivedSmsResponse struct {
	Messages []ReceivedSmsMessage `json:"messages"`
}

// SendSingle sends a single SMS message to one receptor.
func (s *SMSService) SendSingle(ctx context.Context, req *SendSingleSmsRequest) (*SendSingleSmsResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Receptor, "Receptor"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.LineNumber, "LineNumber"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Message, "Message"); err != nil {
		return nil, err
	}
	if err := requireMaxLength(req.Message, maxSmsMessageLength, "Message"); err != nil {
		return nil, err
	}
	if err := validateLocalID(req.LocalID, "LocalID"); err != nil {
		return nil, err
	}

	return doPostJSON[*SendSingleSmsResponse](ctx, s.client, "/v1/sms/single", req)
}

// SendBulk sends the same message to many receptors in one call.
func (s *SMSService) SendBulk(ctx context.Context, req *SendBulkSmsRequest) (*SendBulkSmsResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmptySlice(req.Receptors, "Receptors"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Message, "Message"); err != nil {
		return nil, err
	}
	if err := requireMaxLength(req.Message, maxSmsMessageLength, "Message"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.LineNumber, "LineNumber"); err != nil {
		return nil, err
	}
	for i := range req.Receptors {
		if err := requireNonEmpty(req.Receptors[i].Receptor, "Receptors[].Receptor"); err != nil {
			return nil, err
		}
		if err := validateLocalID(req.Receptors[i].LocalID, "Receptors[].LocalID"); err != nil {
			return nil, err
		}
	}

	return doPostJSON[*SendBulkSmsResponse](ctx, s.client, "/v1/sms/bulk", req)
}

// SendP2P sends distinct, per-receptor messages in one call.
func (s *SMSService) SendP2P(ctx context.Context, req *SendP2PSmsRequest) (*SendP2PSmsResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmptySlice(req.Messages, "Messages"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.LineNumber, "LineNumber"); err != nil {
		return nil, err
	}
	for i := range req.Messages {
		if err := requireNonEmpty(req.Messages[i].Receptor, "Messages[].Receptor"); err != nil {
			return nil, err
		}
		if err := requireNonEmpty(req.Messages[i].Message, "Messages[].Message"); err != nil {
			return nil, err
		}
		if err := requireMaxLength(req.Messages[i].Message, maxSmsMessageLength, "Messages[].Message"); err != nil {
			return nil, err
		}
		if err := validateLocalID(req.Messages[i].LocalID, "Messages[].LocalID"); err != nil {
			return nil, err
		}
	}

	return doPostJSON[*SendP2PSmsResponse](ctx, s.client, "/v1/sms/p2p", req)
}

// SendTemplate sends a pre-approved message template, filled in with
// Parameters, to one receptor.
func (s *SMSService) SendTemplate(ctx context.Context, req *SendTemplateSmsRequest) (*SendTemplateSmsResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.TemplateID, "TemplateID"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Receptor, "Receptor"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.LineNumber, "LineNumber"); err != nil {
		return nil, err
	}
	if err := validateLocalID(req.LocalID, "LocalID"); err != nil {
		return nil, err
	}

	return doPostJSON[*SendTemplateSmsResponse](ctx, s.client, "/v1/sms/template", req)
}

// GetStatus looks up the delivery status of previously sent messages by
// message ID and/or local ID. At least one of messageIDs or localIDs must be
// non-empty, and their combined length must not exceed 2000.
func (s *SMSService) GetStatus(ctx context.Context, messageIDs, localIDs []string) (*GetSmsStatusResponse, error) {
	if err := requireAtLeastOne(len(messageIDs) > 0, len(localIDs) > 0, "at least one of messageIDs or localIDs is required"); err != nil {
		return nil, err
	}
	if err := requireCombinedCountAtMost(distinctCount(messageIDs), distinctCount(localIDs), maxCombinedIDsLookup, "the combined distinct count of messageIDs and localIDs must not exceed 2000"); err != nil {
		return nil, err
	}

	path := "/v1/sms/status" + buildIDsQuery(messageIDs, localIDs)
	return doGet[*GetSmsStatusResponse](ctx, s.client, path)
}

// Cancel cancels previously scheduled messages by message ID and/or local ID.
// At least one of req.MessageIDs or req.LocalIDs must be non-empty.
func (s *SMSService) Cancel(ctx context.Context, req *CancelSmsRequest) (*CancelSmsResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireAtLeastOne(len(req.MessageIDs) > 0, len(req.LocalIDs) > 0, "at least one of MessageIDs or LocalIDs must be non-empty"); err != nil {
		return nil, err
	}

	return doPostJSON[*CancelSmsResponse](ctx, s.client, "/v1/sms/cancel", req)
}

// GetReceived fetches inbound SMS messages received on lineNumber. count (if
// given) must be in [1, 499]; since (if given) filters to messages received
// at or after that time.
func (s *SMSService) GetReceived(ctx context.Context, lineNumber string, count *int, since *time.Time) (*GetReceivedSmsResponse, error) {
	if err := requireNonEmpty(lineNumber, "lineNumber"); err != nil {
		return nil, err
	}
	if count != nil {
		if err := requireInRange(*count, minReceivedCount, maxReceivedCount, "count"); err != nil {
			return nil, err
		}
	}

	values := url.Values{}
	values.Set("line_number", lineNumber)
	if count != nil {
		values.Set("count", strconv.Itoa(*count))
	}
	if since != nil {
		values.Set("since", since.Format(time.RFC3339Nano))
	}

	path := "/v1/sms/receive?" + values.Encode()
	return doGet[*GetReceivedSmsResponse](ctx, s.client, path)
}
