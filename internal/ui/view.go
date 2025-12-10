package ui

import (
	"fmt"
	"strings"
	"time"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func (m *App) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Determine if background should be dimmed (overlay active)
	dimmed := m.activeOverlay != OverlayNone || m.showHelp

	// Apply dimmed palette for background elements when overlay is active
	restoreTheme := useStyleTheme(dimmed)
	defer restoreTheme()

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

	if len(breakdown) > 0 {
		status += " • " + strings.Join(breakdown, " • ")
	}

	// Show view mode indicator when not in default (All) mode
	if m.viewMode != ViewModeAll {
		modeLabel := fmt.Sprintf("[%s]", m.viewMode.String())
		status += " " + styleFilterInfo().Render(modeLabel)
	}

	if m.filterText != "" {
		filterLabel := fmt.Sprintf("Filter: %s", m.filterText)
		status += " " + styleFilterInfo().Render(filterLabel)
	}

	title := "ABACUS"
	if m.version != "" {
		title = fmt.Sprintf("ABACUS v%s", m.version)
	}

	// Build header with repo name on right - all with theme background
	leftContent := styleAppHeader().Render(title) + baseStyle().Render(" ") + styleNormalText().Render(status)
	rightContent := styleNormalText().Render("Repo: ") + styleID().Render(m.repoName)
	availableWidth := m.width - lipgloss.Width(leftContent) - lipgloss.Width(rightContent) - 2
	var header string
	if availableWidth > 0 {
		header = leftContent + styleNormalText().Render(strings.Repeat(" ", availableWidth)) + rightContent
	} else {
		header = leftContent + styleNormalText().Render(" ") + rightContent
	}
	// Ensure header fills full width with background
	header = baseStyle().Width(m.width).Render(header)
	treeViewStr := m.renderTreeView(dimmed)

	var mainBody string
	listHeight := clampDimension(m.height-4, minListHeight, m.height-2)
	if m.ShowDetails {
		leftStyle := stylePane()
		rightStyle := stylePane()
		if m.focus == FocusTree {
			leftStyle = stylePaneFocused()
		} else {
			rightStyle = stylePaneFocused()
		}

		leftWidth := m.width - m.viewport.Width - 4
		if leftWidth < 1 {
			leftWidth = 1
		}
		rightWidth := m.viewport.Width
		if rightWidth < 1 {
			rightWidth = 1
		}

		// Re-render viewport content with current theme (dimmed or bright)
		// This ensures detail pane properly dims when overlay is active
		m.updateViewportContent()

		left := leftStyle.Width(leftWidth).Height(listHeight).Render(treeViewStr)
		right := rightStyle.Width(rightWidth).Height(listHeight).Render(m.viewport.View())
		mainBody = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	} else {
		singleWidth := m.width - 2
		if singleWidth < 1 {
			singleWidth = 1
		}
		mainBody = stylePane().Width(singleWidth).Height(listHeight).Render(treeViewStr)
	}

	var bottomBar string
	if m.searching {
		bottomBar = m.textInput.View()
	} else {
		bottomBar = m.renderFooter()
	}

	wrapWithBackground := func(content string) string {
		return lipgloss.NewStyle().
			Background(theme.Current().Background()).
			Width(m.width).
			Height(m.height).
			Render(content)
	}

	headerHeight := lipgloss.Height(header)
	if headerHeight <= 0 {
		headerHeight = 1
	}
	mainBodyStart := headerHeight
	mainBodyHeight := lipgloss.Height(mainBody)
	if mainBodyHeight <= 0 {
		mainBodyHeight = listHeight
	}
	bottomMargin := lipgloss.Height(bottomBar)
	if bottomMargin <= 0 {
		bottomMargin = 1
	}

	// Determine whether we need to show an overlay (status, labels, create, delete, help)
	var overlayLayers []Layer
	if m.activeOverlay == OverlayStatus && m.statusOverlay != nil {
		if layer := m.statusOverlay.Layer(m.width, m.height, headerHeight, bottomMargin); layer != nil {
			overlayLayers = append(overlayLayers, layer)
		}
	} else if m.activeOverlay == OverlayLabels && m.labelsOverlay != nil {
		if layer := m.labelsOverlay.Layer(m.width, m.height, headerHeight, bottomMargin); layer != nil {
			overlayLayers = append(overlayLayers, layer)
		}
	} else if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
		if layer := m.createOverlay.Layer(m.width, m.height, headerHeight, bottomMargin); layer != nil {
			overlayLayers = append(overlayLayers, layer)
		}
	} else if m.activeOverlay == OverlayDelete && m.deleteOverlay != nil {
		if layer := m.deleteOverlay.Layer(m.width, m.height, headerHeight, bottomMargin); layer != nil {
			overlayLayers = append(overlayLayers, layer)
		}
	} else if m.showHelp {
		overlayLayers = append(overlayLayers, newHelpOverlayLayer(m.keys, m.width, m.height, headerHeight, bottomMargin))
	}

	content := fmt.Sprintf("%s\n%s\n%s", header, mainBody, bottomBar)
	base := wrapWithBackground(content)

	var overlayErrorLayer Layer
	if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
		overlayErrorLayer = m.errorToastLayer(m.width, m.height, mainBodyStart, mainBodyHeight)
	}

	var toastLayer Layer
	toastFactories := []func(int, int, int, int) Layer{
		m.themeToastLayer,
		m.deleteToastLayer,
		m.createToastLayer,
		m.newAssigneeToastLayer,
		m.newLabelToastLayer,
		m.labelsToastLayer,
		m.statusToastLayer,
		m.copyToastLayer,
		m.errorToastLayer,
	}
	for _, factory := range toastFactories {
		if layer := factory(m.width, m.height, mainBodyStart, mainBodyHeight); layer != nil {
			toastLayer = layer
			break
		}
	}

	if len(overlayLayers) > 0 {
		canvas := NewCanvas(m.width, m.height)
		// Background is already rendered with dimmed theme
		canvas.DrawStringAt(0, 0, base)

		// Switch to bright theme for overlay rendering
		restoreBright := useStyleTheme(false)
		for _, layer := range overlayLayers {
			if layer == nil {
				continue
			}
			if c := layer.Render(); c != nil {
				canvas.OverlayCanvas(c)
			}
		}
		if overlayErrorLayer != nil {
			if c := overlayErrorLayer.Render(); c != nil {
				canvas.OverlayCanvas(c)
			}
		}
		if toastLayer != nil {
			if c := toastLayer.Render(); c != nil {
				canvas.OverlayCanvas(c)
			}
		}
		restoreBright()

		return canvas.Render()
	}

	if toastLayer != nil {
		canvas := NewCanvas(m.width, m.height)
		canvas.DrawStringAt(0, 0, base)
		if c := toastLayer.Render(); c != nil {
			canvas.OverlayCanvas(c)
		}
		return canvas.Render()
	}

	return base
}

