Triage all open GitHub issues that have not yet been triaged (no type label applied).

**This command must only be run when explicitly invoked by the user via `/triage-issues`. Never run this autonomously.**

## What "un-triaged" means

An issue is un-triaged if it has none of these type labels: `bug`, `enhancement`, `documentation`, `question`.

## Process

1. Fetch all open issues with no type label:
   ```
   gh issue list --state open --json number,title,body,labels,assignees --limit 100
   ```
   Filter to those where `labels` contains none of: `bug`, `enhancement`, `documentation`, `question`.

2. For each un-triaged issue, read the full body and determine:

   **Type label** — pick exactly one:
   - `bug` — reports incorrect behaviour, crash, or unexpected output
   - `enhancement` — requests new functionality or improvement
   - `documentation` — docs-only issue
   - `question` — asks how something works

   **Area label(s)** — apply all that fit:
   - `area/tui` — tree view, navigation, filter, sort, state, key bindings
   - `area/overlays` — create, edit, status, labels, delete, comment overlays
   - `area/detail` — detail pane
   - `area/backend` — br/bd integration, data layer, SQLite, JSONL
   - `area/cli` — CLI flags, config, startup
   - `area/docs` — documentation

   **Backend label** — apply if the issue is backend-specific:
   - `backend: br`, `backend: bd`, or `backend: both`

3. Apply all determined labels:
   ```
   gh issue edit <number> --add-label "<label1>,<label2>"
   ```

4. Assign to ChrisEdwards if not already assigned:
   ```
   gh issue edit <number> --add-assignee ChrisEdwards
   ```

5. Check completeness and post a comment if needed:

   **For bugs** — is enough information present to reproduce the issue?
   Required: description of what goes wrong. Helpful but not required: steps, versions, OS.
   If the issue is too vague to act on (no description of what actually goes wrong), post a comment asking for clarification and apply `*needs-info`:
   ```
   gh issue comment <number> --body "..."
   gh issue edit <number> --add-label "*needs-info"
   ```

   **For enhancements** — is there enough context to understand what is wanted?
   If the request is too vague (e.g. just a title with no body), post a comment asking for more detail and apply `*needs-info`.

   **For questions** — no completeness check needed.

6. After processing all issues, print a summary table:
   | Issue | Title | Labels applied | Action taken |
   |-------|-------|----------------|--------------|

## Label reference

| Label | Use |
|-------|-----|
| `bug` | Incorrect behaviour |
| `enhancement` | New feature or improvement |
| `documentation` | Docs only |
| `question` | How does X work |
| `area/tui` | Tree, navigation, filter, sort |
| `area/overlays` | All overlay forms |
| `area/detail` | Detail pane |
| `area/backend` | br/bd, data layer |
| `area/cli` | Flags, config, startup |
| `area/docs` | Documentation |
| `backend: br` | br-specific |
| `backend: bd` | bd-specific |
| `backend: both` | Both backends |
| `*needs-info` | Needs more info before action |
