package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---
var (
	// Colors
	cCyan      = lipgloss.Color("39")
	cGreen     = lipgloss.Color("76")
	cRed       = lipgloss.Color("196")
	cGray      = lipgloss.Color("240")
	cLightGray = lipgloss.Color("250")
	cPurple    = lipgloss.Color("99")
	cWhite     = lipgloss.Color("255")
	cGold      = lipgloss.Color("220")
	cHighlight = lipgloss.Color("57") // Bright Purple-Blue for selection

	// Text Styles
	styleInProgress = lipgloss.NewStyle().Foreground(cCyan).Bold(true)
	styleDone       = lipgloss.NewStyle().Foreground(cGray)
	styleReady      = lipgloss.NewStyle().Foreground(cGreen)
	styleBlocked    = lipgloss.NewStyle().Foreground(cRed)
	styleID         = lipgloss.NewStyle().Foreground(cGold).Bold(true).MarginRight(1)
	
	// New Label Style (Pill look)
	styleLabel = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(lipgloss.Color("25")). // Dark Blue pill
			Padding(0, 1).
			MarginRight(1).
			Bold(true)

	// Updated Selection Style (High Visibility)
	styleSelected = lipgloss.NewStyle().
			Background(cHighlight).
			Foreground(cWhite).
			Bold(true)

	// Layout Styles
	styleHeader = lipgloss.NewStyle().Foreground(cWhite).Background(cPurple).Bold(true).Padding(0, 1)
	stylePane   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(cGray)
)

// --- Data Structures ---

type LiteIssue struct {
	ID string `json:"id"`
}

type Comment struct {
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type FullIssue struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Description  string    `json:"description"`
	CreatedAt    string    `json:"created_at"`
	Labels       []string  `json:"labels"`
	Comments     []Comment `json:"comments"` // Added comments support
	
	Dependencies []struct {
		Type     string `json:"type"`
		TargetID string `json:"target_id"`
	} `json:"dependencies"`

	Dependents []struct {
		ID string `json:"id"`
	} `json:"dependents"`
}

type Node struct {
	Issue     FullIssue
	Children  []*Node
	Parent    *Node
	BlockedBy []*Node
	IsBlocked bool

	Expanded      bool
	Depth         int
	HasInProgress bool
	HasReady      bool
}

// --- Model ---

type model struct {
	roots       []*Node
	visibleRows []*Node
	cursor      int

	viewport    viewport.Model
	showDetails bool
	ready       bool

	textInput  textinput.Model
	searching  bool
	filterText string

	width  int
	height int
	err    error
}

// --- Logic: Data Loading (Same as before) ---

func loadData() ([]*Node, error) {
	cmd := exec.Command("bd", "list", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run bd list: %v", err)
	}

	var liteIssues []LiteIssue
	if err := json.Unmarshal(out, &liteIssues); err != nil {
		return nil, err
	}

	if len(liteIssues) == 0 {
		return []*Node{}, nil
	}

	var ids []string
	for _, l := range liteIssues {
		ids = append(ids, l.ID)
	}

	fullIssues, err := batchFetchIssues(ids)
	if err != nil {
		return nil, err
	}

	return buildGraph(fullIssues), nil
}

func batchFetchIssues(ids []string) ([]FullIssue, error) {
	var results []FullIssue
	chunkSize := 20 

	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}

		batchIDs := ids[i:end]
		args := append([]string{"show"}, batchIDs...)
		args = append(args, "--json")

		cmd := exec.Command("bd", args...)
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to run bd show batch: %v", err)
		}

		var batchResult []FullIssue
		if err := json.Unmarshal(out, &batchResult); err != nil {
			return nil, fmt.Errorf("failed to parse batch json: %v", err)
		}
		results = append(results, batchResult...)
	}
	return results, nil
}

func buildGraph(issues []FullIssue) []*Node {
	nodeMap := make(map[string]*Node)
	for _, iss := range issues {
		nodeMap[iss.ID] = &Node{Issue: iss}
	}

	var roots []*Node
	childrenIDs := make(map[string]bool)

	for id, node := range nodeMap {
		for _, dep := range node.Issue.Dependencies {
			if dep.Type == "parent-child" {
				if parent, ok := nodeMap[dep.TargetID]; ok {
					if !isChild(parent, node) {
						parent.Children = append(parent.Children, node)
						node.Parent = parent
						childrenIDs[id] = true
					}
				}
			}
		}

		for _, dep := range node.Issue.Dependents {
			if child, ok := nodeMap[dep.ID]; ok {
				if !isChild(node, child) {
					node.Children = append(node.Children, child)
					child.Parent = node
					childrenIDs[child.Issue.ID] = true
				}
			}
		}
	}

	for _, node := range nodeMap {
		for _, dep := range node.Issue.Dependencies {
			if dep.Type == "blocks" {
				if blocker, ok := nodeMap[dep.TargetID]; ok {
					if blocker.Issue.Status != "closed" {
						node.BlockedBy = append(node.BlockedBy, blocker)
						node.IsBlocked = true
					}
				}
			}
		}
	}

	for id, node := range nodeMap {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Issue.CreatedAt < node.Children[j].Issue.CreatedAt
		})

		if !childrenIDs[id] {
			roots = append(roots, node)
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Issue.CreatedAt < roots[j].Issue.CreatedAt
	})

	for _, root := range roots {
		computeStates(root)
		if root.HasInProgress {
			root.Expanded = true
		}
	}

	return roots
}

