# Keyboard Shortcuts Reference

Quick reference for all keyboard shortcuts in Abacus.

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────────┐
│                    ABACUS KEYBOARD SHORTCUTS                │
├─────────────────────────────────────────────────────────────┤
│ NAVIGATION                                                  │
│   ↑/k         Move up              ←/h       Collapse node  │
│   ↓/j         Move down            →/l       Expand node    │
│   Space       Toggle expand/collapse                        │
│                                                              │
│ VIEWS                                                        │
│   Enter       Toggle detail panel  Tab       Switch focus   │
│   /           Search mode          Esc       Clear search   │
│                                                              │
│ DETAIL PANEL (when focused)                                 │
│   ↑/k         Scroll up            ↓/j       Scroll down    │
│   Ctrl+B      Page up              Ctrl+F    Page down      │
│   PgUp        Page up              PgDn      Page down      │
│   g/Home      Jump to top          G/End     Jump to bottom │
│                                                              │
│ GENERAL                                                      │
│   q           Quit                 Ctrl+C    Quit           │
└─────────────────────────────────────────────────────────────┘
```

## Navigation Shortcuts

### Tree Navigation

| Shortcut | Action | Description |
|----------|--------|-------------|
| `↓` | Move down | Move cursor to the next visible issue |
| `j` | Move down | Vim-style alternative to `↓` |
| `↑` | Move up | Move cursor to the previous visible issue |
| `k` | Move up | Vim-style alternative to `↑` |

**Tips:**
- Works in tree view (default focus)
- Automatically scrolls viewport to keep cursor visible
- Wraps at top and bottom of list

### Tree Expansion

| Shortcut | Action | Description |
|----------|--------|-------------|
| `→` | Expand | Expand the current node to show children |
| `l` | Expand | Vim-style alternative to `→` |
| `←` | Collapse | Collapse the current node to hide children |
| `h` | Collapse | Vim-style alternative to `←` |
| `Space` | Toggle | Toggle between expanded and collapsed |

**Tips:**
- Only works on nodes with children
- Collapsed nodes show `[+N]` badge with child count
- Expansion state persists during search and refresh

## View Management

### Panel Controls

| Shortcut | Action | Description |
|----------|--------|-------------|
| `Enter` | Toggle detail panel | Open/close the detail panel |
| `Tab` | Switch focus | Move focus between tree and detail panel |

**Tips:**
- Detail panel shows full information for selected issue
- Closing detail panel gives full width to tree view
- Focus switching only works when detail panel is open

### Search

| Shortcut | Action | Description |
|----------|--------|-------------|
| `/` | Enter search | Activate search mode |
| `Esc` | Clear search | Exit search and show all issues |

**Search Behavior:**
- Type to filter issues by title
- Results update instantly as you type
- Case-insensitive matching
- Press `Esc` to clear filter and return to full view

## Detail Panel Shortcuts

These shortcuts work when the detail panel is focused (press `Tab` to focus).

### Scrolling

| Shortcut | Action | Description |
|----------|--------|-------------|
| `↓` | Scroll down | Move down one line |
| `j` | Scroll down | Vim-style alternative |
| `↑` | Scroll up | Move up one line |
| `k` | Scroll up | Vim-style alternative |

### Page Navigation

| Shortcut | Action | Description |
|----------|--------|-------------|
| `Ctrl+F` | Page down | Scroll down one page |
| `PgDn` | Page down | Alternative to `Ctrl+F` |
| `Ctrl+B` | Page up | Scroll up one page |
| `PgUp` | Page up | Alternative to `Ctrl+B` |

### Jump Navigation

| Shortcut | Action | Description |
|----------|--------|-------------|
| `g` | Jump to top | Scroll to beginning of detail |
| `Home` | Jump to top | Alternative to `g` |
| `G` | Jump to bottom | Scroll to end of detail |
| `End` | Jump to bottom | Alternative to `G` |

**Tips:**
- Useful for long descriptions
- Works with markdown-rendered content
- Smooth scrolling for better readability

## General Shortcuts

| Shortcut | Action | Description |
|----------|--------|-------------|
| `q` | Quit | Exit Abacus |
| `Ctrl+C` | Quit | Force quit Abacus |

## Shortcut Contexts

Different shortcuts are available depending on context:

### Context 1: Tree View (Default)

**Active when:** Application starts or after focusing tree with `Tab`

**Available shortcuts:**
- All navigation shortcuts (`↑`, `↓`, `←`, `→`, `j`, `k`, `h`, `l`, `Space`)
- View management (`Enter`, `Tab`)
- Search (`/`, `Esc`)
- General (`q`, `Ctrl+C`)

### Context 2: Detail Panel

**Active when:** After pressing `Tab` with detail panel open

**Available shortcuts:**
- Detail panel scrolling (`↑`, `↓`, `j`, `k`)
- Page navigation (`Ctrl+F`, `Ctrl+B`, `PgUp`, `PgDn`)
- Jump navigation (`g`, `G`, `Home`, `End`)
- Switch focus (`Tab`)
- General (`q`, `Ctrl+C`)

### Context 3: Search Mode

**Active when:** After pressing `/`

**Available shortcuts:**
- Type to search
- `Esc` to clear search
- `Enter` to accept search (detail panel stays closed)
- General (`q`, `Ctrl+C`)

**Note:** Most navigation shortcuts are disabled while typing search query.

## Vim-Style Shortcuts

Abacus supports Vim-style navigation for users familiar with Vi/Vim:

| Vim Key | Standard Key | Action |
|---------|--------------|--------|
| `j` | `↓` | Down |
| `k` | `↑` | Up |
| `h` | `←` | Collapse |
| `l` | `→` | Expand |
| `g` | `Home` | Top |
| `G` | `End` | Bottom |

**Note:** Both Vim-style and standard keys work simultaneously.

## Tips and Tricks

### Tip 1: Quick Navigation

Use `j` and `k` for rapid navigation without moving your hand to arrow keys.

### Tip 2: One-Hand Operation

Most operations can be performed with one hand using Vim keys:
- `j`/`k` - Navigate
- `l`/`h` - Expand/collapse
- `/` - Search
- `q` - Quit

### Tip 3: Efficient Searching

```
/auth     → Find authentication issues
Esc       → Clear search
/feat     → Find feature issues
Esc       → Clear search
```

### Tip 4: Reading Long Descriptions

```
Enter     → Open detail panel
Tab       → Focus detail panel
g         → Jump to top
G         → Jump to bottom
Ctrl+F    → Page through content
Tab       → Return to tree
```

### Tip 5: Focus Management

The highlighted border shows which panel is focused:
- **Tree focused:** Navigate and expand issues
- **Detail focused:** Scroll through content
- Press `Tab` to switch

## Keyboard Shortcuts for Different Terminal Emulators

Some shortcuts may behave differently depending on your terminal:

### macOS Terminal

- All shortcuts work as documented
- `Ctrl+C` may show `^C` before quitting

### iTerm2

- All shortcuts work as documented
- Supports smooth scrolling

### Alacritty

- All shortcuts work as documented
- Fast rendering for smooth experience

### Windows Terminal

- Most shortcuts work as documented
- `PgUp`/`PgDn` fully supported
- May need to map `Ctrl+F`/`Ctrl+B` in settings

### tmux

- All shortcuts work as documented
- Ensure tmux prefix doesn't conflict
- Consider mapping tmux prefix to something other than `Ctrl+B`

### GNU Screen

- All shortcuts work as documented
- Screen prefix (`Ctrl+A`) doesn't conflict

## Customizing Shortcuts

Currently, Abacus doesn't support custom keyboard shortcuts. Shortcuts are hardcoded for consistency and simplicity.

If you need different keybindings, consider:
- Using terminal emulator key mapping
- Creating a feature request on GitHub
- Contributing keybinding configuration support

## Accessibility

### For Users Without Arrow Keys

Use Vim-style alternatives:
- `j`, `k`, `h`, `l` instead of arrow keys
- `g`, `G` instead of `Home`, `End`

### For Users Without Function Keys

Use alternatives for page navigation:
- `Ctrl+F`, `Ctrl+B` instead of `PgDn`, `PgUp`

### For Single-Hand Use

All essential functions accessible via left-hand keys:
- `j`/`k` - Navigate
- `l`/`h` - Expand/collapse
- `Space` - Toggle
- `/` - Search
- `Tab` - Switch
- `q` - Quit

## Learning Path

### Day 1: Basic Navigation
Learn: `↑`, `↓`, `Enter`, `q`

### Day 2: Expansion
Learn: `→`, `←`, `Space`

### Day 3: Search
Learn: `/`, `Esc`

### Day 4: Detail Panel
Learn: `Tab`, `j`, `k`, `g`, `G`

### Day 5: Efficiency
Learn: Vim alternatives (`j`, `k`, `h`, `l`)

## Printable Cheat Sheet

```
┌──────────────────────────────────────────────────┐
│              ABACUS CHEAT SHEET                  │
├──────────────────────────────────────────────────┤
│ MOVE      ↑/k ↓/j          EXPAND   →/l  Space  │
│ COLLAPSE  ←/h              DETAIL   Enter        │
│ SEARCH    /                CLEAR    Esc          │
│ FOCUS     Tab              QUIT     q            │
│                                                   │
│ DETAIL PANEL (when focused):                     │
│   SCROLL    ↑/k ↓/j        PAGE    ^F/^B  PgDn/Up│
│   TOP       g Home         BOTTOM  G End         │
└──────────────────────────────────────────────────┘
```

## Next Steps

- Practice shortcuts with the [User Guide](user-guide.md)
- Customize behavior with [Configuration](configuration.md)
- Troubleshoot issues in [Troubleshooting](troubleshooting.md)
