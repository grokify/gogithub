// Package errors provides error types and translation utilities for GitHub API errors.
package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v82/github"
)

// Standard errors for GitHub operations.
var (
	// ErrNotFound indicates that the requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrPermissionDenied indicates insufficient permissions for the operation.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrRateLimited indicates the API rate limit has been exceeded.
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrConflict indicates a conflict (e.g., resource already exists).
	ErrConflict = errors.New("conflict")

	// ErrValidation indicates a validation error in the request.
	ErrValidation = errors.New("validation failed")

	// ErrServerError indicates a GitHub server error.
	ErrServerError = errors.New("server error")
)

// APIError wraps a GitHub API error with additional context.
type APIError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("github api error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("github api error %d: %v", e.StatusCode, e.Err)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// Is implements error matching for APIError.
func (e *APIError) Is(target error) bool {
	if t, ok := target.(*APIError); ok {
		return e.StatusCode == t.StatusCode
	}
	return errors.Is(e.Err, target)
}

// Translate converts a GitHub API error to a standard error.
// It examines both the response status code and the error type
// to return an appropriate standard error.
//
// The returned error wraps the original error, so the caller can
// use errors.Unwrap or errors.Is/As to access the original.
func Translate(err error, resp *github.Response) error {
	if err == nil {
		return nil
	}

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	// Check response status code first
	if statusCode != 0 {
		switch statusCode {
		case http.StatusNotFound:
			return &APIError{StatusCode: statusCode, Err: ErrNotFound}
		case http.StatusUnauthorized, http.StatusForbidden:
			return &APIError{StatusCode: statusCode, Err: ErrPermissionDenied}
		case http.StatusTooManyRequests:
			return &APIError{StatusCode: statusCode, Err: ErrRateLimited}
		case http.StatusConflict:
			return &APIError{StatusCode: statusCode, Err: ErrConflict}
		case http.StatusUnprocessableEntity:
			return &APIError{StatusCode: statusCode, Err: ErrValidation}
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			return &APIError{StatusCode: statusCode, Err: ErrServerError}
		}
	}

	// Check error response type
	var errResp *github.ErrorResponse
	if errors.As(err, &errResp) {
		if errResp.Response != nil {
			statusCode = errResp.Response.StatusCode
			switch statusCode {
			case http.StatusNotFound:
				return &APIError{StatusCode: statusCode, Message: errResp.Message, Err: ErrNotFound}
			case http.StatusUnauthorized, http.StatusForbidden:
				return &APIError{StatusCode: statusCode, Message: errResp.Message, Err: ErrPermissionDenied}
			case http.StatusTooManyRequests:
				return &APIError{StatusCode: statusCode, Message: errResp.Message, Err: ErrRateLimited}
			case http.StatusConflict:
				return &APIError{StatusCode: statusCode, Message: errResp.Message, Err: ErrConflict}
			case http.StatusUnprocessableEntity:
				return &APIError{StatusCode: statusCode, Message: errResp.Message, Err: ErrValidation}
			case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
				return &APIError{StatusCode: statusCode, Message: errResp.Message, Err: ErrServerError}
			}
		}
	}

	// Check for rate limit error
	var rateLimitErr *github.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return &APIError{StatusCode: http.StatusTooManyRequests, Err: ErrRateLimited}
	}

	// Check for abuse rate limit error
	var abuseErr *github.AbuseRateLimitError
	if errors.As(err, &abuseErr) {
		return &APIError{StatusCode: http.StatusTooManyRequests, Err: ErrRateLimited}
	}

	// Return wrapped original error if no specific translation
	return &APIError{StatusCode: statusCode, Err: err}
}

// IsNotFound returns true if the error indicates a not found condition.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsPermissionDenied returns true if the error indicates a permission issue.
func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}

// IsRateLimited returns true if the error indicates rate limiting.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsConflict returns true if the error indicates a conflict.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsValidation returns true if the error indicates a validation failure.
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsServerError returns true if the error indicates a server error.
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerError)
}

// StatusCode extracts the HTTP status code from an error, if available.
// Returns 0 if the status code cannot be determined.
func StatusCode(err error) int {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}

	var errResp *github.ErrorResponse
	if errors.As(err, &errResp) {
		if errResp.Response != nil {
			return errResp.Response.StatusCode
		}
	}

	return 0
}
