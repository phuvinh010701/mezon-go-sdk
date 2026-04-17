// Package auth_test tests the auth package.
package auth_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/phuvinh010701/mezon-go-sdk/auth"
)

// --- APIKeyAuth ---

func TestNewAPIKeyAuth_EmptyKey(t *testing.T) {
	_, err := auth.NewAPIKeyAuth("")
	if err == nil {
		t.Fatal("expected error for empty api key, got nil")
	}
}

func TestNewAPIKeyAuth_ValidKey(t *testing.T) {
	a, err := auth.NewAPIKeyAuth("my-secret-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil authenticator")
	}
}

func TestAuthenticate_SetsAuthorizationHeader(t *testing.T) {
	a, err := auth.NewAPIKeyAuth("test-key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err := a.Authenticate(req); err != nil {
		t.Fatalf("Authenticate returned unexpected error: %v", err)
	}

	want := "Bearer test-key-123"
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization header: want %q, got %q", want, got)
	}
}

func TestAuthenticate_OverwritesPreviousHeader(t *testing.T) {
	a, _ := auth.NewAPIKeyAuth("new-key")

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	req.Header.Set("Authorization", "Bearer old-key")

	if err := a.Authenticate(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "Bearer new-key"
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization header: want %q, got %q", want, got)
	}
}

// --- TokenAuth ---

// makeJWT creates a test JWT with the given exp claim (no real signature).
func makeJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, _ := json.Marshal(map[string]interface{}{"sub": "test", "exp": exp})
	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)
	return fmt.Sprintf("%s.%s.fakesig", header, payloadEnc)
}

func TestTokenAuth_EmptyTokenRejected(t *testing.T) {
	_, err := auth.NewTokenAuth("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestTokenAuth_AuthenticateSetsHeader(t *testing.T) {
	token := makeJWT(time.Now().Add(time.Hour).Unix())
	ta, err := auth.NewTokenAuth(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err := ta.Authenticate(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Bearer " + token
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization header: want %q, got %q", want, got)
	}
}

func TestTokenAuth_IsExpired_WithExpiredToken(t *testing.T) {
	token := makeJWT(time.Now().Add(-time.Hour).Unix())
	ta, _ := auth.NewTokenAuth(token)
	if !ta.IsExpired() {
		t.Fatal("expected token to be expired")
	}
}

func TestTokenAuth_IsExpired_WithFutureToken(t *testing.T) {
	token := makeJWT(time.Now().Add(time.Hour).Unix())
	ta, _ := auth.NewTokenAuth(token)
	if ta.IsExpired() {
		t.Fatal("expected token to not be expired")
	}
}

func TestTokenAuth_IsExpired_InvalidToken(t *testing.T) {
	ta, _ := auth.NewTokenAuth("not.a.jwt")
	// invalid token should be treated as expired
	if !ta.IsExpired() {
		t.Fatal("expected invalid token to be treated as expired")
	}
}

// --- SessionAuth ---

func TestSessionAuth_EmptyClientIDRejected(t *testing.T) {
	_, err := auth.NewSessionAuth("", "apikey")
	if err == nil {
		t.Fatal("expected error for empty clientID")
	}
}

func TestSessionAuth_EmptyAPIKeyRejected(t *testing.T) {
	_, err := auth.NewSessionAuth("clientid", "")
	if err == nil {
		t.Fatal("expected error for empty apiKey")
	}
}

func TestSessionAuth_LazyAuthCalledOnce(t *testing.T) {
	var callCount int32
	token := makeJWT(time.Now().Add(time.Hour).Unix())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}))
	defer srv.Close()

	sa, err := auth.NewSessionAuth("client1", "key1", auth.WithAuthBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
		if err := sa.Authenticate(req); err != nil {
			t.Fatalf("Authenticate error: %v", err)
		}
	}

	if got := atomic.LoadInt32(&callCount); got != 1 {
		t.Fatalf("expected auth endpoint called once, got %d", got)
	}
}

func TestSessionAuth_ReAuthOnExpiry(t *testing.T) {
	var callCount int32

	// First token: already expired. Second token: valid.
	expiredToken := makeJWT(time.Now().Add(-time.Hour).Unix())
	validToken := makeJWT(time.Now().Add(time.Hour).Unix())

	tokens := []string{expiredToken, validToken}
	var idx int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := atomic.AddInt32(&callCount, 1) - 1
		tok := tokens[0]
		if int(i) < len(tokens) {
			tok = tokens[i]
		}
		_ = idx
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tok})
	}))
	defer srv.Close()

	sa, _ := auth.NewSessionAuth("client1", "key1", auth.WithAuthBaseURL(srv.URL))

	// First call: gets expired token, but since it was just obtained it won't re-auth until next call.
	req1, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err := sa.Authenticate(req1); err != nil {
		t.Fatalf("first Authenticate error: %v", err)
	}

	// Second call: token is expired, should re-auth.
	req2, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err := sa.Authenticate(req2); err != nil {
		t.Fatalf("second Authenticate error: %v", err)
	}

	if got := atomic.LoadInt32(&callCount); got != 2 {
		t.Fatalf("expected auth endpoint called twice, got %d", got)
	}

	// Third call: token is valid, no re-auth needed.
	req3, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err := sa.Authenticate(req3); err != nil {
		t.Fatalf("third Authenticate error: %v", err)
	}

	if got := atomic.LoadInt32(&callCount); got != 2 {
		t.Fatalf("expected auth endpoint still called twice, got %d", got)
	}
}

func TestFromEnv_MissingVars(t *testing.T) {
	t.Setenv("MEZON_CLIENT_ID", "")
	t.Setenv("MEZON_API_KEY", "")

	_, err := auth.FromEnv()
	if err == nil {
		t.Fatal("expected error when env vars are missing")
	}
}

func TestFromEnv_MissingAPIKey(t *testing.T) {
	t.Setenv("MEZON_CLIENT_ID", "client1")
	t.Setenv("MEZON_API_KEY", "")

	_, err := auth.FromEnv()
	if err == nil {
		t.Fatal("expected error when MEZON_API_KEY is missing")
	}
}

func TestFromEnv_ValidVars(t *testing.T) {
	t.Setenv("MEZON_CLIENT_ID", "client1")
	t.Setenv("MEZON_API_KEY", "apikey1")

	sa, err := auth.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sa == nil {
		t.Fatal("expected non-nil SessionAuth")
	}
}
