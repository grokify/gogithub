package cliutil

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/grokify/mogo/os/executil"
	"github.com/grokify/mogo/os/osutil"
)

const GitCmdStatusShort = "git status -s"

func GitStatusShortLines(dir string) ([]string, error) {
	ok, err := osutil.IsDir(dir)
	if err != nil {
		return []string{}, err
	} else if !ok {
		return []string{}, errors.New("path is not a directory")
	}
	err = os.Chdir(dir)
	if err != nil {
		return []string{}, err
	}
	stdout, _, err := executil.ExecSimple(GitCmdStatusShort)
	if err != nil {
		return []string{}, err
	}
	lines := strings.Split(stdout.String(), "\n")
	return lines, nil
}

func GitRmDeletedLines(dir string) ([]string, error) {
	lines, err := GitStatusShortLines(dir)
	if err != nil {
		return []string{}, err
	}

	rmlines := []string{}
	rx := regexp.MustCompile(`^\s+D\s+(.+)\s*$`)

	for _, line := range lines {
		m := rx.FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}
		rmlines = append(rmlines, fmt.Sprintf("git rm %s\n", m[1]))
	}
	return rmlines, nil
}

func GitRmDeletedFile(filename, dir string) error {
	rmlines, err := GitRmDeletedLines(dir)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	for _, rmline := range rmlines {
		_, err := file.WriteString(rmline + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}
