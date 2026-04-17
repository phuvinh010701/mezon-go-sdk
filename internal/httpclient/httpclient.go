// Package httpclient provides the internal HTTP transport used by the SDK.
// It is not part of the public API.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/phuvinh010701/mezon-go-sdk/auth"
	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
)

// RetryConfig controls retry behavior for HTTP requests.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// RetryOn is the list of HTTP status codes that should trigger a retry.
	RetryOn []int
}

// DefaultRetryConfig is the default retry configuration.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	RetryOn:     []int{429, 500, 502, 503, 504},
}

// backoffDurations defines the wait time before each retry attempt (index 0 = wait before attempt 2).
var backoffDurations = []time.Duration{500 * time.Millisecond, time.Second, 2 * time.Second}

// Doer is the interface for making HTTP requests, allowing transport injection.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client wraps an HTTP Doer and handles JSON encoding/decoding and error parsing.
// It applies authentication to every outgoing request via the Authenticator.
type Client struct {
	httpClient  Doer
	baseURL     string
	auth        auth.Authenticator
	retryConfig RetryConfig
}

// New creates a new internal HTTP client with default retry configuration.
func New(httpClient Doer, baseURL string, authenticator auth.Authenticator) *Client {
	return &Client{
		httpClient:  httpClient,
		baseURL:     baseURL,
		auth:        authenticator,
		retryConfig: DefaultRetryConfig,
	}
}

// WithRetry returns a new Client with the given retry configuration.
func (c *Client) WithRetry(cfg RetryConfig) *Client {
	cp := *c
	cp.retryConfig = cfg
	return &cp
}

// NewRequest builds an *http.Request with the base URL prepended to path.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
}

// shouldRetry reports whether statusCode is in the retry list.
func (c *Client) shouldRetry(statusCode int) bool {
	for _, code := range c.retryConfig.RetryOn {
		if code == statusCode {
			return true
		}
	}
	return false
}

// Do executes an HTTP request, applies authentication, and decodes the JSON
// response body into out. On non-2xx responses it returns an *sdkerrors.APIError.
// Retries are performed according to the retry configuration.
func (c *Client) Do(ctx context.Context, req *http.Request, out any) error {
	// Buffer the body so it can be re-read on retries.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("read request body: %w", err)
		}
		req.Body.Close()
	}

	maxAttempts := c.retryConfig.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Wait before retries (not before the first attempt).
		if attempt > 0 {
			idx := attempt - 1
			if idx >= len(backoffDurations) {
				idx = len(backoffDurations) - 1
			}
			select {
			case <-ctx.Done():
				return fmt.Errorf("http request: %w", ctx.Err())
			case <-time.After(backoffDurations[idx]):
			}
		}

		// Check context before each attempt.
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("http request: %w", err)
		}

		// Clone the request so headers/context are fresh.
		cloned := req.Clone(ctx)
		if bodyBytes != nil {
			cloned.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			cloned.ContentLength = int64(len(bodyBytes))
		}

		if c.auth != nil {
			if err := c.auth.Authenticate(cloned); err != nil {
				return fmt.Errorf("authenticate request: %w", err)
			}
		}

		resp, err := c.httpClient.Do(cloned)
		if err != nil {
			// Network errors are not retried.
			return fmt.Errorf("http request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var payload struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			code, message := "", string(body)
			if jsonErr := json.Unmarshal(body, &payload); jsonErr == nil {
				code = payload.Code
				message = payload.Message
			}
			apiErr := sdkerrors.ParseAPIError(resp.StatusCode, code, message)
			if c.shouldRetry(resp.StatusCode) && attempt < maxAttempts-1 {
				lastErr = apiErr
				continue
			}
			return apiErr
		}

		if out != nil && len(body) > 0 {
			if err := json.Unmarshal(body, out); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
		}
		return nil
	}

	return lastErr
}

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}
