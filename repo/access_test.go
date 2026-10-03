package repo

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
	ghErrors "github.com/grokify/gogithub/errors"
)

// accessClient stubs the clientv1.Client methods used by the access helpers.
// Calling any other method panics on the nil embedded interface.
type accessClient struct {
	clientv1.Client

	repos          []*gogithub.Repository
	reposErr       error
	memberships    []*gogithub.OrgMembership
	membershipsErr error
	repo           *gogithub.Repository
	repoErr        error
	ownerErr       error
	ownerLookups   int

	gotReposOpts       *clientv1.ListAuthenticatedUserReposOptions
	gotMembershipsOpts *clientv1.ListOrgMembershipsOptions
}

func (c *accessClient) ListAuthenticatedUserRepos(_ context.Context, opts *clientv1.ListAuthenticatedUserReposOptions) ([]*gogithub.Repository, error) {
	c.gotReposOpts = opts
	return c.repos, c.reposErr
}

func (c *accessClient) ListOrgMemberships(_ context.Context, opts *clientv1.ListOrgMembershipsOptions) ([]*gogithub.OrgMembership, error) {
	c.gotMembershipsOpts = opts
	return c.memberships, c.membershipsErr
}

func (c *accessClient) GetRepository(_ context.Context, _, _ string) (*gogithub.Repository, error) {
	return c.repo, c.repoErr
}

func (c *accessClient) GetUser(_ context.Context, login string) (*gogithub.User, error) {
	c.ownerLookups++
	if c.ownerErr != nil {
		return nil, c.ownerErr
	}
	return &gogithub.User{Login: login}, nil
}

func testRepo(owner, ownerType, name string) *gogithub.Repository {
	return &gogithub.Repository{
		Name:     name,
		FullName: owner + "/" + name,
		Owner:    &gogithub.User{Login: owner, Type: ownerType},
	}
}

func testMembership(org string) *gogithub.OrgMembership {
	return &gogithub.OrgMembership{
		Organization: &gogithub.User{Login: org, Type: gogithub.OwnerTypeOrganization},
		Role:         "member",
		State:        gogithub.MembershipStateActive,
	}
}

func fullNames(repos []*gogithub.Repository) []string {
	names := make([]string, 0, len(repos))
	for _, r := range repos {
		names = append(names, r.FullName)
	}
	return names
}