// errorToastLayer renders the error toast as a layer if visible.
func (m *App) errorToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.showErrorToast || m.lastError == "" {
		return nil
	}
	elapsed := time.Since(m.errorToastStart)
	remaining := 10 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Extract a short, user-friendly error message
	errMsg := extractShortError(m.lastError, 80)

	// Build content: title + bd error message + countdown right-aligned
	titleLine := "⚠ Error"
	bdErrLine := fmt.Sprintf("bd: %s", errMsg)
	countdownStr := fmt.Sprintf("[%ds]", remaining)

	// Calculate toast width based on longest line
	toastWidth := 50
	if w := lipgloss.Width(titleLine); w > toastWidth {
		toastWidth = w
	}
	if w := lipgloss.Width(bdErrLine); w > toastWidth {
		toastWidth = w
	}

	padding := toastWidth - len(countdownStr)
	if padding < 0 {
		padding = 0
	}
	content := fmt.Sprintf("%s\n%s\n%s%s", titleLine, bdErrLine, strings.Repeat(" ", padding), countdownStr)

	return newToastLayer(styleErrorToast().Render(content), width, height, mainBodyStart, mainBodyHeight)
}

// copyToastLayer renders the copy success toast if visible.
func (m *App) copyToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.showCopyToast || m.copiedBeadID == "" {
		return nil
	}
	elapsed := time.Since(m.copyToastStart)
	remaining := 5 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Build content: message + right-aligned countdown
	msgLine := fmt.Sprintf("Copied '%s' to clipboard.", m.copiedBeadID)
	countdownStr := fmt.Sprintf("[%ds]", remaining)

	// Calculate toast width based on message
	toastWidth := lipgloss.Width(msgLine)
	if toastWidth < 30 {
		toastWidth = 30
	}

	padding := toastWidth - len(countdownStr)
	if padding < 0 {
		padding = 0
	}
	content := fmt.Sprintf("%s\n%s%s", msgLine, strings.Repeat(" ", padding), countdownStr)

	return newToastLayer(styleSuccessToast().Render(content), width, height, mainBodyStart, mainBodyHeight)
}

