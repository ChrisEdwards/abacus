# Contributing to Abacus

## How This Project Works

Abacus is maintained via AI-assisted implementation. **The primary contribution is a well-specified issue** — not a pull request. A precise issue with clear acceptance criteria can be implemented directly; a vague one cannot.

PRs are not the expected contribution path. If you'd like to see something built, write a detailed issue.

## Filing a Bug Report

Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.yml) template. Include:

- Your Abacus version, backend version, OS, and terminal
- Exact steps to reproduce, starting from launching `abacus`
- What you expected vs. what happened
- Which component is involved, if you know (e.g. "the create overlay", "detail pane")

The more context you provide, the faster the fix.

## Filing a Feature Request

Use the [Feature Request](.github/ISSUE_TEMPLATE/feature_request.yml) template. A useful feature request includes:

- **The problem**, not just the solution — why do you need this?
- **Precise desired behavior** — specific key bindings, UI changes, output format
- **Acceptance criteria** — explicit checkable conditions for "done"
- **What's out of scope** — what you don't want, to prevent scope creep

Vague requests ("it would be nice to have X") are hard to act on. Specific ones ("pressing `f` in the list view should filter issues by label, showing a fuzzy-search overlay, dismissible with Esc") are directly implementable.

## Scope

Abacus is a **read/write TUI for Beads issue databases**. It is intentionally focused on that and nothing else.

Out of scope: general project management features, non-Beads backends, web or GUI interfaces, features that duplicate what `br`/`bd` already provide on the command line.

**br (beads_rust) is the active backend.** New features target br first. bd support is frozen at v0.38.0 — bug fixes are accepted, new bd-only features are unlikely to be merged.

## How Issues Are Closed

Issues are closed for the following reasons, indicated by a label:

| Label | Meaning |
|-------|---------|
| `*duplicate` | Already tracked — see the linked issue |
| `*as-designed` | The behaviour is intentional — a comment explains why |
| `*not-reproducible` | Could not reproduce with the information provided |
| `*out-of-scope` | Outside the project's stated scope |
| `*needs-info` | Closed after requesting more information with no response (14 days) |

If your issue was closed and you disagree, comment with additional context and it will be reconsidered.

## Questions

Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) and the [README](README.md) first. If you're still stuck, open an issue with the `question` label.
