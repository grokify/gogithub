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
