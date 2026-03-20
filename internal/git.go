package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetChangedFiles returns the list of changed files for a given commit ref.
// For merge commits with no file changes, it returns an empty slice.
func GetChangedFiles(commitRef string) ([]string, error) {
	// Verify the commit ref is valid
	if err := verifyCommitRef(commitRef); err != nil {
		return nil, err
	}

	// Check if it's a merge commit
	isMerge, err := isMergeCommit(commitRef)
	if err != nil {
		return nil, err
	}

	if isMerge {
		// For merge commits, use diff-tree which returns empty for merge commits
		// unless there were conflict resolutions
		files, err := diffTree(commitRef)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			return nil, nil
		}
		return files, nil
	}

	return diffTree(commitRef)
}

func verifyCommitRef(ref string) error {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("invalid commit reference '%s': %s", ref, strings.TrimSpace(string(output)))
	}
	return nil
}

func isMergeCommit(ref string) (bool, error) {
	// A merge commit has more than one parent
	cmd := exec.Command("git", "cat-file", "-p", ref)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to inspect commit '%s': %w", ref, err)
	}

	parentCount := 0
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "parent ") {
			parentCount++
		}
		// Stop after the header (empty line separates header from message)
		if line == "" {
			break
		}
	}

	return parentCount > 1, nil
}

func diffTree(ref string) ([]string, error) {
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-only", "-r", "--root", ref)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	raw := strings.TrimSpace(string(output))
	if raw == "" {
		return nil, nil
	}

	files := strings.Split(raw, "\n")
	return files, nil
}
