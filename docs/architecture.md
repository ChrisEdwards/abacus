# Architecture Documentation

Technical overview of Abacus design, architecture, and implementation.

## Table of Contents

- [Overview](#overview)
- [Architecture Principles](#architecture-principles)
- [System Architecture](#system-architecture)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Graph Building](#graph-building)
- [Rendering Pipeline](#rendering-pipeline)
- [State Management](#state-management)
- [Configuration System](#configuration-system)
- [Testing Strategy](#testing-strategy)
- [Performance Considerations](#performance-considerations)
- [Future Enhancements](#future-enhancements)

## Overview

Abacus is a terminal user interface (TUI) application built in Go that provides visualization and navigation for Beads issue tracking databases. It uses the Elm Architecture pattern via Bubble Tea to manage application state and rendering.

### Key Technologies

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework using The Elm Architecture
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - Reusable TUI components (viewport, text input)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Style definitions and layout
- **[Glamour](https://github.com/charmbracelet/glamour)** - Markdown rendering
- **[Viper](https://github.com/spf13/viper)** - Configuration management

## Architecture Principles

### 1. Separation of Concerns

The codebase is organized into logical domains:
- **UI layer** (`internal/ui/`) - Presentation and interaction
- **Domain layer** (`internal/domain/`) - Business logic and rules
- **Graph layer** (`internal/graph/`) - Dependency graph construction
- **Config layer** (`internal/config/`) - Configuration management
- **Beads client** (`internal/beads/`) - Integration with Beads CLI

### 2. Immutability

Following Elm Architecture principles:
- State changes occur through message passing
- Updates create new state rather than mutating
- Pure functions for rendering

### 3. Testability

Components are designed for easy testing:
- Business logic separated from UI
- Dependency injection for external dependencies
- Pure functions where possible

### 4. Performance

Optimizations for large issue sets:
- Lazy rendering (only visible nodes)
- Incremental updates
- Efficient graph algorithms

## System Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                         ABACUS                               │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                     UI Layer (Bubble Tea)               │ │
│  │  ┌─────────┐  ┌────────────┐  ┌──────────────────────┐ │ │
│  │  │  Tree   │  │   Detail   │  │     Search           │ │ │
│  │  │  View   │  │   Panel    │  │     Component        │ │ │
│  │  └─────────┘  └────────────┘  └──────────────────────┘ │ │
│  │                                                         │ │
│  │  ┌─────────────────────────────────────────────────────┤ │
│  │  │           Model (Application State)                 │ │
│  │  └─────────────────────────────────────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────┘ │
│                           │                                   │
│                           ▼                                   │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                  Domain Layer                           │ │
│  │  ┌──────────┐  ┌───────────┐  ┌──────────────────────┐ │ │
│  │  │  Issue   │  │  Status   │  │     Priority         │ │ │
│  │  │  Entity  │  │  Entity   │  │     Entity           │ │ │
│  │  └──────────┘  └───────────┘  └──────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────┘ │
│                           │                                   │
│                           ▼                                   │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                  Graph Layer                            │ │
│  │  ┌───────────────────┐  ┌───────────────────────────┐  │ │
│  │  │   Graph Builder   │  │   Dependency Resolver     │  │ │
│  │  └───────────────────┘  └───────────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────┘ │
│                           │                                   │
│                           ▼                                   │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                 Beads Client                            │ │
│  │  ┌──────────────────────────────────────────────────┐  │ │
│  │  │   CLI Interface (bd list, bd show)               │  │ │
│  │  └──────────────────────────────────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────┘ │
│                           │                                   │
└───────────────────────────┼───────────────────────────────────┘
                            ▼
                  ┌───────────────────┐
                  │   Beads Database  │
                  │   (.beads/)       │
                  └───────────────────┘
```

## Core Components

### 1. Main Entry Point

**Location:** `cmd/abacus/main.go`

**Responsibilities:**
- Parse command-line flags
- Initialize configuration
- Create Beads client
- Start Bubble Tea program

**Key functions:**
```go
func main()
func flagWasExplicitlySet(name string, visited map[string]struct{}) bool
```

### 2. UI Application

**Location:** `internal/ui/app.go`

**Responsibilities:**
- Implement Bubble Tea Model/Update/View pattern
- Manage application state
- Handle user input
- Coordinate rendering

**Key types:**
```go
type App struct {
    // Application state
}

func NewApp(cfg Config) (*App, error)
func (a *App) Init() tea.Cmd
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (a *App) View() string
```

### 3. Tree View

**Location:** `internal/ui/tree.go`

**Responsibilities:**
- Render issue tree
- Handle navigation
- Manage expansion state
- Apply search filters

**Key functions:**
```go
func renderTree(nodes []*graph.Node, cursor int, expanded map[string]bool) string
func findVisibleNodes(nodes []*graph.Node, expanded map[string]bool) []*graph.Node
```

### 4. Detail Panel

**Location:** `internal/ui/detail.go`

**Responsibilities:**
- Render issue details
- Format metadata
- Render markdown descriptions
- Display relationships and comments

**Key functions:**
```go
func renderDetail(issue beads.FullIssue, width, height int) string
func formatMetadata(issue beads.FullIssue) string
func renderMarkdown(content string, width int) string
```

### 5. Graph Builder

**Location:** `internal/graph/builder.go`

**Responsibilities:**
- Build dependency graph from issues
- Resolve parent-child relationships
- Identify root nodes
- Sort nodes intelligently

**Key types:**
```go
type Node struct {
    Issue    domain.Issue
    Children []*Node
    Parent   *Node
    Depth    int
}

func BuildGraph(issues []beads.FullIssue) ([]*Node, error)
func sortNodes(nodes []*Node) []*Node
```

### 6. Configuration System

**Location:** `internal/config/config.go`

**Responsibilities:**
- Load configuration from multiple sources
- Apply precedence rules
- Provide type-safe access
- Support environment variables

**Key functions:**
```go
func Initialize(opts ...Option) error
func GetString(key string) string
func GetBool(key string) bool
func GetDuration(key string) time.Duration
```

### 7. Domain Models

**Location:** `internal/domain/`

**Responsibilities:**
- Define business entities
- Enforce business rules
- Provide validation
- Type-safe operations

**Key types:**
```go
type Issue struct { ... }
type Status int
type Priority int

func NewIssue(...) (Issue, error)
func (i Issue) IsReady() bool
func (s Status) CanTransitionTo(target Status) error
```

## Data Flow

### 1. Application Startup

```
main()
  ↓
Initialize Config
  ↓
Create Beads Client
  ↓
Create UI App
  ↓
Load Issues (bd list)
  ↓
Build Graph
  ↓
Start Bubble Tea Program
  ↓
Render Initial View
```

### 2. User Interaction

```
User Input (Key Press)
  ↓
Bubble Tea Event
  ↓
App.Update(msg)
  ↓
State Change
  ↓
App.View()
  ↓
Render to Terminal
```

### 3. Auto-Refresh

```
Timer Tick (every N seconds)
  ↓
RefreshMsg
  ↓
App.Update(msg)
  ↓
Reload Issues (bd list)
  ↓
Rebuild Graph
  ↓
Preserve UI State (cursor, expanded, search)
  ↓
Re-render
```

## Graph Building

### Algorithm Overview

The graph builder transforms a flat list of issues into a hierarchical tree structure.

### Steps

1. **Parse Issues**
   - Load all issues from Beads
   - Convert to domain entities
   - Validate business rules

2. **Build Adjacency**
   - Create nodes for each issue
   - Map parent-child relationships
   - Map blocking dependencies

3. **Identify Roots**
   - Find issues with no parents
   - Handle orphaned issues
   - Resolve circular dependencies

4. **Sort Intelligently**
   - In-progress issues first
   - Ready issues second
   - Blocked issues last
   - Maintain hierarchy

### Example

**Input Issues:**
```
beads-100: Feature A (open)
beads-101: Subtask 1 (in_progress, parent: beads-100)
beads-102: Subtask 2 (blocked, parent: beads-100, blocked by: beads-101)
beads-103: Feature B (closed)
```

**Output Graph:**
```
○ beads-100: Feature A
  ◐ beads-101: Subtask 1 (in_progress)
    ⛔ beads-102: Subtask 2 (blocked)
✔ beads-103: Feature B
```

### Complexity

- **Time:** O(n log n) where n = number of issues
- **Space:** O(n) for node storage

## Rendering Pipeline

### Tree Rendering

```
1. Filter visible nodes (expanded/collapsed)
2. Apply search filter
3. Calculate viewport (what fits on screen)
4. Render visible nodes only
5. Apply syntax highlighting
6. Apply status colors
```

### Detail Rendering

```
1. Format metadata section
2. Render markdown description
   - Parse markdown (via Glamour)
   - Apply syntax highlighting
   - Wrap to width
3. Format relationships
4. Format comments
5. Combine sections
6. Apply viewport scrolling
```

## State Management

### Application State

```go
type App struct {
    // Data
    nodes     []*graph.Node
    issues    map[string]beads.FullIssue

    // UI State
    cursor    int
    expanded  map[string]bool
    search    string
    focused   FocusMode

    // Components
    viewport  viewport.Model
    input     textinput.Model

    // Config
    config    Config
}
```

### State Transitions

State changes only occur in `Update()`:

```go
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return a.handleKey(msg)
    case RefreshMsg:
        return a.handleRefresh(msg)
    case WindowSizeMsg:
        return a.handleResize(msg)
    }
    return a, nil
}
```

### Immutability

State is never mutated directly:

```go
// Wrong
a.cursor++

// Correct
newApp := *a
newApp.cursor++
return &newApp, nil
```

## Configuration System

### Configuration Sources

1. **Defaults** (hardcoded)
2. **User config** (`~/.config/abacus/config.yaml`)
3. **Project config** (`.abacus/config.yaml`)
4. **Environment variables** (`AB_*`)
5. **CLI flags**

### Precedence Example

```
Default:     refresh-interval = 3s
User config: refresh-interval = 5s
Env var:     AB_REFRESH_INTERVAL = 2s
CLI flag:    --refresh-interval 1s

Result:      1s (CLI wins)
```

### Implementation

Uses Viper for configuration management:

```go
v := viper.New()
v.SetConfigType("yaml")
setDefaults(v)
v.AutomaticEnv()
mergeConfigFile(v, userConfigPath)
mergeConfigFile(v, projectConfigPath)
```

## Testing Strategy

### Unit Tests

Test individual components in isolation:

```go
// internal/domain/issue_test.go
func TestIssue_IsReady(t *testing.T) {
    // Test business logic
}

// internal/graph/builder_test.go
func TestBuildGraph(t *testing.T) {
    // Test graph construction
}
```

### Integration Tests

Test component interaction:

```go
// internal/ui/app_test.go
func TestApp_Navigation(t *testing.T) {
    // Test navigation behavior
}
```

### Test Data

Located in `testdata/`:

```
testdata/
  ├── sample-project/
  │   └── .beads/
  │       └── beads.db
  └── fixtures/
      └── issues.json
```

## Performance Considerations

### Optimizations

1. **Lazy Rendering**
   - Only render visible nodes
   - Skip off-screen content

2. **Incremental Updates**
   - Preserve unchanged state
   - Update only what changed

3. **Efficient Sorting**
   - Cache sort keys
   - Sort once per refresh

4. **Memory Management**
   - Limit history size
   - Clear old render buffers

### Benchmarks

Run benchmarks:

```bash
make bench
```

Example results:

```
BenchmarkBuildGraph-8       1000    1.2ms/op
BenchmarkRenderTree-8       5000    0.3ms/op
BenchmarkSearchFilter-8    10000    0.1ms/op
```

## Project Structure

```
abacus/
├── cmd/
│   └── abacus/
│       └── main.go              # Entry point
├── internal/
│   ├── beads/                   # Beads CLI client
│   │   ├── client.go
│   │   └── types.go
│   ├── config/                  # Configuration
│   │   ├── config.go
│   │   └── config_test.go
│   ├── domain/                  # Business logic
│   │   ├── issue.go
│   │   ├── status.go
│   │   ├── errors.go
│   │   └── *_test.go
│   ├── graph/                   # Dependency graph
│   │   ├── builder.go
│   │   ├── node.go
│   │   ├── errors.go
│   │   └── *_test.go
│   └── ui/                      # User interface
│       ├── app.go               # Main application
│       ├── tree.go              # Tree view
│       ├── detail.go            # Detail panel
│       ├── styles.go            # Styling
│       ├── messages.go          # Bubble Tea messages
│       ├── state.go             # State types
│       ├── helpers.go           # Utilities
│       └── *_test.go
├── docs/                        # Documentation
├── testdata/                    # Test fixtures
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build automation
├── README.md                    # Project overview
└── LICENSE                      # MIT license
```

## Future Enhancements

### Planned Features

1. **Advanced Filtering**
   - Filter by status, priority, labels
   - Save and load filter presets
   - Combine multiple filters

2. **Export Capabilities**
   - Export to markdown
   - Export to JSON
   - Export to HTML

3. **Bulk Operations**
   - Select multiple issues
   - Batch status updates
   - Batch priority changes

4. **Git Integration**
   - Show commit history
   - Link issues to commits
   - Track changes over time

5. **Customization**
   - Custom themes
   - Configurable keybindings
   - Plugin system

### Performance Improvements

1. **Virtualization**
   - Virtual scrolling for very large trees
   - On-demand loading

2. **Caching**
   - Cache rendered nodes
   - Cache markdown parsing

3. **Concurrency**
   - Parallel issue loading
   - Background graph building

## Contributing

See contribution opportunities:
- Implement planned features
- Add tests
- Improve documentation
- Optimize performance
- Fix bugs

## References

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [The Elm Architecture](https://guide.elm-lang.org/architecture/)
- [Beads CLI](https://github.com/beadscli/beads)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

## Questions?

- Check the [User Guide](user-guide.md)
- Review [Troubleshooting](troubleshooting.md)
- Ask on [GitHub Discussions](https://github.com/yourusername/abacus/discussions)
