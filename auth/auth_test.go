package auth

import (
	"errors"
	"testing"
)

func TestAuthErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *AuthError
		expected string
	}{
		{
			name:     "simple message",
			err:      &AuthError{Message: "invalid token"},
			expected: "authentication failed: invalid token",
		},
		{
			name:     "empty message",
			err:      &AuthError{Message: ""},
			expected: "authentication failed: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("AuthError.Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAuthErrorUnwrap(t *testing.T) {
	wrappedErr := errors.New("underlying error")

	t.Run("with wrapped error", func(t *testing.T) {
		authErr := &AuthError{
			Message: "token expired",
			Err:     wrappedErr,
		}

		unwrapped := authErr.Unwrap()
		if unwrapped != wrappedErr {
			t.Errorf("AuthError.Unwrap() = %v, want %v", unwrapped, wrappedErr)
		}
	})

	t.Run("without wrapped error", func(t *testing.T) {
		authErr := &AuthError{Message: "token expired"}

		unwrapped := authErr.Unwrap()
		if unwrapped != nil {
			t.Errorf("AuthError.Unwrap() = %v, want nil", unwrapped)
		}
	})
}

func TestAuthErrorChain(t *testing.T) {
	underlyingErr := errors.New("network timeout")
	authErr := &AuthError{
		Message: "failed to authenticate",
		Err:     underlyingErr,
	}

	// Test errors.Is
	if !errors.Is(authErr, underlyingErr) {
		t.Error("errors.Is should find underlying error in chain")
	}

	// Test errors.As
	var targetErr *AuthError
	if !errors.As(authErr, &targetErr) {
		t.Error("errors.As should find AuthError in chain")
	}
}

func TestBotUserConstants(t *testing.T) {
	// Verify bot user constants are set correctly
	if UsernameDependabot != "dependabot[bot]" {
		t.Errorf("UsernameDependabot = %q, want %q", UsernameDependabot, "dependabot[bot]")
	}

	if UserIDDependabot != 49699333 {
		t.Errorf("UserIDDependabot = %d, want %d", UserIDDependabot, 49699333)
	}
}
