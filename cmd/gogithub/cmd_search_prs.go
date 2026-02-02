package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v82/github"
	"github.com/grokify/gogithub/search"
	"github.com/spf13/cobra"
)

var (
	searchAccounts []string
	searchOutfile  string
)

var searchPRsCmd = &cobra.Command{
	Use:   "search-prs",
	Short: "Search for open pull requests by user",
	Long: `Search for open pull requests across GitHub for specified users.
Results are exported to an Excel file.

Example:
  gogithub search-prs --accounts grokify,octocat --outfile prs.xlsx`,
	RunE: runSearchPRs,
}

func init() {
	searchPRsCmd.Flags().StringSliceVarP(&searchAccounts, "accounts", "a", nil, "GitHub accounts to search (required)")
	searchPRsCmd.Flags().StringVarP(&searchOutfile, "outfile", "o", "githubissues.xlsx", "Output Excel file")
	_ = searchPRsCmd.MarkFlagRequired("accounts")
}

func runSearchPRs(cmd *cobra.Command, args []string) error {
	if len(searchAccounts) == 0 {
		fmt.Println("No accounts specified")
		return nil
	}

	fmt.Printf("Loading public pull requests for (%s)\n", strings.Join(searchAccounts, ", "))

	searchOutfile = strings.TrimSpace(searchOutfile)
	if searchOutfile == "" {
		searchOutfile = "githubissues.xlsx"
	}

	// Create unauthenticated client for public data
	c := search.NewClient(github.NewClient(nil))

	ii := search.Issues{}

	for _, acct := range searchAccounts {
		qry := search.NewQuery().User(acct).StateOpen().IsPR().Build()
		iss, err := c.SearchIssuesAll(context.Background(), qry, nil)
		if err != nil {
			return fmt.Errorf("search issues for %s: %w", acct, err)
		}
		ii = append(ii, iss...)
	}

	ts, err := ii.TableSet()
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	if err := ts.WriteXLSX(searchOutfile); err != nil {
		return fmt.Errorf("write xlsx: %w", err)
	}

	fmt.Printf("Wrote %s\n", searchOutfile)
	fmt.Println("Done")

	return nil
}

// ensureToken gets a GitHub token from environment or exits with error.
func ensureToken() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: GITHUB_TOKEN environment variable not set")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Create a token at: https://github.com/settings/tokens?type=beta")
		fmt.Fprintln(os.Stderr, "For public data, use 'Public Repositories (read-only)' with no permissions.")
		os.Exit(1)
	}
	return token
}
