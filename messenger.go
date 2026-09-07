package adsefid

import (
	"context"
	"io"
	"time"
)

// MessengerService groups the Messenger endpoints of the adsefid.com API.
// Obtain one via Client.Messenger rather than constructing it directly.
type MessengerService struct {
	client *Client
}

// SendSingleMessengerRequest is the request body for MessengerService.SendSingle.
type SendSingleMessengerRequest struct {
	Message  string     `json:"message"`
	Receptor string     `json:"receptor"`
	Profile  string     `json:"profile"`
	Hide     *bool      `json:"hide,omitempty"`
	FileID   *string    `json:"file_id,omitempty"`
	SendTime *time.Time `json:"send_time,omitempty"`
	LocalID  *string    `json:"local_id,omitempty"`
}

// SendSingleMessengerResponse is the response body for MessengerService.SendSingle.
type SendSingleMessengerResponse struct {
	GroupID   string                  `json:"group_id"`
	MessageID string                  `json:"message_id"`
	Status    WebServiceMessageStatus `json:"status"`
	Receptor  string                  `json:"receptor"`
	LocalID   *string                 `json:"local_id"`
	Hide      bool                    `json:"hide"`
	Cost      int64                   `json:"cost"`
	SendTime  *time.Time              `json:"send_time"`
	Profile   string                  `json:"profile"`
	Messenger string                  `json:"messenger"`
}

// BulkMessengerReceptor is one entry of SendBulkMessengerRequest.Receptors.
type BulkMessengerReceptor struct {
	Receptor string  `json:"receptor"`
	LocalID  *string `json:"local_id,omitempty"`
	Hide     *bool   `json:"hide,omitempty"`
}

// SendBulkMessengerRequest is the request body for MessengerService.SendBulk.
type SendBulkMessengerRequest struct {
	Receptors []BulkMessengerReceptor `json:"receptors"`
	Message   string                  `json:"message"`
	SendTime  *time.Time              `json:"send_time,omitempty"`
	Profile   string                  `json:"profile"`
	FileID    *string                 `json:"file_id,omitempty"`
}

// BulkMessengerReceptorResult is one entry of SendBulkMessengerResponse.Receptors.
//
// Status is a plain int, not a typed WebServiceMessageStatus, because a
// failed per-receptor result can carry a WebServiceResponseCode (2000+)
// instead of a WebServiceMessageStatus (1000-1999).
type BulkMessengerReceptorResult struct {
	MessageID *string `json:"message_id"`
	Receptor  string  `json:"receptor"`
	LocalID   *string `json:"local_id"`
	Hide      bool    `json:"hide"`
	Status    int     `json:"status"`
	Cost      int64   `json:"cost"`
}

// SendBulkMessengerResponse is the response body for MessengerService.SendBulk.
type SendBulkMessengerResponse struct {
	GroupID    string                        `json:"group_id"`
	Receptors  []BulkMessengerReceptorResult `json:"receptors"`
	Message    string                        `json:"message"`
	SendTime   *time.Time                    `json:"send_time"`
	TotalCount int                           `json:"total_count"`
	TotalCost  int64                         `json:"total_cost"`
	Counts     map[string]int                `json:"counts"`
	Profile    string                        `json:"profile"`
	Messenger  string                        `json:"messenger"`
}

// P2PMessengerReceptor is one entry of SendP2PMessengerRequest.Receptors.
type P2PMessengerReceptor struct {
	Receptor string  `json:"receptor"`
	Message  string  `json:"message"`
	LocalID  *string `json:"local_id,omitempty"`
	Hide     *bool   `json:"hide,omitempty"`
}

// SendP2PMessengerRequest is the request body for MessengerService.SendP2P.
type SendP2PMessengerRequest struct {
	Receptors []P2PMessengerReceptor `json:"receptors"`
	SendTime  *time.Time             `json:"send_time,omitempty"`
	Profile   string                 `json:"profile"`
	FileID    *string                `json:"file_id,omitempty"`
}

// P2PMessengerReceptorResult is one entry of SendP2PMessengerResponse.Receptors.
//
// Status is a plain int; see BulkMessengerReceptorResult.Status for why.
type P2PMessengerReceptorResult struct {
	MessageID *string `json:"message_id"`
	Receptor  string  `json:"receptor"`
	Message   string  `json:"message"`
	LocalID   *string `json:"local_id"`
	Hide      bool    `json:"hide"`
	Status    int     `json:"status"`
	Cost      int64   `json:"cost"`
}

