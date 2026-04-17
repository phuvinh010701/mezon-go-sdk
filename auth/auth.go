// Package auth provides authentication mechanisms for the Mezon SDK.
package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
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
// to "Bearer <apiKey>". It returns an error if apiKey is empty.
func NewAPIKeyAuth(apiKey string) (*APIKeyAuth, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("api key must not be empty")
	}
	return &APIKeyAuth{apiKey: apiKey}, nil
}

// Authenticate sets the Authorization header on the given request.
func (a *APIKeyAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	return nil
}

// TokenAuth authenticates requests using a pre-obtained JWT bearer token.
type TokenAuth struct {
	token string
}

// NewTokenAuth returns a TokenAuth for the given JWT token.
// Returns an error if token is empty.
func NewTokenAuth(token string) (*TokenAuth, error) {
	if token == "" {
		return nil, fmt.Errorf("token must not be empty")
	}
	return &TokenAuth{token: token}, nil
}

// Authenticate sets the Authorization header to "Bearer <token>".
func (t *TokenAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return nil
}

// Token returns the JWT token string.
func (t *TokenAuth) Token() string { return t.token }

// IsExpired reports whether the token's exp claim is in the past.
// It decodes the JWT payload without verifying the signature.
// Returns true on any parse error (treats unparseable as expired).
func (t *TokenAuth) IsExpired() bool {
	exp, err := jwtExpiry(t.token)
	if err != nil {
		return true
	}
	return time.Now().Unix() >= exp
}

// jwtExpiry parses the exp claim from a JWT without signature verification.
func jwtExpiry(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid jwt format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("decode jwt payload: %w", err)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, fmt.Errorf("parse jwt claims: %w", err)
	}
	return claims.Exp, nil
}

// SessionOption configures a SessionAuth.
type SessionOption func(*SessionAuth)

// WithAuthBaseURL sets the base URL used for the token exchange endpoint.
func WithAuthBaseURL(url string) SessionOption {
	return func(s *SessionAuth) { s.authBaseURL = url }
}

const defaultAuthBaseURL = "https://api.mezon.ai"

// SessionAuth exchanges client_id+api_key for a JWT via the Mezon auth
// endpoint, then uses that token for subsequent requests.
// It is safe for concurrent use.
type SessionAuth struct {
	clientID    string
	apiKey      string
	authBaseURL string
	httpClient  *http.Client

	mu        sync.Mutex
	tokenAuth *TokenAuth
}

// NewSessionAuth creates a SessionAuth. Returns an error if clientID or apiKey
// is empty.
func NewSessionAuth(clientID, apiKey string, opts ...SessionOption) (*SessionAuth, error) {
	if clientID == "" {
		return nil, fmt.Errorf("clientID must not be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey must not be empty")
	}
	s := &SessionAuth{
		clientID:    clientID,
		apiKey:      apiKey,
		authBaseURL: defaultAuthBaseURL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// Authenticate lazily obtains a token and sets the Authorization header.
// If the current token is expired it re-authenticates.
func (s *SessionAuth) Authenticate(req *http.Request) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tokenAuth == nil || s.tokenAuth.IsExpired() {
		if err := s.refreshToken(); err != nil {
			return err
		}
	}
	req.Header.Set("Authorization", "Bearer "+s.tokenAuth.Token())
	return nil
}

// refreshToken performs the token exchange. Caller must hold s.mu.
func (s *SessionAuth) refreshToken() error {
	url := s.authBaseURL + "/v2/apps/authenticate/token"

	body := bytes.NewBufferString("{}")
	httpReq, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("%w: create auth request: %v", sdkerrors.ErrAuthFailed, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	creds := base64.StdEncoding.EncodeToString([]byte(s.clientID + ":" + s.apiKey))
	httpReq.Header.Set("Authorization", "Basic "+creds)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", sdkerrors.ErrAuthFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: auth endpoint returned status %d", sdkerrors.ErrAuthFailed, resp.StatusCode)
	}

	var result struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("%w: decode auth response: %v", sdkerrors.ErrAuthFailed, err)
	}
	if result.Token == "" {
		return fmt.Errorf("%w: empty token in auth response", sdkerrors.ErrAuthFailed)
	}

	ta, err := NewTokenAuth(result.Token)
	if err != nil {
		return fmt.Errorf("%w: %v", sdkerrors.ErrAuthFailed, err)
	}
	s.tokenAuth = ta
	return nil
}

// FromEnv creates a SessionAuth from the MEZON_CLIENT_ID and MEZON_API_KEY
// environment variables.
func FromEnv() (*SessionAuth, error) {
	clientID := os.Getenv("MEZON_CLIENT_ID")
	apiKey := os.Getenv("MEZON_API_KEY")
	if clientID == "" {
		return nil, fmt.Errorf("%w: MEZON_CLIENT_ID not set", sdkerrors.ErrAuthFailed)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("%w: MEZON_API_KEY not set", sdkerrors.ErrAuthFailed)
	}
	return NewSessionAuth(clientID, apiKey)
}