func TestFilterNonMemberOrgRepos(t *testing.T) {
	repos := []*gogithub.Repository{
		testRepo("me", gogithub.OwnerTypeUser, "mine"),
		testRepo("MyOrg", gogithub.OwnerTypeOrganization, "member-repo"),
		testRepo("externalorg", gogithub.OwnerTypeOrganization, "granted"),
		testRepo("otheruser", gogithub.OwnerTypeUser, "personal-collab"),
		testRepo("anotherorg", gogithub.OwnerTypeOrganization, "also-granted"),
		nil,
		{FullName: "noowner/repo"},
	}

	tests := []struct {
		name       string
		memberOrgs []string
		want       []string
	}{
		{
			name:       "excludes member orgs case-insensitively",
			memberOrgs: []string{"myorg"},
			want:       []string{"externalorg/granted", "anotherorg/also-granted"},
		},
		{
			name:       "no memberships keeps all org repos",
			memberOrgs: nil,
			want:       []string{"MyOrg/member-repo", "externalorg/granted", "anotherorg/also-granted"},
		},
		{
			name:       "member of every org",
			memberOrgs: []string{"MYORG", "ExternalOrg", "anotherorg"},
			want:       []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fullNames(FilterNonMemberOrgRepos(repos, tt.memberOrgs))
			if !slices.Equal(got, tt.want) {
				t.Errorf("FilterNonMemberOrgRepos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembershipOrgLogins(t *testing.T) {
	got := MembershipOrgLogins([]*gogithub.OrgMembership{
		testMembership("myorg"),
		nil,
		{Role: "member"},
		{Organization: &gogithub.User{}},
		testMembership("otherorg"),
	})
	want := []string{"myorg", "otherorg"}
	if !slices.Equal(got, want) {
		t.Errorf("MembershipOrgLogins() = %v, want %v", got, want)
	}
}

func TestListNonMemberOrgRepos(t *testing.T) {
	client := &accessClient{
		memberships: []*gogithub.OrgMembership{testMembership("myorg")},
		repos: []*gogithub.Repository{
			testRepo("me", gogithub.OwnerTypeUser, "mine"),
			testRepo("myorg", gogithub.OwnerTypeOrganization, "member-repo"),
			testRepo("externalorg", gogithub.OwnerTypeOrganization, "granted"),
		},
	}

	repos, err := ListNonMemberOrgRepos(context.Background(), client)
	if err != nil {
		t.Fatalf("ListNonMemberOrgRepos() error = %v", err)
	}

	want := []string{"externalorg/granted"}
	if got := fullNames(repos); !slices.Equal(got, want) {
		t.Errorf("ListNonMemberOrgRepos() = %v, want %v", got, want)
	}
	if client.gotMembershipsOpts == nil || client.gotMembershipsOpts.State != gogithub.MembershipStateActive {
		t.Errorf("memberships opts = %+v, want state %q", client.gotMembershipsOpts, gogithub.MembershipStateActive)
	}
	if client.gotReposOpts != nil {
		t.Errorf("repos opts = %+v, want nil (all affiliations)", client.gotReposOpts)
	}
}

func TestListNonMemberOrgReposErrors(t *testing.T) {
	errMemberships := errors.New("memberships failed")
	errRepos := errors.New("repos failed")

	tests := []struct {
		name   string
		client *accessClient
		want   error
	}{
		{"memberships error", &accessClient{membershipsErr: errMemberships}, errMemberships},
		{"repos error", &accessClient{reposErr: errRepos}, errRepos},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ListNonMemberOrgRepos(context.Background(), tt.client)
			if !errors.Is(err, tt.want) {
				t.Errorf("ListNonMemberOrgRepos() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCheckAccess(t *testing.T) {
	const policyMessage = "`exampleorg` forbids access via a personal access token (classic)."

	errNotFound := &ghErrors.APIError{StatusCode: 404, Err: ghErrors.ErrNotFound}
	errPolicy := &ghErrors.APIError{StatusCode: 403, Message: policyMessage, Err: ghErrors.ErrPermissionDenied}
	errForbidden := &ghErrors.APIError{StatusCode: 403, Message: "Forbidden", Err: ghErrors.ErrPermissionDenied}

	granted := testRepo("exampleorg", gogithub.OwnerTypeOrganization, "granted")
	granted.Permissions = &gogithub.RepositoryPermissions{Push: true, Pull: true}

	tests := []struct {
		name             string
		client           *accessClient
		wantStatus       AccessStatus
		wantPermission   string
		wantDetail       string
		wantOwnerLookups int
		wantErr          error
	}{
		{
			name:           "granted",
			client:         &accessClient{repo: granted},
			wantStatus:     AccessGranted,
			wantPermission: gogithub.PermissionPush,
		},
		{
			name:       "granted without permissions reported",
			client:     &accessClient{repo: testRepo("exampleorg", gogithub.OwnerTypeOrganization, "public")},
			wantStatus: AccessGranted,
		},
		{
			name:             "not visible when owner is reachable",
			client:           &accessClient{repoErr: errNotFound},
			wantStatus:       AccessNotVisible,
			wantOwnerLookups: 1,
		},
		{
			name:             "not visible when owner does not exist",
			client:           &accessClient{repoErr: errNotFound, ownerErr: errNotFound},
			wantStatus:       AccessNotVisible,
			wantOwnerLookups: 1,
		},
		{
			name:             "blocked when owner rejects the token",
			client:           &accessClient{repoErr: errNotFound, ownerErr: errPolicy},
			wantStatus:       AccessBlockedByTokenPolicy,
			wantDetail:       policyMessage,
			wantOwnerLookups: 1,
		},
		{
			name:       "blocked when repository request rejects the token",
			client:     &accessClient{repoErr: errPolicy},
			wantStatus: AccessBlockedByTokenPolicy,
			wantDetail: policyMessage,
		},
		{
			name:    "repository error",
			client:  &accessClient{repoErr: errForbidden},
			wantErr: errForbidden,
		},
		{
			name:             "owner error",
			client:           &accessClient{repoErr: errNotFound, ownerErr: errForbidden},
			wantOwnerLookups: 1,
			wantErr:          errForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckAccess(context.Background(), tt.client, "exampleorg", "repo")
			if tt.client.ownerLookups != tt.wantOwnerLookups {
				t.Errorf("owner lookups = %d, want %d", tt.client.ownerLookups, tt.wantOwnerLookups)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CheckAccess() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckAccess() error = %v", err)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.Permission != tt.wantPermission {
				t.Errorf("Permission = %q, want %q", got.Permission, tt.wantPermission)
			}
			if got.Detail != tt.wantDetail {
				t.Errorf("Detail = %q, want %q", got.Detail, tt.wantDetail)
			}
			if (got.Repository != nil) != (tt.wantStatus == AccessGranted) {
				t.Errorf("Repository = %+v, want set only when granted", got.Repository)
			}
		})
	}
}
