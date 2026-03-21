package internal

import (
	"path/filepath"
	"strings"
)

// FilterByExt filters files to only include those matching the given extensions.
// exts is a comma-separated list, e.g. ".go,.ts,.js"
func FilterByExt(files []string, exts string) []string {
	extList := strings.Split(exts, ",")
	extSet := make(map[string]bool, len(extList))
	for _, ext := range extList {
		ext = strings.TrimSpace(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extSet[ext] = true
	}

	var result []string
	for _, f := range files {
		if extSet[filepath.Ext(f)] {
			result = append(result, f)
		}
	}
	return result
}

// FilterByExclude removes files matching the given glob pattern.
// Matches against both the full path and the basename.
func FilterByExclude(files []string, pattern string) []string {
	var result []string
	for _, f := range files {
		// Try matching against full path first, then basename
		matchedFull, _ := filepath.Match(pattern, f)
		matchedBase, _ := filepath.Match(pattern, filepath.Base(f))
		if !matchedFull && !matchedBase {
			result = append(result, f)
		}
	}
	return result
}

// FilterByDir keeps only files under the given directory prefix.
func FilterByDir(files []string, dir string) []string {
	dir = strings.TrimSuffix(dir, "/") + "/"
	var result []string
	for _, f := range files {
		if strings.HasPrefix(f, dir) {
			result = append(result, f)
		}
	}
	return result
}

// FilterChangesByStatus keeps only files matching the given status.
// status: "added" (A), "modified" (M), "deleted" (D), "renamed" (R),
// or a raw letter like "A", "M", "D", "R".
func FilterChangesByStatus(changes []FileChange, status string) []FileChange {
	target := strings.ToUpper(status)
	// Map friendly names to status letters
	nameMap := map[string]string{
		"ADDED":    "A",
		"MODIFIED": "M",
		"DELETED":  "D",
		"RENAMED":  "R",
		"COPIED":   "C",
	}
	if mapped, ok := nameMap[target]; ok {
		target = mapped
	}

	var result []FileChange
	for _, c := range changes {
		if c.Status == target {
			result = append(result, c)
		}
	}
	return result
}
