package internal

import (
	"path/filepath"
	"strings"
)

// FilterByExt filters files to only include those matching the given extensions.
// exts is a comma-separated list of extensions, e.g. ".go,.ts,.js"
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
// pattern uses filepath.Match syntax, e.g. "*_test.go"
func FilterByExclude(files []string, pattern string) []string {
	var result []string
	for _, f := range files {
		base := filepath.Base(f)
		matched, err := filepath.Match(pattern, base)
		if err != nil || !matched {
			result = append(result, f)
		}
	}
	return result
}
