package clientv1

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v89/github"
	"github.com/grokify/gogithub"
)

// rewriteTransport sends every request to the test server, preserving the
// request path and query.
type rewriteTransport struct {
	target *url.URL
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

// newTestClient returns a Client whose requests are served by handler.
func newTestClient(t *testing.T, handler http.Handler) Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	c, err := NewClientWithHTTP(&http.Client{Transport: &rewriteTransport{target: target}})
	if err != nil {
		t.Fatalf("NewClientWithHTTP() error = %v", err)
	}
	return c
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if _, err := fmt.Fprint(w, body); err != nil {
		t.Errorf("write response: %v", err)
	}
}

func TestListAuthenticatedUserRepos(t *testing.T) {
	var queries []url.Values
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/repos" {
			t.Errorf("path = %q, want /user/repos", r.URL.Path)
		}
		q := r.URL.Query()
		queries = append(queries, q)
		if q.Get("page") == "2" {
			writeJSON(t, w, `[{"id":2,"name":"two","full_name":"otheruser/two",
				"owner":{"login":"otheruser","type":"User"},
				"permissions":{"admin":false,"push":false,"pull":true}}]`)
			return
		}
		w.Header().Set("Link", `<https://api.github.com/user/repos?page=2>; rel="next"`)
		writeJSON(t, w, `[{"id":1,"name":"one","full_name":"externalorg/one","private":true,
			"owner":{"login":"externalorg","type":"Organization"},
			"permissions":{"admin":false,"maintain":false,"push":true,"triage":true,"pull":true}}]`)
	}))

	repos, err := c.ListAuthenticatedUserRepos(context.Background(), &ListAuthenticatedUserReposOptions{
		Visibility:   "private",
		Affiliations: []string{AffiliationCollaborator, AffiliationOrganizationMember},
	})
	if err != nil {
		t.Fatalf("ListAuthenticatedUserRepos() error = %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("len(repos) = %d, want 2", len(repos))
	}
	if len(queries) != 2 {
		t.Fatalf("requests = %d, want 2", len(queries))
	}
	if got, want := queries[0].Get("affiliation"), "collaborator,organization_member"; got != want {
		t.Errorf("affiliation = %q, want %q", got, want)
	}
	if got, want := queries[0].Get("visibility"), "private"; got != want {
		t.Errorf("visibility = %q, want %q", got, want)
	}
	if got, want := queries[0].Get("per_page"), "100"; got != want {
		t.Errorf("per_page = %q, want %q", got, want)
	}

	first := repos[0]
	if first.FullName != "externalorg/one" {
		t.Errorf("FullName = %q, want %q", first.FullName, "externalorg/one")
	}
	if first.Owner == nil || first.Owner.Type != gogithub.OwnerTypeOrganization {
		t.Errorf("Owner = %+v, want type %q", first.Owner, gogithub.OwnerTypeOrganization)
	}
	if got := first.Permissions.Highest(); got != gogithub.PermissionPush {
		t.Errorf("Permissions.Highest() = %q, want %q", got, gogithub.PermissionPush)
	}
	if got := repos[1].Permissions.Highest(); got != gogithub.PermissionPull {
		t.Errorf("Permissions.Highest() = %q, want %q", got, gogithub.PermissionPull)
	}
}

func TestListAuthenticatedUserReposNilOptions(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		for _, param := range []string{"affiliation", "visibility", "type"} {
			if q.Has(param) {
				t.Errorf("query has %q = %q, want unset", param, q.Get(param))
			}
		}
		writeJSON(t, w, `[]`)
	}))

	repos, err := c.ListAuthenticatedUserRepos(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListAuthenticatedUserRepos() error = %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("len(repos) = %d, want 0", len(repos))
	}
}

func TestListAuthenticatedUserReposError(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(t, w, `{"message":"Bad credentials"}`)
	}))

	if _, err := c.ListAuthenticatedUserRepos(context.Background(), nil); err == nil {
		t.Fatal("ListAuthenticatedUserRepos() error = nil, want error")
	}
}

func TestListOrgMemberships(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/memberships/orgs" {
			t.Errorf("path = %q, want /user/memberships/orgs", r.URL.Path)
		}
		if got, want := r.URL.Query().Get("state"), gogithub.MembershipStateActive; got != want {
			t.Errorf("state = %q, want %q", got, want)
		}
		writeJSON(t, w, `[
			{"state":"active","role":"admin","organization":{"id":10,"login":"myorg"}},
			{"state":"active","role":"member","organization":{"id":11,"login":"otherorg"}}]`)
	}))

	memberships, err := c.ListOrgMemberships(context.Background(), &ListOrgMembershipsOptions{
		State: gogithub.MembershipStateActive,
	})
	if err != nil {
		t.Fatalf("ListOrgMemberships() error = %v", err)
	}

	if len(memberships) != 2 {
		t.Fatalf("len(memberships) = %d, want 2", len(memberships))
	}
	m := memberships[0]
	if m.Organization == nil || m.Organization.Login != "myorg" {
		t.Errorf("Organization = %+v, want login %q", m.Organization, "myorg")
	}
	if m.Organization != nil && m.Organization.Type != gogithub.OwnerTypeOrganization {
		t.Errorf("Organization.Type = %q, want %q", m.Organization.Type, gogithub.OwnerTypeOrganization)
	}
	if m.Role != "admin" {
		t.Errorf("Role = %q, want %q", m.Role, "admin")
	}
	if m.State != gogithub.MembershipStateActive {
		t.Errorf("State = %q, want %q", m.State, gogithub.MembershipStateActive)
	}
}

func TestListOrgMembershipsError(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(t, w, `{"message":"Resource not accessible by personal access token"}`)
	}))

	if _, err := c.ListOrgMemberships(context.Background(), nil); err == nil {
		t.Fatal("ListOrgMemberships() error = nil, want error")
	}
}

func TestRepositoryPermissionsFromGitHub(t *testing.T) {
	if got := repositoryPermissionsFromGitHub(nil); got != nil {
		t.Errorf("repositoryPermissionsFromGitHub(nil) = %+v, want nil", got)
	}

	got := repositoryPermissionsFromGitHub(&github.RepositoryPermissions{
		Admin: github.Ptr(true),
		Pull:  github.Ptr(true),
	})
	want := gogithub.RepositoryPermissions{Admin: true, Pull: true}
	if got == nil || *got != want {
		t.Errorf("repositoryPermissionsFromGitHub() = %+v, want %+v", got, want)
	}
}

func TestOrgMembershipFromGitHubNil(t *testing.T) {
	if got := orgMembershipFromGitHub(nil); got != nil {
		t.Errorf("orgMembershipFromGitHub(nil) = %+v, want nil", got)
	}
}
