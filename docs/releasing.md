# Releasing wald

This repository uses Conventional Commits, release-please, and GoReleaser.

## Commit and merge conventions

- Use Conventional Commit titles in PRs (enforced by CI).
- Prefer squash merging so the merged commit message matches the PR title.
- Version bumps are driven by commit type:
  - `feat:` -> minor release
  - `fix:` and `perf:` -> patch release
  - append `!` (for example `feat!:`) or add `BREAKING CHANGE:` in the body -> major release

## Required GitHub secrets

- `HOMEBREW_TAP_GITHUB_TOKEN`: token with push access to `felixjung/homebrew-tap`.
- Optional `RELEASE_PLEASE_TOKEN`: PAT with `contents` and `pull_requests` scopes.
  If unset, workflows fall back to `GITHUB_TOKEN`.

## Automated flow

1. Push commits to `main` using Conventional Commit messages.
2. `.github/workflows/release.yml` runs release-please.
3. release-please opens/updates a release PR that bumps versions and updates `CHANGELOG.md`.
4. After the release PR is merged, release-please tags the release.
5. The same workflow runs GoReleaser for that tag.
6. GoReleaser publishes release artifacts and updates the Homebrew formula in `felixjung/homebrew-tap`.

## Local verification

- `make build` embeds local build metadata (`VERSION`, `COMMIT`, `DATE`).
- `wald version` prints embedded metadata.