// SendP2PMessengerResponse is the response body for MessengerService.SendP2P.
type SendP2PMessengerResponse struct {
	GroupID    string                       `json:"group_id"`
	Receptors  []P2PMessengerReceptorResult `json:"receptors"`
	SendTime   *time.Time                   `json:"send_time"`
	TotalCount int                          `json:"total_count"`
	TotalCost  int64                        `json:"total_cost"`
	Counts     map[string]int               `json:"counts"`
	Profile    string                       `json:"profile"`
	Messenger  string                       `json:"messenger"`
}

// SendTemplateMessengerRequest is the request body for MessengerService.SendTemplate.
type SendTemplateMessengerRequest struct {
	TemplateID string                            `json:"template_id"`
	Parameters map[string]TemplateParameterValue `json:"parameters"`
	Receptor   string                            `json:"receptor"`
	LocalID    *string                           `json:"local_id,omitempty"`
	Profile    string                            `json:"profile"`
	ExpiryDate *time.Time                        `json:"expiry_date,omitempty"`
}

// SendTemplateMessengerResponse is the response body for MessengerService.SendTemplate.
type SendTemplateMessengerResponse struct {
	GroupID    string                            `json:"group_id"`
	MessageID  string                            `json:"message_id"`
	Status     WebServiceMessageStatus           `json:"status"`
	LocalID    *string                           `json:"local_id"`
	TemplateID string                            `json:"template_id"`
	SendTime   *time.Time                        `json:"send_time"`
	ExpiryDate *time.Time                        `json:"expiry_date"`
	Cost       int64                             `json:"cost"`
	Receptor   string                            `json:"receptor"`
	Message    string                            `json:"message"`
	Profile    string                            `json:"profile"`
	Messenger  string                            `json:"messenger"`
	Parameters map[string]TemplateParameterValue `json:"parameters"`
}

// UploadMessengerFileResponse is the response body for MessengerService.UploadFile.
type UploadMessengerFileResponse struct {
	FileID string `json:"file_id"`
}

// MessengerStatusItem is one entry of GetMessengerStatusResponse.Receptors.
type MessengerStatusItem struct {
	MessageID    string                  `json:"message_id"`
	LocalID      *string                 `json:"local_id"`
	Status       WebServiceMessageStatus `json:"status"`
	Receptor     string                  `json:"receptor"`
	SendTime     *time.Time              `json:"send_time"`
	DeliveryTime *time.Time              `json:"delivery_time"`
}

// GetMessengerStatusResponse is the response body for MessengerService.GetStatus.
type GetMessengerStatusResponse struct {
	Receptors []MessengerStatusItem `json:"receptors"`
}

// CancelMessengerRequest is the request body for MessengerService.Cancel. At
// least one of MessageIDs or LocalIDs must be non-empty.
type CancelMessengerRequest struct {
	MessageIDs []string `json:"message_ids,omitempty"`
	LocalIDs   []string `json:"local_ids,omitempty"`
}

// CancelledMessengerItem is one entry of CancelMessengerResponse.CancelledMessages
// or CancelMessengerResponse.FailedToCancel.
type CancelledMessengerItem struct {
	MessageID string                  `json:"message_id"`
	LocalID   *string                 `json:"local_id"`
	Status    WebServiceMessageStatus `json:"status"`
}

// CancelMessengerResponse is the response body for MessengerService.Cancel.
type CancelMessengerResponse struct {
	CancelledMessages []CancelledMessengerItem `json:"cancelled_messages"`
	FailedToCancel    []CancelledMessengerItem `json:"failed_to_cancel"`
}

// SendSingle sends a single Messenger message to one receptor.
func (m *MessengerService) SendSingle(ctx context.Context, req *SendSingleMessengerRequest) (*SendSingleMessengerResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Message, "Message"); err != nil {
		return nil, err
	}
	if err := requireMaxLength(req.Message, maxMessengerMessageLength, "Message"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Receptor, "Receptor"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Profile, "Profile"); err != nil {
		return nil, err
	}
	if err := validateLocalID(req.LocalID, "LocalID"); err != nil {
		return nil, err
	}

	return doPostJSON[*SendSingleMessengerResponse](ctx, m.client, "/v1/messenger/single", req)
}

