// Package errors_test tests the errors package.
package errors_test

import (
	"errors"
	"testing"

	sdkerrors "github.com/phuvinh010701/mezon-go-sdk/errors"
)

func TestAPIError_Error(t *testing.T) {
	err := &sdkerrors.APIError{StatusCode: 404, Code: "not_found", Message: "resource not found"}
	want := "mezon api error (status 404, code not_found): resource not found"
	if got := err.Error(); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestIsAPIError(t *testing.T) {
	err := &sdkerrors.APIError{StatusCode: 401, Code: "unauthorized", Message: "unauthorized"}
	if !sdkerrors.IsAPIError(err, 401) {
		t.Fatal("expected IsAPIError to return true for 401")
	}
	if sdkerrors.IsAPIError(err, 403) {
		t.Fatal("expected IsAPIError to return false for 403")
	}
	if sdkerrors.IsAPIError(errors.New("plain error"), 401) {
		t.Fatal("expected IsAPIError to return false for non-APIError")
	}
}

func TestAPIError_ErrorsIs_Sentinel(t *testing.T) {
	cases := []struct {
		name     string
		err      *sdkerrors.APIError
		sentinel error
	}{
		{"unauthorized", sdkerrors.NewUnauthorizedError("unauthorized", "unauthorized"), sdkerrors.ErrUnauthorized},
		{"forbidden", sdkerrors.NewForbiddenError("forbidden", "forbidden"), sdkerrors.ErrForbidden},
		{"not_found", sdkerrors.NewNotFoundError("not_found", "not found"), sdkerrors.ErrNotFound},
		{"rate_limited", sdkerrors.NewRateLimitedError("rate_limited", "rate limited"), sdkerrors.ErrRateLimited},
		{"internal", sdkerrors.NewInternalServerError("internal", "internal error"), sdkerrors.ErrInternalServer},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.err, tc.sentinel) {
				t.Fatalf("expected errors.Is(%v, %v) to be true", tc.err, tc.sentinel)
			}
		})
	}
}

func TestParseAPIError_WrapsCorrectSentinel(t *testing.T) {
	err := sdkerrors.ParseAPIError(404, "not_found", "resource missing")
	if !errors.Is(err, sdkerrors.ErrNotFound) {
		t.Fatalf("expected ParseAPIError(404) to wrap ErrNotFound")
	}
	if !sdkerrors.IsAPIError(err, 404) {
		t.Fatalf("expected IsAPIError(err, 404) to be true")
	}
}

func TestParseAPIError_UnknownStatus_NoSentinel(t *testing.T) {
	err := sdkerrors.ParseAPIError(418, "teapot", "i am a teapot")
	if sdkerrors.IsAPIError(err, 404) {
		t.Fatal("expected IsAPIError(err, 404) to be false for 418")
	}
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}