// statusToastLayer renders the status change success toast if visible.
func (m *App) statusToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.statusToastVisible || m.statusToastNewStatus == "" {
		return nil
	}
	elapsed := time.Since(m.statusToastStart)
	remaining := 7 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Line 1: "Status → ◐ In Progress" - label + new status as hero
	newIcon, newIconStyle, newTextStyle := statusPresentation(m.statusToastNewStatus)
	label := styleStatsDim().Render("Status →")
	status := newIconStyle.Render(newIcon) + " " + newTextStyle.Render(formatStatusLabel(m.statusToastNewStatus))
	heroLine := " " + label + " " + status

	// Line 2: bead ID + right-aligned countdown
	beadID := styleID().Render(m.statusToastBeadID)
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	// Calculate spacing for right-aligned countdown
	leftPart := " " + beadID
	heroWidth := lipgloss.Width(heroLine)
	leftWidth := lipgloss.Width(leftPart)
	countdownWidth := lipgloss.Width(countdownStr)

	// Match hero line width for alignment
	targetWidth := heroWidth
	if targetWidth < 20 {
		targetWidth = 20
	}
	padding := targetWidth - leftWidth - countdownWidth
	if padding < 2 {
		padding = 2
	}

	infoLine := leftPart + strings.Repeat(" ", padding) + countdownStr

	content := heroLine + "\n" + infoLine
	return newToastLayer(styleSuccessToast().Render(content), width, height, mainBodyStart, mainBodyHeight)
}

// statusPresentation returns icon, icon style, and text style for a status.
func statusPresentation(status string) (string, lipgloss.Style, lipgloss.Style) {
	switch status {
	case "in_progress":
		return "◐", styleIconInProgress(), styleInProgressText()
	case "closed":
		return "✔", styleIconDone(), styleDoneText()
	default: // open
		return "○", styleIconOpen(), styleNormalText()
	}
}

// labelsToastLayer renders the labels change success toast if visible.
func (m *App) labelsToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.labelsToastVisible {
		return nil
	}
	elapsed := time.Since(m.labelsToastStart)
	remaining := 7 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Build summary: "+label1, +label2" or "-label1" or both
	// Added labels in green, removed labels in red
	var parts []string
	for _, l := range m.labelsToastAdded {
		parts = append(parts, styleLabelChecked().Render("+"+l))
	}
	for _, l := range m.labelsToastRemoved {
		parts = append(parts, styleBlockedText().Render("-"+l))
	}

	// Line 1: "Labels: +ui, +bug, -old"
	label := styleStatsDim().Render("Labels:")
	changes := strings.Join(parts, styleStatsDim().Render(", "))
	heroLine := " " + label + " " + changes

	// Line 2: bead ID + right-aligned countdown
	beadID := styleID().Render(m.labelsToastBeadID)
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	// Calculate spacing for right-aligned countdown
	leftPart := " " + beadID
	heroWidth := lipgloss.Width(heroLine)
	leftWidth := lipgloss.Width(leftPart)
	countdownWidth := lipgloss.Width(countdownStr)

	// Match hero line width for alignment
	targetWidth := heroWidth
	if targetWidth < 20 {
		targetWidth = 20
	}
	padding := targetWidth - leftWidth - countdownWidth
	if padding < 2 {
		padding = 2
	}

	infoLine := leftPart + strings.Repeat(" ", padding) + countdownStr

	content := heroLine + "\n" + infoLine
	return newToastLayer(styleSuccessToast().Render(content), width, height, mainBodyStart, mainBodyHeight)
}

// createToastLayer renders the bead creation success toast if visible.
func (m *App) createToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.createToastVisible {
		return nil
	}
	elapsed := time.Since(m.createToastStart)
	if elapsed >= 7*time.Second {
		return nil
	}
	remaining := 7 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Line 1: "✓ Created ab-xyz" (or Updated) - bead ID prominent
	beadID := m.createToastBeadID
	if beadID == "" {
		beadID = "..."
	}
	action := "Created"
	if m.createToastIsUpdate {
		action = "Updated"
	}
	heroLine := " ✓ " + styleStatsDim().Render(action) + " " + styleID().Render(beadID)

	// Line 2: title (up to 45 chars) + right-aligned countdown
	titleDisplay := m.createToastTitle
	if len(titleDisplay) > 45 {
		titleDisplay = titleDisplay[:42] + "..."
	}
	titlePart := " " + styleLabelChecked().Render(titleDisplay)
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	// Calculate spacing for right-aligned countdown
	heroWidth := lipgloss.Width(heroLine)
	titleWidth := lipgloss.Width(titlePart)
	countdownWidth := lipgloss.Width(countdownStr)

	// Use wider of hero or title line for alignment
	targetWidth := heroWidth
	if titleWidth > targetWidth {
		targetWidth = titleWidth + countdownWidth + 2
	}
	if targetWidth < 30 {
		targetWidth = 30
	}
	padding := targetWidth - titleWidth - countdownWidth
	if padding < 2 {
		padding = 2
	}

	infoLine := titlePart + strings.Repeat(" ", padding) + countdownStr

	content := heroLine + "\n" + infoLine
	return newToastLayer(styleSuccessToast().Render(content), width, height, mainBodyStart, mainBodyHeight)
}

