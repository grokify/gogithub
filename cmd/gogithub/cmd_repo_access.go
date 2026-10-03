package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/auth/credentialsset"
	"github.com/grokify/gogithub/clientv1"
	"github.com/grokify/gogithub/repo"
	"github.com/spf13/cobra"
)

const (
	repoAccessFormatText = "text"
	repoAccessFormatJSON = "json"
)

var (
	repoAccessNonMemberOrgs bool
	repoAccessFormat        string
	repoAccessRepos         []string
	repoAccessCreds         string
	repoAccessAccount       string
)

var repoAccessCmd = &cobra.Command{
	Use:   "repo-access",
	Short: "List repositories the authenticated user can access",
	Long: `List repositories the authenticated user has access to, with the
permission level granted on each.

Use --non-member-orgs to show only repositories owned by organizations the
user is not a member of, such as outside-collaborator grants.

Use --repo to check specific repositories instead of listing. A repository
whose owner rejects the token is reported as blocked_by_token_policy rather
than not_visible, since access cannot be determined with that token.

Examples:
  gogithub repo-access                              # All accessible repositories
  gogithub repo-access --non-member-orgs            # Grants in organizations you don't belong to
  gogithub repo-access --non-member-orgs -f json    # JSON output
  gogithub repo-access --repo owner/name            # Check one repository

  # Authenticate as an OAuth app from a goauth credentials set file
  gogithub repo-access --creds credentials.json --account github --non-member-orgs

Authentication:
  GITHUB_TOKEN    Classic personal access token or OAuth token with the 'repo'
                  and 'read:org' scopes. Fine-grained tokens are bound to one
                  resource owner and cannot list access across organizations.

  --creds and --account select an OAuth app from a goauth credentials set
  file instead. Organizations that forbid personal access tokens accept an
  OAuth app token, provided they allow the OAuth app. Without a stored token,
  the authorization code grant is used: open the printed URL, authorize, and
  enter the code shown on the redirect page.`,
	RunE: runRepoAccess,
}

func init() {
	repoAccessCmd.Flags().BoolVar(&repoAccessNonMemberOrgs, "non-member-orgs", false, "Only repositories in organizations the user is not a member of")
	repoAccessCmd.Flags().StringVarP(&repoAccessFormat, "format", "f", repoAccessFormatText, "Output format: text or json")
	repoAccessCmd.Flags().StringSliceVar(&repoAccessRepos, "repo", nil, "Check access to specific repositories (owner/name) instead of listing")
	repoAccessCmd.Flags().StringVar(&repoAccessCreds, "creds", "", "goauth credentials set file")
	repoAccessCmd.Flags().StringVar(&repoAccessAccount, "account", "", "Account key in the credentials set file")
	repoAccessCmd.MarkFlagsRequiredTogether("creds", "account")
	repoAccessCmd.MarkFlagsMutuallyExclusive("repo", "non-member-orgs")
}

// repoAccessEntry is one row of repo-access output.
type repoAccessEntry struct {
	Repository string `json:"repository"`
	Status     string `json:"status"`
	Owner      string `json:"owner,omitempty"`
	OwnerType  string `json:"ownerType,omitempty"`
	Visibility string `json:"visibility,omitempty"`
	Permission string `json:"permission,omitempty"`
	Archived   bool   `json:"archived"`
	Detail     string `json:"detail,omitempty"`
}

func newRepoAccessEntry(r *gogithub.Repository) repoAccessEntry {
	entry := repoAccessEntry{
		Repository: r.FullName,
		Status:     string(repo.AccessGranted),
		Visibility: r.Visibility,
		Permission: r.Permissions.Highest(),
		Archived:   r.Archived,
	}
	if r.Owner != nil {
		entry.Owner = r.Owner.Login
		entry.OwnerType = r.Owner.Type
	}
	return entry
}

func runRepoAccess(cmd *cobra.Command, args []string) error {
	if repoAccessFormat != repoAccessFormatText && repoAccessFormat != repoAccessFormatJSON {
		return fmt.Errorf("unsupported format %q (use %s or %s)", repoAccessFormat, repoAccessFormatText, repoAccessFormatJSON)
	}

	ctx := context.Background()
	client, err := newRepoAccessClient(ctx)
	if err != nil {
		return err
	}

	user, err := client.GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("get authenticated user: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Authenticated as %s\n", user.Login)

	var entries []repoAccessEntry
	if len(repoAccessRepos) > 0 {
		entries, err = checkRepoAccess(ctx, client, repoAccessRepos)
	} else {
		entries, err = listRepoAccess(ctx, client, repoAccessNonMemberOrgs)
	}
	if err != nil {
		return err
	}
	return writeRepoAccess(entries, repoAccessFormat)
}

// newRepoAccessClient creates a client from a goauth credentials set when
// --creds is given, and from GITHUB_TOKEN otherwise.
func newRepoAccessClient(ctx context.Context) (clientv1.Client, error) {
	if repoAccessCreds == "" {
		client, err := clientv1.NewClient(ctx, ensureToken())
		if err != nil {
			return nil, fmt.Errorf("creating github client: %w", err)
		}
		return client, nil
	}
	client, err := credentialsset.NewClient(ctx, repoAccessCreds, repoAccessAccount,
		credentialsset.PromptReadWriter(os.Stderr, os.Stdin))
	if err != nil {
		return nil, fmt.Errorf("creating github client from credentials set: %w", err)
	}
	return client, nil
}

func listRepoAccess(ctx context.Context, client clientv1.Client, nonMemberOrgs bool) ([]repoAccessEntry, error) {
	var repos []*gogithub.Repository
	var err error
	if nonMemberOrgs {
		repos, err = repo.ListNonMemberOrgRepos(ctx, client)
	} else {
		repos, err = client.ListAuthenticatedUserRepos(ctx, nil)
	}
	if err != nil {
		return nil, err
	}
	entries := make([]repoAccessEntry, 0, len(repos))
	for _, r := range repos {
		entries = append(entries, newRepoAccessEntry(r))
	}
	return entries, nil
}

func checkRepoAccess(ctx context.Context, client clientv1.Client, fullNames []string) ([]repoAccessEntry, error) {
	entries := make([]repoAccessEntry, 0, len(fullNames))
	for _, fullName := range fullNames {
		owner, name, err := repo.ParseRepoName(fullName)
		if err != nil {
			return nil, err
		}
		result, err := repo.CheckAccess(ctx, client, owner, name)
		if err != nil {
			return nil, err
		}
		if result.Repository == nil {
			entries = append(entries, repoAccessEntry{
				Repository: fullName,
				Status:     string(result.Status),
				Owner:      owner,
				Detail:     result.Detail,
			})
			continue
		}
		entries = append(entries, newRepoAccessEntry(result.Repository))
	}
	return entries, nil
}

func writeRepoAccess(entries []repoAccessEntry, format string) error {
	if format == repoAccessFormatJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "REPOSITORY\tSTATUS\tOWNER TYPE\tVISIBILITY\tPERMISSION\tARCHIVED"); err != nil {
		return err
	}
	var details []error
	for _, e := range entries {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\n", e.Repository, e.Status, e.OwnerType, e.Visibility, e.Permission, e.Archived); err != nil {
			return err
		}
		if e.Detail != "" {
			details = append(details, fmt.Errorf("%s: %s", e.Repository, e.Detail))
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d repositories\n", len(entries))
	if len(details) > 0 {
		fmt.Fprintln(os.Stderr, errors.Join(details...))
	}
	return nil
}
