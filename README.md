# gitpeek

Batch open files changed in a git commit, in your preferred editor.

## Install

```bash
go install github.com/Cactusinhand/gitpeek@latest
```

Or build from source:

```bash
git clone https://github.com/Cactusinhand/gitpeek.git
cd gitpeek
go build -o gitpeek .
```

## Usage

```bash
# Open files in latest commit
gitpeek HEAD

# Open files in previous commit
gitpeek HEAD^

# Open files in a specific commit
gitpeek abc1234

# Open files changed in last 3 commits
gitpeek HEAD~3..HEAD

# Open files in a stash
gitpeek stash@{0}

# Open staged files
gitpeek --staged

# Open unstaged/untracked files
gitpeek --unstaged
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--terminal` | `-t` | Editor to use (vscode, cursor, zed, sublime, vim, nvim) |
| `--ext` | | Filter by file extensions, comma-separated (e.g. `.go,.ts`) |
| `--exclude` | | Exclude files matching glob pattern (e.g. `*_test.go`) |
| `--dir` | | Only open files under this directory |
| `--status` | | Filter by change status: `added`, `modified`, `deleted`, `renamed` |
| `--goto-line` | | Open files at the first changed line |
| `--dry-run` | | List files without opening |
| `--version` | `-v` | Show version |

## Examples

```bash
# Open only .go files, exclude tests
gitpeek HEAD --ext .go --exclude "*_test.go"

# Open newly added files under src/
gitpeek HEAD --status added --dir src/

# Preview which files would be opened
gitpeek HEAD --dry-run

# Open files at their first changed line in Cursor
gitpeek HEAD --goto-line --terminal cursor
```

## Config

Create `~/.gitpeekrc` to set defaults:

```json
{
  "terminal": "cursor",
  "ext": ".go,.ts",
  "exclude": "*_test.go"
}
```

CLI flags override config values.

## Supported Editors

| Name | Command | Goto-line support |
|------|---------|-------------------|
| VS Code | `code` | `-g file:line` |
| Cursor | `cursor` | `-g file:line` |
| Zed | `zed` | `file:line` |
| Sublime Text | `subl` | `file:line` |
| Neovim | `nvim` | `+line file` |
| Vim | `vim` | `+line file` |

If `--terminal` is not specified, gitpeek auto-detects by checking `$EDITOR`, then trying editors in order: cursor, code, zed, subl, nvim, vim.

## License

MIT
