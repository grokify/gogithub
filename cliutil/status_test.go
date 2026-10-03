package cliutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// newTestRepo creates a git repository with committed files, then deletes
// some in the working tree and leaves others modified or untracked.
func newTestRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...) //nolint:gosec // Test helper; args are fixed git subcommands
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	git("init", "-q")
	for _, name := range []string{"kept.txt", "deleted-a.txt", "deleted-b.txt", "with space.txt", "modified.txt"} {
		write(name, "original\n")
	}
	git("add", ".")
	git("commit", "-q", "-m", "initial")

	for _, name := range []string{"deleted-a.txt", "deleted-b.txt", "with space.txt"} {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	write("modified.txt", "changed\n")
	write("untracked.txt", "new\n")
	return dir
}

func TestGitStatusShortLines(t *testing.T) {
	dir := newTestRepo(t)

	lines, err := GitStatusShortLines(dir)
	if err != nil {
		t.Fatalf("GitStatusShortLines() error = %v", err)
	}

	if len(lines) != 5 {
		t.Errorf("len(lines) = %d, want 5: %q", len(lines), lines)
	}
	for _, line := range lines {
		if line == "" {
			t.Error("GitStatusShortLines() returned an empty line")
		}
	}
	if !slices.Contains(lines, "?? untracked.txt") {
		t.Errorf("lines = %q, want an untracked entry", lines)
	}
}

func TestGitStatusShortLinesClean(t *testing.T) {
	dir := newTestRepo(t)
	for _, name := range []string{"deleted-a.txt", "deleted-b.txt", "with space.txt", "modified.txt"} {
		cmd := exec.Command("git", "-C", dir, "checkout", "--", name)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git checkout: %v\n%s", err, out)
		}
	}
	if err := os.Remove(filepath.Join(dir, "untracked.txt")); err != nil {
		t.Fatal(err)
	}

	lines, err := GitStatusShortLines(dir)
	if err != nil {
		t.Fatalf("GitStatusShortLines() error = %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("lines = %q, want none for a clean tree", lines)
	}
}

func TestGitStatusShortLinesErrors(t *testing.T) {
	tests := []struct {
		name string
		dir  string
	}{
		{"missing directory", filepath.Join(t.TempDir(), "missing")},
		{"not a repository", t.TempDir()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GitStatusShortLines(tt.dir); err == nil {
				t.Error("GitStatusShortLines() error = nil, want error")
			}
		})
	}
}

func TestGitStatusShortLinesNotADirectory(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := GitStatusShortLines(file); err == nil {
		t.Error("GitStatusShortLines() error = nil, want error")
	}
}

func TestGitRmDeletedLines(t *testing.T) {
	dir := newTestRepo(t)

	rmlines, err := GitRmDeletedLines(dir)
	if err != nil {
		t.Fatalf("GitRmDeletedLines() error = %v", err)
	}

	want := []string{"git rm deleted-a.txt", "git rm deleted-b.txt", `git rm "with space.txt"`}
	if !slices.Equal(rmlines, want) {
		t.Errorf("GitRmDeletedLines() = %q, want %q", rmlines, want)
	}
}

func TestGitRmDeletedLinesStagesDeletions(t *testing.T) {
	dir := newTestRepo(t)
	rmlines, err := GitRmDeletedLines(dir)
	if err != nil {
		t.Fatalf("GitRmDeletedLines() error = %v", err)
	}

	script := strings.Join(rmlines, "\n")
	cmd := exec.Command("sh", "-c", script)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run generated commands: %v\n%s", err, out)
	}

	remaining, err := GitRmDeletedLines(dir)
	if err != nil {
		t.Fatalf("GitRmDeletedLines() after staging error = %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("GitRmDeletedLines() after staging = %q, want none", remaining)
	}
}

func TestGitRmDeletedFile(t *testing.T) {
	dir := newTestRepo(t)
	filename := filepath.Join(t.TempDir(), "rm-deleted.sh")
	if err := os.WriteFile(filename, []byte(strings.Repeat("stale content\n", 20)), 0600); err != nil {
		t.Fatal(err)
	}

	if err := GitRmDeletedFile(filename, dir); err != nil {
		t.Fatalf("GitRmDeletedFile() error = %v", err)
	}

	got, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	want := "git rm deleted-a.txt\ngit rm deleted-b.txt\ngit rm \"with space.txt\"\n"
	if string(got) != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}
