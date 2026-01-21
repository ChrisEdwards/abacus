READ ~/Users/chrisedwards~/projects/chris/agent-shared/AGENTS.md BEFORE ANYTHING (skip if missing).

# Agent Development Guidelines

## RULE 1 – ABSOLUTE (DO NOT EVER VIOLATE THIS)

You may NOT delete any file or directory unless I explicitly give the exact command **in this session**.

- This includes files you just created (tests, tmp files, scripts, etc.).
- You do not get to decide that something is "safe" to remove.
- If you think something should be removed, stop and ask. You must receive clear written approval **before** any deletion command is even proposed.

Treat "never delete files without permission" as a hard invariant.

---

## IRREVERSIBLE GIT & FILESYSTEM ACTIONS

Absolutely forbidden unless I give the **exact command and explicit approval** in the same message:

- `git reset --hard`
- `git clean -fd`
- `rm -rf`
- Any command that can delete or overwrite code/data

Rules:

1. If you are not 100% sure what a command will delete, do not propose or run it. Ask first.
2. Prefer safe tools: `git status`, `git diff`, `git stash`, copying to backups, etc.
3. After approval, restate the command verbatim, list what it will affect, and wait for confirmation.
4. When a destructive command is run, record in your response:
   - The exact user text authorizing it
   - The command run
   - When you ran it

If that audit trail is missing, then you must act as if the operation never happened.

---

## Code Editing Discipline

- Do **not** run scripts that bulk-modify code (codemods, invented one-off scripts, giant `sed`/regex refactors).
- Large mechanical changes: break into smaller, explicit edits and review diffs.
- Subtle/complex changes: edit by hand, file-by-file, with careful reasoning.

---

## Backwards Compatibility & File Sprawl

We optimize for a clean architecture now, not backwards compatibility.

- No "compat shims" or "v2" file clones.
- When changing behavior, migrate callers and remove old code.
- New files are only for genuinely new domains that don't fit existing modules.
- The bar for adding files is very high.

---

## Development Commands

Use these make targets for all checks and tests:

```bash
make check            # Run linting and static analysis (quiet output)
make test             # Run unit tests only (fast, excludes integration tests)
make test-integration # Run integration tests only (requires bd/br binaries)
make test-all         # Run all tests (unit + integration)
make check-test       # Run checks and unit tests

# Verbose output when debugging failures
make check VERBOSE=1
make test VERBOSE=1
make test-integration VERBOSE=1
```

Integration tests are separated using Go build tags (`//go:build integration`).
They require the `bd` and/or `br` binaries to be installed.

## Quick Reference: bd Commands

```bash
# Adding comments - use subcommand syntax, NOT flags
bd comments add <issue-id> "comment text"   # CORRECT
bd comments <issue-id> --add "text"         # WRONG - --add is not a flag

# Labels
bd label add <issue-id> <label>
bd label remove <issue-id> <label>
```

## Go Code Size Limits

Keep code modular and maintainable by respecting these limits:

| Metric | Production | Test Files |
|--------|------------|------------|
| File length | 500 lines | 800 lines |
| Function length | 60 lines | 80 lines |
| Function statements | 40 | 60 |
| Line length | 120 chars | 120 chars |
| Cyclomatic complexity | 10 | 15 |

If a file or function exceeds these limits, decompose it:
- Split files by domain concept, lifecycle, or abstraction level
- Extract helper functions for repeated logic
- Use Go naming conventions: `_view.go`, `_handlers.go`, `_types.go`

## Testing Requirements
Write tests for any changes made in this codebase. All code must build successfully, pass linting, and all tests must pass before marking a bead as closed.

### TUI Design Principles
When building UI features, follow the design principles in [`docs/UI_PRINCIPLES.md`](docs/UI_PRINCIPLES.md). This includes:
- Visual hierarchy and consistent styling
- Toast and overlay patterns
- Context-aware footer hints
- Hotkey design guidelines

### TUI Visual Testing
When making UI changes, use `scripts/tui-test.sh` to verify the layout visually:
```bash
make build                         # Build first after code changes
./scripts/tui-test.sh start        # Launch in tmux
./scripts/tui-test.sh keys 'jjjl'  # Navigate (j=down, l=expand)
./scripts/tui-test.sh enter        # Open detail pane
./scripts/tui-test.sh view         # Capture current state
./scripts/tui-test.sh quit         # Clean up
```
Always verify UI changes look correct before marking work complete.

