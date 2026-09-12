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
// instead of a WebServiceMessageStatus (1000-1999). MessageStatus and
// ErrorCode split it into the typed enum for its range.
type BulkSmsReceptorResult struct {
	MessageID *string `json:"message_id"`
	Receptor  string  `json:"receptor"`
	LocalID   *string `json:"local_id"`
	Status    int     `json:"status"`
	Hide      bool    `json:"hide"`
	Cost      float64 `json:"cost"`
}

// MessageStatus returns the typed message status when Status is in
// 1000-1999 and names a status this SDK knows; see MessageStatusOf.
func (r BulkSmsReceptorResult) MessageStatus() (WebServiceMessageStatus, bool) {
	return MessageStatusOf(r.Status)
}

// ErrorCode returns the typed error code when Status is 2000 or above and
// names an error this SDK knows, meaning this one item was rejected; see
// ErrorCodeOf.
func (r BulkSmsReceptorResult) ErrorCode() (WebServiceResponseCode, bool) {
	return ErrorCodeOf(r.Status)
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

// MessageStatus returns the typed message status when Status is in
// 1000-1999 and names a status this SDK knows; see MessageStatusOf.
func (r P2PSmsMessageResult) MessageStatus() (WebServiceMessageStatus, bool) {
	return MessageStatusOf(r.Status)
}

// ErrorCode returns the typed error code when Status is 2000 or above and
// names an error this SDK knows, meaning this one item was rejected; see
// ErrorCodeOf.
func (r P2PSmsMessageResult) ErrorCode() (WebServiceResponseCode, bool) { return ErrorCodeOf(r.Status) }

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

// GetSmsStatusRequest is the query for SMSService.GetStatus. At least one of
// MessageIDs or LocalIDs must be non-empty, and their combined distinct count
// must not exceed 2000.
type GetSmsStatusRequest struct {
	MessageIDs []string
	LocalIDs   []string
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

// GetReceivedSmsRequest is the query for SMSService.GetReceived. LineNumber is
// required; Count (if set) must be in [1, 499]; Since (if set) filters to
// messages received at or after that time.
type GetReceivedSmsRequest struct {
	LineNumber string
	Count      *int
	Since      *time.Time
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

// SendBulk sends the same message to many receptors in one call. Item errors
// are returned in the partial API response.
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
	return doPostJSON[*SendBulkSmsResponse](ctx, s.client, "/v1/sms/bulk", req)
}

// SendP2P sends distinct, per-receptor messages in one call. Item errors are
// returned in the partial API response.
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
// message ID and/or local ID.
func (s *SMSService) GetStatus(ctx context.Context, req *GetSmsStatusRequest) (*GetSmsStatusResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := validateIDsLookup(req.MessageIDs, req.LocalIDs); err != nil {
		return nil, err
	}

	path := "/v1/sms/status" + buildIDsQuery(req.MessageIDs, req.LocalIDs)
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

// GetReceived fetches inbound SMS messages received on req.LineNumber.
func (s *SMSService) GetReceived(ctx context.Context, req *GetReceivedSmsRequest) (*GetReceivedSmsResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.LineNumber, "LineNumber"); err != nil {
		return nil, err
	}
	if req.Count != nil {
		if err := requireInRange(*req.Count, minReceivedCount, maxReceivedCount, "Count"); err != nil {
			return nil, err
		}
	}

	values := url.Values{}
	values.Set("line_number", req.LineNumber)
	if req.Count != nil {
		values.Set("count", strconv.Itoa(*req.Count))
	}
	if req.Since != nil {
		values.Set("since", req.Since.Format(time.RFC3339Nano))
	}

	path := "/v1/sms/receive?" + values.Encode()
	return doGet[*GetReceivedSmsResponse](ctx, s.client, path)
}
