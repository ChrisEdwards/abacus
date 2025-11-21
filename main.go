package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---
var (
	// Theme Colors
	cPurple    = lipgloss.Color("99")
	cCyan      = lipgloss.Color("39")
	cNeonGreen = lipgloss.Color("118") // Icon only
	cRed       = lipgloss.Color("196")
	cOrange    = lipgloss.Color("208")
	cGold      = lipgloss.Color("220")
	cGray      = lipgloss.Color("240")
	cLightGray = lipgloss.Color("250")
	cWhite     = lipgloss.Color("255")
	cHighlight = lipgloss.Color("57")
	cField     = lipgloss.Color("63") // Consistent color for field names (Blue-ish purple)

	// Tree Text Styles
	styleInProgressText = lipgloss.NewStyle().Foreground(cCyan).Bold(true)
	styleNormalText     = lipgloss.NewStyle().Foreground(cWhite)
	styleDoneText       = lipgloss.NewStyle().Foreground(cGray)
	styleBlockedText    = lipgloss.NewStyle().Foreground(cRed)

	// Tree Icon Styles (Separate from text)
	styleIconOpen       = lipgloss.NewStyle().Foreground(cWhite)
	styleIconInProgress = lipgloss.NewStyle().Foreground(cNeonGreen) // THE NEON GREEN ICON
	styleIconDone       = lipgloss.NewStyle().Foreground(cGray)
	styleIconBlocked    = lipgloss.NewStyle().Foreground(cRed)

	styleID = lipgloss.NewStyle().Foreground(cGold).Bold(true).MarginRight(1)

	// Selection Style
	styleSelected = lipgloss.NewStyle().
			Background(cHighlight).
			Foreground(cWhite).
			Bold(true)

	// Header & Layout
	styleAppHeader = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(cPurple).
			Bold(true).
			Padding(0, 1)

	stylePane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(cGray)

	// --- Detail Pane Specifics ---
	
	// The merged ID + Title bar
	styleDetailHeader = lipgloss.NewStyle().
			Background(cHighlight).
			Foreground(cWhite).
			Bold(true).
			Padding(0, 1)

	// Metadata Fields
	styleField = lipgloss.NewStyle().Foreground(cField).Width(10) // Fixed width for alignment
	styleVal   = lipgloss.NewStyle().Foreground(cWhite)
	
	styleSectionHeader = lipgloss.NewStyle().
			Foreground(cGold).
			Bold(true).
			MarginTop(1).
			MarginBottom(0)

	styleLabel = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(lipgloss.Color("25")).
			Padding(0, 1).
			MarginRight(1).
			Bold(true)
			
	stylePrio = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(cOrange).
			Padding(0, 1).
			Bold(true)
)

// --- Data Structures (Unchanged) ---

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
	Priority     int       `json:"priority"`
	Description  string    `json:"description"`
	CreatedAt    string    `json:"created_at"`
	Labels       []string  `json:"labels"`
	Comments     []Comment `json:"comments"`
	
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
	repoName    string

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

// --- Logic: Helpers ---

func formatTime(isoStr string) string {
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return isoStr
	}
	return t.Local().Format("Jan 02, 3:04 PM")
}

func getRepoName() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return filepath.Base(dir)
}

// --- Logic: Data Loading (Unchanged) ---

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
		repoName:  getRepoName(),
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

// --- REFINED Viewport Renderer ---

