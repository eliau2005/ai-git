package git

import (
	"bytes"
	"os/exec"
	"strings"
)

// GetConflictingFiles returns a list of files that currently have merge conflicts.
func GetConflictingFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	var files []string
	for _, line := range lines {
		files = append(files, strings.TrimSpace(line))
	}
	return files, nil
}
