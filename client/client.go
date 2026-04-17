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
	defaultBaseURL = "https://api.mezon.ai/v1"
	defaultTimeout = 30 * time.Second
)

// Client is the main entry point for the Mezon SDK.
// Use New to construct one.
type Client struct {
	baseURL    string
	httpClient *httpclient.Client
	logger     *slog.Logger
	auth       auth.Authenticator
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
func WithAPIKey(apiKey string) Option {
	return func(c *config) { c.auth = auth.NewAPIKeyAuth(apiKey) }
}

// WithAuthenticator sets a custom Authenticator implementation.
func WithAuthenticator(a auth.Authenticator) Option {
	return func(c *config) { c.auth = a }
}

// New constructs a new Client. An API key must be provided via WithAPIKey or
// WithAuthenticator, otherwise New returns ErrMissingAPIKey.
func New(opts ...Option) (*Client, error) {
	cfg := &config{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
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
		baseURL:    cfg.baseURL,
		httpClient: httpclient.New(cfg.httpClient, cfg.baseURL),
		logger:     cfg.logger,
		auth:       cfg.auth,
	}, nil
}

// Logger returns the client's structured logger.
func (c *Client) Logger() *slog.Logger { return c.logger }

// BaseURL returns the configured API base URL.
func (c *Client) BaseURL() string { return c.baseURL }
