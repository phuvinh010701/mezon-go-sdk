// Package client provides the top-level Mezon SDK client.
package client

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/phuvinh010701/mezon-go-sdk/auth"
	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
	"github.com/phuvinh010701/mezon-go-sdk/internal/httpclient"
)

const (
	// DefaultBaseURL is the default Mezon API endpoint.
	DefaultBaseURL    = "https://api.mezon.ai/v1"
	defaultHTTPTimeout = 30 * time.Second
)

// Client is the main entry point for the Mezon SDK.
// Use New to construct one.
type Client struct {
	http   *httpclient.Client
	logger *slog.Logger
}

// Option is a functional option for configuring a Client.
type Option func(*config)

type config struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
	auth       auth.Authenticator
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *config) { c.baseURL = url }
}

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) { c.httpClient = hc }
}

// WithLogger sets a custom structured logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *config) { c.logger = l }
}

// WithAPIKey configures bearer-token authentication.
// Returns an option that errors during New if the key is empty.
func WithAPIKey(apiKey string) Option {
	return func(c *config) {
		a, err := auth.NewAPIKeyAuth(apiKey)
		if err != nil {
			// Store nil; New will detect missing auth and report clearly.
			return
		}
		c.auth = a
	}
}

// WithAuthenticator sets a custom Authenticator implementation.
func WithAuthenticator(a auth.Authenticator) Option {
	return func(c *config) { c.auth = a }
}

// New constructs a new Client. An API key must be provided via WithAPIKey or
// WithAuthenticator, otherwise New returns ErrMissingAPIKey.
func New(opts ...Option) (*Client, error) {
	cfg := &config{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		logger: slog.Default(),
	}

	for _, o := range opts {
		o(cfg)
	}

	if cfg.auth == nil {
		return nil, fmt.Errorf("%w: provide WithAPIKey or WithAuthenticator", sdkerrors.ErrMissingAPIKey)
	}

	return &Client{
		http:   httpclient.New(cfg.httpClient, cfg.baseURL, cfg.auth),
		logger: cfg.logger,
	}, nil
}

// Logger returns the client's structured logger.
func (c *Client) Logger() *slog.Logger { return c.logger }

// BaseURL returns the configured API base URL.
func (c *Client) BaseURL() string { return c.http.BaseURL() }

// HTTP returns the internal httpclient for use by sub-resource clients.
// It is exported for use within the SDK packages only.
func (c *Client) HTTP() *httpclient.Client { return c.http }
