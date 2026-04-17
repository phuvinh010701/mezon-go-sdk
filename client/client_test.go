// Package client_test tests the public API of the client package.
package client_test

import (
	"testing"

	"github.com/phuvinh010701/mezon-go-sdk/client"
	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
	"errors"
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
