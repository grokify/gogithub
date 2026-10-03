// Package cliutil provides helpers for scripting git operations on a local
// working tree, such as staging files that were deleted outside of git.
package cliutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/grokify/mogo/os/osutil"
)

// GitCmdStatusShort is the git command whose output GitStatusShortLines
// returns. The porcelain format is stable across git versions.
const GitCmdStatusShort = "git status --porcelain"

// deletedInWorktree matches a porcelain status line for a file deleted in the
// working tree but still tracked in the index, capturing its path.
var deletedInWorktree = regexp.MustCompile(`^\s+D\s+(.+?)\s*$`)

// GitStatusShortLines returns the non-empty lines of `git status --porcelain`
// for the repository at dir.
func GitStatusShortLines(dir string) ([]string, error) {
	ok, err := osutil.IsDir(dir)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("path is not a directory")
	}
	cmd := exec.Command("git", "-C", dir, "status", "--porcelain")
	stdout, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("%s: %w: %s", GitCmdStatusShort, err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("%s: %w", GitCmdStatusShort, err)
	}
	var lines []string
	for line := range strings.SplitSeq(strings.TrimRight(string(stdout), "\n"), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

// GitRmDeletedLines returns one `git rm <path>` command per file that was
// deleted in the working tree of the repository at dir but is still tracked,
// so that running them stages the deletions. Paths are quoted as git prints
// them.
func GitRmDeletedLines(dir string) ([]string, error) {
	lines, err := GitStatusShortLines(dir)
	if err != nil {
		return nil, err
	}
	var rmlines []string
	for _, line := range lines {
		m := deletedInWorktree.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		rmlines = append(rmlines, "git rm "+m[1])
	}
	return rmlines, nil
}

// GitRmDeletedFile writes the commands returned by GitRmDeletedLines to
// filename, one per line, replacing any existing file.
func GitRmDeletedFile(filename, dir string) error {
	rmlines, err := GitRmDeletedLines(dir)
	if err != nil {
		return err
	}
	var sb strings.Builder
	for _, rmline := range rmlines {
		sb.WriteString(rmline)
		sb.WriteByte('\n')
	}
	return os.WriteFile(filename, []byte(sb.String()), 0600)
}
