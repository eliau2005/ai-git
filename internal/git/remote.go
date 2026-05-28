package git

import (
	"bytes"
	"os/exec"
	"strings"
)

type RemoteInfo struct {
	Platform string // "github", "gitlab", "bitbucket", or "unknown"
	Owner    string
	Repo     string
}

// GetRemoteInfo gets the origin remote and parses the provider, owner, and repo.
func GetRemoteInfo() (*RemoteInfo, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	url := strings.TrimSpace(out.String())
	return ParseRemoteURL(url), nil
}

// ParseRemoteURL parses a git remote URL into RemoteInfo
func ParseRemoteURL(url string) *RemoteInfo {
	info := &RemoteInfo{Platform: "unknown"}

	// Determine platform
	if strings.Contains(url, "github.com") {
		info.Platform = "github"
	} else if strings.Contains(url, "gitlab.com") {
		info.Platform = "gitlab"
	} else if strings.Contains(url, "bitbucket.org") {
		info.Platform = "bitbucket"
	}

	// Extract Owner and Repo
	// Formats:
	// git@github.com:owner/repo.git
	// https://github.com/owner/repo.git
	
	var pathPart string
	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			pathPart = parts[1]
		}
	} else if strings.HasPrefix(url, "http") {
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			pathPart = strings.Join(parts[len(parts)-2:], "/")
		}
	}

	if pathPart != "" {
		pathPart = strings.TrimSuffix(pathPart, ".git")
		parts := strings.Split(pathPart, "/")
		if len(parts) >= 2 {
			info.Owner = parts[0]
			info.Repo = strings.Join(parts[1:], "/")
		}
	}

	return info
}
