package adsefid

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"
)

// UserService groups the User endpoints of the adsefid.com API. Obtain one
// via Client.User rather than constructing it directly.
type UserService struct {
	client *Client
}

// UserInfo is the response body for UserService.GetInfo.
type UserInfo struct {
	Name          string  `json:"name"`
	CompanyName   *string `json:"company_name"`
	CreditLeft    int64   `json:"credit_left"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	AccountStatus string  `json:"account_status"`
}

// UserLine is one entry of the slice returned by UserService.GetLines.
type UserLine struct {
	LineNumber   string       `json:"line_number"`
	LineSelector LineSelector `json:"line_selector"`
	LineName     string       `json:"line_name"`
	Enabled      bool         `json:"enabled"`
}

// UserProfile is one entry of the slice returned by UserService.GetProfiles.
type UserProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Messenger string `json:"messenger"`
}

// UserTemplate is one entry of GetUserTemplatesResponse.Items.
type UserTemplate struct {
	TemplateID  string                           `json:"template_id"`
	Content     string                           `json:"content"`
	Parameters  map[string]TemplateParameterType `json:"parameters"`
	State       TemplateState                    `json:"state"`
	Description *string                          `json:"description"`
	CreatedAt   time.Time                        `json:"created_at"`
	UpdatedAt   time.Time                        `json:"updated_at"`
}

// UnmarshalJSON keeps the documented public parameter-type set forward-compatible
// when the service emits an undocumented value.
func (u *UserTemplate) UnmarshalJSON(data []byte) error {
	var wire struct {
		TemplateID  string            `json:"template_id"`
		Content     string            `json:"content"`
		Parameters  map[string]string `json:"parameters"`
		State       TemplateState     `json:"state"`
		Description *string           `json:"description"`
		CreatedAt   time.Time         `json:"created_at"`
		UpdatedAt   time.Time         `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	parameters := make(map[string]TemplateParameterType, len(wire.Parameters))
	for name, value := range wire.Parameters {
		typeValue := TemplateParameterType(value)
		if typeValue == TemplateParameterTypeString || typeValue == TemplateParameterTypeNumber {
			parameters[name] = typeValue
		}
	}
	*u = UserTemplate{
		TemplateID: wire.TemplateID, Content: wire.Content, Parameters: parameters,
		State: wire.State, Description: wire.Description, CreatedAt: wire.CreatedAt, UpdatedAt: wire.UpdatedAt,
	}
	return nil
}

// GetUserTemplatesResponse is the response body for UserService.GetTemplates.
type GetUserTemplatesResponse struct {
	Items []UserTemplate `json:"items"`
	Total int            `json:"total"`
}

// GetInfo fetches the authenticated account's general information.
func (u *UserService) GetInfo(ctx context.Context) (*UserInfo, error) {
	return doGet[*UserInfo](ctx, u.client, "/v1/user/info")
}

// GetLines fetches the authenticated account's SMS lines.
func (u *UserService) GetLines(ctx context.Context) ([]UserLine, error) {
	return doGet[[]UserLine](ctx, u.client, "/v1/user/lines")
}

// GetProfiles fetches the authenticated account's Messenger profiles.
func (u *UserService) GetProfiles(ctx context.Context) ([]UserProfile, error) {
	return doGet[[]UserProfile](ctx, u.client, "/v1/user/profiles")
}

// GetTemplates fetches a page of the authenticated account's message
// templates. state (if given) filters by moderation state. skip must be >= 0
// if given; take must be in [1, 100] if given.
func (u *UserService) GetTemplates(ctx context.Context, state *TemplateState, skip, take *int) (*GetUserTemplatesResponse, error) {
	if skip != nil {
		if err := requireInRange(*skip, 0, int(^uint(0)>>1), "skip"); err != nil {
			return nil, err
		}
	}
	if take != nil {
		if err := requireInRange(*take, 1, maxTemplatesTake, "take"); err != nil {
			return nil, err
		}
	}

	values := url.Values{}
	if state != nil {
		values.Set("state", string(*state))
	}
	if skip != nil {
		values.Set("skip", strconv.Itoa(*skip))
	}
	if take != nil {
		values.Set("take", strconv.Itoa(*take))
	}

	path := "/v1/user/templates"
	if len(values) > 0 {
		path += "?" + values.Encode()
	}
	return doGet[*GetUserTemplatesResponse](ctx, u.client, path)
}
