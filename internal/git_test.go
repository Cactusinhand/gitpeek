package internal

import (
	"reflect"
	"testing"
)

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []string
	}{
		{"normal", []byte("a.go\nb.go\nc.go\n"), []string{"a.go", "b.go", "c.go"}},
		{"no trailing newline", []byte("a.go\nb.go"), []string{"a.go", "b.go"}},
		{"single line", []byte("a.go\n"), []string{"a.go"}},
		{"empty", []byte(""), nil},
		{"whitespace only", []byte("  \n  "), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitLines(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseNameStatus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []FileChange
	}{
		{
			"basic",
			"A\tnew.go\nM\tchanged.go\nD\tremoved.go\n",
			[]FileChange{
				{Status: "A", Path: "new.go"},
				{Status: "M", Path: "changed.go"},
				{Status: "D", Path: "removed.go"},
			},
		},
		{
			"rename with score",
			"R100\told.go\tnew.go\n",
			[]FileChange{
				{Status: "R", Path: "new.go"},
			},
		},
		{
			"empty",
			"",
			nil,
		},
		{
			"paths with directories",
			"M\tcmd/root.go\nA\tinternal/config.go\n",
			[]FileChange{
				{Status: "M", Path: "cmd/root.go"},
				{Status: "A", Path: "internal/config.go"},
			},
		},
		{
			"file with spaces",
			"M\tmy file.go\nA\tpath/to/another file.ts\n",
			[]FileChange{
				{Status: "M", Path: "my file.go"},
				{Status: "A", Path: "path/to/another file.ts"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNameStatus([]byte(tt.input))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseNameStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangesToPaths(t *testing.T) {
	tests := []struct {
		name    string
		changes []FileChange
		want    []string
	}{
		{
			"normal",
			[]FileChange{
				{Status: "A", Path: "a.go"},
				{Status: "M", Path: "b.go"},
			},
			[]string{"a.go", "b.go"},
		},
		{"nil input", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := changesToPaths(tt.changes)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("changesToPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHunkLine(t *testing.T) {
	tests := []struct {
		name string
		hunk string
		want int
	}{
		{"standard hunk", "@@ -10,3 +15,5 @@", 15},
		{"single line add", "@@ -0,0 +1 @@", 1},
		{"single line change", "@@ -5 +5 @@", 5},
		{"large line number", "@@ -100,20 +200,30 @@", 200},
		{"with context", "@@ -10,3 +15,5 @@ func main()", 15},
		{"no plus sign", "@@ -10,3 @@", 0},
		{"zero line", "@@ -0,0 +0,0 @@", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHunkLine(tt.hunk)
			if got != tt.want {
				t.Errorf("parseHunkLine(%q) = %d, want %d", tt.hunk, got, tt.want)
			}
		})
	}
}

func TestParseDiffForLines(t *testing.T) {
	diff := `commit abc123
diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -5,3 +5,4 @@ package main
@@ -20,2 +21,3 @@ func init()
diff --git a/cmd/root.go b/cmd/root.go
--- a/cmd/root.go
+++ b/cmd/root.go
@@ -10,0 +11,5 @@ func run()
`
	got := parseDiffForLines(diff)

	want := map[string]int{
		"main.go":     5,
		"cmd/root.go": 11,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDiffForLines() = %v, want %v", got, want)
	}
}

func TestParseDiffForLines_Empty(t *testing.T) {
	got := parseDiffForLines("")
	if len(got) != 0 {
		t.Errorf("parseDiffForLines(\"\") = %v, want empty map", got)
	}
}

func TestParseDiffForLines_PathContainsSpaceB(t *testing.T) {
	// File path containing " b/" in its directory name
	diff := `diff --git a/a b/config.go b/a b/config.go
--- a/a b/config.go
+++ b/a b/config.go
@@ -1,3 +1,5 @@ package config
`
	got := parseDiffForLines(diff)
	want := map[string]int{
		"a b/config.go": 1,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDiffForLines() = %v, want %v", got, want)
	}
}

func TestParseDiffForLines_NewFile(t *testing.T) {
	diff := `diff --git a/new_file.go b/new_file.go
new file mode 100644
--- /dev/null
+++ b/new_file.go
@@ -0,0 +1,10 @@
`
	got := parseDiffForLines(diff)
	want := map[string]int{
		"new_file.go": 1,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDiffForLines() = %v, want %v", got, want)
	}
}