func isChild(parent, potentialChild *Node) bool {
	for _, c := range parent.Children {
		if c.Issue.ID == potentialChild.Issue.ID {
			return true
		}
	}
	return false
}

func computeStates(n *Node) {
	if n.Issue.Status == "in_progress" {
		n.HasInProgress = true
	}
	if n.Issue.Status == "open" && !n.IsBlocked {
		n.HasReady = true
	}

	for _, child := range n.Children {
		child.Depth = n.Depth + 1
		computeStates(child)
		if child.HasInProgress {
			n.HasInProgress = true
			n.Expanded = true
		}
		if child.HasReady {
			n.HasReady = true
		}
	}
}

func (m *model) recalcVisibleRows() {
	m.visibleRows = []*Node{}

	matches := func(n *Node) bool {
		if m.filterText == "" {
			return true
		}
		return strings.Contains(strings.ToLower(n.Issue.Title), strings.ToLower(m.filterText))
	}

	var traverse func(nodes []*Node)
	traverse = func(nodes []*Node) {
		for _, node := range nodes {
			isMatch := matches(node)
			hasMatchingChild := false
			if m.filterText != "" {
				var checkChildren func([]*Node) bool
				checkChildren = func(kids []*Node) bool {
					for _, k := range kids {
						if matches(k) || checkChildren(k.Children) { return true }
					}
					return false
				}
				hasMatchingChild = checkChildren(node.Children)
			}

			shouldShow := true
			if m.filterText != "" {
				shouldShow = isMatch || hasMatchingChild
			}

			if shouldShow {
				m.visibleRows = append(m.visibleRows, node)
				if (m.filterText == "" && node.Expanded) || (m.filterText != "" && hasMatchingChild) {
					traverse(node.Children)
				}
			}
		}
	}

	traverse(m.roots)
}

func (m *model) jumpToBlocker() {
	if m.cursor >= len(m.visibleRows) { return }
	node := m.visibleRows[m.cursor]
	if len(node.BlockedBy) == 0 { return }
	target := node.BlockedBy[0]
	curr := target.Parent
	for curr != nil {
		curr.Expanded = true
		curr = curr.Parent
	}
	if m.filterText != "" {
		m.filterText = ""
		m.textInput.SetValue("")
		m.searching = false
	}
	m.recalcVisibleRows()
	for i, n := range m.visibleRows {
		if n.Issue.ID == target.Issue.ID {
			m.cursor = i
			m.updateViewportContent()
			return
		}
	}
}

func initialModel() model {
	roots, err := loadData()
	ti := textinput.New()
	ti.Placeholder = "Filter..."
	ti.Prompt = "/ "
	ti.CharLimit = 50

	m := model{
		roots:     roots,
		err:       err,
		textInput: ti,
		searching: false,
	}
	m.recalcVisibleRows()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.viewport.Width = int(float64(msg.Width)*0.4) - 2
		m.viewport.Height = msg.Height - 4
		m.updateViewportContent()

	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "enter", "esc":
				m.searching = false
				m.textInput.Blur()
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				m.filterText = m.textInput.Value()
				m.recalcVisibleRows()
				m.cursor = 0
				return m, cmd
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			m.searching = true
			m.textInput.Focus()
			return m, textinput.Blink
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateViewportContent()
			}
		case "down", "j":
			if m.cursor < len(m.visibleRows)-1 {
				m.cursor++
				m.updateViewportContent()
			}
		case "right", "l", "space":
			node := m.visibleRows[m.cursor]
			if len(node.Children) > 0 {
				node.Expanded = !node.Expanded
				m.recalcVisibleRows()
			}
		case "enter":
			m.showDetails = !m.showDetails
			m.updateViewportContent()
		case "b":
			m.jumpToBlocker()
		case "ctrl+j":
			m.viewport.LineDown(1)
		case "ctrl+k":
			m.viewport.LineUp(1)
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// --- UPDATED VIEWPORT RENDERER ---

