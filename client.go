package adsefid

import (
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

// DefaultBaseURL is the default base URL used by NewClient when
// WithBaseURL is not supplied.
const DefaultBaseURL = "https://api.adsefid.com"

const sdkModulePath = "github.com/adsefid/sdk-go"

// Client is a client for the adsefid.com SMS Web Service API. Construct one
// with NewClient. A *Client is safe for concurrent use by multiple
// goroutines, as long as the *http.Client it holds is.
type Client struct {
	SMS       *SMSService
	Messenger *MessengerService
	User      *UserService

	apiKey     string
	baseURL    string
	userAgent  string
	httpClient *http.Client
}

// ClientOption configures a Client constructed by NewClient.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL    string
	userAgent  string
	httpClient *http.Client
	timeout    time.Duration
}

// WithBaseURL overrides the default API base URL (https://api.adsefid.com).
// This is mainly useful for testing against a mock server.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient supplies a custom *http.Client for the SDK to use. When set,
// it takes precedence over WithTimeout:
// WithTimeout only affects the *http.Client this package builds internally
// when WithHTTPClient is not used.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *clientConfig) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the timeout of the default *http.Client this package
// builds when WithHTTPClient is not supplied. It has no effect if
// WithHTTPClient is also supplied.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithUserAgent replaces the default adsefid-go/<SDK_VERSION> User-Agent value.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *clientConfig) {
		c.userAgent = userAgent
	}
}

// NewClient constructs a Client authenticated with apiKey, which is sent on
// every request as the X-API-KEY header. It returns a *ValidationError if
// apiKey is blank.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, &ValidationError{Field: "apiKey", Message: "is required"}
	}

	cfg := &clientConfig{
		baseURL:   DefaultBaseURL,
		userAgent: defaultUserAgent(),
		timeout:   30 * time.Second,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.baseURL = strings.TrimRight(cfg.baseURL, "/")
	if parsed, err := url.Parse(cfg.baseURL); err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, &ValidationError{Field: "baseURL", Message: "must be an absolute http(s) URL"}
	}
	if strings.TrimSpace(cfg.userAgent) == "" || strings.ContainsAny(cfg.userAgent, "\r\n") {
		return nil, &ValidationError{Field: "userAgent", Message: "must be non-blank and contain no line breaks"}
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.timeout}
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    cfg.baseURL,
		userAgent:  cfg.userAgent,
		httpClient: httpClient,
	}
	c.SMS = &SMSService{client: c}
	c.Messenger = &MessengerService{client: c}
	c.User = &UserService{client: c}
	return c, nil
}

func defaultUserAgent() string {
	version := "0+unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Path == sdkModulePath {
			version = info.Main.Version
		} else {
			for _, dependency := range info.Deps {
				if dependency.Path == sdkModulePath {
					version = dependency.Version
					break
				}
			}
		}
	}
	version = strings.TrimPrefix(version, "v")
	if version == "" || version == "(devel)" {
		version = "0+unknown"
	}
	return "adsefid-go/" + version
}
