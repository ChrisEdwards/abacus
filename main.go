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

	// Icon Styles
	styleIconOpen       = lipgloss.NewStyle().Foreground(cWhite)
	styleIconInProgress = lipgloss.NewStyle().Foreground(cNeonGreen)
	styleIconDone       = lipgloss.NewStyle().Foreground(cBrightGray)
	styleIconBlocked    = lipgloss.NewStyle().Foreground(cRed)

	styleID = lipgloss.NewStyle().Foreground(cGold).Bold(true)

	// Used for the selection cursor (now only applied to the arrow/indent)
	styleSelected = lipgloss.NewStyle().
			Background(cHighlight).
			Foreground(cWhite).
			Bold(true)

	styleAppHeader = lipgloss.NewStyle().
			Foreground(cWhite).
			Background(cPurple).
			Bold(true).
			Padding(0, 1)

	stylePane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(cGray)

	// --- Detail Pane Specifics ---

	styleDetailHeaderBlock = lipgloss.NewStyle().
				Background(cHighlight).
				Foreground(cWhite).
				Bold(true).
				Padding(0, 1)

	// Specific styles to ensure background color continuity in Header
	styleDetailHeaderID = lipgloss.NewStyle().
				Background(cHighlight).
				Foreground(cGold).
				Bold(true)

	styleDetailHeaderText = lipgloss.NewStyle().
				Background(cHighlight).
				Foreground(cWhite).
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

// --- Data Structures ---

type LiteIssue struct {
	ID string `json:"id"`
}

type Comment struct {
	ID        int    `json:"id"`
	IssueID   string `json:"issue_id"`
	Author    string `json:"author"`
	Text      string `json:"text"` // Changed from Body to Text to match JSON
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
	Parents   []*Node // Used for depth calc
	Parent    *Node   // Used for visual tree (points to deepest parent)

	BlockedBy []*Node
	Blocks    []*Node

	IsBlocked      bool
	CommentsLoaded bool // Flag for lazy loading

	Expanded      bool
	Depth         int
	TreeDepth     int
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
	} // Safety

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
	ti.Placeholder = "Filter..."
	ti.Prompt = "/ "

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

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.viewport.Width = int(float64(msg.Width)*0.45) - 2
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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.searching = true
			m.textInput.Focus()
			return m, textinput.Blink
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateViewportContent()
				m.viewport.GotoTop()
			}
		case "down", "j":
			if m.cursor < len(m.visibleRows)-1 {
				m.cursor++
				m.updateViewportContent()
				m.viewport.GotoTop()
			}
		case "space", "right", "l": // Added right/l for expand
			node := m.visibleRows[m.cursor]
			if len(node.Children) > 0 {
				node.Expanded = !node.Expanded
				m.recalcVisibleRows()
			}
		case "left", "h": // Added left/h for collapse (good practice)
			node := m.visibleRows[m.cursor]
			if node.Expanded {
				node.Expanded = false
				m.recalcVisibleRows()
			}
		case "enter":
			m.showDetails = !m.showDetails
			m.updateViewportContent()
		case "ctrl+j":
			m.viewport.LineDown(1)
		case "ctrl+k":
			m.viewport.LineUp(1)
		}
	}
	return m, cmd
}

// --- Viewport Rendering ---

