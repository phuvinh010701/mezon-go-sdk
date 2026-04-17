// Package types defines shared types used across the Mezon Go SDK.
package types

// PaginationParams holds common pagination parameters for list requests.
type PaginationParams struct {
	Limit  int
	Offset int
}

// Response is a generic wrapper for API responses with metadata.
type Response[T any] struct {
	Data    T      `json:"data"`
	Total   int    `json:"total,omitempty"`
	HasMore bool   `json:"has_more,omitempty"`
}
