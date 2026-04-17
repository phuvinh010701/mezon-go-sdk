// Package auth provides authentication mechanisms for the Mezon SDK.
package auth

import (
	"net/http"
)

// Authenticator is implemented by any value that can attach authentication
// credentials to an HTTP request.
type Authenticator interface {
	Authenticate(req *http.Request) error
}

// APIKeyAuth authenticates requests using a static API key in the
// Authorization header.
type APIKeyAuth struct {
	apiKey string
}

// NewAPIKeyAuth returns an Authenticator that sets the Authorization header
// to "Bearer <apiKey>".
func NewAPIKeyAuth(apiKey string) *APIKeyAuth {
	return &APIKeyAuth{apiKey: apiKey}
}

// Authenticate sets the Authorization header on the given request.
func (a *APIKeyAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	return nil
}
