// Package httpclient provides the internal HTTP transport used by the SDK.
// It is not part of the public API.
package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
)

// Doer is the interface for making HTTP requests, allowing transport injection.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client wraps an HTTP Doer and handles JSON encoding/decoding and error parsing.
type Client struct {
	httpClient Doer
	baseURL    string
}

// New creates a new internal HTTP client.
func New(httpClient Doer, baseURL string) *Client {
	return &Client{httpClient: httpClient, baseURL: baseURL}
}

// Do executes an HTTP request and decodes the JSON response body into out.
// On non-2xx responses it returns an *sdkerrors.APIError.
func (c *Client) Do(ctx context.Context, req *http.Request, out any) error {
	req = req.WithContext(ctx)

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
		apiErr := &sdkerrors.APIError{StatusCode: resp.StatusCode}
		// Attempt to parse a structured error body.
		var payload struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if jsonErr := json.Unmarshal(body, &payload); jsonErr == nil {
			apiErr.Code = payload.Code
			apiErr.Message = payload.Message
		} else {
			apiErr.Message = string(body)
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

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}
