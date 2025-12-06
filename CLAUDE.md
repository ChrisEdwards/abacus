# Claude Development Instructions

**You MUST Read [AGENTS.md](AGENTS.md) for complete development guidelines.**

The AGENTS.md file contains essential information about:
- Testing requirements
- Issue tracking with beads
- Complete workflow for starting and completing work
- Rules for closing beads and parent beads

Always refer to AGENTS.md before beginning any work in this codebase.

## Quick Reference: bd Commands

```bash
# Adding comments - use subcommand syntax, NOT flags
bd comments add <issue-id> "comment text"   # CORRECT
bd comments <issue-id> --add "text"         # WRONG - --add is not a flag

# Labels
bd label add <issue-id> <label>
bd label remove <issue-id> <label>

# Sync - only run right before commit, not after every bead change
bd sync
```
