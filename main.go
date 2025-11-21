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
	"github.com/muesli/reflow/wordwrap"
)

// --- Styles ---
var (
	// Theme Colors
	cPurple     = lipgloss.Color("99")
	cCyan       = lipgloss.Color("39")
	cNeonGreen  = lipgloss.Color("118")
	cRed        = lipgloss.Color("196")
	cOrange     = lipgloss.Color("208")
	cGold       = lipgloss.Color("220")
	cGray       = lipgloss.Color("240")
	cBrightGray = lipgloss.Color("246")
	cLightGray  = lipgloss.Color("250")
	cWhite      = lipgloss.Color("255")
	cHighlight  = lipgloss.Color("57")
	cField      = lipgloss.Color("63")

	// Text Styles
	styleInProgressText = lipgloss.NewStyle().Foreground(cCyan).Bold(true)
	styleNormalText     = lipgloss.NewStyle().Foreground(cWhite)
	styleDoneText       = lipgloss.NewStyle().Foreground(cBrightGray)
	styleBlockedText    = lipgloss.NewStyle().Foreground(cRed)
	styleStatsDim = lipgloss.NewStyle().Foreground(cBrightGray)

	// Icon Styles
	styleIconOpen       = lipgloss.NewStyle().Foreground(cWhite)
	styleIconInProgress = lipgloss.NewStyle().Foreground(cNeonGreen)
	styleIconDone       = lipgloss.NewStyle().Foreground(cBrightGray)
	styleIconBlocked    = lipgloss.NewStyle().Foreground(cRed)

	styleID = lipgloss.NewStyle().Foreground(cGold).Bold(true)

	// Tree Selection Style
	styleSelected = lipgloss.NewStyle().
			Background(cHighlight).
			Foreground(cWhite).
			Bold(true)

	styleAppHeader = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(cPurple).
			Bold(true).
			Padding(0, 1)

	styleFilterInfo = lipgloss.NewStyle().
			Foreground(cLightGray).
			Background(cPurple)

	// Border Styles for Focus toggling
	stylePane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(cGray)

	stylePaneFocused = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(cPurple)

	// --- Detail Pane Specifics ---

	styleDetailHeaderBlock = lipgloss.NewStyle().
				Background(cHighlight).
				Foreground(cWhite).
				Bold(true).
				Padding(0, 1)

	styleDetailHeaderCombined = lipgloss.NewStyle().
					Background(cHighlight).
					Bold(true)

	styleField = lipgloss.NewStyle().
			Foreground(cField).
			Bold(true).
			Width(12)

	styleVal = lipgloss.NewStyle().Foreground(cWhite)

	styleSectionHeader = lipgloss.NewStyle().
				Foreground(cGold).
				Bold(true).
				MarginTop(1).
				MarginBottom(0).
				MarginLeft(1)

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

	// Comment Styles
	styleCommentHeader = lipgloss.NewStyle().
				Foreground(cBrightGray).
				Bold(true)
)

// --- Data Structures ---

type LiteIssue struct {
	ID string `json:"id"`
}

