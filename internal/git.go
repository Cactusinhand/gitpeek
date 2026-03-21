package internal

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FileChange represents a changed file with its git status.
type FileChange struct {
	Status string // A (added), M (modified), D (deleted), R (renamed), etc.
	Path   string
}

// GetChangedFiles returns the list of changed files for a given commit ref.
func GetChangedFiles(commitRef string) ([]string, error) {
	changes, err := GetChangedFilesWithStatus(commitRef)
	if err != nil {
		return nil, err
	}
	return changesToPaths(changes), nil
}

// GetChangedFilesWithStatus returns files with their change status.
func GetChangedFilesWithStatus(commitRef string) ([]FileChange, error) {
	// Stash ref
	if strings.HasPrefix(commitRef, "stash@{") {
		return getStashFiles(commitRef)
	}

	// Commit range
	if strings.Contains(commitRef, "..") {
		return getRangeFilesWithStatus(commitRef)
	}

	if err := verifyCommitRef(commitRef); err != nil {
		return nil, err
	}

	isMerge, err := isMergeCommit(commitRef)
	if err != nil {
		return nil, err
	}

	changes, err := diffTreeWithStatus(commitRef)
	if err != nil {
		return nil, err
	}

	if isMerge && len(changes) == 0 {
		return nil, nil
	}
	return changes, nil
}

// GetStagedFiles returns files that have been staged but not yet committed.
func GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}
	return splitLines(output), nil
}

// GetUnstagedFiles returns files modified but not staged, plus untracked files.
func GetUnstagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get unstaged files: %w", err)
	}
	files := splitLines(output)

	cmd2 := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output2, err := cmd2.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get untracked files: %w", err)
	}
	untracked := splitLines(output2)

	seen := make(map[string]bool, len(files))
	for _, f := range files {
		seen[f] = true
	}
	for _, f := range untracked {
		if !seen[f] {
			files = append(files, f)
		}
	}
	return files, nil
}

// GetFirstChangedLines parses the diff of a commit and returns the first changed
// line number for each file. Returns map[filepath]lineNumber.
func GetFirstChangedLines(commitRef string) (map[string]int, error) {
	var cmd *exec.Cmd

	if strings.HasPrefix(commitRef, "stash@{") {
		cmd = exec.Command("git", "stash", "show", "-p", "--unified=0", commitRef)
	} else if strings.Contains(commitRef, "..") {
		cmd = exec.Command("git", "diff", "--unified=0", commitRef)
	} else {
		cmd = exec.Command("git", "diff-tree", "-p", "--unified=0", "--root", "-r", commitRef)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff for line numbers: %w", err)
	}

	return parseDiffForLines(string(output)), nil
}

func parseDiffForLines(diff string) map[string]int {
	result := make(map[string]int)
	var currentFile string

	for _, line := range strings.Split(diff, "\n") {
		// Match: diff --git a/path b/path
		if strings.HasPrefix(line, "diff --git ") {
			parts := strings.Split(line, " b/")
			if len(parts) == 2 {
				currentFile = parts[1]
			}
		}
		// Match: @@ -old,count +new,count @@
		if strings.HasPrefix(line, "@@") && currentFile != "" {
			if _, exists := result[currentFile]; !exists {
				if lineNum := parseHunkLine(line); lineNum > 0 {
					result[currentFile] = lineNum
				}
			}
		}
	}
	return result
}

func parseHunkLine(hunk string) int {
	// Parse "@@ -a,b +c,d @@" to extract c (the new file line number)
	plusIdx := strings.Index(hunk, "+")
	if plusIdx < 0 {
		return 0
	}
	rest := hunk[plusIdx+1:]
	// rest is like "10,5 @@" or "10 @@"
	end := strings.IndexAny(rest, ", @")
	if end < 0 {
		return 0
	}
	n, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0
	}
	return n
}

func getStashFiles(stashRef string) ([]FileChange, error) {
	cmd := exec.Command("git", "stash", "show", "--name-status", stashRef)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get stash files for '%s': %s", stashRef, strings.TrimSpace(string(output)))
	}
	return parseNameStatus(output), nil
}

func getRangeFilesWithStatus(rangeRef string) ([]FileChange, error) {
	parts := strings.SplitN(rangeRef, "..", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid commit range '%s', expected format: ref1..ref2", rangeRef)
	}
	for _, ref := range parts {
		if err := verifyCommitRef(ref); err != nil {
			return nil, err
		}
	}

	cmd := exec.Command("git", "diff", "--name-status", rangeRef)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get files for range '%s': %w", rangeRef, err)
	}
	return parseNameStatus(output), nil
}

func diffTreeWithStatus(ref string) ([]FileChange, error) {
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-status", "-r", "--root", ref)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}
	return parseNameStatus(output), nil
}

func parseNameStatus(output []byte) []FileChange {
	raw := strings.TrimSpace(string(output))
	if raw == "" {
		return nil
	}
	var changes []FileChange
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			status := string(fields[0][0]) // Take first char (R100 → R)
			path := fields[len(fields)-1]  // For renames, take the new path
			changes = append(changes, FileChange{Status: status, Path: path})
		}
	}
	return changes
}

func changesToPaths(changes []FileChange) []string {
	if changes == nil {
		return nil
	}
	paths := make([]string, len(changes))
	for i, c := range changes {
		paths[i] = c.Path
	}
	return paths
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
