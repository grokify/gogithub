package search

import "strings"

// Query represents a GitHub search query.
type Query map[string]string

// Encode implements GitHub API query encoding.
func (q Query) Encode() string {
	var parts []string
	for k, v := range q {
		parts = append(parts, k+":"+v)
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
