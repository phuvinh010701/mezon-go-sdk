// Package httpclient_test tests the internal httpclient package.
package httpclient_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phuvinh010701/mezon-go-sdk/auth"
	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
	"github.com/phuvinh010701/mezon-go-sdk/internal/httpclient"
)

func TestDo_SuccessDecodes(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload{Name: "mezon"})
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("test-key")
	c := httpclient.New(srv.Client(), srv.URL, a)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	var out payload
	if err := c.Do(context.Background(), req, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Name != "mezon" {
		t.Fatalf("want Name=mezon, got %q", out.Name)
	}
}

func TestDo_SetsAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("secret-key")
	c := httpclient.New(srv.Client(), srv.URL, a)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/ping", nil)
	_ = c.Do(context.Background(), req, nil)

	want := "Bearer secret-key"
	if gotAuth != want {
		t.Fatalf("Authorization header: want %q, got %q", want, gotAuth)
	}
}

func TestDo_NonSuccessReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"code": "not_found", "message": "item missing"})
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("key")
	c := httpclient.New(srv.Client(), srv.URL, a)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/missing", nil)
	err := c.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !sdkerrors.IsAPIError(err, http.StatusNotFound) {
		t.Fatalf("expected 404 APIError, got: %v", err)
	}
	if !errors.Is(err, sdkerrors.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, ErrNotFound) to be true")
	}
}

func TestDo_UnstructuredErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("plain text error"))
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("key")
	c := httpclient.New(srv.Client(), srv.URL, a)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/crash", nil)
	err := c.Do(context.Background(), req, nil)

	var apiErr *sdkerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Message != "plain text error" {
		t.Fatalf("expected message 'plain text error', got %q", apiErr.Message)
	}
}

func TestDo_NilOutSkipsDecoding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("key")
	c := httpclient.New(srv.Client(), srv.URL, a)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/item", nil)
	if err := c.Do(context.Background(), req, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
