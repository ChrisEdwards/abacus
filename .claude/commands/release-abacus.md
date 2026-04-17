You are performing a full release of abacus. Walk through each step below in order. Run everything you can automatically; pause and ask the user only when a decision or confirmation is needed.

---

## Step 1 — Pre-flight checks

Run these automatically and report pass/fail for each:

```bash
# 1a. On main branch?
git branch --show-current

# 1b. Up to date with origin?
git fetch origin && git status

# 1c. Working directory clean?
git status --short

# 1d. Tests pass?
make test

# 1e. Build succeeds?
make build
```

If anything fails, stop and tell the user what needs fixing before continuing.

---

## Step 2 — Analyze changes since last release

Run automatically:

```bash
# What was the last release?
git describe --tags --abbrev=0

# What changed since then?
git log --oneline --no-decorate $(git describe --tags --abbrev=0)..HEAD

# Which files changed?
git diff --name-only $(git describe --tags --abbrev=0)..HEAD

# What's in the [Unreleased] section already?
# Read CHANGELOG.md

# What version is planned?
cat next-version.txt
```

Also read the closed beads since the last tag date:
```bash
br list --status=closed --json
```

Use the commit messages, changed files, and closed beads to determine:
- The right **semantic version bump** (patch / minor / major) with rationale
- **Changelog entries** in Keep a Changelog format (Added / Changed / Fixed / Removed)

Write from the user's perspective. Reference bead IDs where relevant. Be concise.

Present your recommendation:

```
## Proposed Release

Current version: X.Y.Z (last tag)
Proposed version: A.B.C (PATCH|MINOR|MAJOR)

Rationale: [one sentence]

### Added
- ...

### Changed
- ...

### Fixed
- ...
```

**Ask the user:** "Does this version and changelog look right? Any edits before I write the files?"

Wait for confirmation or edits before proceeding.

---

## Step 3 — Write changelog files

Once the user approves, update these files:

1. **`next-version.txt`** — write the new version number (just `X.Y.Z`, no `v` prefix)
2. **`CHANGELOG.md`** — populate the `[Unreleased]` section with the approved entries (do not finalize the heading yet — `release.sh` does that)

Show a `git diff` of both files, then ask: "Looks good? I'll commit these and move on."

---

## Step 4 — Commit changelog prep

```bash
git add CHANGELOG.md next-version.txt
git commit -m "chore: prepare release vX.Y.Z"
git push origin main
```

Confirm push succeeded before moving on.

---

## Step 5 — Dry run

Show the user what the release script will do:

```bash
./scripts/release.sh --dry-run
```

Present the output and ask: "Ready to execute the release?"

---

## Step 6 — Execute the release

Run:

```bash
./scripts/release.sh --execute
```

The script will:
- Finalize `CHANGELOG.md` (`[Unreleased]` → `[X.Y.Z] - YYYY-MM-DD`)
- Create a release commit and annotated tag `vX.Y.Z`
- Ask you to confirm before pushing (answer **y** when prompted)
- Push commit + tag → triggers GitHub Actions
- Auto-bump `next-version.txt` to next patch and push

---

## Step 7 — Monitor CI

After the push, check the release workflow:

```bash
# Watch for the release workflow run
gh run list --workflow release.yml --limit 3

# Get the run ID of the triggered release and watch it
gh run watch <run-id>
```

Wait for the workflow to complete. It runs two jobs:
- **GoReleaser** — builds cross-platform binaries and creates the GitHub Release
- **Update Homebrew** — updates `ChrisEdwards/homebrew-tap`

If either job fails, report the error and the relevant logs.

---

## Step 8 — Verify the release

Once CI passes, run these checks automatically:

```bash
# 8a. Release exists on GitHub?
gh release view vX.Y.Z

# 8b. All expected assets present?
gh release view vX.Y.Z --json assets --jq '[.assets[].name] | sort'
```

Expected assets: `darwin_amd64`, `darwin_arm64`, `linux_amd64`, `linux_arm64`, `windows_amd64` tarballs + `checksums.txt`.

Report the results and give the user a final summary:

```
✓ Release vX.Y.Z published
✓ Binaries available for darwin/linux/windows
✓ Homebrew formula updated (or: needs manual check)

Release: https://github.com/ChrisEdwards/abacus/releases/tag/vX.Y.Z
Actions: https://github.com/ChrisEdwards/abacus/actions
```

---

## Important rules

- Never skip a step — each one gates the next.
- Never force-push or delete tags without explicit user approval.
- If `release.sh --execute` fails partway through, report exactly what happened and what state git is in before suggesting recovery steps.
- If CI fails, show the failing job logs and ask the user how to proceed.
