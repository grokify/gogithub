package repo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
	ghErrors "github.com/grokify/gogithub/errors"
)

// AccessStatus describes the outcome of CheckAccess.
type AccessStatus string

const (
	// AccessGranted means the authenticated user can access the repository.
	AccessGranted AccessStatus = "granted"
	// AccessNotVisible means the repository does not exist or the
	// authenticated user cannot access it. GitHub does not distinguish the
	// two for private repositories.
	AccessNotVisible AccessStatus = "not_visible"
	// AccessBlockedByTokenPolicy means the repository owner rejects the type
	// of token used, so access cannot be determined with this token. The
	// user may still have access through another credential.
	AccessBlockedByTokenPolicy AccessStatus = "blocked_by_token_policy"
)

// AccessResult is the authenticated user's access to a single repository.
type AccessResult struct {
	Status AccessStatus
	// Repository is set when Status is AccessGranted.
	Repository *gogithub.Repository
	// Permission is the highest permission level granted, when Status is
	// AccessGranted and GitHub reported permissions.
	Permission string
	// Detail is the message GitHub returned, when Status is
	// AccessBlockedByTokenPolicy.
	Detail string
}

// CheckAccess reports the authenticated user's access to a repository.
//
// A private repository owned by an organization whose policy rejects the
// token returns a plain 404, indistinguishable from a repository the user
// cannot access. CheckAccess therefore probes the owner on a 404 and reports
// AccessBlockedByTokenPolicy rather than AccessNotVisible when the owner
// rejects the token.
func CheckAccess(ctx context.Context, client clientv1.Client, owner, name string) (*AccessResult, error) {
	r, err := client.GetRepository(ctx, owner, name)
	if err == nil {
		return &AccessResult{
			Status:     AccessGranted,
			Repository: r,
			Permission: r.Permissions.Highest(),
		}, nil
	}
	if ghErrors.IsTokenPolicyError(err) {
		return &AccessResult{Status: AccessBlockedByTokenPolicy, Detail: ghErrors.Message(err)}, nil
	}
	if ghErrors.StatusCode(err) != http.StatusNotFound {
		return nil, fmt.Errorf("get repository %s/%s: %w", owner, name, err)
	}

	_, err = client.GetUser(ctx, owner)
	switch {
	case err == nil, ghErrors.StatusCode(err) == http.StatusNotFound:
		return &AccessResult{Status: AccessNotVisible}, nil
	case ghErrors.IsTokenPolicyError(err):
		return &AccessResult{Status: AccessBlockedByTokenPolicy, Detail: ghErrors.Message(err)}, nil
	}
	return nil, fmt.Errorf("get owner %s: %w", owner, err)
}

// ListNonMemberOrgRepos lists repositories the authenticated user has been
// granted access to that are owned by organizations the user is not an
// active member of (e.g. outside-collaborator grants).
//
// Repositories are listed across all affiliations and then filtered by
// owner, so a repository granted directly in an organization the user also
// belongs to is excluded.
func ListNonMemberOrgRepos(ctx context.Context, client clientv1.Client) ([]*gogithub.Repository, error) {
	memberships, err := client.ListOrgMemberships(ctx, &clientv1.ListOrgMembershipsOptions{
		State: gogithub.MembershipStateActive,
	})
	if err != nil {
		return nil, err
	}
	repos, err := client.ListAuthenticatedUserRepos(ctx, nil)
	if err != nil {
		return nil, err
	}
	return FilterNonMemberOrgRepos(repos, MembershipOrgLogins(memberships)), nil
}

// MembershipOrgLogins returns the organization logins for the given memberships.
func MembershipOrgLogins(memberships []*gogithub.OrgMembership) []string {
	logins := make([]string, 0, len(memberships))
	for _, m := range memberships {
		if m == nil || m.Organization == nil || m.Organization.Login == "" {
			continue
		}
		logins = append(logins, m.Organization.Login)
	}
	return logins
}

// FilterNonMemberOrgRepos returns the repositories owned by an organization
// whose login is not in memberOrgs. Repositories owned by user accounts are
// excluded. Logins are compared case-insensitively.
func FilterNonMemberOrgRepos(repos []*gogithub.Repository, memberOrgs []string) []*gogithub.Repository {
	member := make(map[string]struct{}, len(memberOrgs))
	for _, org := range memberOrgs {
		member[strings.ToLower(org)] = struct{}{}
	}
	var filtered []*gogithub.Repository
	for _, r := range repos {
		if r == nil || r.Owner == nil || r.Owner.Type != gogithub.OwnerTypeOrganization {
			continue
		}
		if _, ok := member[strings.ToLower(r.Owner.Login)]; ok {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}