### Testing with br Backend

The abacus repository itself uses `bd` (beads Go CLI), but abacus supports both `bd` and `br` (beads_rust) backends. To test br functionality, you must use a separate directory:

```bash
# 1. Create a test directory outside the abacus repo
mkdir /tmp/abacus-br-test && cd /tmp/abacus-br-test

# 2. Initialize a br workspace
br init

# 3. Create test beads
br create "Test bead 1" --type task --priority 2
br create "Test bead 2" --type epic --priority 1

# 4. Configure for br backend
mkdir -p .abacus
cat > .abacus/config.yaml << 'EOF'
beads:
  backend: br
EOF

# 5. Run abacus from that directory
/path/to/abacus/bin/abacus
```

The TUI will show `[br]` in the status bar when using the br backend.

**Why a separate directory?**
- The abacus repo has `.beads/` configured for bd
- br and bd have different database schemas
- Testing br in the abacus repo would corrupt the beads database
- A separate `/tmp` directory keeps tests isolated

**Verified br operations (via TUI):**
- Create overlay (n) - creates beads via `br create`
- Edit overlay (e) - updates beads via `br update`
- Status change (s) - changes status via `br update --status`
- Delete (Del) - deletes beads via `br delete`
- Labels (L) - adds/removes labels via `br label add/remove`
- Comments (m) - adds comments via `br comments add`

## Available Tools

### ripgrep (rg)
Fast code search tool available via command line. Common patterns:
- `rg "pattern"` - search all files
- `rg "pattern" -t go` - search only Go files
- `rg "pattern" -g "*.go"` - search files matching glob
- `rg "pattern" -l` - list matching files only
- `rg "pattern" -C 3` - show 3 lines of context

### ast-grep (sg)
Structural code search using AST patterns. Use when text search is fragile (formatting varies, need semantic matches).
```bash
sg -p 'func $NAME($$$) { $$$BODY }' -l go       # Find functions
sg -p '$VAR.Error()' -l go                      # Find method calls
```

### Go LSP (gopls)
Available via the LSP tool for code intelligence: `goToDefinition`, `findReferences`, `hover`, `documentSymbol`, `workspaceSymbol`, `goToImplementation`, `incomingCalls`, `outgoingCalls`.

### Bash etiquette
- Do not prefix bash commands with comments. 

---

## MCP Agent Mail: coordination for multi-agent workflows

What it is
- A mail-like layer that lets coding agents coordinate asynchronously via MCP tools and resources.
- Provides identities, inbox/outbox, searchable threads, and advisory file reservations, with human-auditable artifacts in Git.

Why it's useful
- Prevents agents from stepping on each other with explicit file reservations (leases) for files/globs.
- Keeps communication out of your token budget by storing messages in a per-project archive.
- Offers quick reads (`resource://inbox/...`, `resource://thread/...`) and macros that bundle common flows.

How to use effectively
1) Same repository
   - Register an identity: call `ensure_project`, then `register_agent` using this repo's absolute path as `project_key`.
   - Reserve files before you edit: `file_reservation_paths(project_key, agent_name, ["src/**"], ttl_seconds=3600, exclusive=true)` to signal intent and avoid conflict.
   - Communicate with threads: use `send_message(..., thread_id="FEAT-123")`; check inbox with `fetch_inbox` and acknowledge with `acknowledge_message`.
   - Read fast: `resource://inbox/{Agent}?project=<abs-path>&limit=20` or `resource://thread/{id}?project=<abs-path>&include_bodies=true`.
   - Tip: set `AGENT_NAME` in your environment so the pre-commit guard can block commits that conflict with others' active exclusive file reservations.

2) Across different repos in one project (e.g., Next.js frontend + FastAPI backend)
   - Option A (single project bus): register both sides under the same `project_key` (shared key/path). Keep reservation patterns specific (e.g., `frontend/**` vs `backend/**`).
   - Option B (separate projects): each repo has its own `project_key`; use `macro_contact_handshake` or `request_contact`/`respond_contact` to link agents, then message directly. Keep a shared `thread_id` (e.g., ticket key) across repos for clean summaries/audits.

Macros vs granular tools
- Prefer macros when you want speed or are on a smaller model: `macro_start_session`, `macro_prepare_thread`, `macro_file_reservation_cycle`, `macro_contact_handshake`.
- Use granular tools when you need control: `register_agent`, `file_reservation_paths`, `send_message`, `fetch_inbox`, `acknowledge_message`.