func (m *model) updateViewportContent() {
	if !m.showDetails || m.cursor >= len(m.visibleRows) {
		return
	}
	node := m.visibleRows[m.cursor]

	// Lazy Load Comments: If we are showing details and comments aren't loaded, get them now.
	if !node.CommentsLoaded {
		fetchCommentsForNode(node)
	}

	iss := node.Issue

	vpWidth := m.viewport.Width - 2

	// 1. HEADER
	idStr := iss.ID
	prefixWidth := len(idStr) + 2
	wrappedTitle := wrapWithHangingIndent(prefixWidth, iss.Title, vpWidth)
	titleLines := strings.Split(wrappedTitle, "\n")

	// Fix Issue 1: Construct header line with styles applied individually so background is consistent
	// We style the ID and the first line of text with specific styles that share the background color.
	headerContent := fmt.Sprintf("%s  %s",
		styleDetailHeaderID.Render(idStr),
		styleDetailHeaderText.Render(titleLines[0]))

	for i := 1; i < len(titleLines); i++ {
		// For wrapped lines in header, ensure they have the background too
		headerContent += "\n" + styleDetailHeaderBlock.Render(titleLines[i])
	}

	// Use the container block for outer padding/margins
	headerBlock := styleDetailHeaderBlock.Width(vpWidth).Render(headerContent)

	// 2. METADATA GRID
	makeRow := func(k, v string) string {
		return lipgloss.JoinHorizontal(lipgloss.Left, styleField.Render(k), styleVal.Render(v))
	}

	// Column 1
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

	// Column 2
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
	metaBlock = lipgloss.NewStyle().Padding(1, 1).Render(metaBlock)

	// 3. RELATIONSHIPS
	relBuilder := strings.Builder{}

	if iss.ExternalRef != "" {
		relBuilder.WriteString(styleSectionHeader.Render("External Reference") + "\n")
		relBuilder.WriteString(fmt.Sprintf("🔗 %s\n\n", iss.ExternalRef))
	}

	if node.Parent != nil {
		relBuilder.WriteString(styleSectionHeader.Render("Parent") + "\n")
		pTitle := node.Parent.Issue.Title
		pId := node.Parent.Issue.ID
		pPrefixW := len(pId) + 2
		pWrapped := wrapWithHangingIndent(pPrefixW, pTitle, vpWidth-5)
		pLines := strings.Split(pWrapped, "\n")
		relBuilder.WriteString(fmt.Sprintf("%s  %s\n", styleID.Render(pId), pLines[0]))
		for i := 1; i < len(pLines); i++ {
			relBuilder.WriteString("      " + pLines[i] + "\n")
		}
	}

	if node.IsBlocked {
		relBuilder.WriteString(styleSectionHeader.Render("Blocked By") + "\n")
		for _, b := range node.BlockedBy {
			relBuilder.WriteString(fmt.Sprintf("%s  %s\n", styleID.Render(b.Issue.ID), b.Issue.Title))
		}
	}

	if len(node.Children) > 0 {
		relBuilder.WriteString(styleSectionHeader.Render(fmt.Sprintf("Depends On (%d)", len(node.Children))) + "\n")
		for _, child := range node.Children {
			cTitle := child.Issue.Title
			cId := child.Issue.ID
			cPrefixW := len(cId) + 2
			cWrapped := wrapWithHangingIndent(cPrefixW, cTitle, vpWidth-5)
			cLines := strings.Split(cWrapped, "\n")
			relBuilder.WriteString(fmt.Sprintf("%s  %s\n", styleID.Render(cId), cLines[0]))
			for i := 1; i < len(cLines); i++ {
				relBuilder.WriteString("      " + cLines[i] + "\n")
			}
		}
	}

	if len(node.Blocks) > 0 {
		relBuilder.WriteString(styleSectionHeader.Render(fmt.Sprintf("Blocks (%d)", len(node.Blocks))) + "\n")
		for _, child := range node.Blocks {
			relBuilder.WriteString(fmt.Sprintf("%s  %s\n", styleID.Render(child.Issue.ID), child.Issue.Title))
		}
	}

	relBlock := ""
	if relBuilder.Len() > 0 {
		relBlock = lipgloss.NewStyle().Padding(0, 1).Render(relBuilder.String())
	}

	// 4. DESCRIPTION & COMMENTS
	desc := strings.ReplaceAll(iss.Description, "• ", "- ")
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(vpWidth-4),
	)
	bodyStr, _ := renderer.Render(desc)

	if len(iss.Comments) > 0 {
		cBuilder := strings.Builder{}
		cBuilder.WriteString("\n### Comments\n")
		for _, c := range iss.Comments {
			cBuilder.WriteString(fmt.Sprintf("**%s** (%s):\n%s\n\n", c.Author, formatTime(c.CreatedAt), c.Text))
		}
		cStr, _ := renderer.Render(cBuilder.String())
		bodyStr += "\n" + cStr
	}

	finalContent := lipgloss.JoinVertical(lipgloss.Left,
		headerBlock,
		metaBlock,
		relBlock,
		bodyStr,
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

	var treeLines []string
	listHeight := m.height - 3
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

	for i := start; i < end; i++ {
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
		// Fix Issue 2: Adjust prefixWidth for wrapping (-4 correction)
		totalPrefixWidth := len(prefixRaw) + len(node.Issue.ID) + 1 - 4
		if totalPrefixWidth < 0 {
			totalPrefixWidth = 0
		}

		wrappedTitle := wrapWithHangingIndent(totalPrefixWidth, node.Issue.Title, treeWidth)
		titleLines := strings.Split(wrappedTitle, "\n")

		// Fix Issue 4: Highlight logic
		if i == m.cursor {
			// Only highlight up to the arrow (indent + marker)
			// The rest (Icon + ID + Title) has normal background
			highlightedPrefix := styleSelected.Render(fmt.Sprintf("%s%s", indent, marker))
			// Reconstruct the line without the prefix inside
			line1Rest := fmt.Sprintf(" %s %s %s", iconStyle.Render(iconStr), styleID.Render(node.Issue.ID), textStyle.Render(titleLines[0]))
			treeLines = append(treeLines, highlightedPrefix+line1Rest)
		} else {
			line1Prefix := fmt.Sprintf("%s%s %s ", indent, iconStyle.Render(marker), iconStyle.Render(iconStr))
			line1 := fmt.Sprintf("%s%s %s", line1Prefix, styleID.Render(node.Issue.ID), textStyle.Render(titleLines[0]))
			treeLines = append(treeLines, " "+line1)
		}

		for k := 1; k < len(titleLines); k++ {
			// Fix Issue 2b: Ensure wrapped lines keep the text color (e.g. Cyan)
			treeLines = append(treeLines, " "+textStyle.Render(titleLines[k]))
		}
	}

	treeViewStr := strings.Join(treeLines, "\n")

	var mainBody string
	if m.showDetails {
		left := stylePane.Width(m.width - m.viewport.Width - 4).Height(m.height - 3).Render(treeViewStr)
		right := stylePane.Width(m.viewport.Width).Height(m.height - 3).Render(m.viewport.View())
		mainBody = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	} else {
		mainBody = stylePane.Width(m.width - 2).Height(m.height - 3).Render(treeViewStr)
	}

	footer := lipgloss.NewStyle().Foreground(cLightGray).Render(
		fmt.Sprintf(" [ / ] Search  [ b ] Blockers  [ enter ] Detail  [ space/→ ] Expand  [ q ] Quit   %s",
			lipgloss.PlaceHorizontal(m.width-60, lipgloss.Right, "Repo: "+m.repoName)))

	return fmt.Sprintf("%s\n%s\n%s", header, mainBody, footer)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}