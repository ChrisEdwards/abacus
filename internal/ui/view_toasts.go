package ui

import (
	"fmt"
	"strings"
	"time"

	"abacus/internal/update"

	"github.com/charmbracelet/lipgloss"
)

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

// commentToastLayer renders the comment added success toast if visible.
func (m *App) commentToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.commentToastVisible || m.commentToastBeadID == "" {
		return nil
	}
	elapsed := time.Since(m.commentToastStart)
	remaining := 7 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Line 1: "✓ Comment added"
	heroLine := " ✓ " + styleStatsDim().Render("Comment added")

	// Line 2: bead ID + countdown
	beadID := styleID().Render(m.commentToastBeadID)
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	// Calculate spacing for right-aligned countdown
	beadIDWidth := lipgloss.Width(beadID)
	countdownWidth := lipgloss.Width(countdownStr)
	heroWidth := lipgloss.Width(heroLine)

	targetWidth := heroWidth
	if targetWidth < 25 {
		targetWidth = 25
	}
	padding := targetWidth - beadIDWidth - countdownWidth - 1
	if padding < 2 {
		padding = 2
	}

	infoLine := " " + beadID + strings.Repeat(" ", padding) + countdownStr
	content := heroLine + "\n" + infoLine
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

// updateToastLayer renders the update notification toast if visible.
func (m *App) updateToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.updateToastVisible || m.updateInfo == nil || !m.updateInfo.UpdateAvailable {
		return nil
	}
	elapsed := time.Since(m.updateToastStart)
	remaining := 10 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Line 1: "⬆ Update available: v0.6.2"
	icon := baseStyle().Render(" ⬆ ")
	label := styleStatsDim().Render("Update available:")
	space := baseStyle().Render(" ")
	version := styleID().Render(m.updateInfo.LatestVersion.String())
	heroLine := lipgloss.JoinHorizontal(lipgloss.Left, icon, label, space, version)

	// Line 2: Action instruction + countdown
	// Different content based on install method
	var actionText string
	if m.updateInfo.InstallMethod == update.InstallHomebrew {
		actionText = "Run: " + m.updateInfo.UpdateCommand
	} else {
		actionText = "Press [U] to update"
	}
	actionPart := " " + styleStatsDim().Render(actionText)
	countdownStr := styleStatsDim().Render(fmt.Sprintf("[%ds]", remaining))

	// Calculate spacing for right-aligned countdown
	heroWidth := lipgloss.Width(heroLine)
	actionWidth := lipgloss.Width(actionPart)
	countdownWidth := lipgloss.Width(countdownStr)

	// Use wider of hero or action line for alignment
	targetWidth := heroWidth
	if actionWidth > targetWidth {
		targetWidth = actionWidth + countdownWidth + 2
	}
	if targetWidth < 30 {
		targetWidth = 30
	}
	padding := targetWidth - actionWidth - countdownWidth
	if padding < 2 {
		padding = 2
	}

	paddingSpaces := ""
	if padding > 0 {
		paddingSpaces = baseStyle().Render(strings.Repeat(" ", padding))
	}
	infoLine := actionPart + paddingSpaces + countdownStr

	content := heroLine + "\n" + infoLine
	return newToastLayer(styleInfoToast().Render(content), width, height, mainBodyStart, mainBodyHeight)
}

// columnsToastLayer renders the columns toggle toast if visible.
func (m *App) columnsToastLayer(width, height, mainBodyStart, mainBodyHeight int) Layer {
	if !m.columnsToastVisible {
		return nil
	}
	elapsed := time.Since(m.columnsToastStart)
	remaining := 3 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Line 1: "Columns: ON" or "Columns: OFF"
	icon := baseStyle().Render(" ☰ ")
	label := styleStatsDim().Render("Columns:")
	space := baseStyle().Render(" ")
	var state string
	if m.columnsToastEnabled {
		state = styleLabelChecked().Render("ON")
	} else {
		state = styleStatsDim().Render("OFF")
	}
	heroLine := lipgloss.JoinHorizontal(lipgloss.Left, icon, label, space, state)
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
