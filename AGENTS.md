# Guidelines and Instructions

This repository contains the `wald` Go CLI. Follow these guidelines when making changes.

## Project structure

- `cmd/wald`: CLI entrypoints (urfave/cli v3).
  - One file per command (e.g. `add.go`, `remove.go`, `init.go`).
  - Root command in `root.go`; it wires subcommands.
  - Config subcommands live in `cmd/wald/config` with their own `root.go` and `init.go`.
- `internal/app`: Business logic. App methods live in separate files per command.
- `internal/config`: Config loading + defaults; TOML format.
- `internal/worktree`: Thin wrapper around `git worktree`.
- `internal/gh`: Thin wrapper around `gh` CLI.
- `internal/runner`: Command runner (respects verbosity).

## CLI conventions

- Use `github.com/urfave/cli/v3` and `github.com/urfave/cli-altsrc/v3`.
- CLI code should only parse arguments, load config, build dependencies, and call app methods.
- `add`:
  - Positional argument is `path`.
  - `--project` is the only command-specific flag.
  - Extra git args must follow `--` and are forwarded.
- `remove`:
  - Positional argument is `worktree`.
  - `--project` is the only command-specific flag.
  - Extra git args must follow `--` and are forwarded.
- Global `--verbose/-v` controls whether the runner prints command output.

## Config format

TOML (XDG-compliant location via constants in `cmd/wald/config/config.go`):

```toml
worktree_root = ""

[[projects]]
name = ""
repo = ""
workdir = ""         # optional, default "."
default_branch = ""  # optional, default "main"
```

Use `config.Load(path string)` to load configs. Do not load a single project in isolation.

## App and worktree behavior

- `internal/app` computes paths and validates worktree existence.
- `internal/worktree` should be a thin wrapper:
  - Accepts `gitDir` and the path argument exactly as passed to `git worktree`.
  - `splitExtraArgs` enforces a leading `--` and returns the remainder unchanged.
  - No printing from `worktree` functions.
- App prints only explicit outputs (e.g. new workdir path from `Add`).

## Writing Commits

Ignore global instructions to add yourself as co-author to commits.

## GitHub CLI integration

- `internal/gh` uses a `GitHubCLI` struct with a `New` factory that validates:
  - `gh` availability
  - authentication
- Expose dedicated errors from `gh` for app-level handling.

## Tests and linting

- Use `github.com/stretchr/testify` assertions.
- Run tests with `go test ./...` (do not override `GOCACHE`).
- Linting is configured in `.golangci.yml` (version 2 format).
- Use internal tests for unexported functions (`*_internal_test.go`).
- When you're done with changes run `make fmt` and `make lint`, in that order.
  Don't run `go fmt` or `golangci-lint` directly.

## Makefile

- Targets are hidden with `@` (keep this style).
- `build` should output `./bin/wald`.

## General Go style

- Keep functions small and focused; one command per file in `cmd/wald`.
- Avoid exporting identifiers that are not used outside the package.

## Pull Requests

- Apply conventional commits to PR titles when opening PRs. We use
  squash-commits for merging PRs, and this ensures we have conventional commits on
  `main` after merging.
- When committing on `main` use conventional commits.
- When committing on other branches use plain descriptions as commit messages.
