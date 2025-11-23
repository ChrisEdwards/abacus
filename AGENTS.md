# Agent Development Guidelines

## Testing Requirements
Write tests for any changes made in this codebase. All code must build successfully and all tests must pass before marking a bead as closed.

## Issue Tracking with Beads
We use beads for issue tracking and work planning. If you need more information, execute `bd quickstart`

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
3. **Run Tests**: Ensure all tests pass (`go test ./...`)
4. **Commit Changes**: Commit your changes with a detailed commit message explaining:
   - What was changed and why
   - Any architectural decisions made
   - How the changes relate to the bead requirements
5. **Comment on Bead**: Add a comment to the bead (`bd comments <bead-id> --add`) with:
   - Summary of what you did
   - The commit hash
   - Any relevant notes or considerations
6. **Close Bead**: Only after completing steps 1-5, mark the bead as closed

### Parent Beads (Epics)
**IMPORTANT**: Do not mark parent beads as closed until ALL child beads are closed. Parent beads represent collections of work and can only be considered complete when all subtasks are finished.

Other AI's may be working in this codebase in parallel on other files. Do not revert those files. Ignore them. Only stage and commit the files you changed or added.