// newLabelToastLayer renders the new label toast if visible.
// Shown when a label is created that wasn't in the existing options.
func (m *App) newLabelToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.newLabelToastVisible || m.newLabelToastLabel == "" {
		return nil
	}
	elapsed := time.Since(m.newLabelToastStart)
	if elapsed >= 3*time.Second {
		return nil
	}
	remaining := 3 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Simple one-line toast: "New Label Added: [labelname]"
	content := " ✓ New Label Added: " + styleLabelChecked().Render(m.newLabelToastLabel) + " "
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	return newToastLayer(styleSuccessToast().Render(content+countdownStr), width, height, mainBodyStart, mainBodyHeight)
}

// newAssigneeToastLayer renders the new assignee toast if visible.
// Shown when an assignee is created that wasn't in the existing options.
func (m *App) newAssigneeToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.newAssigneeToastVisible || m.newAssigneeToastAssignee == "" {
		return nil
	}
	elapsed := time.Since(m.newAssigneeToastStart)
	if elapsed >= 3*time.Second {
		return nil
	}
	remaining := 3 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Simple one-line toast: "New Assignee Added: [name]"
	content := " ✓ New Assignee Added: " + styleLabelChecked().Render(m.newAssigneeToastAssignee) + " "
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	return newToastLayer(styleSuccessToast().Render(content+countdownStr), width, height, mainBodyStart, mainBodyHeight)
}

// deleteToastLayer renders the delete success toast if visible.
func (m *App) deleteToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.deleteToastVisible || m.deleteToastBeadID == "" {
		return nil
	}
	elapsed := time.Since(m.deleteToastStart)
	remaining := 5 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Line 1: "✓ Deleted ab-xyz" (+ optional child count)
	heroLine := " ✓ " + styleStatsDim().Render("Deleted") + " " + styleID().Render(m.deleteToastBeadID)
	if m.deleteToastCascade && m.deleteToastChildCount > 0 {
		heroLine += styleStatsDim().Render(fmt.Sprintf(" (+%d %s)", m.deleteToastChildCount, childWord(m.deleteToastChildCount)))
	}
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	// Calculate spacing for right-aligned countdown
	heroWidth := lipgloss.Width(heroLine)
	countdownWidth := lipgloss.Width(countdownStr)

	targetWidth := heroWidth
	if targetWidth < 25 {
		targetWidth = 25
	}
	padding := targetWidth - countdownWidth
	if padding < 2 {
		padding = 2
	}

	content := heroLine + "\n" + strings.Repeat(" ", padding) + countdownStr
	return newToastLayer(styleSuccessToast().Render(content), width, height, mainBodyStart, mainBodyHeight)
}

// themeToastLayer renders the theme change toast if visible.
func (m *App) themeToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.themeToastVisible || m.themeToastName == "" {
		return nil
	}
	elapsed := time.Since(m.themeToastStart)
	remaining := 3 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Format theme name nicely (capitalize first letter)
	themeName := m.themeToastName
	if len(themeName) > 0 {
		themeName = strings.ToUpper(themeName[:1]) + themeName[1:]
	}

	// Line 1: "Theme: Dracula" with background-safe spacing
	icon := baseStyle().Render(" 🎨 ")
	label := styleStatsDim().Render("Theme:")
	space := baseStyle().Render(" ")
	name := styleID().Render(themeName)
	heroLine := lipgloss.JoinHorizontal(lipgloss.Left, icon, label, space, name)
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	// Calculate spacing for right-aligned countdown
	heroWidth := lipgloss.Width(heroLine)
	countdownWidth := lipgloss.Width(countdownStr)

	targetWidth := heroWidth
	if targetWidth < 25 {
		targetWidth = 25
	}
	padding := targetWidth - countdownWidth
	if padding < 2 {
		padding = 2
	}

	paddingSpaces := ""
	if padding > 0 {
		paddingSpaces = baseStyle().Render(strings.Repeat(" ", padding))
	}
	content := heroLine + "\n" + paddingSpaces + countdownStr
	return newToastLayer(styleSuccessToast().Render(content), width, height, mainBodyStart, mainBodyHeight)
}