Common pitfalls
- "from_agent not registered": always `register_agent` in the correct `project_key` first.
- "FILE_RESERVATION_CONFLICT": adjust patterns, wait for expiry, or use a non-exclusive reservation when appropriate.
- Auth errors: if JWT+JWKS is enabled, include a bearer token with a `kid` that matches server JWKS; static bearer is used only when JWT is disabled.

---

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

---

## Issue Tracking with Beads

We use beads for issue tracking and work planning. If you need more information, execute `bd quickstart`

**IMPORTANT**: Beads (`bd` CLI) is a third-party tool we do not maintain. Do not propose changes to the beads codebase. The beads source may be in a sibling folder for reference, but we cannot modify it.

**Related Codebases (read-only reference):**
- **beads (Go)**: `../beads` - Original Go implementation (`bd` CLI)
- **beads_rust**: `../beads_rust` - Rust port (`br` CLI)

### Dependencies
```bash
bd dep add <child> <parent> --type parent-child   # Make child a subtask of parent
bd dep add <blocked> <blocker> --type blocks      # blocker blocks blocked
bd dep remove <from> <to>                         # Remove dependency
```
**Note**: Use `bd dep add`, not `bd dep` directly. First arg depends on second arg.

### Bead ID Format
**IMPORTANT**: Always use standard bead IDs (e.g., `ab-xyz`, `ab-4aw`). Do NOT use dotted notation like `ab-4aw.1` or `ab-4aw.2` for bead names. Each bead should have its own unique ID from the beads system.

### Bead Workflow

#### When Starting Work
1. **Read the bead details**: Use `bd show <bead-id>` to view the full bead information
2. **Read the comments**: Use `bd comments <bead-id>` to read all comments on the bead
   - Comments often contain important context, analysis, or clarifications
   - Prior discussions may provide insights into requirements or constraints
   - Reviewers may have added specific guidance or considerations
3. **IMPORTANT!** Set the bead status to `in_progress` when you start work

#### Before Closing a Bead
You must complete ALL of the following steps before marking a bead as closed:

1. **Write Tests**: Write comprehensive tests for any code you added or changed
2. **Run Checks and Tests**: Run `make check-test` and fix all issues before committing
   - Remove unused variables and styles
   - Use `//nolint:unparam` only when parameter is used in tests
3. If you discover new work, create a new bead with `discovered-from:<parent-id>`.
4. **Commit Changes**: Only commit files you created or changed (use `git add <specific-files>`, not `git add .`)
5. Commit `.beads/` in the same commit as code changes.
6. **Push and Verify GitHub Build**: Push and wait for GitHub Actions build to pass before closing
7. **Comment on Bead**: Add a comment with summary and commit hash
8. **Close Bead**: Only after sucessful push

### Auto-sync:
- bd exports to `.beads/issues.jsonl` after changes (debounced).
- It imports from JSONL when newer (e.g. after `git pull`).

### Never:
- Use markdown TODO lists.
- Use other trackers.
- Duplicate tracking.

### Parent Beads (Epics)
**IMPORTANT**: Do not mark parent beads as closed until ALL child beads are closed. Parent beads represent collections of work and can only be considered complete when all subtasks are finished.

### Working with Other Agents
Other agents may be working in parallel. Only commit files you created or changed - ignore other modified files.

### Testing Beads
If you need to create or modify beads to test some functionality, do it in a bead that is a child (or descendant) of ab-cj3. That is the test beads parent.

---

### Using bv as an AI sidecar

bv is a graph-aware triage engine for Beads projects (.beads/beads.jsonl). Instead of parsing JSONL or hallucinating graph traversal, use robot flags for deterministic, dependency-aware outputs with precomputed metrics (PageRank, betweenness, critical path, cycles, HITS, eigenvector, k-core).

