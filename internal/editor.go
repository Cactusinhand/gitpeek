package internal

import (
	"fmt"
	"os"
	"os/exec"
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

// DetectEditor tries to find an available editor, checking $EDITOR env var first,
// then common GUI editors in order of preference.
func DetectEditor() string {
	if env := os.Getenv("EDITOR"); env != "" {
		if _, err := exec.LookPath(env); err == nil {
			return env
		}
	}
	// Try common editors in order
	for _, bin := range []string{"cursor", "code", "zed", "subl", "nvim", "vim"} {
		if _, err := exec.LookPath(bin); err == nil {
			return bin
		}
	}
	return ""
}

// OpenFiles opens the given files in the specified editor.
func OpenFiles(editor string, files []string) error {
	bin, ok := editors[editor]
	if !ok {
		bin = editor
	}

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
