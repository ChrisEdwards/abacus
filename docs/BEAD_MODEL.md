# Bead Data Model

This document describes the data model for beads (issues) in the bd issue tracker.

## Issue (Bead)

A bead is a trackable work item with the following fields:

### Core Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier (e.g., `ab-xyz`) |
| `title` | string | Yes | Short description of the work item (max 500 chars) |
| `description` | string | No | Detailed description of the work |
| `status` | Status | Yes | Current state of the issue |
| `priority` | int | Yes | Priority level (0-4, where 0 is highest) |
| `issue_type` | IssueType | Yes | Category of work |

### Extended Fields

| Field | Type | Description |
|-------|------|-------------|
| `design` | string | Design notes or technical approach |
| `acceptance_criteria` | string | Criteria for considering work complete |
| `notes` | string | Additional notes or context |
| `assignee` | string | Person assigned to the work |
| `estimated_minutes` | int | Time estimate in minutes |
| `external_ref` | string | External reference (e.g., `gh-9`, `jira-ABC`) |

### Timestamps

| Field | Type | Description |
|-------|------|-------------|
| `created_at` | timestamp | When the issue was created |
| `updated_at` | timestamp | When the issue was last modified |
| `closed_at` | timestamp | When the issue was closed (only set if status is `closed`) |

### Compaction Fields (Internal)

These fields are used for issue compaction (summarizing old closed issues):

| Field | Type | Description |
|-------|------|-------------|
| `compaction_level` | int | Level of compaction applied |
| `compacted_at` | timestamp | When the issue was compacted |
| `compacted_at_commit` | string | Git commit hash when compacted |
| `original_size` | int | Original size before compaction |

### Relationship Fields

These are populated during export/import operations:

| Field | Type | Description |
|-------|------|-------------|
| `labels` | []string | Tags attached to the issue |
| `dependencies` | []Dependency | Relationships to other issues |
| `comments` | []Comment | Discussion threads |

---

## Status

The current state of an issue:

| Value | Description |
|-------|-------------|
| `open` | Ready to be worked on |
| `in_progress` | Currently being worked on |
| `blocked` | Cannot proceed due to blockers |
| `closed` | Work is complete |

**Invariant**: `closed_at` timestamp must be set if and only if status is `closed`.

---

## Issue Type

Categories of work:

| Value | Description |
|-------|-------------|
| `task` | A small unit of work |
| `feature` | New functionality for users |
| `bug` | Something that's broken |
| `epic` | A large initiative with subtasks |
| `chore` | Maintenance or housekeeping |

---

## Priority

Priority levels from 0 (highest) to 4 (lowest):

| Value | Label | Description |
|-------|-------|-------------|
| 0 | Critical | Must be addressed immediately |
| 1 | High | Important, should be done soon |
| 2 | Medium | Normal priority (default) |
| 3 | Low | Can wait, nice to have |
| 4 | Backlog | Someday/maybe |

---

## Dependency

Represents a relationship between two issues:

| Field | Type | Description |
|-------|------|-------------|
| `issue_id` | string | The issue that has the dependency |
| `depends_on_id` | string | The issue being depended on |
| `type` | DependencyType | Type of relationship |
| `created_at` | timestamp | When the dependency was created |
| `created_by` | string | Who created the dependency |

### Dependency Types

| Value | Description |
|-------|-------------|
| `blocks` | Hard blocker - work cannot proceed until dependency is resolved |
| `related` | Soft link - issues are related but not blocking |
| `parent-child` | Hierarchical relationship (epic/subtask) |
| `discovered-from` | Issue was discovered while working on another |

### Dependency Semantics

- **blocks**: If A blocks B, then B cannot be started until A is closed
- **parent-child**: Creates a hierarchy where the parent (epic) tracks completion of children
- **related**: Informational link, no workflow implications
- **discovered-from**: Tracks provenance of issues found during work

---

## Label

A tag attached to an issue:

| Field | Type | Description |
|-------|------|-------------|
| `issue_id` | string | The issue with the label |
| `label` | string | The label text |

Labels can be used for filtering and organizing issues.

---

## Comment

A discussion entry on an issue:

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Unique comment identifier |
| `issue_id` | string | The issue being commented on |
| `author` | string | Who wrote the comment |
| `text` | string | The comment content |
| `created_at` | timestamp | When the comment was created |

---

## Event (Audit Trail)

Tracks changes to issues:

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Unique event identifier |
| `issue_id` | string | The issue that changed |
| `event_type` | EventType | Type of change |
| `actor` | string | Who made the change |
| `old_value` | string | Previous value (if applicable) |
| `new_value` | string | New value (if applicable) |
| `comment` | string | Optional explanation |
| `created_at` | timestamp | When the change occurred |

### Event Types

| Value | Description |
|-------|-------------|
| `created` | Issue was created |
| `updated` | Issue fields were modified |
| `status_changed` | Status was changed |
| `commented` | Comment was added |
| `closed` | Issue was closed |
| `reopened` | Issue was reopened |
| `dependency_added` | Dependency was added |
| `dependency_removed` | Dependency was removed |
| `label_added` | Label was added |
| `label_removed` | Label was removed |
| `compacted` | Issue was compacted |

---

## Epic Status

Represents an epic with its completion metrics:

| Field | Type | Description |
|-------|------|-------------|
| `epic` | Issue | The epic issue |
| `total_children` | int | Total number of child issues |
| `closed_children` | int | Number of closed children |
| `eligible_for_close` | bool | Whether all children are closed |

---

## Statistics

Aggregate metrics for the project:

| Field | Type | Description |
|-------|------|-------------|
| `total_issues` | int | Total number of issues |
| `open_issues` | int | Issues with status `open` |
| `in_progress_issues` | int | Issues with status `in_progress` |
| `closed_issues` | int | Issues with status `closed` |
| `blocked_issues` | int | Issues with status `blocked` |
| `ready_issues` | int | Open issues with no blockers |
| `epics_eligible_for_closure` | int | Epics where all children are closed |
| `average_lead_time_hours` | float | Average time from creation to closure |

---

## Sort Policies

When querying ready work, these policies determine ordering:

| Value | Description |
|-------|-------------|
| `hybrid` | Recent issues (48h) by priority, older by age (default) |
| `priority` | Always sort by priority first, then creation date |
| `oldest` | Always sort by creation date (oldest first) |

---

## Validation Rules

1. **Title**: Required, max 500 characters
2. **Priority**: Must be 0-4
3. **Status**: Must be a valid status value
4. **Issue Type**: Must be a valid issue type value
5. **Estimated Minutes**: Cannot be negative
6. **Closed At**: Must be set iff status is `closed`
