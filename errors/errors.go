// Package errors defines SDK-level error types and sentinel errors.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the Mezon API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mezon api error (status %d, code %s): %s", e.StatusCode, e.Code, e.Message)
}

// Sentinel errors for common failure modes.
var (
	ErrUnauthorized     = &APIError{StatusCode: http.StatusUnauthorized, Code: "unauthorized", Message: "unauthorized"}
	ErrForbidden        = &APIError{StatusCode: http.StatusForbidden, Code: "forbidden", Message: "forbidden"}
	ErrNotFound         = &APIError{StatusCode: http.StatusNotFound, Code: "not_found", Message: "resource not found"}
	ErrRateLimited      = &APIError{StatusCode: http.StatusTooManyRequests, Code: "rate_limited", Message: "rate limit exceeded"}
	ErrInternalServer   = &APIError{StatusCode: http.StatusInternalServerError, Code: "internal", Message: "internal server error"}
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrMissingAPIKey    = errors.New("missing api key")
)

// IsAPIError reports whether err is an *APIError with the given HTTP status code.
func IsAPIError(err error, statusCode int) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == statusCode
	}
	return false
}
