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
bd dep add <parent> <child> --type parent-child   # Make child a subtask of parent
bd dep add <blocker> <blocked> --type blocks      # blocker blocks blocked
bd dep remove <from> <to>                         # Remove dependency
```
**Note**: Use `bd dep add`, not `bd dep` directly.

### Bead ID Format
**IMPORTANT**: Always use standard bead IDs (e.g., `ab-xyz`, `ab-4aw`). Do NOT use dotted notation like `ab-4aw.1` or `ab-4aw.2` for bead names. Each bead should have its own unique ID from the beads system.

## Bead Workflow

### When Starting Work
1. **Read the bead details**: Use `bd show <bead-id>` to view the full bead information
2. **Read the comments**: Use `bd comments <bead-id>` to read all comments on the bead
   - Comments often contain important context, analysis, or clarifications
   - Prior discussions may provide insights into requirements or constraints
   - Reviewers may have added specific guidance or considerations
3. Set the bead status to `in_progress` when you start work

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
6. **Comment on Bead**: Add a comment to the bead (`bd comments <bead-id> --add`) with:
   - Summary of what you did
   - The commit hash
   - Any relevant notes or considerations
7. **Close Bead**: Only after completing steps 1-6, mark the bead as closed

### Parent Beads (Epics)
**IMPORTANT**: Do not mark parent beads as closed until ALL child beads are closed. Parent beads represent collections of work and can only be considered complete when all subtasks are finished.

Other AI's may be working in this codebase in parallel on other files. Do not revert those files. Ignore them. Only stage and commit the files you changed or added.

### Testing Beads
If you need to create or modify beads to test some functionality, do it in a bead that is a child (or descendant) of ab-cj3. That is the test beads parent.