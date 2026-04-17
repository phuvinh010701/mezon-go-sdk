// Package transport defines the Transport interface and a default HTTP transport.
package transport

import (
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

// Transport is an interface for HTTP transports, allowing injection of custom
// round-trippers (e.g., for testing or adding middleware).
type Transport interface {
	RoundTrip(req *http.Request) (*http.Response, error)
}

// NewDefaultHTTPClient returns an *http.Client configured with sensible defaults.
func NewDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultTimeout,
	}
}
