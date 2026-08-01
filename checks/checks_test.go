package checks

import (
	"testing"

	"github.com/grokify/gogithub"
)

func TestGetChecksStatus(t *testing.T) {
	tests := []struct {
		name       string
		checks     []*gogithub.CheckRun
		wantTotal  int
		wantPassed int
		wantFailed int
		wantPend   int
		wantAll    bool
		wantAnyF   bool
		wantAnyP   bool
	}{
		{
			name:       "empty checks",
			checks:     []*gogithub.CheckRun{},
			wantTotal:  0,
			wantPassed: 0,
			wantFailed: 0,
			wantPend:   0,
			wantAll:    false,
			wantAnyF:   false,
			wantAnyP:   false,
		},
		{
			name: "all passed",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "success"},
			},
			wantTotal:  2,
			wantPassed: 2,
			wantFailed: 0,
			wantPend:   0,
			wantAll:    true,
			wantAnyF:   false,
			wantAnyP:   false,
		},
		{
			name: "some failed",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "failure"},
			},
			wantTotal:  2,
			wantPassed: 1,
			wantFailed: 1,
			wantPend:   0,
			wantAll:    false,
			wantAnyF:   true,
			wantAnyP:   false,
		},
		{
			name: "some pending",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "success"},
				{Status: "in_progress", Conclusion: ""},
			},
			wantTotal:  2,
			wantPassed: 1,
			wantFailed: 0,
			wantPend:   1,
			wantAll:    false,
			wantAnyF:   false,
			wantAnyP:   true,
		},
		{
			name: "all pending",
			checks: []*gogithub.CheckRun{
				{Status: "queued", Conclusion: ""},
				{Status: "in_progress", Conclusion: ""},
			},
			wantTotal:  2,
			wantPassed: 0,
			wantFailed: 0,
			wantPend:   2,
			wantAll:    false,
			wantAnyF:   false,
			wantAnyP:   true,
		},
		{
			name: "mixed status",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "failure"},
				{Status: "in_progress", Conclusion: ""},
			},
			wantTotal:  3,
			wantPassed: 1,
			wantFailed: 1,
			wantPend:   1,
			wantAll:    false,
			wantAnyF:   true,
			wantAnyP:   true,
		},
		{
			name: "cancelled counts as failed",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "cancelled"},
			},
			wantTotal:  1,
			wantPassed: 0,
			wantFailed: 1,
			wantPend:   0,
			wantAll:    false,
			wantAnyF:   true,
			wantAnyP:   false,
		},
		{
			name: "skipped counts as failed",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "skipped"},
			},
			wantTotal:  1,
			wantPassed: 0,
			wantFailed: 1,
			wantPend:   0,
			wantAll:    false,
			wantAnyF:   true,
			wantAnyP:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := GetChecksStatus(tt.checks)

			if status.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", status.Total, tt.wantTotal)
			}
			if status.Passed != tt.wantPassed {
				t.Errorf("Passed = %d, want %d", status.Passed, tt.wantPassed)
			}
			if status.Failed != tt.wantFailed {
				t.Errorf("Failed = %d, want %d", status.Failed, tt.wantFailed)
			}
			if status.Pending != tt.wantPend {
				t.Errorf("Pending = %d, want %d", status.Pending, tt.wantPend)
			}
			if status.AllPassed != tt.wantAll {
				t.Errorf("AllPassed = %v, want %v", status.AllPassed, tt.wantAll)
			}
			if status.AnyFailed != tt.wantAnyF {
				t.Errorf("AnyFailed = %v, want %v", status.AnyFailed, tt.wantAnyF)
			}
			if status.AnyPending != tt.wantAnyP {
				t.Errorf("AnyPending = %v, want %v", status.AnyPending, tt.wantAnyP)
			}
		})
	}
}

func TestAllChecksPassed(t *testing.T) {
	tests := []struct {
		name     string
		checks   []*gogithub.CheckRun
		expected bool
	}{
		{
			name:     "empty checks",
			checks:   []*gogithub.CheckRun{},
			expected: false,
		},
		{
			name: "all passed",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "success"},
			},
			expected: true,
		},
		{
			name: "one passed",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "success"},
			},
			expected: true,
		},
		{
			name: "one failed",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "failure"},
			},
			expected: false,
		},
		{
			name: "one pending",
			checks: []*gogithub.CheckRun{
				{Status: "in_progress", Conclusion: ""},
			},
			expected: false,
		},
		{
			name: "mixed passed and failed",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "failure"},
			},
			expected: false,
		},
		{
			name: "passed with pending",
			checks: []*gogithub.CheckRun{
				{Status: "completed", Conclusion: "success"},
				{Status: "queued", Conclusion: ""},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AllChecksPassed(tt.checks)
			if result != tt.expected {
				t.Errorf("AllChecksPassed() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChecksStatusStruct(t *testing.T) {
	status := &ChecksStatus{
		Total:      5,
		Passed:     3,
		Failed:     1,
		Pending:    1,
		AllPassed:  false,
		AnyFailed:  true,
		AnyPending: true,
	}

	if status.Total != 5 {
		t.Errorf("Total = %d, want %d", status.Total, 5)
	}
	if status.Passed != 3 {
		t.Errorf("Passed = %d, want %d", status.Passed, 3)
	}
	if status.Failed != 1 {
		t.Errorf("Failed = %d, want %d", status.Failed, 1)
	}
	if status.Pending != 1 {
		t.Errorf("Pending = %d, want %d", status.Pending, 1)
	}
}
