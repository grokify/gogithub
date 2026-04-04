package search

import (
	"testing"
)

func TestQueryEncode(t *testing.T) {
	tests := []struct {
		name     string
		query    Query
		expected string
	}{
		{
			name:     "empty query",
			query:    Query{},
			expected: "",
		},
		{
			name:     "single parameter",
			query:    Query{"user": "grokify"},
			expected: "user:grokify",
		},
		{
			name:     "multiple parameters sorted",
			query:    Query{"user": "grokify", "state": "open", "is": "pr"},
			expected: "is:pr state:open user:grokify",
		},
		{
			name:     "parameters with special values",
			query:    Query{"repo": "owner/repo", "label": "bug"},
			expected: "label:bug repo:owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.query.Encode()
			if result != tt.expected {
				t.Errorf("Query.Encode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestQueryEncodeDeterministic(t *testing.T) {
	// Run multiple times to verify deterministic output
	query := Query{
		"user":  "grokify",
		"state": "open",
		"is":    "pr",
		"org":   "myorg",
		"type":  "issue",
	}
	expected := "is:pr org:myorg state:open type:issue user:grokify"

	for i := 0; i < 100; i++ {
		result := query.Encode()
		if result != expected {
			t.Errorf("Query.Encode() iteration %d = %q, want %q", i, result, expected)
		}
	}
}

func TestNewQuery(t *testing.T) {
	qb := NewQuery()
	if qb == nil {
		t.Fatal("NewQuery() returned nil")
	}
	if qb.q == nil {
		t.Fatal("NewQuery().q is nil")
	}
	if len(qb.q) != 0 {
		t.Errorf("NewQuery() should create empty query, got %d entries", len(qb.q))
	}
}

func TestQueryBuilderSet(t *testing.T) {
	qb := NewQuery().Set("custom", "value")

	if qb.q["custom"] != "value" {
		t.Errorf("Set() did not set value, got %q", qb.q["custom"])
	}
}

func TestQueryBuilderSetChaining(t *testing.T) {
	qb := NewQuery().
		Set("key1", "value1").
		Set("key2", "value2").
		Set("key3", "value3")

	if len(qb.q) != 3 {
		t.Errorf("Set() chaining: expected 3 entries, got %d", len(qb.q))
	}
}

func TestQueryBuilderBuild(t *testing.T) {
	qb := NewQuery().Set("user", "grokify")
	q := qb.Build()

	if q["user"] != "grokify" {
		t.Errorf("Build() did not return correct query")
	}

	// Verify Build() returns the same underlying map
	qb.Set("state", "open")
	if q["state"] != "open" {
		t.Errorf("Build() should return reference to underlying map")
	}
}

func TestQueryBuilderUser(t *testing.T) {
	q := NewQuery().User("grokify").Build()
	if q[ParamUser] != "grokify" {
		t.Errorf("User() = %q, want %q", q[ParamUser], "grokify")
	}
}

func TestQueryBuilderOrg(t *testing.T) {
	q := NewQuery().Org("myorg").Build()
	if q[ParamOrg] != "myorg" {
		t.Errorf("Org() = %q, want %q", q[ParamOrg], "myorg")
	}
}

func TestQueryBuilderRepo(t *testing.T) {
	q := NewQuery().Repo("owner/repo").Build()
	if q[ParamRepo] != "owner/repo" {
		t.Errorf("Repo() = %q, want %q", q[ParamRepo], "owner/repo")
	}
}

func TestQueryBuilderState(t *testing.T) {
	q := NewQuery().State("open").Build()
	if q[ParamState] != "open" {
		t.Errorf("State() = %q, want %q", q[ParamState], "open")
	}
}

func TestQueryBuilderStateOpen(t *testing.T) {
	q := NewQuery().StateOpen().Build()
	if q[ParamState] != ParamStateValueOpen {
		t.Errorf("StateOpen() = %q, want %q", q[ParamState], ParamStateValueOpen)
	}
}

func TestQueryBuilderStateClosed(t *testing.T) {
	q := NewQuery().StateClosed().Build()
	if q[ParamState] != ParamStateValueClosed {
		t.Errorf("StateClosed() = %q, want %q", q[ParamState], ParamStateValueClosed)
	}
}

func TestQueryBuilderIs(t *testing.T) {
	q := NewQuery().Is("merged").Build()
	if q[ParamIs] != "merged" {
		t.Errorf("Is() = %q, want %q", q[ParamIs], "merged")
	}
}

func TestQueryBuilderIsPR(t *testing.T) {
	q := NewQuery().IsPR().Build()
	if q[ParamIs] != ParamIsValuePR {
		t.Errorf("IsPR() = %q, want %q", q[ParamIs], ParamIsValuePR)
	}
}

func TestQueryBuilderIsIssue(t *testing.T) {
	q := NewQuery().IsIssue().Build()
	if q[ParamIs] != ParamIsValueIssue {
		t.Errorf("IsIssue() = %q, want %q", q[ParamIs], ParamIsValueIssue)
	}
}

func TestQueryBuilderType(t *testing.T) {
	q := NewQuery().Type("pr").Build()
	if q[ParamType] != "pr" {
		t.Errorf("Type() = %q, want %q", q[ParamType], "pr")
	}
}

func TestQueryBuilderAuthor(t *testing.T) {
	q := NewQuery().Author("grokify").Build()
	if q["author"] != "grokify" {
		t.Errorf("Author() = %q, want %q", q["author"], "grokify")
	}
}

func TestQueryBuilderAssignee(t *testing.T) {
	q := NewQuery().Assignee("grokify").Build()
	if q["assignee"] != "grokify" {
		t.Errorf("Assignee() = %q, want %q", q["assignee"], "grokify")
	}
}

func TestQueryBuilderLabel(t *testing.T) {
	q := NewQuery().Label("bug").Build()
	if q["label"] != "bug" {
		t.Errorf("Label() = %q, want %q", q["label"], "bug")
	}
}

func TestQueryBuilderMentions(t *testing.T) {
	q := NewQuery().Mentions("grokify").Build()
	if q["mentions"] != "grokify" {
		t.Errorf("Mentions() = %q, want %q", q["mentions"], "grokify")
	}
}

func TestQueryBuilderInvolves(t *testing.T) {
	q := NewQuery().Involves("grokify").Build()
	if q["involves"] != "grokify" {
		t.Errorf("Involves() = %q, want %q", q["involves"], "grokify")
	}
}

func TestQueryBuilderComplexChain(t *testing.T) {
	q := NewQuery().
		User("grokify").
		StateOpen().
		IsPR().
		Label("enhancement").
		Build()

	expected := Query{
		ParamUser:  "grokify",
		ParamState: ParamStateValueOpen,
		ParamIs:    ParamIsValuePR,
		"label":    "enhancement",
	}

	if len(q) != len(expected) {
		t.Errorf("Complex chain: got %d entries, want %d", len(q), len(expected))
	}

	for k, v := range expected {
		if q[k] != v {
			t.Errorf("Complex chain: q[%q] = %q, want %q", k, q[k], v)
		}
	}
}

func TestQueryBuilderEncodeIntegration(t *testing.T) {
	encoded := NewQuery().
		User("grokify").
		StateOpen().
		IsPR().
		Build().
		Encode()

	expected := "is:pr state:open user:grokify"
	if encoded != expected {
		t.Errorf("QueryBuilder -> Encode() = %q, want %q", encoded, expected)
	}
}

func TestQueryBuilderOverwrite(t *testing.T) {
	q := NewQuery().
		State("open").
		State("closed"). // Should overwrite
		Build()

	if q[ParamState] != "closed" {
		t.Errorf("Overwrite: State = %q, want %q", q[ParamState], "closed")
	}
}