type Comment struct {
	ID        int    `json:"id"`
	IssueID   string `json:"issue_id"`
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type FullIssue struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	IssueType   string    `json:"issue_type"`
	Priority    int       `json:"priority"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	ClosedAt    string    `json:"closed_at"`
	ExternalRef string    `json:"external_ref"`
	Labels      []string  `json:"labels"`
	Comments    []Comment `json:"comments"`

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
	Parents   []*Node
	Parent    *Node

	BlockedBy []*Node
	Blocks    []*Node

	IsBlocked      bool
	CommentsLoaded bool

	Expanded      bool
	Depth         int
	TreeDepth     int
	HasInProgress bool
	HasReady      bool
}

type FocusArea int

const (
	FocusTree FocusArea = iota
	FocusDetails
)

type Stats struct {
	Total      int
	InProgress int
	Ready      int
	Blocked    int
	Closed     int
}

// --- Model ---

type model struct {
	roots       []*Node
	visibleRows []*Node
	cursor      int
	repoName    string

	viewport    viewport.Model
	showDetails bool
	focus       FocusArea
	ready       bool

	textInput  textinput.Model
	searching  bool
	filterText string

	width  int
	height int
	err    error
}

// --- Helpers ---

func formatTime(isoStr string) string {
	if isoStr == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return isoStr
	}
	return t.Local().Format("Jan 02, 3:04 PM")
}

func wrapWithHangingIndent(prefixWidth int, text string, maxWidth int) string {
	if maxWidth <= prefixWidth {
		return text
	}

	contentWidth := maxWidth - prefixWidth
	if contentWidth <= 0 {
		contentWidth = 10
	}

	wrapped := wordwrap.String(text, contentWidth)

	lines := strings.Split(wrapped, "\n")
	if len(lines) <= 1 {
		return text
	}

	var sb strings.Builder
	sb.WriteString(lines[0])

	padding := strings.Repeat(" ", prefixWidth)
	for i := 1; i < len(lines); i++ {
		sb.WriteString("\n")
		sb.WriteString(padding)
		sb.WriteString(lines[i])
	}
	return sb.String()
}

func indentBlock(text string, spaces int) string {
	padding := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = padding + line
		}
	}
	return strings.Join(lines, "\n")
}

// trimGlamourOutput removes leading/trailing whitespace/newlines that Glamour adds by default
func trimGlamourOutput(s string) string {
	return strings.TrimSpace(s)
}

func fetchCommentsForNode(n *Node) {
	if n.CommentsLoaded {
		return
	}
	cmd := exec.Command("bd", "comments", n.Issue.ID, "--json")
	out, err := cmd.Output()
	if err == nil {
		var comments []Comment
		if json.Unmarshal(out, &comments) == nil {
			n.Issue.Comments = comments
		}
	}
	n.CommentsLoaded = true
}

// --- Data Loading & Graph Logic ---

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

	roots := buildDeepestParentGraph(fullIssues)

	sort.SliceStable(roots, func(i, j int) bool {
		a, b := roots[i], roots[j]
		rankA, rankB := 2, 2
		if a.HasInProgress {
			rankA = 0
		} else if a.HasReady {
			rankA = 1
		}
		if b.HasInProgress {
			rankB = 0
		} else if b.HasReady {
			rankB = 1
		}
		if rankA != rankB {
			return rankA < rankB
		}
		return a.Issue.CreatedAt < b.Issue.CreatedAt
	})

	return roots, nil
}

func batchFetchIssues(ids []string) ([]FullIssue, error) {
	var results []FullIssue
	chunkSize := 20
	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}

		args := append([]string{"show"}, ids[i:end]...)
		args = append(args, "--json")
		out, _ := exec.Command("bd", args...).Output()

		var batch []FullIssue
		json.Unmarshal(out, &batch)
		results = append(results, batch...)
	}
	return results, nil
}

func buildDeepestParentGraph(issues []FullIssue) []*Node {
	nodeMap := make(map[string]*Node)
	for _, iss := range issues {
		nodeMap[iss.ID] = &Node{Issue: iss}
	}

	for _, node := range nodeMap {
		for _, dep := range node.Issue.Dependencies {
			if dep.Type == "parent-child" {
				if parent, ok := nodeMap[dep.TargetID]; ok {
					node.Parents = append(node.Parents, parent)
				}
			} else if dep.Type == "blocks" {
				if blocker, ok := nodeMap[dep.TargetID]; ok {
					if blocker.Issue.Status != "closed" {
						node.BlockedBy = append(node.BlockedBy, blocker)
						node.IsBlocked = true
						blocker.Blocks = append(blocker.Blocks, node)
					}
				}
			}
		}
		for _, dep := range node.Issue.Dependents {
			if child, ok := nodeMap[dep.ID]; ok {
				child.Parents = append(child.Parents, node)
			}
		}
	}

	var getDepth func(n *Node, visited map[string]bool) int
	getDepth = func(n *Node, visited map[string]bool) int {
		if visited[n.Issue.ID] {
			return 0
		}
		visited[n.Issue.ID] = true
		if len(n.Parents) == 0 {
			return 0
		}
		maxP := 0
		for _, p := range n.Parents {
			d := getDepth(p, visited)
			if d > maxP {
				maxP = d
			}
		}
		delete(visited, n.Issue.ID)
		return maxP + 1
	}

	for _, node := range nodeMap {
		node.TreeDepth = getDepth(node, make(map[string]bool))
	}

	var roots []*Node
	childrenIDs := make(map[string]bool)

	for _, node := range nodeMap {
		if len(node.Parents) == 0 {
			roots = append(roots, node)
			continue
		}

		maxParentDepth := -1
		for _, p := range node.Parents {
			if p.TreeDepth > maxParentDepth {
				maxParentDepth = p.TreeDepth
			}
		}

		for _, p := range node.Parents {
			if p.TreeDepth == maxParentDepth {
				p.Children = append(p.Children, node)
				node.Parent = p // Set Visual Parent
				childrenIDs[node.Issue.ID] = true
			}
		}
	}

	for _, node := range nodeMap {
		if len(node.Parents) > 0 && !childrenIDs[node.Issue.ID] {
			roots = append(roots, node)
		}
	}

	for _, node := range nodeMap {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Issue.CreatedAt < node.Children[j].Issue.CreatedAt
		})
		sort.Slice(node.Blocks, func(i, j int) bool {
			return node.Blocks[i].Issue.CreatedAt < node.Blocks[j].Issue.CreatedAt
		})
	}

	for _, root := range roots {
		computeStates(root)
		if root.HasInProgress {
			root.Expanded = true
		}
	}

	return roots
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

// --- UI Logic ---

func initialModel() model {
	roots, err := loadData()
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = "/"

	repo := "abacus"
	wd, _ := os.Getwd()
	if wd != "" {
		repo = filepath.Base(wd)
	}

	m := model{
		roots:     roots,
		err:       err,
		textInput: ti,
		repoName:  repo,
		focus:     FocusTree,
	}
	m.recalcVisibleRows()
	return m
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
						if matches(k) || checkChildren(k.Children) {
							return true
						}
					}
					return false
				}
				hasMatchingChild = checkChildren(node.Children)
			}

			if isMatch || hasMatchingChild {
				m.visibleRows = append(m.visibleRows, node)
				if (m.filterText == "" && node.Expanded) || (m.filterText != "" && hasMatchingChild) {
					traverse(node.Children)
				}
			}
		}
	}
	traverse(m.roots)
}

func (m *model) getStats() Stats {
	s := Stats{}

	var traverse func(nodes []*Node)
	traverse = func(nodes []*Node) {
		for _, n := range nodes {
			matches := m.filterText == "" || strings.Contains(strings.ToLower(n.Issue.Title), strings.ToLower(m.filterText))

			if matches {
				s.Total++
				if n.Issue.Status == "in_progress" {
					s.InProgress++
				} else if n.Issue.Status == "closed" {
					s.Closed++
				} else if n.IsBlocked {
					s.Blocked++
				} else {
					s.Ready++
				}
			}
			traverse(n.Children)
		}
	}
	traverse(m.roots)
	return s
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.viewport.Width = int(float64(msg.Width)*0.45) - 2
		m.viewport.Height = msg.Height - 5
		m.updateViewportContent()

	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "enter":
				m.searching = false
				m.textInput.Blur()
				return m, nil
			case "esc":
				m.searching = false
				m.textInput.Blur()
				m.textInput.Reset()
				m.filterText = ""
				m.recalcVisibleRows()
				m.updateViewportContent()
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				m.filterText = m.textInput.Value()
				m.recalcVisibleRows()
				m.cursor = 0
				m.updateViewportContent()
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.filterText != "" {
				m.filterText = ""
				m.textInput.Reset()
				m.recalcVisibleRows()
				m.updateViewportContent()
			}
			return m, nil
		case "/":
			m.searching = true
			m.textInput.Focus()
			m.textInput.SetValue("")
			m.filterText = ""
			m.recalcVisibleRows()
			return m, textinput.Blink
		case "tab":
			if m.showDetails {
				if m.focus == FocusTree {
					m.focus = FocusDetails
				} else {
					m.focus = FocusTree
				}
			}
		}

		if m.focus == FocusDetails && m.showDetails {
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		} else {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					m.updateViewportContent()
					if m.focus == FocusTree {
						m.viewport.GotoTop()
					}
				}
			case "down", "j":
				if m.cursor < len(m.visibleRows)-1 {
					m.cursor++
					m.updateViewportContent()
					if m.focus == FocusTree {
						m.viewport.GotoTop()
					}
				}
			case "space", "right", "l":
				node := m.visibleRows[m.cursor]
				if len(node.Children) > 0 {
					node.Expanded = !node.Expanded
					m.recalcVisibleRows()
				}
			case "left", "h":
				node := m.visibleRows[m.cursor]
				if node.Expanded {
					node.Expanded = false
					m.recalcVisibleRows()
				}
			case "enter":
				m.showDetails = !m.showDetails
				m.updateViewportContent()
			}
		}
	}
	return m, cmd
}

// renderRefRow creates a strict two-column layout: [ID] [Gap] [TitleBlock].
// It handles background colors explicitly to prevent "black gaps" in headers.
func renderRefRow(id, title string, targetWidth int, idStyle, titleStyle lipgloss.Style, bgColor lipgloss.Color) string {
	// 1. Define the Gap
	gap := "  "
	gapWidth := 2

	// 2. Render ID (single line)
	// We apply background to ID immediately
	idRendered := idStyle.Background(bgColor).Render(id)
	idWidth := lipgloss.Width(idRendered)

	// 3. Calculate Title Width
	// SAFETY BUFFER: Subtract 4 extra characters to prevent terminal hard-wrapping
	titleWidth := targetWidth - idWidth - gapWidth - 4
	if titleWidth < 10 {
		titleWidth = 10
	}

	// 4. Wrap and Render Title
	// We force the width on the style so it fills the block
	titleRendered := titleStyle.
		Background(bgColor).
		Width(titleWidth).
		Render(wordwrap.String(title, titleWidth))

	// 5. Equalize Heights (Crucial for the solid header bar)
	// If the title wraps to 3 lines, we need the ID and Gap to be 3 lines tall 
	// so the background color extends down.
	h := lipgloss.Height(titleRendered)
	idRendered = idStyle.Background(bgColor).Height(h).Render(id)
	gapRendered := lipgloss.NewStyle().Background(bgColor).Height(h).Render(gap)

	// 6. Join
	return lipgloss.JoinHorizontal(lipgloss.Top, idRendered, gapRendered, titleRendered)
}

func (m *model) updateViewportContent() {
	if !m.showDetails || m.cursor >= len(m.visibleRows) {
		return
	}
	node := m.visibleRows[m.cursor]

	if !node.CommentsLoaded {
		fetchCommentsForNode(node)
	}

	iss := node.Issue
	vpWidth := m.viewport.Width

	// --- 1. HEADER ---
	// We pass the full width. The helper now subtracts a safety buffer.
	// We pass cHighlight (Purple) for the background.
	headerContent := renderRefRow(
		iss.ID,
		iss.Title,
		vpWidth,
		styleDetailHeaderCombined.Foreground(cGold),
		styleDetailHeaderCombined.Foreground(cWhite),
		cHighlight, 
	)
	// Render inside the block style to ensure full width background extension
	headerBlock := styleDetailHeaderBlock.Width(vpWidth).Render(headerContent)

	// --- 2. METADATA ---
	makeRow := func(k, v string) string {
		return lipgloss.JoinHorizontal(lipgloss.Left, styleField.Render(k), styleVal.Render(v))
	}

	col1 := []string{
		makeRow("Status:", iss.Status),
		makeRow("Type:", iss.IssueType),
		makeRow("Created:", formatTime(iss.CreatedAt)),
	}
	if iss.UpdatedAt != iss.CreatedAt {
		col1 = append(col1, makeRow("Updated:", formatTime(iss.UpdatedAt)))
	}
	if iss.Status == "closed" {
		col1 = append(col1, makeRow("Closed:", formatTime(iss.ClosedAt)))
	}

	prioLabel := fmt.Sprintf("P%d", iss.Priority)
	col2 := []string{
		makeRow("Priority:", stylePrio.Render(prioLabel)),
	}
	if iss.ExternalRef != "" {
		col2 = append(col2, makeRow("Ext Ref:", iss.ExternalRef))
	}

	if len(iss.Labels) > 0 {
		var labelRows []string
		var currentRow string
		currentLen := 0
		labelPrefixWidth := 12
		availableLabelWidth := (vpWidth / 2) - labelPrefixWidth
		if availableLabelWidth < 10 {
			availableLabelWidth = 10
		}

		for _, l := range iss.Labels {
			rendered := styleLabel.Render(l)
			w := lipgloss.Width(rendered)
			if currentLen+w > availableLabelWidth && currentLen > 0 {
				labelRows = append(labelRows, currentRow)
				currentRow = ""
				currentLen = 0
			}
			currentRow += rendered
			currentLen += w
		}
		if currentRow != "" {
			labelRows = append(labelRows, currentRow)
		}

		firstRow := lipgloss.JoinHorizontal(lipgloss.Left, styleField.Render("Labels:"), labelRows[0])
		finalLabelBlock := firstRow
		padding := strings.Repeat(" ", labelPrefixWidth)
		for i := 1; i < len(labelRows); i++ {
			finalLabelBlock += "\n" + padding + labelRows[i]
		}
		col2 = append(col2, finalLabelBlock)
	} else {
		col2 = append(col2, makeRow("Labels:", "-"))
	}

	leftStack := lipgloss.JoinVertical(lipgloss.Left, col1...)
	rightStack := lipgloss.JoinVertical(lipgloss.Left, col2...)

	var metaBlock string
	if vpWidth < 60 {
		metaBlock = lipgloss.JoinVertical(lipgloss.Left, leftStack, rightStack)
	} else {
		metaBlock = lipgloss.JoinHorizontal(lipgloss.Top, leftStack, "    ", rightStack)
	}
	metaBlock = lipgloss.NewStyle().MarginLeft(1).PaddingTop(1).PaddingBottom(1).Render(metaBlock)

	// --- 3. RELATIONSHIPS ---
	relBuilder := strings.Builder{}

	if iss.ExternalRef != "" {
		relBuilder.WriteString(styleSectionHeader.Render("External Reference") + "\n")
		relBuilder.WriteString(fmt.Sprintf("  🔗 %s\n\n", iss.ExternalRef))
	}

	// Helper for the lists (Parents, Depends On, etc)
	renderRelSection := func(title string, items []*Node) {
		relBuilder.WriteString(styleSectionHeader.Render(title) + "\n")
		for _, item := range items {
			// We indent the whole row by 2 spaces manually
			indent := "  "
			
			// We reduce the target width by (Indent + Safety Buffer) to ensure alignment
			// Passing lipgloss.Color("") ensures no background is applied
			row := renderRefRow(
				item.Issue.ID,
				item.Issue.Title,
				vpWidth - 6, 
				styleID,
				styleVal,
				lipgloss.Color(""), 
			)
			relBuilder.WriteString(indent + row + "\n")
		}
	}

	if node.Parent != nil {
		renderRelSection("Parent", []*Node{node.Parent})
	}
	if len(node.Children) > 0 {
		renderRelSection(fmt.Sprintf("Depends On (%d)", len(node.Children)), node.Children)
	}
	if node.IsBlocked {
		renderRelSection("Blocked By", node.BlockedBy)
	}
	if len(node.Blocks) > 0 {
		renderRelSection(fmt.Sprintf("Blocks (%d)", len(node.Blocks)), node.Blocks)
	}

	relBlock := ""
	if relBuilder.Len() > 0 {
		relBlock = lipgloss.NewStyle().Render(relBuilder.String())
	}

	// --- 4. DESCRIPTION & COMMENTS ---
	// Fix for Double Wrapping:
	// We reserve 10 chars of space. 
	// (2 for margin + 2 for indent + 2 for scrollbar + 4 for safety)
	safeWidth := vpWidth - 10
	if safeWidth < 10 {
		safeWidth = 10
	}

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(safeWidth),
	)

	descBuilder := strings.Builder{}
	descBuilder.WriteString(styleSectionHeader.Render("Description") + "\n")
	desc := strings.ReplaceAll(iss.Description, "• ", "- ")

	renderedDesc, _ := renderer.Render(desc)
	renderedDesc = trimGlamourOutput(renderedDesc)
	descBuilder.WriteString(indentBlock(renderedDesc, 1))

	if len(iss.Comments) > 0 {
		descBuilder.WriteString("\n" + styleSectionHeader.Render("Comments") + "\n")
		for _, c := range iss.Comments {
			header := fmt.Sprintf("  %s  %s", c.Author, formatTime(c.CreatedAt))
			descBuilder.WriteString(styleCommentHeader.Render(header) + "\n")

			renderedComment, _ := renderer.Render(c.Text)
			renderedComment = trimGlamourOutput(renderedComment)
			// The text is already wrapped to safeWidth (vp - 10).
			// Indenting it by 1 or 2 spaces is now safe.
			descBuilder.WriteString(indentBlock(renderedComment, 1) + "\n\n")
		}
	}

	finalContent := lipgloss.JoinVertical(lipgloss.Left,
		headerBlock,
		metaBlock,
		relBlock,
		descBuilder.String(),
	)

	m.viewport.SetContent(finalContent)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	if !m.ready {
		return "Initializing..."
	}

	stats := m.getStats()
	status := fmt.Sprintf("Beads: %d", stats.Total)

	breakdown := []string{}
	if stats.InProgress > 0 {
		breakdown = append(breakdown, fmt.Sprintf("%d In Progress", stats.InProgress))
	}
	if stats.Ready > 0 {
		breakdown = append(breakdown, fmt.Sprintf("%d Ready", stats.Ready))
	}
	if stats.Blocked > 0 {
		breakdown = append(breakdown, fmt.Sprintf("%d Blocked", stats.Blocked))
	}
	if stats.Closed > 0 {
		breakdown = append(breakdown, fmt.Sprintf("%d Closed", stats.Closed))
	}

	// Apply the Dim style to the breakdown list
	if len(breakdown) > 0 {
		status += " " + styleStatsDim.Render("("+strings.Join(breakdown, ", ")+")")
	}

	if m.filterText != "" {
		filterInfo := fmt.Sprintf("(Filtered: '%s' - [ESC] to Clear)", m.filterText)
		status += " " + styleFilterInfo.Render(filterInfo)
	}
	header := styleAppHeader.Render("ABACUS") + " " + status

	// ... (Rest of the View function logic regarding Tree generation remains exactly the same) ...
	
	var treeLines []string
	listHeight := m.height - 4
	start, end := 0, len(m.visibleRows)

	if end > listHeight {
		if m.cursor > listHeight/2 {
			start = m.cursor - listHeight/2
		}
		if start+listHeight < end {
			end = start + listHeight
		} else {
			start = end - listHeight
			if start < 0 {
				start = 0
			}
		}
	}

	treeWidth := m.width - 2
	if m.showDetails {
		treeWidth = m.width - m.viewport.Width - 4
	}

	visualLinesCount := 0

	for i := start; i < end; i++ {
		if visualLinesCount >= listHeight {
			break
		}

		node := m.visibleRows[i]

		indent := strings.Repeat("  ", node.Depth)
		marker := " •"
		if len(node.Children) > 0 {
			if node.Expanded {
				marker = " ▼"
			} else {
				marker = " ▶"
			}
		}

		iconStr, iconStyle, textStyle := "○", styleNormalText, styleNormalText
		if node.Issue.Status == "in_progress" {
			iconStr, iconStyle, textStyle = "◐", styleIconInProgress, styleInProgressText
		} else if node.Issue.Status == "closed" {
			iconStr, iconStyle, textStyle = "✔", styleIconDone, styleDoneText
		} else if node.IsBlocked {
			iconStr, iconStyle, textStyle = "⛔", styleIconBlocked, styleBlockedText
		}

		prefixRaw := fmt.Sprintf("%s%s %s ", indent, marker, iconStr)
		totalPrefixWidth := len(prefixRaw) + len(node.Issue.ID) + 1 - 4
		if totalPrefixWidth < 0 {
			totalPrefixWidth = 0
		}

		wrappedTitle := wrapWithHangingIndent(totalPrefixWidth, node.Issue.Title, treeWidth-4)
		titleLines := strings.Split(wrappedTitle, "\n")

		if i == m.cursor {
			highlightedPrefix := styleSelected.Render(fmt.Sprintf(" %s%s", indent, marker))
			line1Rest := fmt.Sprintf(" %s %s %s", iconStyle.Render(iconStr), styleID.Render(node.Issue.ID), textStyle.Render(titleLines[0]))
			treeLines = append(treeLines, highlightedPrefix+line1Rest)
			visualLinesCount++
		} else {
			line1Prefix := fmt.Sprintf(" %s%s %s ", indent, iconStyle.Render(marker), iconStyle.Render(iconStr))
			line1 := fmt.Sprintf("%s%s %s", line1Prefix, styleID.Render(node.Issue.ID), textStyle.Render(titleLines[0]))
			treeLines = append(treeLines, line1)
			visualLinesCount++
		}

		for k := 1; k < len(titleLines); k++ {
			if visualLinesCount >= listHeight {
				break
			}
			treeLines = append(treeLines, " "+textStyle.Render(titleLines[k]))
			visualLinesCount++
		}
	}

	for visualLinesCount < listHeight {
		treeLines = append(treeLines, "")
		visualLinesCount++
	}

	treeViewStr := strings.Join(treeLines, "\n")

	var mainBody string
	if m.showDetails {
		leftStyle := stylePane
		rightStyle := stylePane
		if m.focus == FocusTree {
			leftStyle = stylePaneFocused
		} else {
			rightStyle = stylePaneFocused
		}

		left := leftStyle.Width(m.width - m.viewport.Width - 4).Height(listHeight).Render(treeViewStr)
		right := rightStyle.Width(m.viewport.Width).Height(listHeight).Render(m.viewport.View())
		mainBody = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	} else {
		mainBody = stylePane.Width(m.width - 2).Height(listHeight).Render(treeViewStr)
	}

	var bottomBar string
	if m.searching {
		bottomBar = m.textInput.View()
	} else {
		footerStr := " [ / ] Search  [ enter ] Detail  [ tab ] Switch Focus  [ q ] Quit"
		if m.showDetails && m.focus == FocusDetails {
			footerStr += "  [ j/k ] Scroll Details"
		} else {
			footerStr += "  [ space ] Expand"
		}
		bottomBar = lipgloss.NewStyle().Foreground(cLightGray).Render(
			fmt.Sprintf("%s   %s", footerStr,
				lipgloss.PlaceHorizontal(m.width-len(footerStr)-5, lipgloss.Right, "Repo: "+m.repoName)))
	}

	return fmt.Sprintf("%s\n%s\n%s", header, mainBody, bottomBar)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}