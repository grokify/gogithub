package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-github/v84/github"
)

func TestTranslateNotFound(t *testing.T) {
	resp := &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}
	err := Translate(errors.New("original"), resp)

	if !IsNotFound(err) {
		t.Error("expected IsNotFound to return true")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected APIError")
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}

func TestTranslatePermissionDenied(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"unauthorized", http.StatusUnauthorized},
		{"forbidden", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &github.Response{Response: &http.Response{StatusCode: tt.statusCode}}
			err := Translate(errors.New("original"), resp)

			if !IsPermissionDenied(err) {
				t.Error("expected IsPermissionDenied to return true")
			}
		})
	}
}

func TestTranslateRateLimited(t *testing.T) {
	resp := &github.Response{Response: &http.Response{StatusCode: http.StatusTooManyRequests}}
	err := Translate(errors.New("original"), resp)

	if !IsRateLimited(err) {
		t.Error("expected IsRateLimited to return true")
	}
}

func TestTranslateConflict(t *testing.T) {
	resp := &github.Response{Response: &http.Response{StatusCode: http.StatusConflict}}
	err := Translate(errors.New("original"), resp)

	if !IsConflict(err) {
		t.Error("expected IsConflict to return true")
	}
}

func TestTranslateValidation(t *testing.T) {
	resp := &github.Response{Response: &http.Response{StatusCode: http.StatusUnprocessableEntity}}
	err := Translate(errors.New("original"), resp)

	if !IsValidation(err) {
		t.Error("expected IsValidation to return true")
	}
}

func TestTranslateServerError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"internal server error", http.StatusInternalServerError},
		{"bad gateway", http.StatusBadGateway},
		{"service unavailable", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &github.Response{Response: &http.Response{StatusCode: tt.statusCode}}
			err := Translate(errors.New("original"), resp)

			if !IsServerError(err) {
				t.Error("expected IsServerError to return true")
			}
		})
	}
}

func TestTranslateNilError(t *testing.T) {
	err := Translate(nil, nil)
	if err != nil {
		t.Errorf("Translate(nil, nil) = %v, want nil", err)
	}
}

func TestTranslateUnknownError(t *testing.T) {
	original := errors.New("unknown error")
	err := Translate(original, nil)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected APIError")
	}

	if !errors.Is(apiErr.Err, original) {
		t.Error("expected Err to be the original error")
	}
}

func TestAPIErrorError(t *testing.T) {
	tests := []struct {
		name    string
		apiErr  *APIError
		wantMsg string
	}{
		{
			name:    "with message",
			apiErr:  &APIError{StatusCode: 404, Message: "resource not found"},
			wantMsg: "github api error 404: resource not found",
		},
		{
			name:    "without message",
			apiErr:  &APIError{StatusCode: 500, Err: errors.New("internal")},
			wantMsg: "github api error 500: internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.apiErr.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestAPIErrorUnwrap(t *testing.T) {
	inner := errors.New("inner error")
	apiErr := &APIError{StatusCode: 500, Err: inner}

	if !errors.Is(apiErr, inner) {
		t.Error("expected APIError to unwrap to inner error")
	}
}

func TestAPIErrorIs(t *testing.T) {
	err1 := &APIError{StatusCode: 404, Err: ErrNotFound}
	err2 := &APIError{StatusCode: 404, Err: ErrServerError}

	if !err1.Is(err2) {
		t.Error("expected Is to return true for same status code")
	}

	err3 := &APIError{StatusCode: 500, Err: ErrServerError}
	if err1.Is(err3) {
		t.Error("expected Is to return false for different status code")
	}

	// Test Is with underlying error
	if !errors.Is(err1, ErrNotFound) {
		t.Error("expected Is to work with underlying error")
	}
}

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "APIError",
			err:  &APIError{StatusCode: 404},
			want: 404,
		},
		{
			name: "plain error",
			err:  errors.New("plain"),
			want: 0,
		},
		{
			name: "nil",
			err:  nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusCode(tt.err); got != tt.want {
				t.Errorf("StatusCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsHelpers(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		helper func(error) bool
		want   bool
	}{
		{"IsNotFound true", ErrNotFound, IsNotFound, true},
		{"IsNotFound false", ErrPermissionDenied, IsNotFound, false},
		{"IsPermissionDenied true", ErrPermissionDenied, IsPermissionDenied, true},
		{"IsPermissionDenied false", ErrNotFound, IsPermissionDenied, false},
		{"IsRateLimited true", ErrRateLimited, IsRateLimited, true},
		{"IsRateLimited false", ErrNotFound, IsRateLimited, false},
		{"IsConflict true", ErrConflict, IsConflict, true},
		{"IsConflict false", ErrNotFound, IsConflict, false},
		{"IsValidation true", ErrValidation, IsValidation, true},
		{"IsValidation false", ErrNotFound, IsValidation, false},
		{"IsServerError true", ErrServerError, IsServerError, true},
		{"IsServerError false", ErrNotFound, IsServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.helper(tt.err); got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsHelpersWithWrappedErrors(t *testing.T) {
	wrapped := &APIError{StatusCode: 404, Err: ErrNotFound}

	if !IsNotFound(wrapped) {
		t.Error("IsNotFound should return true for wrapped ErrNotFound")
	}
}
