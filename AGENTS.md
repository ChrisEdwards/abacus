# Agent Development Guidelines

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

## Available Tools

### ripgrep (rg)
Fast code search tool available via command line. Common patterns:
- `rg "pattern"` - search all files
- `rg "pattern" -t go` - search only Go files
- `rg "pattern" -g "*.go"` - search files matching glob
- `rg "pattern" -l` - list matching files only
- `rg "pattern" -C 3` - show 3 lines of context

## Issue Tracking with Beads
We use beads for issue tracking and work planning. If you need more information, execute `bd quickstart`

### Dependencies
```bash
bd dep add <child> <parent> --type parent-child   # Make child a subtask of parent
bd dep add <blocked> <blocker> --type blocks      # blocker blocks blocked
bd dep remove <from> <to>                         # Remove dependency
```
**Note**: Use `bd dep add`, not `bd dep` directly. First arg depends on second arg.

### Bead ID Format
**IMPORTANT**: Always use standard bead IDs (e.g., `ab-xyz`, `ab-4aw`). Do NOT use dotted notation like `ab-4aw.1` or `ab-4aw.2` for bead names. Each bead should have its own unique ID from the beads system.

## Bead Workflow

### When Starting Work
1. **Read the bead details**: Use `bd show <bead-id>` to view the full bead information
2. **Read the comments**: Use `bd comments <bead-id>` to read all comments on the bead
   - Comments often contain important context, analysis, or clarifications
   - Prior discussions may provide insights into requirements or constraints
   - Reviewers may have added specific guidance or considerations
3. **IMPORTANT!** Set the bead status to `in_progress` when you start work

### Before Closing a Bead
You must complete ALL of the following steps before marking a bead as closed:

1. **Write Tests**: Write comprehensive tests for any code you added or changed
2. **Verify Build**: Ensure the code builds successfully (`go build`)
3. **Run Linter**: Run `make lint` and fix all issues before committing
   - Remove unused variables and styles
   - Use `//nolint:unparam` only when parameter is used in tests
4. **Run Tests**: Ensure all tests pass (`go test ./...`)
5. **Commit Changes**: Commit your changes with a detailed commit message explaining:
   - What was changed and why
   - Any architectural decisions made
   - How the changes relate to the bead requirements
   - **IMPORTANT**: Only stage and commit files you created or changed. Do not commit files modified by other agents.
6. **Push and Verify GitHub Build**: Push your changes and wait for the GitHub Actions build to pass
   - Run `git push` to push your commits
   - Check the build status in GitHub Actions
   - Do NOT close the bead until the build passes
   - If the build fails, fix the issues and repeat steps 2-6
7. **Comment on Bead**: IMPORTANT! Add a comment to the bead (`bd comments <bead-id> --add`) with:
   - Summary of what you did
   - The commit hash
   - Any relevant notes or considerations
8. **Close Bead**: Only after completing steps 1-7 (including a passing GitHub build), mark the bead as closed

### Parent Beads (Epics)
**IMPORTANT**: Do not mark parent beads as closed until ALL child beads are closed. Parent beads represent collections of work and can only be considered complete when all subtasks are finished.

### Working with Other Agents
Other AI agents may be working in this codebase in parallel. When you see modified files that you didn't change:
- Do not revert or stage those files
- Do not include them in your commits
- Only stage and commit files you created or changed
- If `git status` shows modified files you didn't touch, use `git add <specific-files>` instead of `git add .`

### Testing Beads
If you need to create or modify beads to test some functionality, do it in a bead that is a child (or descendant) of ab-cj3. That is the test beads parent.


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


## Integrating with Beads (dependency-aware task planning)

Beads provides a lightweight, dependency-aware issue database and a CLI (`bd`) for selecting "ready work," setting priorities, and tracking status. It complements MCP Agent Mail's messaging, audit trail, and file-reservation signals. Project: [steveyegge/beads](https://github.com/steveyegge/beads)

Recommended conventions
- **Single source of truth**: Use **Beads** for task status/priority/dependencies; use **Agent Mail** for conversation, decisions, and attachments (audit).
- **Shared identifiers**: Use the Beads issue id (e.g., `bd-123`) as the Mail `thread_id` and prefix message subjects with `[bd-123]`.
- **Reservations**: When starting a `bd-###` task, call `file_reservation_paths(...)` for the affected paths; include the issue id in the `reason` and release on completion.

Typical flow (agents)
1) **Pick ready work** (Beads)
   - `bd ready --json` → choose one item (highest priority, no blockers). Make sure it does not have  the `requires-human-review` label. If it does, do not work on it.
2) **Reserve edit surface** (Mail)
   - `file_reservation_paths(project_key, agent_name, ["src/**"], ttl_seconds=3600, exclusive=true, reason="bd-123")`
3) **Announce start** (Mail)
   - `send_message(..., thread_id="bd-123", subject="[bd-123] Start: <short title>", ack_required=true)`
4) **Work and update**
   - Reply in-thread with progress and attach artifacts/images; keep the discussion in one thread per issue id
5) **Complete and release**
   - `bd close bd-123 --reason "Completed"` (Beads is status authority)
   - `release_file_reservations(project_key, agent_name, paths=["src/**"])`
   - Final Mail reply: `[bd-123] Completed` with summary and links

Mapping cheat-sheet
- **Mail `thread_id`** ↔ `bd-###`
- **Mail subject**: `[bd-###] …`
- **File reservation `reason`**: `bd-###`
- **Commit messages (optional)**: include `bd-###` for traceability

Event mirroring (optional automation)
- On `bd update --status blocked`, send a high-importance Mail message in thread `bd-###` describing the blocker.
- On Mail "ACK overdue" for a critical decision, add a Beads label (e.g., `needs-ack`) or bump priority to surface it in `bd ready`.

Pitfalls to avoid
- Don't create or manage tasks in Mail; treat Beads as the single task queue.
- Always include `bd-###` in message `thread_id` to avoid ID drift across tools.

