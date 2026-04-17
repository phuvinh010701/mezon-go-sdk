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
	// Unwrap allows errors.Is matching against sentinel errors (e.g. ErrUnauthorized).
	sentinel error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mezon api error (status %d, code %s): %s", e.StatusCode, e.Code, e.Message)
}

// Unwrap returns the sentinel error this APIError wraps, enabling errors.Is support.
func (e *APIError) Unwrap() error { return e.sentinel }

// sentinelError is an immutable error value used as the target for errors.Is.
type sentinelError string

func (s sentinelError) Error() string { return string(s) }

// Sentinel errors for common failure modes.
// These are immutable values; use errors.Is to test for them.
var (
	ErrUnauthorized    = sentinelError("unauthorized")
	ErrForbidden       = sentinelError("forbidden")
	ErrNotFound        = sentinelError("not found")
	ErrRateLimited     = sentinelError("rate limited")
	ErrInternalServer  = sentinelError("internal server error")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrMissingAPIKey   = errors.New("missing api key")
	ErrAuthFailed      = errors.New("authentication failed")
	ErrTokenExpired    = sentinelError("token expired")
)

// newAPIError constructs an APIError that wraps the given sentinel so that
// errors.Is(err, sentinel) returns true.
func newAPIError(statusCode int, code, message string, sentinel error) *APIError {
	return &APIError{StatusCode: statusCode, Code: code, Message: message, sentinel: sentinel}
}

// NewUnauthorizedError returns an APIError wrapping ErrUnauthorized.
func NewUnauthorizedError(code, message string) *APIError {
	return newAPIError(http.StatusUnauthorized, code, message, ErrUnauthorized)
}

// NewForbiddenError returns an APIError wrapping ErrForbidden.
func NewForbiddenError(code, message string) *APIError {
	return newAPIError(http.StatusForbidden, code, message, ErrForbidden)
}

// NewNotFoundError returns an APIError wrapping ErrNotFound.
func NewNotFoundError(code, message string) *APIError {
	return newAPIError(http.StatusNotFound, code, message, ErrNotFound)
}

// NewRateLimitedError returns an APIError wrapping ErrRateLimited.
func NewRateLimitedError(code, message string) *APIError {
	return newAPIError(http.StatusTooManyRequests, code, message, ErrRateLimited)
}

// NewInternalServerError returns an APIError wrapping ErrInternalServer.
func NewInternalServerError(code, message string) *APIError {
	return newAPIError(http.StatusInternalServerError, code, message, ErrInternalServer)
}

// ParseAPIError creates an APIError from an HTTP status code and optional
// structured payload. It automatically wraps the appropriate sentinel.
func ParseAPIError(statusCode int, code, message string) *APIError {
	sentinel := sentinelForStatus(statusCode)
	return newAPIError(statusCode, code, message, sentinel)
}

func sentinelForStatus(code int) error {
	switch code {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusInternalServerError:
		return ErrInternalServer
	default:
		return nil
	}
}

// IsAPIError reports whether err is an *APIError with the given HTTP status code.
func IsAPIError(err error, statusCode int) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == statusCode
	}
	return false
}