// SendBulk sends the same message to many receptors in one call.
func (m *MessengerService) SendBulk(ctx context.Context, req *SendBulkMessengerRequest) (*SendBulkMessengerResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmptySlice(req.Receptors, "Receptors"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Message, "Message"); err != nil {
		return nil, err
	}
	if err := requireMaxLength(req.Message, maxMessengerMessageLength, "Message"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Profile, "Profile"); err != nil {
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

	return doPostJSON[*SendBulkMessengerResponse](ctx, m.client, "/v1/messenger/bulk", req)
}

// SendP2P sends distinct, per-receptor messages in one call.
func (m *MessengerService) SendP2P(ctx context.Context, req *SendP2PMessengerRequest) (*SendP2PMessengerResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmptySlice(req.Receptors, "Receptors"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Profile, "Profile"); err != nil {
		return nil, err
	}
	for i := range req.Receptors {
		if err := requireNonEmpty(req.Receptors[i].Receptor, "Receptors[].Receptor"); err != nil {
			return nil, err
		}
		if err := requireNonEmpty(req.Receptors[i].Message, "Receptors[].Message"); err != nil {
			return nil, err
		}
		if err := requireMaxLength(req.Receptors[i].Message, maxMessengerMessageLength, "Receptors[].Message"); err != nil {
			return nil, err
		}
		if err := validateLocalID(req.Receptors[i].LocalID, "Receptors[].LocalID"); err != nil {
			return nil, err
		}
	}

	return doPostJSON[*SendP2PMessengerResponse](ctx, m.client, "/v1/messenger/p2p", req)
}

// UploadFile uploads a file (image, video, or document, per the API's
// supported-type list) to be attached to a subsequent Messenger send via its
// returned FileID. r is read to completion but never closed by this method.
func (m *MessengerService) UploadFile(ctx context.Context, r io.Reader, fileName, contentType string) (*UploadMessengerFileResponse, error) {
	if r == nil {
		return nil, &ValidationError{Field: "file", Message: "is required"}
	}
	if err := requireNonEmpty(fileName, "fileName"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(contentType, "contentType"); err != nil {
		return nil, err
	}

	return doPostMultipart[*UploadMessengerFileResponse](ctx, m.client, "/v1/messenger/file", fileName, contentType, r)
}

// Cancel cancels previously scheduled messages by message ID and/or local ID.
// At least one of req.MessageIDs or req.LocalIDs must be non-empty.
func (m *MessengerService) Cancel(ctx context.Context, req *CancelMessengerRequest) (*CancelMessengerResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireAtLeastOne(len(req.MessageIDs) > 0, len(req.LocalIDs) > 0, "at least one of MessageIDs or LocalIDs must be non-empty"); err != nil {
		return nil, err
	}

	return doPostJSON[*CancelMessengerResponse](ctx, m.client, "/v1/messenger/cancel", req)
}

// SendTemplate sends a pre-approved message template, filled in with
// Parameters, to one receptor.
func (m *MessengerService) SendTemplate(ctx context.Context, req *SendTemplateMessengerRequest) (*SendTemplateMessengerResponse, error) {
	if err := requireRequest(req); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.TemplateID, "TemplateID"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Receptor, "Receptor"); err != nil {
		return nil, err
	}
	if err := requireNonEmpty(req.Profile, "Profile"); err != nil {
		return nil, err
	}
	if err := validateLocalID(req.LocalID, "LocalID"); err != nil {
		return nil, err
	}

	return doPostJSON[*SendTemplateMessengerResponse](ctx, m.client, "/v1/messenger/template", req)
}

// GetStatus looks up the delivery status of previously sent messages by
// message ID and/or local ID. At least one of messageIDs or localIDs must be
// non-empty, and their combined length must not exceed 2000.
func (m *MessengerService) GetStatus(ctx context.Context, messageIDs, localIDs []string) (*GetMessengerStatusResponse, error) {
	if err := requireAtLeastOne(len(messageIDs) > 0, len(localIDs) > 0, "at least one of messageIDs or localIDs is required"); err != nil {
		return nil, err
	}
	if err := requireCombinedCountAtMost(distinctCount(messageIDs), distinctCount(localIDs), maxCombinedIDsLookup, "the combined distinct count of messageIDs and localIDs must not exceed 2000"); err != nil {
		return nil, err
	}

	path := "/v1/messenger/status" + buildIDsQuery(messageIDs, localIDs)
	return doGet[*GetMessengerStatusResponse](ctx, m.client, path)
}
