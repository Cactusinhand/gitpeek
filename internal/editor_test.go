package internal

import (
	"testing"
)

func TestResolveBin(t *testing.T) {
	tests := []struct {
		editor string
		want   string
	}{
		{"vscode", "code"},
		{"code", "code"},
		{"cursor", "cursor"},
		{"zed", "zed"},
		{"sublime", "subl"},
		{"subl", "subl"},
		{"vim", "vim"},
		{"nvim", "nvim"},
		{"neovim", "nvim"},
		{"unknown-editor", "unknown-editor"},
	}

	for _, tt := range tests {
		t.Run(tt.editor, func(t *testing.T) {
			got := resolveBin(tt.editor)
			if got != tt.want {
				t.Errorf("resolveBin(%q) = %q, want %q", tt.editor, got, tt.want)
			}
		})
	}
}

func TestGotoEditorCategories(t *testing.T) {
	// Verify that code/cursor use -g flag style
	for _, bin := range []string{"code", "cursor"} {
		if !gotoFlagEditors[bin] {
			t.Errorf("expected %q to be in gotoFlagEditors", bin)
		}
	}

	// Verify that zed/subl use file:line arg style
	for _, bin := range []string{"zed", "subl"} {
		if !gotoArgEditors[bin] {
			t.Errorf("expected %q to be in gotoArgEditors", bin)
		}
	}

	// Verify that vim/nvim use +line style
	for _, bin := range []string{"vim", "nvim"} {
		if !gotoPlusEditors[bin] {
			t.Errorf("expected %q to be in gotoPlusEditors", bin)
		}
	}
}

func TestTerminalEditorsAreSubsetOfPlus(t *testing.T) {
	// All terminal editors should also be in gotoPlusEditors (they use +line syntax)
	for bin := range terminalEditors {
		if !gotoPlusEditors[bin] {
			t.Errorf("terminal editor %q should also be in gotoPlusEditors", bin)
		}
	}
}

func TestEditorMapCompleteness(t *testing.T) {
	// Every editor binary should be in exactly one goto category (or none for unknown)
	allBins := make(map[string]bool)
	for _, bin := range editors {
		allBins[bin] = true
	}

	for bin := range allBins {
		categories := 0
		if gotoFlagEditors[bin] {
			categories++
		}
		if gotoArgEditors[bin] {
			categories++
		}
		if gotoPlusEditors[bin] {
			categories++
		}
		if categories > 1 {
			t.Errorf("binary %q appears in %d goto categories, expected at most 1", bin, categories)
		}
	}
}
