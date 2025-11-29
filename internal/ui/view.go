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
		status += " " + styleFilterInfo.Render(filterLabel)
	}

	if m.lastRefreshStats != "" {
		refreshStr := fmt.Sprintf(" Δ %s", m.lastRefreshStats)
		if m.showRefreshFlash && time.Since(m.lastRefreshTime) < refreshFlashDuration {
			refreshStr = styleSelected.Render(refreshStr)
		} else {
			refreshStr = styleStatsDim.Render(refreshStr)
			m.showRefreshFlash = false
		}
		status += " " + refreshStr
	}

	title := "ABACUS"
	if m.version != "" {
		title = fmt.Sprintf("ABACUS v%s", m.version)
	}

	// Build header with right-aligned error indicator if present
	leftContent := styleAppHeader.Render(title) + " " + status
	var header string
	if m.lastError != "" {
		rightContent := styleErrorIndicator.Render("⚠ Refresh error (e)")
		availableWidth := m.width - lipgloss.Width(leftContent) - lipgloss.Width(rightContent) - 2
		if availableWidth > 0 {
			header = leftContent + strings.Repeat(" ", availableWidth) + rightContent
		} else {
			header = leftContent + " " + rightContent
		}
	} else {
		header = leftContent
	}
	treeViewStr := m.renderTreeView()

	var mainBody string
	listHeight := clampDimension(m.height-4, minListHeight, m.height-2)
	if m.ShowDetails {
		leftStyle := stylePane
		rightStyle := stylePane
		if m.focus == FocusTree {
			leftStyle = stylePaneFocused
		} else {
			rightStyle = stylePaneFocused
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
		mainBody = stylePane.Width(singleWidth).Height(listHeight).Render(treeViewStr)
	}

	var bottomBar string
	if m.searching {
		bottomBar = m.textInput.View()
	} else {
		bottomBar = m.renderFooter()
	}

	// Help overlay takes visual precedence over everything
	if m.showHelp {
		helpOverlay := renderHelpOverlay(m.keys, m.width, m.height-2)
		return fmt.Sprintf("%s\n%s\n%s", header, helpOverlay, bottomBar)
	}

	// Overlay toast on mainBody if visible (copy toast takes priority)
	if toast := m.renderCopyToast(); toast != "" {
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
	titleLine := "⚠ Refresh Error"
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

	return styleErrorToast.Render(content)
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

	return styleSuccessToast.Render(content)
}
