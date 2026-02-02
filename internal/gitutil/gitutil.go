package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsGitRepository checks if the current directory or any parent is a git repository
func IsGitRepository(dir string) bool {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return false
		}
	}

	// Walk up the directory tree looking for .git
	current := dir
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return true
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			return false
		}
		current = parent
	}
}

// GetCurrentBranch returns the current git branch name
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemotes returns a map of git remote names to their URLs
func GetRemotes() (map[string]string, error) {
	cmd := exec.Command("git", "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	remotes := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: origin	git@github.com:user/repo.git (fetch)
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			url := parts[1]
			// Only store once (skip duplicate fetch/push entries)
			if _, exists := remotes[name]; !exists {
				remotes[name] = url
			}
		}
	}

	return remotes, nil
}
