// Package httpclient_test tests the internal httpclient package.
package httpclient_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/phuvinh010701/mezon-go-sdk/auth"
	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
	"github.com/phuvinh010701/mezon-go-sdk/internal/httpclient"
)

func newTestClient(t *testing.T, srv *httptest.Server, maxRetries int) *httpclient.Client {
	t.Helper()
	a, err := auth.NewAPIKeyAuth("test-key")
	if err != nil {
		t.Fatalf("NewAPIKeyAuth: %v", err)
	}
	return httpclient.New(srv.Client(), srv.URL, a, maxRetries)
}

func TestDo_SuccessDecodes(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload{Name: "mezon"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv, 0)

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
	c := httpclient.New(srv.Client(), srv.URL, a, 0)

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

	c := newTestClient(t, srv, 0)

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

	// maxRetries=0 so the test is deterministic (no sleeps).
	c := newTestClient(t, srv, 0)

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

	c := newTestClient(t, srv, 0)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/item", nil)
	if err := c.Do(context.Background(), req, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDo_RetryOn5xx verifies that 5xx responses are retried up to maxRetries times
// and that the final wrapped error is returned when all attempts fail.
func TestDo_RetryOn5xx(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":"server_error","message":"boom"}`))
	}))
	defer srv.Close()

	const maxRetries = 2
	c := newTestClient(t, srv, maxRetries)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/flaky", nil)
	err := c.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error after retries, got nil")
	}

	// 1 initial attempt + maxRetries retries.
	if got := int(callCount.Load()); got != maxRetries+1 {
		t.Fatalf("expected %d calls, got %d", maxRetries+1, got)
	}

	var apiErr *sdkerrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected wrapped *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", apiErr.StatusCode)
	}
}

// TestDo_NoRetryOn4xx verifies that 4xx responses are NOT retried.
func TestDo_NoRetryOn4xx(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"bad_request","message":"invalid"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, 3)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/bad", nil)
	err := c.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got := int(callCount.Load()); got != 1 {
		t.Fatalf("expected exactly 1 call (no retry on 4xx), got %d", got)
	}
}

// TestDo_RetrySucceedsEventually verifies that a request that fails on the first
// attempt but succeeds on a later one returns the decoded response.
func TestDo_RetrySucceedsEventually(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv, 3)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/recover", nil)
	var out map[string]string
	if err := c.Do(context.Background(), req, &out); err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if out["result"] != "ok" {
		t.Fatalf("unexpected response: %v", out)
	}
	if got := int(callCount.Load()); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

// TestDo_ContextCancellation verifies that a cancelled context causes Do to
// return promptly without completing all retries.
func TestDo_ContextCancellation(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow server — sleep until client disconnects.
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer slow.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	c := newTestClient(t, slow, 5)

	req, _ := http.NewRequest(http.MethodGet, slow.URL+"/slow", nil)
	start := time.Now()
	err := c.Do(ctx, req, nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
	// Should have returned well before the full retry sequence (5 * backoff).
	if elapsed > 3*time.Second {
		t.Fatalf("Do took too long (%v); context cancellation may not be working", elapsed)
	}
}

// TestDo_RateLimited_RetriesWithRetryAfterSeconds verifies that a 429 response
// with a numeric Retry-After header is retried and the wait is honoured.
func TestDo_RateLimited_RetriesWithRetryAfterSeconds(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0") // 0s so the test doesn't actually wait
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"code":"rate_limited","message":"slow down"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv, 2)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/resource", nil)
	var out map[string]string
	if err := c.Do(context.Background(), req, &out); err != nil {
		t.Fatalf("expected success after 429 retry, got: %v", err)
	}
	if got := int(callCount.Load()); got != 2 {
		t.Fatalf("expected 2 calls (1 rate-limited + 1 success), got %d", got)
	}
}

// TestDo_RateLimited_ExhaustsRetries verifies that when all attempts return 429
// the final error is a rate-limit APIError.
func TestDo_RateLimited_ExhaustsRetries(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"code":"rate_limited","message":"slow down"}`))
	}))
	defer srv.Close()

	const maxRetries = 2
	c := newTestClient(t, srv, maxRetries)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/resource", nil)
	err := c.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error after exhausted retries, got nil")
	}
	if !errors.Is(err, sdkerrors.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited in error chain, got: %v", err)
	}
	if got := int(callCount.Load()); got != maxRetries+1 {
		t.Fatalf("expected %d calls, got %d", maxRetries+1, got)
	}
}
