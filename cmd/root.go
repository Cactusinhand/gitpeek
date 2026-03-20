package cmd

import (
	"fmt"
	"os"

	"github.com/lilinchao/gopen/internal"
	"github.com/spf13/cobra"
)

var terminal string

var rootCmd = &cobra.Command{
	Use:   "gopen <commit-ref>",
	Short: "Batch open files changed in a git commit",
	Long: `gopen is a CLI tool that batch opens files from a git commit
in your preferred editor.

Examples:
  gopen HEAD                    # open files in latest commit
  gopen HEAD^                   # open files in previous commit
  gopen abc1234                 # open files in a specific commit
  gopen HEAD --terminal cursor  # open files in Cursor editor`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		commitRef := args[0]

		// Get changed files
		files, err := internal.GetChangedFiles(commitRef)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			fmt.Println("No changed files found (might be a merge commit).")
			return nil
		}

		fmt.Printf("Found %d changed file(s):\n", len(files))
		for _, f := range files {
			fmt.Printf("  %s\n", f)
		}

		// Determine editor
		editor := terminal
		if editor == "" {
			// Default to VS Code
			editor = "code"
		}

		// Open files
		if err := internal.OpenFiles(editor, files); err != nil {
			return err
		}

		fmt.Printf("Opened in %s.\n", editor)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&terminal, "terminal", "t", "", "editor to open files in (vscode, cursor, zed, sublime, vim, nvim)")
}
