// Package main provides the gogithub CLI tool.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "dev"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gogithub",
	Short: "GitHub CLI utilities",
	Long: `gogithub is a CLI tool for GitHub operations including
searching pull requests and fetching user contribution statistics.

Environment Variables:
  GITHUB_TOKEN    GitHub personal access token (required for most commands)`,
	Version: Version,
}

func init() {
	rootCmd.AddCommand(searchPRsCmd)
	rootCmd.AddCommand(profileCmd)
}
