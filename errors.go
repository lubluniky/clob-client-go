package client

import (
	"errors"
	"fmt"
)

// APIError represents an error response from the Polymarket API.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("polymarket: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Message)
}

// Sentinel errors for common HTTP status codes.
var (
	ErrUnauthorized = errors.New("polymarket: unauthorized (401)")
	ErrForbidden    = errors.New("polymarket: forbidden (403)")
	ErrNotFound     = errors.New("polymarket: not found (404)")
	ErrRateLimited  = errors.New("polymarket: rate limited (429)")
)

// AuthError indicates an authentication/signing failure.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("polymarket auth: %s", e.Message)
}

// ValidationError indicates invalid input parameters.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("polymarket validation: %s: %s", e.Field, e.Message)
}

// IsRetryable returns true if the error is transient and the request can be retried.
func IsRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500 || apiErr.StatusCode == 429
	}
	return false
}
