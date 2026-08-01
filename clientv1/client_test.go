package clientv1

import (
	"testing"

	"github.com/google/go-github/v89/github"
	"github.com/grokify/gogithub"
)

func TestUserFromGitHub(t *testing.T) {
	ghUser := &github.User{
		ID:        github.Ptr(int64(12345)),
		Login:     github.Ptr("testuser"),
		Name:      github.Ptr("Test User"),
		Email:     github.Ptr("test@example.com"),
		AvatarURL: github.Ptr("https://example.com/avatar.png"),
		HTMLURL:   github.Ptr("https://github.com/testuser"),
		Type:      github.Ptr("User"),
		Bio:       github.Ptr("A test user"),
		Company:   github.Ptr("Test Corp"),
		Location:  github.Ptr("Test City"),
		Blog:      github.Ptr("https://example.com"),
		Followers: github.Ptr(100),
		Following: github.Ptr(50),
	}

	user := userFromGitHub(ghUser)

	if user.ID != 12345 {
		t.Errorf("ID = %d, want 12345", user.ID)
	}
	if user.Login != "testuser" {
		t.Errorf("Login = %q, want %q", user.Login, "testuser")
	}
	if user.Name != "Test User" {
		t.Errorf("Name = %q, want %q", user.Name, "Test User")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "test@example.com")
	}
	if user.AvatarURL != "https://example.com/avatar.png" {
		t.Errorf("AvatarURL = %q, want %q", user.AvatarURL, "https://example.com/avatar.png")
	}
	if user.Type != "User" {
		t.Errorf("Type = %q, want %q", user.Type, "User")
	}
	if user.Followers != 100 {
		t.Errorf("Followers = %d, want 100", user.Followers)
	}
}

func TestUserFromGitHubNil(t *testing.T) {
	user := userFromGitHub(nil)
	if user != nil {
		t.Error("userFromGitHub(nil) should return nil")
	}
}

func TestRepositoryFromGitHub(t *testing.T) {
	ghRepo := &github.Repository{
		ID:            github.Ptr(int64(67890)),
		Name:          github.Ptr("testrepo"),
		FullName:      github.Ptr("testuser/testrepo"),
		Description:   github.Ptr("A test repository"),
		HTMLURL:       github.Ptr("https://github.com/testuser/testrepo"),
		CloneURL:      github.Ptr("https://github.com/testuser/testrepo.git"),
		SSHURL:        github.Ptr("git@github.com:testuser/testrepo.git"),
		DefaultBranch: github.Ptr("main"),
		Private:       github.Ptr(false),
		Fork:          github.Ptr(false),
		Archived:      github.Ptr(false),
		Language:      github.Ptr("Go"),
		ForksCount:    github.Ptr(10),
		Owner: &github.User{
			ID:    github.Ptr(int64(12345)),
			Login: github.Ptr("testuser"),
		},
	}

	repo := repositoryFromGitHub(ghRepo)

	if repo.ID != 67890 {
		t.Errorf("ID = %d, want 67890", repo.ID)
	}
	if repo.Name != "testrepo" {
		t.Errorf("Name = %q, want %q", repo.Name, "testrepo")
	}
	if repo.FullName != "testuser/testrepo" {
		t.Errorf("FullName = %q, want %q", repo.FullName, "testuser/testrepo")
	}
	if repo.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", repo.DefaultBranch, "main")
	}
	if repo.Private {
		t.Error("Private should be false")
	}
	if repo.Owner == nil {
		t.Error("Owner should not be nil")
	} else if repo.Owner.Login != "testuser" {
		t.Errorf("Owner.Login = %q, want %q", repo.Owner.Login, "testuser")
	}
}

func TestRepositoryFromGitHubNil(t *testing.T) {
	repo := repositoryFromGitHub(nil)
	if repo != nil {
		t.Error("repositoryFromGitHub(nil) should return nil")
	}
}

func TestReferenceFromGitHub(t *testing.T) {
	ghRef := &github.Reference{
		Ref: github.Ptr("refs/heads/main"),
		URL: github.Ptr("https://api.github.com/repos/owner/repo/git/refs/heads/main"),
		Object: &github.GitObject{
			Type: github.Ptr("commit"),
			SHA:  github.Ptr("abc123def456"),
			URL:  github.Ptr("https://api.github.com/repos/owner/repo/git/commits/abc123def456"),
		},
	}

	ref := referenceFromGitHub(ghRef)

	if ref.Ref != "refs/heads/main" {
		t.Errorf("Ref = %q, want %q", ref.Ref, "refs/heads/main")
	}
	if ref.SHA != "abc123def456" {
		t.Errorf("SHA = %q, want %q", ref.SHA, "abc123def456")
	}
	if ref.Object == nil {
		t.Error("Object should not be nil")
	} else {
		if ref.Object.Type != "commit" {
			t.Errorf("Object.Type = %q, want %q", ref.Object.Type, "commit")
		}
		if ref.Object.SHA != "abc123def456" {
			t.Errorf("Object.SHA = %q, want %q", ref.Object.SHA, "abc123def456")
		}
	}
}

func TestNewClientFromRaw(t *testing.T) {
	ghClient, err := github.NewClient()
	if err != nil {
		t.Fatalf("github.NewClient() error = %v", err)
	}
	client := NewClientFromRaw(ghClient)

	if client == nil {
		t.Fatal("NewClientFromRaw returned nil")
	}

	raw := client.Raw()
	if raw == nil {
		t.Fatal("Raw() returned nil")
	}
	if _, ok := raw.(*github.Client); !ok {
		t.Errorf("Raw() returned %T, want *github.Client", raw)
	}
}

func TestClientInterfaceImplementation(t *testing.T) {
	var _ Client = &client{}
}

func TestTypesAreFromRootPackage(t *testing.T) {
	// Verify that the types returned are from the gogithub package
	var _ *gogithub.User = userFromGitHub(nil)
	var _ *gogithub.Repository = repositoryFromGitHub(nil)
	var _ *gogithub.Reference = referenceFromGitHub(nil)
}
