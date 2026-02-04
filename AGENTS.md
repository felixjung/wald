# AGENTS.md

This repository contains the `trees` Go CLI. Follow these guidelines when making changes.

## Project structure

- `cmd/trees`: CLI entrypoints (urfave/cli v3).
  - One file per command (e.g. `add.go`, `remove.go`, `init.go`).
  - Root command in `root.go`; it wires subcommands.
  - Config subcommands live in `cmd/trees/config` with their own `root.go` and `init.go`.
- `internal/app`: Business logic. App methods live in separate files per command.
- `internal/config`: Config loading + defaults; YAML format.
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

YAML (XDG-compliant location via constants in `cmd/trees/config/config.go`):

```
worktree_root: ""
projects:
  - name: ""
    repo: ""
    workdir: ""          # optional, default "."
    default_branch: ""   # optional, default "main"
```

Use `config.Load(path string)` to load configs. Do not load a single project in isolation.

## App and worktree behavior

- `internal/app` computes paths and validates worktree existence.
- `internal/worktree` should be a thin wrapper:
  - Accepts `gitDir` and the path argument exactly as passed to `git worktree`.
  - `splitExtraArgs` enforces a leading `--` and returns the remainder unchanged.
  - No printing from `worktree` functions.
- App prints only explicit outputs (e.g. new workdir path from `Add`).

## GitHub CLI integration

- `internal/gh` uses a `GitHubCLI` struct with a `New` factory that validates:
  - `gh` availability
  - authentication
- Expose dedicated errors from `gh` for app-level handling.

## Tests and linting

- Use `github.com/stretchr/testify` assertions.
- Run tests with `go test ./...` (do not override `GOCACHE`).
- Linting is configured in `.golangci.yml` (version 2 format).
- Prefer internal tests for unexported functions (`*_internal_test.go`).

## Makefile

- Targets are hidden with `@` (keep this style).
- `build` should output `./bin/trees`.

## General Go style

- Keep functions small and focused; one command per file in `cmd/trees`.
- Avoid exporting identifiers that are not used outside the package.