**Scope boundary:** bv handles *what to work on* (triage, priority, planning). For agent-to-agent coordination (messaging, work claiming, file reservations), use [MCP Agent Mail](https://github.com/Dicklesworthstone/mcp_agent_mail).

**⚠️ CRITICAL: Use ONLY `--robot-*` flags. Bare `bv` launches an interactive TUI that blocks your session.**

#### The Workflow: Start With Triage

**`bv --robot-triage` is your single entry point.** It returns everything you need in one call:
- `quick_ref`: at-a-glance counts + top 3 picks
- `recommendations`: ranked actionable items with scores, reasons, unblock info
- `quick_wins`: low-effort high-impact items
- `blockers_to_clear`: items that unblock the most downstream work
- `project_health`: status/type/priority distributions, graph metrics
- `commands`: copy-paste shell commands for next steps

bv --robot-triage        # THE MEGA-COMMAND: start here
bv --robot-next          # Minimal: just the single top pick + claim command

#### Other Commands

**Planning:**
| Command | Returns |
|---------|---------|
| `--robot-plan` | Parallel execution tracks with `unblocks` lists |
| `--robot-priority` | Priority misalignment detection with confidence |

**Graph Analysis:**
| Command | Returns |
|---------|---------|
| `--robot-insights` | Full metrics: PageRank, betweenness, HITS (hubs/authorities), eigenvector, critical path, cycles, k-core, articulation points, slack |
| `--robot-label-health` | Per-label health: `health_level` (healthy\|warning\|critical), `velocity_score`, `staleness`, `blocked_count` |
| `--robot-label-flow` | Cross-label dependency: `flow_matrix`, `dependencies`, `bottleneck_labels` |
| `--robot-label-attention [--attention-limit=N]` | Attention-ranked labels by: (pagerank × staleness × block_impact) / velocity |

**History & Change Tracking:**
| Command | Returns |
|---------|---------|
| `--robot-history` | Bead-to-commit correlations: `stats`, `histories` (per-bead events/commits/milestones), `commit_index` |
| `--robot-diff --diff-since <ref>` | Changes since ref: new/closed/modified issues, cycles introduced/resolved |

**Other Commands:**
| Command | Returns |
|---------|---------|
| `--robot-burndown <sprint>` | Sprint burndown, scope changes, at-risk items |
| `--robot-forecast <id\|all>` | ETA predictions with dependency-aware scheduling |
| `--robot-alerts` | Stale issues, blocking cascades, priority mismatches |
| `--robot-suggest` | Hygiene: duplicates, missing deps, label suggestions, cycle breaks |
| `--robot-graph [--graph-format=json\|dot\|mermaid]` | Dependency graph export |
| `--export-graph <file.html>` | Self-contained interactive HTML visualization |

#### Scoping & Filtering

bv --robot-plan --label backend              # Scope to label's subgraph
bv --robot-insights --as-of HEAD~30          # Historical point-in-time
bv --recipe actionable --robot-plan          # Pre-filter: ready to work (no blockers)
bv --recipe high-impact --robot-triage       # Pre-filter: top PageRank scores
bv --robot-triage --robot-triage-by-track    # Group by parallel work streams
bv --robot-triage --robot-triage-by-label    # Group by domain

#### Understanding Robot Output

**All robot JSON includes:**
- `data_hash` — Fingerprint of source beads.jsonl (verify consistency across calls)
- `status` — Per-metric state: `computed|approx|timeout|skipped` + elapsed ms
- `as_of` / `as_of_commit` — Present when using `--as-of`; contains ref and resolved SHA

**Two-phase analysis:**
- **Phase 1 (instant):** degree, topo sort, density — always available immediately
- **Phase 2 (async, 500ms timeout):** PageRank, betweenness, HITS, eigenvector, cycles — check `status` flags

**For large graphs (>500 nodes):** Some metrics may be approximated or skipped. Always check `status`.

#### jq Quick Reference

bv --robot-triage | jq '.quick_ref'                        # At-a-glance summary
bv --robot-triage | jq '.recommendations[0]'               # Top recommendation
bv --robot-plan | jq '.plan.summary.highest_impact'        # Best unblock target
bv --robot-insights | jq '.status'                         # Check metric readiness
bv --robot-insights | jq '.Cycles'                         # Circular deps (must fix!)
bv --robot-label-health | jq '.results.labels[] | select(.health_level == "critical")'

**Performance:** Phase 1 instant, Phase 2 async (500ms timeout). Prefer `--robot-plan` over `--robot-insights` when speed matters. Results cached by data hash.

Use bv instead of parsing beads.jsonl—it computes PageRank, critical paths, cycles, and parallel tracks deterministically.

---

### cass — Cross-Agent Search

`cass` indexes prior agent conversations (Claude Code, Codex, Cursor, Gemini, ChatGPT, etc.) so we can reuse solved problems.

Rules:

- Never run bare `cass` (TUI). Always use `--robot` or `--json`.

Examples:

```bash
cass health
cass search "authentication error" --robot --limit 5
cass view /path/to/session.jsonl -n 42 --json
cass expand /path/to/session.jsonl -n 42 -C 3 --json
cass capabilities --json
cass robot-docs guide
```

Tips:

- Use `--fields minimal` for lean output.
- Filter by agent with `--agent`.
- Use `--days N` to limit to recent history.

stdout is data-only, stderr is diagnostics; exit code 0 means success.

Treat cass as a way to avoid re-solving problems other agents already handled.

---

## Memory System: cass-memory

The Cass Memory System (cm) is a tool for giving agents an effective memory based on the ability to quickly search across previous coding agent sessions across an array of different coding agent tools (e.g., Claude Code, Codex, Gemini-CLI, Cursor, etc) and projects (and even across multiple machines, optionally) and then reflect on what they find and learn in new sessions to draw out useful lessons and takeaways; these lessons are then stored and can be queried and retrieved later, much like how human memory works.

The `cm onboard` command guides you through analyzing historical sessions and extracting valuable rules.

### Quick Start

```bash
# 1. Check status and see recommendations
cm onboard status

# 2. Get sessions to analyze (filtered by gaps in your playbook)
cm onboard sample --fill-gaps

# 3. Read a session with rich context
cm onboard read /path/to/session.jsonl --template

# 4. Add extracted rules (one at a time or batch)
cm playbook add "Your rule content" --category "debugging"
# Or batch add:
cm playbook add --file rules.json

# 5. Mark session as processed
cm onboard mark-done /path/to/session.jsonl
```

Before starting complex tasks, retrieve relevant context:

```bash
cm context "<task description>" --json
```

This returns:
- **relevantBullets**: Rules that may help with your task
- **antiPatterns**: Pitfalls to avoid
- **historySnippets**: Past sessions that solved similar problems
- **suggestedCassQueries**: Searches for deeper investigation

### Protocol

1. **START**: Run `cm context "<task>" --json` before non-trivial work
2. **WORK**: Reference rule IDs when following them (e.g., "Following b-8f3a2c...")
3. **FEEDBACK**: Leave inline comments when rules help/hurt:
   - `// [cass: helpful b-xyz] - reason`
   - `// [cass: harmful b-xyz] - reason`
4. **END**: Just finish your work. Learning happens automatically.

### Key Flags

| Flag | Purpose |
|------|---------|
| `--json` | Machine-readable JSON output (required!) |
| `--limit N` | Cap number of rules returned |
| `--no-history` | Skip historical snippets for faster response |

stdout = data only, stderr = diagnostics. Exit 0 = success.


---

## UBS Quick Reference for AI Agents

UBS stands for "Ultimate Bug Scanner": **The AI Coding Agent's Secret Weapon: Flagging Likely Bugs for Fixing Early On**

**Golden Rule:** `ubs <changed-files>` before every commit. Exit 0 = safe. Exit >0 = fix & re-run.

**Commands:**
```bash
ubs file.ts file2.py                    # Specific files (< 1s) — USE THIS
ubs $(git diff --name-only --cached)    # Staged files — before commit
ubs --only=js,python src/               # Language filter (3-5x faster)
ubs --ci --fail-on-warning .            # CI mode — before PR
ubs --help                              # Full command reference
ubs sessions --entries 1                # Tail the latest install session log
ubs .                                   # Whole project (ignores things like .venv and node_modules automatically)
```

**Output Format:**
```
⚠️  Category (N errors)
    file.ts:42:5 – Issue description
    💡 Suggested fix
Exit code: 1
```
Parse: `file:line:col` → location | 💡 → how to fix | Exit 0/1 → pass/fail

**Fix Workflow:**
1. Read finding → category + fix suggestion
2. Navigate `file:line:col` → view context
3. Verify real issue (not false positive)
4. Fix root cause (not symptom)
5. Re-run `ubs <file>` → exit 0
6. Commit

**Speed Critical:** Scope to changed files. `ubs src/file.ts` (< 1s) vs `ubs .` (30s). Never full scan for small edits.

**Bug Severity:**
- **Critical** (always fix): Null safety, XSS/injection, async/await, memory leaks
- **Important** (production): Type narrowing, division-by-zero, resource leaks
- **Contextual** (judgment): TODO/FIXME, console logs

**Anti-Patterns:**
- ❌ Ignore findings → ✅ Investigate each
- ❌ Full scan per edit → ✅ Scope to file
- ❌ Fix symptom (`if (x) { x.y }`) → ✅ Root cause (`x?.y`)