package ui

import (
	"fmt"
	"strings"

	"abacus/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func (m *App) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Determine if background should be dimmed (overlay active)
	dimmed := m.activeOverlay != OverlayNone || m.showHelp

	// Apply dimmed palette for background elements when overlay is active.
	// We restore to standard theme before rendering overlays - dimming only
	// affects the background content, not the overlays themselves.
	var restoreTheme func()
	if dimmed {
		restoreTheme = useDimmedTheme()
	} else {
		restoreTheme = func() {} // No theme override when no overlay
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
	} else if m.activeOverlay == OverlayComment && m.commentOverlay != nil {
		if layer := m.commentOverlay.Layer(m.width, m.height, headerHeight, bottomMargin); layer != nil {
			overlayLayers = append(overlayLayers, layer)
		}
	} else if m.showHelp {
		overlayLayers = append(overlayLayers, newHelpOverlayLayer(m.keys, m.width, m.height, headerHeight, bottomMargin))
	}

	content := fmt.Sprintf("%s\n%s\n%s", header, mainBody, bottomBar)
	base := wrapWithBackground(content)

	// Restore standard theme before rendering overlays and toasts.
	// Dimming only applies to background content.
	restoreTheme()

	var overlayErrorLayer Layer
	if m.activeOverlay == OverlayCreate && m.createOverlay != nil {
		overlayErrorLayer = m.errorToastLayer(m.width, m.height, mainBodyStart, mainBodyHeight)
	}

	var toastLayer Layer
	toastFactories := []func(int, int, int, int) Layer{
		m.themeToastLayer,
		m.columnsToastLayer,
		m.updateSuccessToastLayer,
		m.updateFailureToastLayer,
		m.updateToastLayer,
		m.deleteToastLayer,
		m.createToastLayer,
		m.commentToastLayer,
		m.newAssigneeToastLayer,
		m.newLabelToastLayer,
		m.labelsToastLayer,
		m.statusToastLayer,
		m.copyToastLayer,
	}
	// Only add errorToastLayer if not already handled by overlayErrorLayer
	// to avoid double-rendering error toasts when create overlay is open
	if overlayErrorLayer == nil {
		toastFactories = append(toastFactories, m.errorToastLayer)
	}
	for _, factory := range toastFactories {
		if layer := factory(m.width, m.height, mainBodyStart, mainBodyHeight); layer != nil {
			toastLayer = layer
			break
		}
	}

	if len(overlayLayers) > 0 {
		canvas := NewCanvas(m.width, m.height)
		canvas.DrawStringAt(0, 0, base)

		// Overlays render with standard theme (restored above)
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
