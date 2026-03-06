<p align="center">
  <img src="docs/logo.png" alt="wald logo" width="560" />
</p>

# wald

`wald` is a CLI for managing Git worktrees across multiple projects from one config file.
It's another worktree CLI, but tuned for my workflow, and it may not fit yours.

## Features

- Manage multiple repositories/projects from a single config.
- Create, switch, list, and remove worktrees with interactive prompts when needed.
- Optional shell integration so `wald switch`, `wald add`, and `wald remove` can change your current shell directory.
- Optional hooks for `post-add`, `pre-remove`, `post-remove`, and global `post-switch`.

## Requirements

- `git`
- [`gh` (GitHub CLI)](https://cli.github.com/) for `wald init`
- Authenticated `gh` session (`gh auth login`) for repositories you clone
- Go `1.26+` if building from source

## Installation

### Go install

```bash
go install github.com/felixjung/wald/cmd/wald@latest
```

### Install from source (make)

```bash
git clone https://github.com/felixjung/wald.git
cd wald
make install
wald --help
```

If `/usr/local/bin` is not writable, use:

```bash
sudo make install
# or
make install INSTALL_DIR="$HOME/.local/bin"
```

## Quick Start

### 1. Create config

```bash
wald config init --worktree-root ~/worktrees
```

This writes config to:

- `$XDG_CONFIG_HOME/wald/config.toml` (or `~/.config/wald/config.toml`)
- Falls back to `~/.wald.toml` when present

### 2. Add a project

```bash
wald config add my-api --repo my-org/my-api --workdir .
```

### 3. Initialize default worktrees

```bash
wald init
```

This creates `<worktree_root>/<project_name>/<default_branch>` by cloning with `gh repo clone`.

### 4. Enable shell integration (recommended)

Add one of these to your shell rc file:

```bash
# zsh
eval "$(wald shell init zsh)"
```

```bash
# bash
eval "$(wald shell init bash)"
```

```bash
# fish
wald shell init fish | source
```

After this, `wald switch`, `wald add`, and `wald remove` can update your current directory in the same shell session.

### 5. Daily usage

```bash
# list projects/worktrees
wald list

# create and switch to a new worktree path
wald add --project my-api feature/IDT-1234

# create from a specific base ref
wald add -p my-api feature/IDT-1235 --base origin/main

# forward extra args to git worktree add (must come after --)
wald add -p my-api feature/IDT-1236 -- --track

# switch to existing worktree
wald switch -p my-api -w feature/IDT-1234

# create+switch in one step if missing
wald switch -p my-api -w feature/IDT-1237 --create --base origin/main

# remove worktree (extra args forwarded to git worktree remove)
wald remove -p my-api feature/IDT-1234 -- --force
```

## Configuration

Example `config.toml`:

```toml
worktree_root = "~/worktrees"

[theme]
name = "default"  # optional, default "default"
mode = "auto"     # optional, one of "auto", "light", "dark"

[hooks.post-switch]
01_set_title = "echo switched to {{project}}/{{worktree}}"

[[projects]]
name = "my-api"
repo = "my-org/my-api"
workdir = "."          # optional, default "."
default_branch = "main" # optional, default "main"

[projects.hooks.post-add]
01_setup = "cp .env.example .env"

[projects.hooks.pre-remove]
01_confirm = "echo removing {{worktree_path}}"

[projects.hooks.post-remove]
01_cleanup = "echo removed {{worktree}}"
```

Notes:

- `workdir` and `working-dir` overrides must be relative paths.
- `config add` supports `name`, `repo`, and `workdir`; set `default_branch` or hooks by editing `config.toml`.
- Hook commands run via `sh -c ...` in deterministic name order (sorted by hook key).
- If `[theme]` is omitted, `wald` uses `name = "default"` and `mode = "auto"`.

### Themes

- Theme files are loaded from `$XDG_CONFIG_HOME/wald/themes/<name>.toml`.
- When `XDG_CONFIG_HOME` is unset, the fallback location is `~/.config/wald/themes/<name>.toml`.
- Theme files use TOML and must include `name`, plus both `[light]` and `[dark]` color sets.
- Supported color values:
  - `default` (terminal default color)
  - ANSI names (`red`, `bright_blue`, etc.) or ANSI indexes (`0` to `255`)
  - Hex colors (`#RGB` or `#RRGGBB`)

Example theme file:

```toml
name = "solarized"
description = "Solarized-inspired palette"

[light]
title = "default"
label = "8"
label_focused = "4"
required = "1"
prompt = "8"
prompt_focused = "4"
text = "default"
text_focused = "default"
placeholder = "8"
help = "8"
error = "1"

[dark]
title = "default"
label = "8"
label_focused = "12"
required = "9"
prompt = "8"
prompt_focused = "12"
text = "default"
text_focused = "default"
placeholder = "8"
help = "8"
error = "9"
```

If a configured theme file is missing or invalid, `wald` prints a warning and falls back to the built-in `default` theme.

### Hook template variables

Available in hooks:

- `{{project}}`
- `{{worktree}}`
- `{{repo}}`
- `{{default_branch}}`
- `{{project_workdir}}`
- `{{worktree_path}}`
- `{{target_path}}`

## Commands

- `wald add <path> [-- <git worktree add args>]`
- `wald init`
- `wald list`
- `wald remove <worktree> [-- <git worktree remove args>]`
- `wald switch`
- `wald shell init <fish|zsh|bash>`
- `wald config init`
- `wald config add <name>`
- `wald version`

Run `wald <command> --help` for flags and examples.

## Development

```bash
make fmt
make lint
go test ./...
make build
```

## Troubleshooting

- `config not found ...`
  - Run `wald config init --worktree-root <path>` first.
- `cannot initialize default branches: gh CLI not found in PATH`
  - Install `gh`.
- `cannot initialize default branches: gh is not authenticated; run 'gh auth login'`
  - Authenticate with `gh auth login`.
- `extra args must start with --`
  - Pass forwarded Git args only after `--`.

## Acknowledgments

`wald` borrows inspiration from [Worktrunk](https://worktrunk.dev/), especially around hooks and related worktree workflow concepts.

## License

MIT. See [LICENSE](LICENSE).
