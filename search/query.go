package search

import (
	"sort"
	"strings"
)

// Query represents a GitHub search query.
type Query map[string]string

// Encode implements GitHub API query encoding.
// Keys are sorted to ensure deterministic output.
func (q Query) Encode() string {
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(q))
	for _, k := range keys {
		parts = append(parts, k+":"+q[k])
	}
	return strings.Join(parts, " ")
}

// QueryBuilder provides a fluent interface for constructing search queries.
// It wraps the Query map type, providing type-safe methods for common qualifiers
// while preserving flexibility through the Set() method for any qualifier.
type QueryBuilder struct {
	q Query
}

// NewQuery creates a new QueryBuilder for constructing search queries.
func NewQuery() *QueryBuilder {
	return &QueryBuilder{q: make(Query)}
}

// Set adds any key-value pair to the query. Use this for qualifiers not covered
// by the typed methods, or for new GitHub search qualifiers.
func (qb *QueryBuilder) Set(key, value string) *QueryBuilder {
	qb.q[key] = value
	return qb
}

// Build returns the constructed Query map.
func (qb *QueryBuilder) Build() Query {
	return qb.q
}

// User filters by repository owner username.
func (qb *QueryBuilder) User(username string) *QueryBuilder {
	qb.q[ParamUser] = username
	return qb
}

// Org filters by organization.
func (qb *QueryBuilder) Org(org string) *QueryBuilder {
	qb.q[ParamOrg] = org
	return qb
}

// Repo filters by specific repository (owner/repo format).
func (qb *QueryBuilder) Repo(repo string) *QueryBuilder {
	qb.q[ParamRepo] = repo
	return qb
}

// State filters by issue/PR state (open, closed).
func (qb *QueryBuilder) State(state string) *QueryBuilder {
	qb.q[ParamState] = state
	return qb
}

// StateOpen filters for open issues/PRs.
func (qb *QueryBuilder) StateOpen() *QueryBuilder {
	return qb.State(ParamStateValueOpen)
}

// StateClosed filters for closed issues/PRs.
func (qb *QueryBuilder) StateClosed() *QueryBuilder {
	return qb.State(ParamStateValueClosed)
}

// Is filters by type or state (pr, issue, open, closed, merged, etc.).
func (qb *QueryBuilder) Is(value string) *QueryBuilder {
	qb.q[ParamIs] = value
	return qb
}

// IsPR filters for pull requests only.
func (qb *QueryBuilder) IsPR() *QueryBuilder {
	return qb.Is(ParamIsValuePR)
}

// IsIssue filters for issues only.
func (qb *QueryBuilder) IsIssue() *QueryBuilder {
	return qb.Is(ParamIsValueIssue)
}

// Type filters by type (pr, issue).
func (qb *QueryBuilder) Type(typ string) *QueryBuilder {
	qb.q[ParamType] = typ
	return qb
}

// Author filters by item creator.
func (qb *QueryBuilder) Author(username string) *QueryBuilder {
	qb.q["author"] = username
	return qb
}

// Assignee filters by assigned user.
func (qb *QueryBuilder) Assignee(username string) *QueryBuilder {
	qb.q["assignee"] = username
	return qb
}

// Label filters by label.
func (qb *QueryBuilder) Label(label string) *QueryBuilder {
	qb.q["label"] = label
	return qb
}

// Mentions filters by mentioned user.
func (qb *QueryBuilder) Mentions(username string) *QueryBuilder {
	qb.q["mentions"] = username
	return qb
}

// Involves filters by user involvement (author, assignee, mentions, or commenter).
func (qb *QueryBuilder) Involves(username string) *QueryBuilder {
	qb.q["involves"] = username
	return qb
}

// Query parameter names.
const (
	ParamUser  = "user"
	ParamState = "state"
	ParamIs    = "is"
	ParamOrg   = "org"
	ParamRepo  = "repo"
	ParamType  = "type"
)

// Query parameter values.
const (
	ParamStateValueOpen   = "open"
	ParamStateValueClosed = "closed"
	ParamIsValuePR        = "pr"
	ParamIsValueIssue     = "issue"
)

// Pagination constants.
const (
	ParamPerPageValueMax     = 100
	ParamPerPageValueDefault = 30
)
