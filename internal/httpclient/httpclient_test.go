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

// --- Retry tests ---

// noBackoffClient uses a RetryConfig with zero backoff by overriding backoff via
// a custom config. To speed tests we use MaxAttempts but rely on test server to
// control responses.

func newFastRetryClient(srv *httptest.Server, a auth.Authenticator, maxAttempts int) *httpclient.Client {
	c := httpclient.New(srv.Client(), srv.URL, a)
	return c.WithRetry(httpclient.RetryConfig{
		MaxAttempts: maxAttempts,
		RetryOn:     []int{429, 500, 502, 503, 504},
	})
}

func TestRetry_429RetriedThenSucceeds(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("key")
	// Use a client that retries; we accept the real backoff here but with t.Parallel this is fine
	// In CI we want fast tests; use a small MaxAttempts
	c := httpclient.New(srv.Client(), srv.URL, a).WithRetry(httpclient.RetryConfig{
		MaxAttempts: 3,
		RetryOn:     []int{429, 500, 502, 503, 504},
	})

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)

	// Override the backoff by using context with generous timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var out map[string]string
	err := c.Do(ctx, req, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&callCount); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}
}

func TestRetry_500ExhaustedReturnsError(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("key")
	c := httpclient.New(srv.Client(), srv.URL, a).WithRetry(httpclient.RetryConfig{
		MaxAttempts: 3,
		RetryOn:     []int{500},
	})

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.Do(ctx, req, nil)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if !sdkerrors.IsAPIError(err, http.StatusInternalServerError) {
		t.Fatalf("expected 500 APIError, got %v", err)
	}
	if got := atomic.LoadInt32(&callCount); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestRetry_ContextCancellationStopsRetry(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("key")
	c := httpclient.New(srv.Client(), srv.URL, a).WithRetry(httpclient.RetryConfig{
		MaxAttempts: 5,
		RetryOn:     []int{429},
	})

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately after first call completes.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	err := c.Do(ctx, req, nil)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
	// Should have fewer than 5 calls.
	if got := atomic.LoadInt32(&callCount); got >= 5 {
		t.Fatalf("expected fewer than 5 calls due to cancellation, got %d", got)
	}
}

func TestRetry_NonRetryable400NotRetried(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	a, _ := auth.NewAPIKeyAuth("key")
	c := httpclient.New(srv.Client(), srv.URL, a).WithRetry(httpclient.RetryConfig{
		MaxAttempts: 3,
		RetryOn:     []int{429, 500, 502, 503, 504},
	})

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	err := c.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error for 400")
	}
	if got := atomic.LoadInt32(&callCount); got != 1 {
		t.Fatalf("expected exactly 1 call (no retry for 400), got %d", got)
	}
}
