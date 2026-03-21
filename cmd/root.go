package cmd

import (
	"fmt"
	"os"

	"github.com/lilinchao/gopen/internal"
	"github.com/spf13/cobra"
)

var (
	terminal string
	staged   bool
	unstaged bool
	ext      string
	exclude  string
	dryRun   bool
)

var rootCmd = &cobra.Command{
	Use:   "gopen [commit-ref]",
	Short: "Batch open files changed in a git commit",
	Long: `gopen is a CLI tool that batch opens files from a git commit
in your preferred editor.

Examples:
  gopen HEAD                        # open files in latest commit
  gopen HEAD^                       # open files in previous commit
  gopen abc1234                     # open files in a specific commit
  gopen HEAD~3..HEAD                # open files changed in last 3 commits
  gopen --staged                    # open staged files
  gopen --unstaged                  # open unstaged modified files
  gopen HEAD --terminal cursor      # open in Cursor editor
  gopen HEAD --ext .go,.ts          # only open .go and .ts files
  gopen HEAD --exclude "*_test.go"  # exclude test files
  gopen HEAD --dry-run              # list files without opening`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
	var files []string
	var err error

	switch {
	case staged:
		files, err = internal.GetStagedFiles()
		if err != nil {
			return err
		}
	case unstaged:
		files, err = internal.GetUnstagedFiles()
		if err != nil {
			return err
		}
	default:
		if len(args) == 0 {
			return fmt.Errorf("requires a commit reference (e.g. HEAD) or use --staged/--unstaged")
		}
		files, err = internal.GetChangedFiles(args[0])
		if err != nil {
			return err
		}
	}

	if len(files) == 0 {
		fmt.Println("No changed files found.")
		return nil
	}

	// Apply filters
	if ext != "" {
		files = internal.FilterByExt(files, ext)
	}
	if exclude != "" {
		files = internal.FilterByExclude(files, exclude)
	}

	if len(files) == 0 {
		fmt.Println("No files match the given filters.")
		return nil
	}

	// Print file list
	fmt.Printf("Found %d file(s):\n", len(files))
	for _, f := range files {
		fmt.Printf("  %s\n", f)
	}

	// Dry run: stop here
	if dryRun {
		return nil
	}

	// Determine editor
	editor := terminal
	if editor == "" {
		editor = "code"
	}

	if err := internal.OpenFiles(editor, files); err != nil {
		return err
	}

	fmt.Printf("Opened in %s.\n", editor)
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	f := rootCmd.Flags()
	f.StringVarP(&terminal, "terminal", "t", "", "editor to open files in (vscode, cursor, zed, sublime, vim, nvim)")
	f.BoolVar(&staged, "staged", false, "open staged (git add) files")
	f.BoolVar(&unstaged, "unstaged", false, "open unstaged modified files")
	f.StringVar(&ext, "ext", "", "filter by file extensions, comma-separated (e.g. .go,.ts)")
	f.StringVar(&exclude, "exclude", "", "exclude files matching glob pattern (e.g. *_test.go)")
	f.BoolVar(&dryRun, "dry-run", false, "list files without opening")
}
