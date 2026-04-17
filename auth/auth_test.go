// Package auth_test tests the auth package.
package auth_test

import (
	"net/http"
	"testing"

	"github.com/phuvinh010701/mezon-go-sdk/auth"
)

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
