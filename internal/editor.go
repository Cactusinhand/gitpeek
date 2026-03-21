package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Supported editor configurations
var editors = map[string]string{
	"vscode":  "code",
	"code":    "code",
	"cursor":  "cursor",
	"zed":     "zed",
	"sublime": "subl",
	"subl":    "subl",
	"vim":     "vim",
	"nvim":    "nvim",
	"neovim":  "nvim",
}

// Editors that support -g file:line syntax
var gotoFlagEditors = map[string]bool{
	"code":   true,
	"cursor": true,
}

// Editors that support file:line as argument directly
var gotoArgEditors = map[string]bool{
	"zed":  true,
	"subl": true,
}

// Editors that support +line file syntax
var gotoPlusEditors = map[string]bool{
	"vim":  true,
	"nvim": true,
}

// DetectEditor tries to find an available editor, checking $EDITOR env var first,
// then common GUI editors in order of preference.
func DetectEditor() string {
	if env := os.Getenv("EDITOR"); env != "" {
		if _, err := exec.LookPath(env); err == nil {
			return env
		}
	}
	for _, bin := range []string{"cursor", "code", "zed", "subl", "nvim", "vim"} {
		if _, err := exec.LookPath(bin); err == nil {
			return bin
		}
	}
	return ""
}

// OpenFiles opens the given files in the specified editor.
func OpenFiles(editor string, files []string) error {
	bin := resolveBin(editor)

	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("editor '%s' (command: '%s') not found in PATH, use --terminal to specify one", editor, bin)
	}

	cmd := exec.Command(bin, files...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open editor '%s': %w", bin, err)
	}

	return nil
}

// OpenFilesWithLines opens files at specific line numbers.
// lines maps file path to line number.
func OpenFilesWithLines(editor string, files []string, lines map[string]int) error {
	bin := resolveBin(editor)

	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("editor '%s' (command: '%s') not found in PATH, use --terminal to specify one", editor, bin)
	}

	var args []string

	switch {
	case gotoFlagEditors[bin]:
		// code/cursor: -g file:line -g file:line
		for _, f := range files {
			if line, ok := lines[f]; ok && line > 0 {
				args = append(args, "-g", f+":"+strconv.Itoa(line))
			} else {
				args = append(args, f)
			}
		}

	case gotoArgEditors[bin]:
		// zed/subl: file:line file:line
		for _, f := range files {
			if line, ok := lines[f]; ok && line > 0 {
				args = append(args, f+":"+strconv.Itoa(line))
			} else {
				args = append(args, f)
			}
		}

	case gotoPlusEditors[bin]:
		// vim/nvim: +line file (only supports one file with goto)
		// For multiple files, open first file at its line, rest normally
		if len(files) > 0 {
			first := files[0]
			if line, ok := lines[first]; ok && line > 0 {
				args = append(args, "+"+strconv.Itoa(line))
			}
			args = append(args, files...)
		}

	default:
		// Unknown editor, just pass file paths
		args = files
	}

	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open editor '%s': %w", bin, err)
	}

	return nil
}

func resolveBin(editor string) string {
	if bin, ok := editors[editor]; ok {
		return bin
	}
	return editor
}

// ListEditors returns the list of supported editor names.
func ListEditors() []string {
	seen := make(map[string]bool)
	var result []string
	for name, bin := range editors {
		if !seen[bin] {
			seen[bin] = true
			result = append(result, name)
		}
	}
	return result
}
