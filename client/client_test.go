// Package client_test tests the public API of the client package.
package client_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phuvinh010701/mezon-go-sdk/auth"
	"github.com/phuvinh010701/mezon-go-sdk/client"
	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
)

func TestNew_MissingAPIKey(t *testing.T) {
	_, err := client.New()
	if err == nil {
		t.Fatal("expected error for missing api key, got nil")
	}
	if !errors.Is(err, sdkerrors.ErrMissingAPIKey) {
		t.Fatalf("expected ErrMissingAPIKey, got: %v", err)
	}
}

func TestNew_EmptyAPIKey(t *testing.T) {
	// WithAPIKey("") now propagates a descriptive error through config.err.
	_, err := client.New(client.WithAPIKey(""))
	if err == nil {
		t.Fatal("expected error for empty api key, got nil")
	}
	// The error message should mention the empty key, not the generic missing-key sentinel.
	if errors.Is(err, sdkerrors.ErrMissingAPIKey) {
		t.Fatalf("expected a specific WithAPIKey validation error, not ErrMissingAPIKey; got: %v", err)
	}
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestNew_WithAPIKey(t *testing.T) {
	c, err := client.New(client.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNew_WithBaseURL(t *testing.T) {
	c, err := client.New(
		client.WithAPIKey("test-key"),
		client.WithBaseURL("https://custom.api.example.com"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.BaseURL(); got != "https://custom.api.example.com" {
		t.Fatalf("expected custom base URL, got: %s", got)
	}
}

func TestNew_DefaultBaseURL(t *testing.T) {
	c, err := client.New(client.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.BaseURL(); got != client.DefaultBaseURL {
		t.Fatalf("expected default base URL %q, got %q", client.DefaultBaseURL, got)
	}
}

func TestNew_WithAuthenticator(t *testing.T) {
	a, err := auth.NewAPIKeyAuth("custom-auth-key")
	if err != nil {
		t.Fatalf("unexpected error building authenticator: %v", err)
	}
	c, err := client.New(client.WithAuthenticator(a))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNew_WithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5}
	c, err := client.New(
		client.WithAPIKey("test-key"),
		client.WithHTTPClient(custom),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNew_WithMaxRetries(t *testing.T) {
	c, err := client.New(
		client.WithAPIKey("test-key"),
		client.WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// TestIntegration_AuthHeaderSent verifies that the SDK sends the Authorization
// header on every outgoing request end-to-end.
func TestIntegration_AuthHeaderSent(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := client.New(
		client.WithAPIKey("integration-key"),
		client.WithBaseURL(srv.URL),
		client.WithHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/ping", nil)
	if err := c.HTTP().Do(context.Background(), req, nil); err != nil {
		t.Fatalf("Do returned unexpected error: %v", err)
	}

	want := "Bearer integration-key"
	if gotAuth != want {
		t.Fatalf("Authorization header: want %q, got %q", want, gotAuth)
	}
}
