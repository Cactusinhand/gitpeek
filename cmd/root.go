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
	dir      string
	status   string
	gotoLine bool
	dryRun   bool
)

var rootCmd = &cobra.Command{
	Use:   "gopen [commit-ref]",
	Short: "Batch open files changed in a git commit",
	Long: `gopen is a CLI tool that batch opens files from a git commit
in your preferred editor.

Examples:
  gopen HEAD                         # open files in latest commit
  gopen HEAD^                        # open files in previous commit
  gopen abc1234                      # open files in a specific commit
  gopen HEAD~3..HEAD                 # open files changed in last 3 commits
  gopen stash@{0}                    # open files in a stash
  gopen --staged                     # open staged files
  gopen --unstaged                   # open unstaged/untracked files
  gopen HEAD --terminal cursor       # open in Cursor editor
  gopen HEAD --ext .go,.ts           # only open .go and .ts files
  gopen HEAD --exclude "*_test.go"   # exclude test files
  gopen HEAD --dir src/              # only open files under src/
  gopen HEAD --status added          # only open newly added files
  gopen HEAD --goto-line             # open files at first changed line
  gopen HEAD --dry-run               # list files without opening

Config file (~/.gopenrc):
  {"terminal": "cursor", "ext": ".go", "exclude": "*_test.go"}`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
	// Load config as defaults
	cfg := internal.LoadConfig()
	if terminal == "" && cfg.Terminal != "" {
		terminal = cfg.Terminal
	}
	if ext == "" && cfg.Ext != "" {
		ext = cfg.Ext
	}
	if exclude == "" && cfg.Exclude != "" {
		exclude = cfg.Exclude
	}

	var files []string
	var err error

	// For --status filter, we need the status-aware path
	useStatusFilter := status != ""

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
		if useStatusFilter {
			changes, err := internal.GetChangedFilesWithStatus(args[0])
			if err != nil {
				return err
			}
			changes = internal.FilterChangesByStatus(changes, status)
			for _, c := range changes {
				files = append(files, c.Path)
			}
		} else {
			files, err = internal.GetChangedFiles(args[0])
			if err != nil {
				return err
			}
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
	if dir != "" {
		files = internal.FilterByDir(files, dir)
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

	if dryRun {
		return nil
	}

	// Determine editor
	editor := terminal
	if editor == "" {
		editor = internal.DetectEditor()
		if editor == "" {
			return fmt.Errorf("no editor found in PATH, use --terminal to specify one")
		}
	}

	// Open with or without goto-line
	if gotoLine && len(args) > 0 {
		lines, err := internal.GetFirstChangedLines(args[0])
		if err != nil {
			return err
		}
		if err := internal.OpenFilesWithLines(editor, files, lines); err != nil {
			return err
		}
	} else {
		if err := internal.OpenFiles(editor, files); err != nil {
			return err
		}
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
	f.BoolVar(&unstaged, "unstaged", false, "open unstaged/untracked files")
	f.StringVar(&ext, "ext", "", "filter by file extensions, comma-separated (e.g. .go,.ts)")
	f.StringVar(&exclude, "exclude", "", "exclude files matching glob pattern (e.g. *_test.go)")
	f.StringVar(&dir, "dir", "", "only open files under this directory")
	f.StringVar(&status, "status", "", "filter by change status: added, modified, deleted, renamed (or A, M, D, R)")
	f.BoolVar(&gotoLine, "goto-line", false, "open files at the first changed line")
	f.BoolVar(&dryRun, "dry-run", false, "list files without opening")
}
