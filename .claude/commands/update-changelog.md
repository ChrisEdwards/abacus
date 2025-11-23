You are tasked with preparing a release by analyzing changes since the last release and generating a comprehensive changelog with an appropriate semantic version bump.

## Your Task

1. **Gather information about changes since last release:**
   - Get the last git tag: `git describe --tags --abbrev=0 2>/dev/null`
   - Get all commits since that tag: `git log --oneline --no-decorate LAST_TAG..HEAD`
   - Get all closed beads since the last release: `bd list --status=closed`
   - Analyze which files were changed: `git diff --name-only LAST_TAG..HEAD`

2. **Read existing context:**
   - Read the `[Unreleased]` section in CHANGELOG.md
   - If there are manual entries, treat them as important context (don't discard them)
   - Read the current `next-version.txt` to see what version is currently planned

3. **Check if this command has already been run:**
   - Compare `next-version.txt` content with the last git tag
   - Check if `next-version.txt` was modified in the most recent commit
   - If it appears `/update-changelog` was run without committing, acknowledge this and ask if user wants to regenerate

4. **Analyze changes and determine semantic version bump:**

   Consider ALL of the following factors:

   - **Commit message patterns:**
     - `BREAKING CHANGE:` or `!` → MAJOR
     - `feat:` or new features → MINOR
     - `fix:` or bug fixes → PATCH
     - `chore:`, `docs:`, `refactor:` → typically PATCH

   - **Bead analysis:**
     - Closed bugs → PATCH
     - Closed features → MINOR
     - Breaking changes or API redesigns → MAJOR

   - **Code changes:**
     - New public APIs or commands → MINOR
     - Changes to existing APIs/interfaces → MAJOR
     - Internal refactoring only → PATCH
     - Bug fixes → PATCH

   - **Manual CHANGELOG entries:**
     - If user wrote "Breaking: ..." → MAJOR
     - If user added major new features → MINOR

   **Important semver rules:**
   - Before 1.0.0: Major version 0 signals "initial development", breaking changes can be MINOR
   - After 1.0.0: Breaking changes MUST be MAJOR
   - When in doubt between MINOR and PATCH: choose MINOR
   - When in doubt between MAJOR and MINOR: ask the user

5. **Generate Keep a Changelog format entries:**

   Create well-organized changelog entries under these categories (only include non-empty ones):
   - `### Added` - New features, capabilities
   - `### Changed` - Changes to existing functionality
   - `### Fixed` - Bug fixes
   - `### Removed` - Removed features
   - `### Security` - Security improvements

   **Guidelines for good changelog entries:**
   - Write from user perspective (what they experience, not implementation details)
   - Start with action verbs (Add, Fix, Change, Update, Remove)
   - Be concise but descriptive
   - Include bead IDs where relevant: `- Feature description (ab-123)`
   - One logical change per line
   - Multiple beads can contribute to one changelog line if they're related
   - Merge insights from: manual entries + commits + beads + code analysis

6. **Present your recommendation:**

   Show the user:
   ```
   ## Proposed Version Bump

   Current version: X.Y.Z (from last tag)
   Recommended: A.B.C (PATCH|MINOR|MAJOR)

   Rationale:
   - [Explain why this version bump based on the analysis]

   ## Generated Changelog

   ### Added
   - [entries]

   ### Changed
   - [entries]

   ### Fixed
   - [entries]

   [etc.]
   ```

7. **Get user confirmation:**

   Ask: "Does this version bump and changelog look correct? I'll update next-version.txt and CHANGELOG.md accordingly."

   If user approves:
   - Update `next-version.txt` with the new version number (just the number, e.g., `0.2.0`)
   - Update the `[Unreleased]` section in CHANGELOG.md with the generated entries
   - Show a git diff of the changes
   - Remind user: "Please review the changes and commit them when ready. Then run `./scripts/release.sh --execute` to complete the release."

   If user wants changes:
   - Ask what they'd like adjusted
   - Regenerate and present again

## Important Notes

- You are NOT committing anything - user commits manually after review
- You are preparing files for the release, not performing the release
- Be thoughtful about version bumps - they communicate API stability to users
- The changelog should tell a story of what changed from the user's perspective
- Quality over quantity - merge related changes into clear, concise entries

Begin your analysis now.
