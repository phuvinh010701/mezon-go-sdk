// Package errors_test tests the errors package.
package errors_test

import (
	"testing"
	"errors"

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
