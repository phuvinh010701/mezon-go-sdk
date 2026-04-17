// Package httpclient provides the internal HTTP transport used by the SDK.
// It is not part of the public API.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/phuvinh010701/mezon-go-sdk/auth"
	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
)

// Doer is the interface for making HTTP requests, allowing transport injection.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client wraps an HTTP Doer and handles JSON encoding/decoding and error parsing.
// It applies authentication to every outgoing request via the Authenticator.
type Client struct {
	httpClient Doer
	baseURL    string
	auth       auth.Authenticator
	maxRetries int
}

// New creates a new internal HTTP client.
// maxRetries controls how many times a transient request is retried (0 = no retries).
func New(httpClient Doer, baseURL string, authenticator auth.Authenticator, maxRetries int) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		auth:       authenticator,
		maxRetries: maxRetries,
	}
}

// Do executes an HTTP request, applies authentication, and decodes the JSON
// response body into out. On non-2xx responses it returns an *sdkerrors.APIError.
// Transient failures (5xx, 429, timeout network errors) are retried up to
// maxRetries times. For 429 responses the Retry-After header is honoured when
// present; all other retries use exponential backoff. Context cancellation is
// respected at every wait point.
func (c *Client) Do(ctx context.Context, req *http.Request, out any) error {
	// Buffer the body so it can be replayed on retries.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return fmt.Errorf("read request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Rebuild a fresh request for each attempt (body reader is consumed after first use).
		r := req.Clone(ctx)
		if bodyBytes != nil {
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			r.ContentLength = int64(len(bodyBytes))
		}

		if c.auth != nil {
			if err := c.auth.Authenticate(r); err != nil {
				return fmt.Errorf("authenticate request: %w", err)
			}
		}

		resp, err := c.httpClient.Do(r)
		if err != nil {
			// Network-level error — retry if transient.
			if isRetryableNetworkError(err) {
				lastErr = fmt.Errorf("http request (attempt %d): %w", attempt+1, err)
				if !c.sleep(ctx, backoff(attempt+1)) {
					return ctx.Err()
				}
				continue
			}
			return fmt.Errorf("http request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// Attempt to parse a structured error body.
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

			// Retry on 5xx and 429 (rate limited).
			if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
				lastErr = apiErr
				wait := backoff(attempt + 1)
				if resp.StatusCode == http.StatusTooManyRequests {
					// Prefer Retry-After over exponential backoff for rate limiting.
					wait = retryAfter(resp.Header.Get("Retry-After"))
				}
				if !c.sleep(ctx, wait) {
					return ctx.Err()
				}
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

	return fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// sleep waits for d or until ctx is cancelled. Returns false if ctx is done.
func (c *Client) sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// backoff returns the wait duration before the given retry attempt (1-based).
// It uses exponential backoff: 100ms, 200ms, 400ms, …, capped at 10s.
func backoff(attempt int) time.Duration {
	ms := 100 * math.Pow(2, float64(attempt-1))
	if ms > 10_000 {
		ms = 10_000
	}
	return time.Duration(ms) * time.Millisecond
}

// retryAfter parses the value of a Retry-After header and returns the
// corresponding wait duration. It supports both integer seconds ("120") and
// HTTP-date formats. Falls back to exponential backoff attempt-1 (100ms) when
// the header is absent or unparseable.
func retryAfter(header string) time.Duration {
	if header == "" {
		return backoff(1)
	}
	// Try seconds (most common).
	if secs, err := strconv.ParseFloat(header, 64); err == nil && secs > 0 {
		return time.Duration(secs * float64(time.Second))
	}
	// Try HTTP-date.
	if t, err := http.ParseTime(header); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return backoff(1)
}

// isRetryableNetworkError reports whether a transport-level error is transient
// and worth retrying.
func isRetryableNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}
