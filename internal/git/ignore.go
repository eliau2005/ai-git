package git

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func LoadIgnoreRules(root string) ([]string, error) {
	ignorePath := filepath.Join(root, ".aiignore")
	file, err := os.Open(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No ignore file
		}
		return nil, err
	}
	defer file.Close()

	var rules []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			rules = append(rules, line)
		}
	}
	return rules, scanner.Err()
}

// ShouldIgnore checks if a file matches any rule.
// This is a naive implementation (contains check). 
// For robust glob matching, we'd need `path/filepath.Match` or a library.
// Let's use `filepath.Match`.
func ShouldIgnore(file string, rules []string) bool {
	for _, rule := range rules {
		// Handle directory matches like "lockfiles/"
		if strings.HasSuffix(rule, "/") {
			if strings.HasPrefix(file, rule) {
				return true
			}
		}
		
		matched, _ := filepath.Match(rule, file)
		if matched {
			return true
		}
		// Check basic suffix (e.g. .lock)
		if strings.HasSuffix(file, rule) {
			return true
		}
	}
	// Default hardcoded rules
	defaults := []string{"package-lock.json", "yarn.lock", "go.sum", "pnpm-lock.yaml"}
	for _, d := range defaults {
		if strings.HasSuffix(file, d) {
			return true
		}
	}
	return false
}
