Implement the specified bead. If one was not specified, list all beads that are open and have a `human-reviewed` label. Show them in a numbered list and ask the user to type the number or bead id of the one to implement.

**FIRST**: Follow Agent Mail protocol (AGENTS.md) - `macro_start_session`, reserve files, send start message with `thread_id` = bead ID.

Enter plan mode and make a plan. Use context7 for idiomatic library usage.
Ensure it is a delightful experience for the user.
Implement clean code, well factored. Favor simplicity, intuitiveness, and functionality.

Write tests for any code you change. Update documentation if applicable. Test with tmux if applicable.

**Before closing the bead:**
1. Commit all changes with a descriptive message
2. Push to the remote branch
3. Check GitHub build status with `gh run list --branch <branch> --limit 1`
4. Wait for the build to complete - poll every minute if still running
5. If the build **passes**: close the bead and complete
6. If the build **fails**: investigate the failure, fix it, and repeat from step 1

On completion: comment on bead with what you did to implement and include the commit hash, release file reservations, send completion message.