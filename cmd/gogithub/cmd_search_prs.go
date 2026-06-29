package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v88/github"
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

By default, results are displayed as an ASCII table to stdout.
Use -o/--outfile to write to a file (format auto-detected from extension).

Supported formats:
  .xlsx  Excel spreadsheet
  .md    Markdown table
  .csv   CSV file

Examples:
  gogithub search-prs -a grokify                    # ASCII table to stdout
  gogithub search-prs -a grokify -o prs.xlsx        # Excel file
  gogithub search-prs -a grokify,octocat -o prs.md  # Markdown file`,
	RunE: runSearchPRs,
}

func init() {
	searchPRsCmd.Flags().StringSliceVarP(&searchAccounts, "accounts", "a", nil, "GitHub accounts to search (required)")
	searchPRsCmd.Flags().StringVarP(&searchOutfile, "outfile", "o", "", "Output file (format from extension: .xlsx, .md, .csv)")
	_ = searchPRsCmd.MarkFlagRequired("accounts")
}

func runSearchPRs(cmd *cobra.Command, args []string) error {
	if len(searchAccounts) == 0 {
		fmt.Println("No accounts specified")
		return nil
	}

	searchOutfile = strings.TrimSpace(searchOutfile)

	// Only print status messages when writing to file (not stdout)
	if searchOutfile != "" {
		fmt.Fprintf(os.Stderr, "Loading public pull requests for (%s)\n", strings.Join(searchAccounts, ", "))
	}

	// Create unauthenticated client for public data
	ghClient, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("creating github client: %w", err)
	}
	c := search.NewClient(ghClient)

	ii := search.Issues{}

	for _, acct := range searchAccounts {
		qry := search.NewQuery().User(acct).StateOpen().IsPR().Build()
		iss, err := c.SearchIssuesAll(context.Background(), qry, nil)
		if err != nil {
			return fmt.Errorf("search issues for %s: %w", acct, err)
		}
		ii = append(ii, iss...)
	}

	// Output to stdout if no outfile specified
	if searchOutfile == "" {
		tbl, err := ii.Table("Pull Requests")
		if err != nil {
			return fmt.Errorf("create table: %w", err)
		}
		return tbl.Text(os.Stdout)
	}

	// Write to file based on extension
	ext := strings.ToLower(filepath.Ext(searchOutfile))
	switch ext {
	case ".xlsx":
		ts, err := ii.TableSet()
		if err != nil {
			return fmt.Errorf("create table: %w", err)
		}
		if err := ts.WriteXLSX(searchOutfile); err != nil {
			return fmt.Errorf("write xlsx: %w", err)
		}
	case ".md":
		tbl, err := ii.Table("Pull Requests")
		if err != nil {
			return fmt.Errorf("create table: %w", err)
		}
		if err := tbl.WriteMarkdown(searchOutfile, 0644, "\n", true); err != nil {
			return fmt.Errorf("write markdown: %w", err)
		}
	case ".csv":
		tbl, err := ii.Table("Pull Requests")
		if err != nil {
			return fmt.Errorf("create table: %w", err)
		}
		if err := tbl.WriteCSV(searchOutfile); err != nil {
			return fmt.Errorf("write csv: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file extension %q (use .xlsx, .md, or .csv)", ext)
	}

	fmt.Fprintf(os.Stderr, "Wrote %s\n", searchOutfile)
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
