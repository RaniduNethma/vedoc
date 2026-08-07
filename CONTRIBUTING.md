# Contributing to Vedoc

Vedoc uses a small, review-first pull request workflow so changes stay easy to reason about and easy to merge.

## Source of truth

`RaniduNethma/vedoc:main` is the canonical branch. Fork branches are temporary work branches, not long-lived integration branches.

Before starting a new pull request, sync your fork and create the branch from the current upstream `main`.

```bash
git remote add upstream https://github.com/RaniduNethma/vedoc.git 2>/dev/null || true
git fetch upstream
git switch main
git reset --hard upstream/main
git push origin main --force-with-lease
git switch -c <type>/<short-scope>
```

Do not start dependent work from an unmerged feature branch unless the maintainers explicitly agree to a stacked-PR workflow.

## Pull request rules

1. Keep each PR focused on one meaningful concern.
2. Do not include unrelated refactors, formatting churn, or cleanup.
3. Run the relevant validation before pushing.
4. Open the PR against `RaniduNethma/vedoc:main`.
5. Address review findings on the same branch and PR.
6. Dependent work starts only after its prerequisite PR is merged into upstream `main`.
7. Maintainers own the final merge decision. Contributors should not merge their own PR unless explicitly delegated.
8. GitHub is the evidence trail: commits, diff, CI, and review discussion are sufficient. Do not add separate evidence documents unless a change genuinely needs documentation.

## Validation

For Go changes, run `gofmt -w` on every Go file you changed, then run:

```bash
go mod verify
go vet ./...
go test ./...
CGO_ENABLED=1 go build ./...
```

For npm wrapper changes, also run:

```bash
node --check npm-wrapper/bin/vedoc.js
node --check npm-wrapper/install.js
(
  cd npm-wrapper
  npm pack --dry-run
)
```

The pull request CI repeats repository-wide correctness checks and enforces formatting on Go files changed by the PR. This avoids mixing pre-existing formatting cleanup into unrelated work while preventing new formatting debt.

A green CI run is required for a change to be considered ready for review.

## After merge

Treat the merged upstream `main` as the new baseline. Sync the fork again, retire the merged branch, and create the next dependent branch from the updated upstream `main`.
