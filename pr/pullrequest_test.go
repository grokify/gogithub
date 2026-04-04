package pr

import (
	"errors"
	"testing"
)

func TestPRErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *PRError
		expected string
	}{
		{
			name: "with title and error",
			err: &PRError{
				Title: "Add new feature",
				Err:   errors.New("permission denied"),
			},
			expected: "failed to create PR 'Add new feature': permission denied",
		},
		{
			name: "empty title",
			err: &PRError{
				Title: "",
				Err:   errors.New("network error"),
			},
			expected: "failed to create PR '': network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("PRError.Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPRErrorUnwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	prErr := &PRError{
		Title: "Test PR",
		Err:   underlyingErr,
	}

	unwrapped := prErr.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("PRError.Unwrap() = %v, want %v", unwrapped, underlyingErr)
	}
}

func TestPRErrorChain(t *testing.T) {
	underlyingErr := errors.New("API rate limit exceeded")
	prErr := &PRError{
		Title: "Fix bug",
		Err:   underlyingErr,
	}

	// Test errors.Is
	if !errors.Is(prErr, underlyingErr) {
		t.Error("errors.Is should find underlying error in chain")
	}

	// Test errors.As
	var targetErr *PRError
	if !errors.As(prErr, &targetErr) {
		t.Error("errors.As should find PRError in chain")
	}
	if targetErr.Title != "Fix bug" {
		t.Errorf("Title = %q, want %q", targetErr.Title, "Fix bug")
	}
}

func TestReviewEventConstants(t *testing.T) {
	tests := []struct {
		name     string
		event    ReviewEvent
		expected string
	}{
		{
			name:     "approve",
			event:    ReviewEventApprove,
			expected: "APPROVE",
		},
		{
			name:     "request changes",
			event:    ReviewEventRequestChanges,
			expected: "REQUEST_CHANGES",
		},
		{
			name:     "comment",
			event:    ReviewEventComment,
			expected: "COMMENT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.event) != tt.expected {
				t.Errorf("ReviewEvent = %q, want %q", string(tt.event), tt.expected)
			}
		})
	}
}

func TestMergeableStateStruct(t *testing.T) {
	state := &MergeableState{
		Mergeable: true,
		State:     "clean",
		Message:   "PR is ready to merge",
	}

	if !state.Mergeable {
		t.Error("Mergeable should be true")
	}
	if state.State != "clean" {
		t.Errorf("State = %q, want %q", state.State, "clean")
	}
	if state.Message != "PR is ready to merge" {
		t.Errorf("Message = %q, want %q", state.Message, "PR is ready to merge")
	}
}

func TestMergeableStateVariants(t *testing.T) {
	tests := []struct {
		name      string
		state     *MergeableState
		mergeable bool
	}{
		{
			name:      "clean state",
			state:     &MergeableState{Mergeable: true, State: "clean"},
			mergeable: true,
		},
		{
			name:      "blocked state",
			state:     &MergeableState{Mergeable: false, State: "blocked"},
			mergeable: false,
		},
		{
			name:      "behind state",
			state:     &MergeableState{Mergeable: false, State: "behind"},
			mergeable: false,
		},
		{
			name:      "dirty state",
			state:     &MergeableState{Mergeable: false, State: "dirty"},
			mergeable: false,
		},
		{
			name:      "unstable state",
			state:     &MergeableState{Mergeable: true, State: "unstable"},
			mergeable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.state.Mergeable != tt.mergeable {
				t.Errorf("Mergeable = %v, want %v", tt.state.Mergeable, tt.mergeable)
			}
		})
	}
}

func TestReviewEventTypeConversion(t *testing.T) {
	// Test that ReviewEvent can be used as a string
	event := ReviewEventApprove
	var s string = string(event)
	if s != "APPROVE" {
		t.Errorf("string(ReviewEventApprove) = %q, want %q", s, "APPROVE")
	}
}