func (m *model) updateViewportContent() {
	if !m.showDetails || m.cursor >= len(m.visibleRows) {
		return
	}
	node := m.visibleRows[m.cursor]
	iss := node.Issue

	sb := strings.Builder{}

	// 1. Header
	sb.WriteString(fmt.Sprintf("# %s\n", iss.Title))
	sb.WriteString(fmt.Sprintf("**ID**: `%s`  |  **Status**: `%s`  |  **Created**: `%s`\n", iss.ID, iss.Status, iss.CreatedAt))

	// 2. Labels (Standalone line, Pill style)
	if len(iss.Labels) > 0 {
		sb.WriteString("\n")
		for _, l := range iss.Labels {
			sb.WriteString(styleLabel.Render(l))
		}
		sb.WriteString("\n")
	}
	
	// 3. Context Box (Relationships)
	sb.WriteString("\n---\n")
	
	// Parent: ID First
	if node.Parent != nil {
		p := node.Parent.Issue
		sb.WriteString(fmt.Sprintf("**Parent**: `%s` %s\n", p.ID, p.Title))
	}

	// Blockers
	if node.IsBlocked {
		sb.WriteString("**Blocked By**:\n")
		for _, b := range node.BlockedBy {
			sb.WriteString(fmt.Sprintf("- `%s` %s\n", b.Issue.ID, b.Issue.Title))
		}
	}

	// Children
	if len(node.Children) > 0 {
		sb.WriteString(fmt.Sprintf("**Children**: %d items\n", len(node.Children)))
	}
	sb.WriteString("---\n\n")

	// 4. Description
	desc := strings.ReplaceAll(iss.Description, "• ", "- ")
	sb.WriteString(desc)

	// 5. Comments (New Section)
	if len(iss.Comments) > 0 {
		sb.WriteString("\n\n---\n### 💬 Comments\n")
		for _, c := range iss.Comments {
			sb.WriteString(fmt.Sprintf("**%s** (%s):\n%s\n\n", c.Author, c.CreatedAt, c.Body))
		}
	}

	// Render Markdown
	width := int(float64(m.width)*0.4) - 4
	if width < 10 {
		width = 10
	}

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)

	str, err := renderer.Render(sb.String())
	if err != nil {
		str = "Error rendering markdown"
	}

	m.viewport.SetContent(str)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	if !m.ready {
		return "Initializing..."
	}

	header := ""
	if m.searching {
		header = styleHeader.Render("SEARCH") + " " + m.textInput.View()
	} else {
		status := fmt.Sprintf("Items: %d", len(m.visibleRows))
		if m.filterText != "" {
			status += fmt.Sprintf(" (Filtered: '%s')", m.filterText)
		}
		header = styleHeader.Render("ABACUS") + " " + status
	}

	treeView := strings.Builder{}
	listHeight := m.height - 3
	start := 0
	end := len(m.visibleRows)

	if end > listHeight {
		if m.cursor > listHeight/2 {
			start = m.cursor - listHeight/2
		}
		if start+listHeight < end {
			end = start + listHeight
		} else {
			start = end - listHeight
			if start < 0 { start = 0 }
		}
	}

	for i := start; i < end; i++ {
		node := m.visibleRows[i]
		indent := strings.Repeat("  ", node.Depth)
		marker := " "
		if len(node.Children) > 0 {
			if node.Expanded {
				marker = "▼"
			} else {
				marker = "▶"
			}
		} else {
			marker = "•"
		}

		icon := "○"
		s := styleReady
		if node.Issue.Status == "in_progress" {
			icon = "🔥"
			s = styleInProgress
		} else if node.Issue.Status == "closed" {
			icon = "✔"
			s = styleDone
		} else if node.IsBlocked {
			icon = "⛔"
			s = styleBlocked
		}

		idStr := styleID.Render(node.Issue.ID)
		titleStr := s.Render(node.Issue.Title)

		line := fmt.Sprintf("%s%s %s %s %s", indent, marker, icon, idStr, titleStr)

		if i == m.cursor {
			// New High-Visibility Selection
			line = styleSelected.Render(fmt.Sprintf(" %s%s %s %s %s ", indent, marker, icon, node.Issue.ID, node.Issue.Title))
		} else {
			line = " " + line
		}

		treeView.WriteString(line + "\n")
	}

	var mainBody string
	if m.showDetails {
		left := stylePane.
			Width(m.width - m.viewport.Width - 4).
			Height(m.height - 3).
			Render(treeView.String())

		right := stylePane.
			Width(m.viewport.Width).
			Height(m.height - 3).
			Render(m.viewport.View())

		mainBody = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	} else {
		mainBody = stylePane.
			Width(m.width - 2).
			Height(m.height - 3).
			Render(treeView.String())
	}

	footerText := " [ / ] Search  [ b ] Blockers  [ enter ] Toggle Details  [ space ] Expand  [ q ] Quit"
	footer := lipgloss.NewStyle().Foreground(cLightGray).Render(footerText)

	return fmt.Sprintf("%s\n%s\n%s", header, mainBody, footer)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}