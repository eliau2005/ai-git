package git

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

func IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func Status() (string, error) {
	cmd := exec.Command("git", "status")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func StatusShort() (string, error) {
	cmd := exec.Command("git", "status", "--short")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func DiffStaged() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func Add(path string) error {
	cmd := exec.Command("git", "add", path)
	return cmd.Run()
}

func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

func Push() error {
	cmd := exec.Command("git", "push")
	return cmd.Run()
}

func PushInteractive() error {
	cmd := exec.Command("git", "push")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Pull() error {
	cmd := exec.Command("git", "pull")
	return cmd.Run()
}

func PullInteractive() error {
	cmd := exec.Command("git", "pull")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetBranches() ([]string, string, error) {
	cmd := exec.Command("git", "branch")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, "", err
	}

	var branches []string
	var current string
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "* ") {
			current = strings.TrimPrefix(trimmed, "* ")
			branches = append(branches, current)
		} else {
			branches = append(branches, trimmed)
		}
	}
	return branches, current, nil
}

func Checkout(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	return cmd.Run()
}

func CreateBranch(branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	return cmd.Run()
}

func DeleteBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	return cmd.Run()
}
