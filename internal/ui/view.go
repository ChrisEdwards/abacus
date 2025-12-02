package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m *App) View() string {
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

	if len(breakdown) > 0 {
		status += " • " + strings.Join(breakdown, " • ")
	}

	if m.filterText != "" {
		filterLabel := fmt.Sprintf("Filter: %s", m.filterText)
		status += " " + styleFilterInfo().Render(filterLabel)
	}

	title := "ABACUS"
	if m.version != "" {
		title = fmt.Sprintf("ABACUS v%s", m.version)
	}

	// Build header with repo name on right
	leftContent := styleAppHeader().Render(title) + " " + status
	rightContent := styleFooterMuted().Render("Repo: " + m.repoName)
	availableWidth := m.width - lipgloss.Width(leftContent) - lipgloss.Width(rightContent) - 2
	var header string
	if availableWidth > 0 {
		header = leftContent + strings.Repeat(" ", availableWidth) + rightContent
	} else {
		header = leftContent + " " + rightContent
	}
	treeViewStr := m.renderTreeView()

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

	// Overlays take visual precedence over help
	if m.activeOverlay == OverlayStatus && m.statusOverlay != nil {
		overlay := m.statusOverlay.View()
		centered := lipgloss.Place(m.width, m.height-2,
			lipgloss.Center, lipgloss.Center,
			overlay,
			lipgloss.WithWhitespaceChars(" "),
		)
		return fmt.Sprintf("%s\n%s\n%s", header, centered, bottomBar)
	}

	if m.activeOverlay == OverlayLabels && m.labelsOverlay != nil {
		overlay := m.labelsOverlay.View()
		centered := lipgloss.Place(m.width, m.height-2,
			lipgloss.Center, lipgloss.Center,
			overlay,
			lipgloss.WithWhitespaceChars(" "),
		)
		return fmt.Sprintf("%s\n%s\n%s", header, centered, bottomBar)
	}

	if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
		overlay := m.createOverlay.View()
		centered := lipgloss.Place(m.width, m.height-2,
			lipgloss.Center, lipgloss.Center,
			overlay,
			lipgloss.WithWhitespaceChars(" "),
		)
		// Show error toast over the overlay if visible
		if toast := m.renderErrorToast(); toast != "" {
			containerWidth := lipgloss.Width(centered)
			centered = overlayBottomRight(centered, toast, containerWidth, 1)
		}
		return fmt.Sprintf("%s\n%s\n%s", header, centered, bottomBar)
	}

	if m.activeOverlay == OverlayDelete && m.deleteOverlay != nil {
		overlay := m.deleteOverlay.View()
		centered := lipgloss.Place(m.width, m.height-2,
			lipgloss.Center, lipgloss.Center,
			overlay,
			lipgloss.WithWhitespaceChars(" "),
		)
		return fmt.Sprintf("%s\n%s\n%s", header, centered, bottomBar)
	}

	// Help overlay takes visual precedence over everything else
	if m.showHelp {
		helpOverlay := renderHelpOverlay(m.keys, m.width, m.height-2)
		return fmt.Sprintf("%s\n%s\n%s", header, helpOverlay, bottomBar)
	}

	// Overlay toast on mainBody if visible (theme toast > delete toast > create toast > new assignee toast > new label toast > labels toast > status toast > copy toast > error toast)
	if toast := m.renderThemeToast(); toast != "" {
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	} else if toast := m.renderDeleteToast(); toast != "" {
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	} else if toast := m.renderCreateToast(); toast != "" {
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	} else if toast := m.renderNewAssigneeToast(); toast != "" {
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	} else if toast := m.renderNewLabelToast(); toast != "" {
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	} else if toast := m.renderLabelsToast(); toast != "" {
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	} else if toast := m.renderStatusToast(); toast != "" {
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	} else if toast := m.renderCopyToast(); toast != "" {
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	} else if toast := m.renderErrorToast(); toast != "" {
		// Measure actual rendered width for proper right-alignment
		containerWidth := lipgloss.Width(mainBody)
		mainBody = overlayBottomRight(mainBody, toast, containerWidth, 1)
	}

	return fmt.Sprintf("%s\n%s\n%s", header, mainBody, bottomBar)
}

