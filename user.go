package adsefid

import (
	"context"
	"encoding/json"
	"math"
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
	CreditLeft    float64 `json:"credit_left"`
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

// GetUserTemplatesRequest is the query for UserService.GetTemplates. Every
// field is optional: State filters by moderation state; Skip (if set) must be
// non-negative; Take (if set) must be in [1, 100].
type GetUserTemplatesRequest struct {
	State *TemplateState
	Skip  *int
	Take  *int
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
// templates. A nil req lists the first page in every state.
func (u *UserService) GetTemplates(ctx context.Context, req *GetUserTemplatesRequest) (*GetUserTemplatesResponse, error) {
	if req == nil {
		req = &GetUserTemplatesRequest{}
	}
	if req.Skip != nil {
		if err := requireInRange(*req.Skip, 0, math.MaxInt, "Skip"); err != nil {
			return nil, err
		}
	}
	if req.Take != nil {
		if err := requireInRange(*req.Take, minTemplatesTake, maxTemplatesTake, "Take"); err != nil {
			return nil, err
		}
	}

	values := url.Values{}
	if req.State != nil {
		values.Set("state", string(*req.State))
	}
	if req.Skip != nil {
		values.Set("skip", strconv.Itoa(*req.Skip))
	}
	if req.Take != nil {
		values.Set("take", strconv.Itoa(*req.Take))
	}

	path := "/v1/user/templates"
	if len(values) > 0 {
		path += "?" + values.Encode()
	}
	return doGet[*GetUserTemplatesResponse](ctx, u.client, path)
}
