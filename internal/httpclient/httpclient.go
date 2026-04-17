// Package httpclient provides the internal HTTP transport used by the SDK.
// It is not part of the public API.
package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
}

// New creates a new internal HTTP client.
func New(httpClient Doer, baseURL string, authenticator auth.Authenticator) *Client {
	return &Client{httpClient: httpClient, baseURL: baseURL, auth: authenticator}
}

// Do executes an HTTP request, applies authentication, and decodes the JSON
// response body into out. On non-2xx responses it returns an *sdkerrors.APIError.
func (c *Client) Do(ctx context.Context, req *http.Request, out any) error {
	req = req.WithContext(ctx)

	if c.auth != nil {
		if err := c.auth.Authenticate(req); err != nil {
			return fmt.Errorf("authenticate request: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
		return sdkerrors.ParseAPIError(resp.StatusCode, code, message)
	}

	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}