// renderErrorToast renders the error toast content if visible.
func (m *App) renderErrorToast() string {
	if !m.showErrorToast || m.lastError == "" {
		return ""
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

	return styleErrorToast().Render(content)
}

// renderCopyToast renders the copy success toast content if visible.
func (m *App) renderCopyToast() string {
	if !m.showCopyToast || m.copiedBeadID == "" {
		return ""
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

	return styleSuccessToast().Render(content)
}

// renderStatusToast renders the status change success toast if visible.
func (m *App) renderStatusToast() string {
	if !m.statusToastVisible || m.statusToastNewStatus == "" {
		return ""
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
	return styleSuccessToast().Render(content)
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

// renderLabelsToast renders the labels change success toast if visible.
func (m *App) renderLabelsToast() string {
	if !m.labelsToastVisible {
		return ""
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
	return styleSuccessToast().Render(content)
}

// renderCreateToast renders the bead creation success toast if visible.
func (m *App) renderCreateToast() string {
	if !m.createToastVisible {
		return ""
	}
	elapsed := time.Since(m.createToastStart)
	remaining := 7 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Line 1: "✓ Created ab-xyz" - bead ID prominent
	beadID := m.createToastBeadID
	if beadID == "" {
		beadID = "..."
	}
	heroLine := " ✓ " + styleStatsDim().Render("Created") + " " + styleID().Render(beadID)

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
	return styleSuccessToast().Render(content)
}

// renderNewLabelToast renders the new label toast if visible.
// Shown when a label is created that wasn't in the existing options.
func (m *App) renderNewLabelToast() string {
	if !m.newLabelToastVisible || m.newLabelToastLabel == "" {
		return ""
	}
	elapsed := time.Since(m.newLabelToastStart)
	if elapsed >= 3*time.Second {
		return ""
	}
	remaining := 3 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Simple one-line toast: "New Label Added: [labelname]"
	content := " ✓ New Label Added: " + styleLabelChecked().Render(m.newLabelToastLabel) + " "
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	return styleSuccessToast().Render(content + countdownStr)
}

// renderNewAssigneeToast renders the new assignee toast if visible.
// Shown when an assignee is created that wasn't in the existing options.
func (m *App) renderNewAssigneeToast() string {
	if !m.newAssigneeToastVisible || m.newAssigneeToastAssignee == "" {
		return ""
	}
	elapsed := time.Since(m.newAssigneeToastStart)
	if elapsed >= 3*time.Second {
		return ""
	}
	remaining := 3 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Simple one-line toast: "New Assignee Added: [name]"
	content := " ✓ New Assignee Added: " + styleLabelChecked().Render(m.newAssigneeToastAssignee) + " "
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	return styleSuccessToast().Render(content + countdownStr)
}

// renderDeleteToast renders the delete success toast if visible.
func (m *App) renderDeleteToast() string {
	if !m.deleteToastVisible || m.deleteToastBeadID == "" {
		return ""
	}
	elapsed := time.Since(m.deleteToastStart)
	remaining := 5 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Line 1: "✓ Deleted ab-xyz"
	heroLine := " ✓ " + styleStatsDim().Render("Deleted") + " " + styleID().Render(m.deleteToastBeadID)
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
	return styleSuccessToast().Render(content)
}

// renderThemeToast renders the theme change toast if visible.
func (m *App) renderThemeToast() string {
	if !m.themeToastVisible || m.themeToastName == "" {
		return ""
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

	// Line 1: "Theme: Dracula"
	heroLine := " 🎨 " + styleStatsDim().Render("Theme:") + " " + styleID().Render(themeName)
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
	return styleSuccessToast().Render(content)
}
