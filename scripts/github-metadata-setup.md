# GitHub Repository Metadata Setup

## ✅ Completed via gh CLI

### Repository Description
Updated to: "Interactive terminal UI for visualizing and navigating Beads issue tracking projects"

```bash
gh repo edit --description "Interactive terminal UI for visualizing and navigating Beads issue tracking projects"
```

### Repository Topics
Added topics:
- `tui`
- `go`
- `beads`
- `terminal`
- `issue-tracker`
- `bubbletea`
- `golang`
- `terminal-ui`
- `project-management`

```bash
gh repo edit --add-topic tui --add-topic go --add-topic beads --add-topic terminal --add-topic issue-tracker --add-topic bubbletea --add-topic golang --add-topic terminal-ui --add-topic project-management
```

### Repository Features
- ✅ Issues: Enabled
- ❌ Discussions: Disabled (can enable if desired)
- ❌ Wiki: Disabled (documentation is in `docs/`)

To enable discussions:
```bash
gh repo edit --enable-discussions
```

## 📋 Manual Configuration Required

The following items require manual configuration through the GitHub web interface:

### Social Preview Image
1. Navigate to: https://github.com/ChrisEdwards/abacus/settings
2. Scroll to "Social preview"
3. Click "Upload an image..."
4. Upload: `assets/abacus-preview.png` (594KB PNG)
5. Click "Save"

**Image specifications:**
- Min: 640×320px
- Max: 1280×640px
- Max file size: 1MB
- Current image: 594KB ✓

### Optional: Repository Website URL
If you want to add a homepage URL:
1. Navigate to: https://github.com/ChrisEdwards/abacus
2. Click the gear icon ⚙️ next to "About"
3. Add website URL (e.g., documentation site or project page)
4. Save changes

### Optional: Security Policy
Consider adding `.github/SECURITY.md` if not already present to define vulnerability reporting process.

### Optional: Issue/PR Templates
Consider adding:
- `.github/ISSUE_TEMPLATE/bug_report.md`
- `.github/ISSUE_TEMPLATE/feature_request.md`
- `.github/pull_request_template.md`

## Current Repository Settings

```json
{
  "description": "Interactive terminal UI for visualizing and navigating Beads issue tracking projects",
  "topics": ["beads", "bubbletea", "go", "golang", "issue-tracker", "project-management", "terminal", "terminal-ui", "tui"],
  "hasIssuesEnabled": true,
  "hasDiscussionsEnabled": false,
  "hasWikiEnabled": false
}
```

## Verification

To verify current settings:
```bash
gh repo view --json description,repositoryTopics,hasIssuesEnabled,hasDiscussionsEnabled,hasWikiEnabled
```
