package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetChangedFiles returns the list of changed files for a given commit ref.
// For merge commits with no file changes, it returns an empty slice.
func GetChangedFiles(commitRef string) ([]string, error) {
	// Check if it's a commit range (contains "..")
	if strings.Contains(commitRef, "..") {
		return getRangeFiles(commitRef)
	}

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

// GetStagedFiles returns files that have been staged (git add) but not yet committed.
func GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}
	return splitLines(output), nil
}

// GetUnstagedFiles returns files that have been modified but not staged.
func GetUnstagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get unstaged files: %w", err)
	}
	return splitLines(output), nil
}

func getRangeFiles(rangeRef string) ([]string, error) {
	// Verify both ends of the range
	parts := strings.SplitN(rangeRef, "..", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid commit range '%s', expected format: ref1..ref2", rangeRef)
	}

	for _, ref := range parts {
		if err := verifyCommitRef(ref); err != nil {
			return nil, err
		}
	}

	cmd := exec.Command("git", "diff", "--name-only", rangeRef)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get files for range '%s': %w", rangeRef, err)
	}
	return splitLines(output), nil
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
	return splitLines(output), nil
}

func splitLines(output []byte) []string {
	raw := strings.TrimSpace(string(output))
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}
