package internal

import (
	"reflect"
	"testing"
)

func TestFilterByExt(t *testing.T) {
	files := []string{
		"main.go",
		"utils.ts",
		"style.css",
		"cmd/root.go",
		"README.md",
	}

	tests := []struct {
		name string
		exts string
		want []string
	}{
		{"single ext with dot", ".go", []string{"main.go", "cmd/root.go"}},
		{"single ext without dot", "go", []string{"main.go", "cmd/root.go"}},
		{"multiple exts", ".go,.ts", []string{"main.go", "utils.ts", "cmd/root.go"}},
		{"no match", ".py", nil},
		{"with spaces", ".go, .ts", []string{"main.go", "utils.ts", "cmd/root.go"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByExt(files, tt.exts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterByExt(%q) = %v, want %v", tt.exts, got, tt.want)
			}
		})
	}
}

func TestFilterByExt_EmptyInput(t *testing.T) {
	got := FilterByExt(nil, ".go")
	if got != nil {
		t.Errorf("FilterByExt(nil) = %v, want nil", got)
	}
}

func TestFilterByExclude(t *testing.T) {
	files := []string{
		"main.go",
		"main_test.go",
		"cmd/root.go",
		"cmd/root_test.go",
		"go.sum",
	}

	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{"exclude test files", "*_test.go", []string{"main.go", "cmd/root.go", "go.sum"}},
		{"exclude sum", "*.sum", []string{"main.go", "main_test.go", "cmd/root.go", "cmd/root_test.go"}},
		{"no match", "*.py", files},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByExclude(files, tt.pattern)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterByExclude(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestFilterByExclude_DirPattern(t *testing.T) {
	files := []string{
		"main.go",
		"internal/git.go",
		"internal/editor.go",
		"cmd/root.go",
	}

	// User expects "internal/*" to exclude internal/ files, but current
	// implementation only matches against basename, so this won't work.
	got := FilterByExclude(files, "internal/*")
	wantIfWorking := []string{"main.go", "cmd/root.go"}

	if !reflect.DeepEqual(got, wantIfWorking) {
		t.Errorf("FilterByExclude with dir pattern: got %v, want %v (dir-level exclude not working)", got, wantIfWorking)
	}
}

func TestFilterByExclude_EmptyInput(t *testing.T) {
	got := FilterByExclude(nil, "*_test.go")
	if got != nil {
		t.Errorf("FilterByExclude(nil) = %v, want nil", got)
	}
}

func TestFilterByDir(t *testing.T) {
	files := []string{
		"main.go",
		"cmd/root.go",
		"cmd/sub/help.go",
		"internal/git.go",
		"internal/editor.go",
	}

	tests := []struct {
		name string
		dir  string
		want []string
	}{
		{"cmd dir", "cmd", []string{"cmd/root.go", "cmd/sub/help.go"}},
		{"cmd dir with slash", "cmd/", []string{"cmd/root.go", "cmd/sub/help.go"}},
		{"internal dir", "internal", []string{"internal/git.go", "internal/editor.go"}},
		{"nested dir", "cmd/sub", []string{"cmd/sub/help.go"}},
		{"no match", "pkg", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByDir(files, tt.dir)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterByDir(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestFilterChangesByStatus(t *testing.T) {
	changes := []FileChange{
		{Status: "A", Path: "new.go"},
		{Status: "M", Path: "changed.go"},
		{Status: "D", Path: "removed.go"},
		{Status: "M", Path: "also_changed.go"},
		{Status: "R", Path: "renamed.go"},
	}

	tests := []struct {
		name   string
		status string
		want   []FileChange
	}{
		{"by letter A", "A", []FileChange{{Status: "A", Path: "new.go"}}},
		{"by letter M", "M", []FileChange{{Status: "M", Path: "changed.go"}, {Status: "M", Path: "also_changed.go"}}},
		{"by name added", "added", []FileChange{{Status: "A", Path: "new.go"}}},
		{"by name modified", "modified", []FileChange{{Status: "M", Path: "changed.go"}, {Status: "M", Path: "also_changed.go"}}},
		{"by name deleted", "deleted", []FileChange{{Status: "D", Path: "removed.go"}}},
		{"by name renamed", "renamed", []FileChange{{Status: "R", Path: "renamed.go"}}},
		{"case insensitive", "Added", []FileChange{{Status: "A", Path: "new.go"}}},
		{"no match", "C", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterChangesByStatus(changes, tt.status)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterChangesByStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
