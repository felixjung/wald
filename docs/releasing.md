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

Use two separate fine-grained PATs. Do not fall back to `GITHUB_TOKEN`.

- `RELEASE_PLEASE_TOKEN` (repository access: `felixjung/wald`):
  - `Contents`: Read and write
  - `Pull requests`: Read and write
  - `Issues`: Read and write
- `HOMEBREW_TAP_GITHUB_TOKEN` (repository access: `felixjung/homebrew-tap`):
  - `Contents`: Read and write

## Automated flow

1. Push commits to `main` using Conventional Commit messages.
2. `.github/workflows/ci.yml` runs on the push.
3. If CI succeeds, `.github/workflows/release.yml` runs and performs token preflight checks.
4. release-please opens/updates a release PR that bumps versions and updates `CHANGELOG.md`.
5. After the release PR is merged and CI succeeds on that merge commit, release-please creates the tag/release.
6. The same release workflow runs GoReleaser for the new tag.
7. GoReleaser publishes release artifacts and updates the Homebrew formula in `felixjung/homebrew-tap`.

## Debug token setup

Use `.github/workflows/release-permissions-debug.yml` (`workflow_dispatch`) to validate:

- both required secrets are configured,
- each token can authenticate to GitHub API,
- `RELEASE_PLEASE_TOKEN` can access `felixjung/wald` with repository write access,
- `RELEASE_PLEASE_TOKEN` can access pull request and issue endpoints in `felixjung/wald`,
- `HOMEBREW_TAP_GITHUB_TOKEN` can access `felixjung/homebrew-tap` with repository write access.

## Local verification

- `make build` embeds local build metadata (`VERSION`, `COMMIT`, `DATE`).
- `wald version` prints embedded metadata.
