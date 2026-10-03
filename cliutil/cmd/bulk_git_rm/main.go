// Command bulk_git_rm prints a `git rm` command for every tracked file that
// was deleted in a repository's working tree, so the deletions can be staged
// in one step.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/grokify/gogithub/cliutil"
)

var (
	dir     string
	outfile string
)

var rootCmd = &cobra.Command{
	Use:   "bulk_git_rm",
	Short: "Generate git rm commands for files deleted in the working tree",
	Long: `Generate a git rm command for every tracked file that was deleted in a
repository's working tree, so the deletions can be staged in one step.

Commands are printed to stdout, or written to a file with --out.

Examples:
  bulk_git_rm                        # Current directory
  bulk_git_rm --dir ../other-repo    # Another repository
  bulk_git_rm --out rm-deleted.sh    # Write a script instead of printing
  bulk_git_rm | sh                   # Stage the deletions`,
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&dir, "dir", "d", ".", "Repository directory")
	rootCmd.Flags().StringVarP(&outfile, "out", "o", "", "Write commands to this file instead of stdout")
}

func run(cmd *cobra.Command, args []string) error {
	if outfile != "" {
		if err := cliutil.GitRmDeletedFile(outfile, dir); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Wrote %s\n", outfile)
		return nil
	}
	rmlines, err := cliutil.GitRmDeletedLines(dir)
	if err != nil {
		return err
	}
	for _, rmline := range rmlines {
		fmt.Println(rmline)
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
