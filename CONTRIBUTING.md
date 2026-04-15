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

Abacus is a TUI viewer for Beads issue tracking databases. It is intentionally focused. Before filing a feature request, consider whether it fits that scope.

**BR (beads_rust) is the active backend.** New features target BR first. BD support is frozen at v0.38.0 — bug fixes are accepted, new BD-only features are unlikely to be merged.

## Questions

Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) and the [README](README.md) first. If you're still stuck, open an issue with the `question` label.
