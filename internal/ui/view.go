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

	title := "ABACUS"
	if m.version != "" {
		title = fmt.Sprintf("ABACUS v%s", m.version)
	}

	// Build header with repo name on right
	leftContent := styleAppHeader.Render(title) + " " + status
	rightContent := styleFooterMuted.Render("Repo: " + m.repoName)
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

	// Status overlay takes visual precedence over help
	if m.activeOverlay == OverlayStatus && m.statusOverlay != nil {
		overlay := m.statusOverlay.View()
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

	// Overlay toast on mainBody if visible (status toast > copy toast > error toast)
	if toast := m.renderStatusToast(); toast != "" {
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

// renderStatusToast renders the status change success toast if visible.
func (m *App) renderStatusToast() string {
	if !m.statusToastVisible || m.statusToastMessage == "" {
		return ""
	}
	elapsed := time.Since(m.statusToastStart)
	remaining := 7 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// Build content with bead context
	titleLine := fmt.Sprintf("Status -> %s", m.statusToastMessage)

	// Truncate title if too long
	title := m.statusToastTitle
	if len(title) > 35 {
		title = title[:32] + "..."
	}
	beadLine := fmt.Sprintf("%s %s", m.statusToastBeadID, title)

	countdownStr := fmt.Sprintf("[%ds]", remaining)

	// Calculate toast width
	toastWidth := lipgloss.Width(beadLine)
	if w := lipgloss.Width(titleLine); w > toastWidth {
		toastWidth = w
	}
	if toastWidth < 30 {
		toastWidth = 30
	}

	padding := toastWidth - len(countdownStr)
	if padding < 0 {
		padding = 0
	}

	content := fmt.Sprintf("%s\n%s\n%s%s", titleLine, beadLine, strings.Repeat(" ", padding), countdownStr)
	return styleSuccessToast.Render(content)
}
