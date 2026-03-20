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

// OpenFiles opens the given files in the specified editor.
func OpenFiles(editor string, files []string) error {
	bin, ok := editors[editor]
	if !ok {
		// If not in the map, try using the value directly as a command
		bin = editor
	}

	// Verify the editor command exists
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("editor '%s' (command: '%s') not found in PATH", editor, bin)
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