func (m *model) updateViewportContent() {
	if !m.showDetails || m.cursor >= len(m.visibleRows) {
		return
	}
	node := m.visibleRows[m.cursor]
	iss := node.Issue

	// 1. HEADER: [ID] Title (Full width background)
	// We construct a single string for the header content
	headerContent := fmt.Sprintf("%s %s", 
		styleID.Render(iss.ID), 
		lipgloss.NewStyle().Foreground(cWhite).Render(iss.Title),
	)
	// Apply the background style to the whole block, ensuring full width
	headerBlock := styleDetailHeader.Width(m.viewport.Width - 2).Render(headerContent)

	// 2. METADATA GRID (Fixed width keys for alignment)
	// Row 1
	statusKey := styleField.Render("Status:")
	statusVal := styleVal.Render(iss.Status)
	
	prioKey := styleField.Render("Priority:")
	prioVal := stylePrio.Render(fmt.Sprintf("P%d", iss.Priority))

	// Row 2
	createdKey := styleField.Render("Created:")
	createdVal := styleVal.Render(formatTime(iss.CreatedAt))
	
	// Labels (if any)
	labelsVal := ""
	if len(iss.Labels) > 0 {
		var pills []string
		for _, l := range iss.Labels {
			pills = append(pills, styleLabel.Render(l))
		}
		labelsVal = strings.Join(pills, "")
	} else {
		labelsVal = styleVal.Render("-")
	}
	labelsKey := styleField.Render("Labels:")

	// Construct Rows
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, statusKey, statusVal, "    ", prioKey, prioVal)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, createdKey, createdVal, "    ", labelsKey, labelsVal)
	
	metaBlock := lipgloss.JoinVertical(lipgloss.Left, row1, row2)
	metaBlock = lipgloss.NewStyle().Padding(1, 1).Render(metaBlock)

	// 3. RELATIONSHIPS (Lipgloss, no markdown artifacts)
	relBuilder := strings.Builder{}
	
	if node.Parent != nil {
		relBuilder.WriteString(styleSectionHeader.Render("Parent") + "\n")
		relBuilder.WriteString(fmt.Sprintf("%s %s\n", styleID.Render(node.Parent.Issue.ID), node.Parent.Issue.Title))
	}

	if node.IsBlocked {
		relBuilder.WriteString(styleSectionHeader.Render("Blocked By") + "\n")
		for _, b := range node.BlockedBy {
			relBuilder.WriteString(fmt.Sprintf("%s %s\n", styleID.Render(b.Issue.ID), b.Issue.Title))
		}
	}

	if len(node.Children) > 0 {
		relBuilder.WriteString(styleSectionHeader.Render(fmt.Sprintf("Children (%d)", len(node.Children))) + "\n")
		// Optional: List first few children? For now just the header is clean.
	}
	
	relBlock := ""
	if relBuilder.Len() > 0 {
		relBlock = lipgloss.NewStyle().Padding(0, 1).Render(relBuilder.String())
	}

	// 4. DESCRIPTION (Glamour)
	desc := strings.ReplaceAll(iss.Description, "• ", "- ")
	
	// Render Markdown
	width := int(float64(m.width)*0.4) - 6
	if width < 10 { width = 10 }

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	
	bodyStr, _ := renderer.Render(desc)
	
	// Add Comments if they exist
	if len(iss.Comments) > 0 {
		commentBuilder := strings.Builder{}
		commentBuilder.WriteString("\n### Comments\n")
		for _, c := range iss.Comments {
			commentBuilder.WriteString(fmt.Sprintf("**%s** (%s):\n%s\n\n", c.Author, formatTime(c.CreatedAt), c.Body))
		}
		commentsStr, _ := renderer.Render(commentBuilder.String())
		bodyStr += "\n" + commentsStr
	}

	// Assemble Final
	finalContent := lipgloss.JoinVertical(lipgloss.Left,
		headerBlock,
		metaBlock,
		relBlock,
		bodyStr,
	)

	m.viewport.SetContent(finalContent)
}

// --- Main View ---

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	if !m.ready {
		return "Initializing..."
	}

	// --- Header ---
	header := ""
	if m.searching {
		header = styleAppHeader.Render("SEARCH") + " " + m.textInput.View()
	} else {
		status := fmt.Sprintf("Items: %d", len(m.visibleRows))
		if m.filterText != "" {
			status += fmt.Sprintf(" (Filtered: '%s')", m.filterText)
		}
		header = styleAppHeader.Render("ABACUS") + " " + status
	}

	// --- Tree View ---
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

		// Icon vs Text Logic
		iconStr := "○"
		iconStyle := styleNormalText
		textStyle := styleNormalText // Default white for ready

		if node.Issue.Status == "in_progress" {
			iconStr = "◐"
			iconStyle = styleIconInProgress // Neon Green
			textStyle = styleInProgressText // Cyan
		} else if node.Issue.Status == "closed" {
			iconStr = "✔"
			iconStyle = styleIconDone
			textStyle = styleDoneText
		} else if node.IsBlocked {
			iconStr = "⛔"
			iconStyle = styleIconBlocked
			textStyle = styleBlockedText
		}

		// Elements
		idStr := styleID.Render(node.Issue.ID)
		titleStr := textStyle.Render(node.Issue.Title)
		markerStr := iconStyle.Render(marker + " " + iconStr)

		// Line Construction
		rawContent := fmt.Sprintf("%s%s %s %s", indent, markerStr, idStr, titleStr)

		if i == m.cursor {
			treeView.WriteString(styleSelected.Render(" " + rawContent))
		} else {
			treeView.WriteString(" " + rawContent)
		}
		treeView.WriteString("\n")
	}

	// --- Pane Assembly ---
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

	// --- Footer ---
	leftFooter := lipgloss.NewStyle().Foreground(cLightGray).Render(" [ / ] Search  [ b ] Blockers  [ enter ] Detail  [ space ] Expand  [ q ] Quit")
	rightFooter := lipgloss.NewStyle().Foreground(cLightGray).Render("Repo: " + m.repoName + " ")
	
	// Calculate spacer
	availableWidth := m.width - lipgloss.Width(leftFooter) - lipgloss.Width(rightFooter)
	if availableWidth < 0 { availableWidth = 0 }
	spacer := strings.Repeat(" ", availableWidth)
	
	footer := lipgloss.JoinHorizontal(lipgloss.Top, leftFooter, spacer, rightFooter)

	return fmt.Sprintf("%s\n%s\n%s", header, mainBody, footer)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}