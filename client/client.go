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
	DefaultBaseURL     = "https://api.mezon.ai/v1"
	defaultHTTPTimeout = 30 * time.Second
	// DefaultMaxRetries is the default number of retry attempts on transient errors.
	DefaultMaxRetries = 3
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
	maxRetries int
	// err holds the first error produced by an option, surfaced by New.
	err error
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
// If apiKey is empty, New will return a descriptive error.
func WithAPIKey(apiKey string) Option {
	return func(c *config) {
		if c.err != nil {
			return // already failing; don't overwrite the first error
		}
		a, err := auth.NewAPIKeyAuth(apiKey)
		if err != nil {
			c.err = fmt.Errorf("WithAPIKey: %w", err)
			return
		}
		c.auth = a
	}
}

// WithAuthenticator sets a custom Authenticator implementation.
func WithAuthenticator(a auth.Authenticator) Option {
	return func(c *config) { c.auth = a }
}

// WithMaxRetries sets the maximum number of retry attempts for transient
// errors (5xx responses and network-level failures). A value of 0 disables
// retries. The default is DefaultMaxRetries.
func WithMaxRetries(n int) Option {
	return func(c *config) { c.maxRetries = n }
}

// New constructs a new Client. An API key must be provided via WithAPIKey or
// WithAuthenticator, otherwise New returns ErrMissingAPIKey.
func New(opts ...Option) (*Client, error) {
	cfg := &config{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		logger:     slog.Default(),
		maxRetries: DefaultMaxRetries,
	}

	for _, o := range opts {
		o(cfg)
	}

	// Surface any option-level error (e.g. empty API key) before proceeding.
	if cfg.err != nil {
		return nil, cfg.err
	}

	if cfg.auth == nil {
		return nil, fmt.Errorf("%w: provide WithAPIKey or WithAuthenticator", sdkerrors.ErrMissingAPIKey)
	}

	return &Client{
		http:   httpclient.New(cfg.httpClient, cfg.baseURL, cfg.auth, cfg.maxRetries),
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
