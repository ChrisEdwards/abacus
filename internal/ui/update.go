package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all tea messages for the App model.
// It dispatches to focused handlers for overlay, background, and key messages.
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Overlay messages have highest priority
	if model, cmd, handled := m.handleOverlayMsg(msg); handled {
		return model, cmd
	}

	// Background and system messages
	if model, cmd, handled := m.handleBackgroundMsg(msg); handled {
		return model, cmd
	}

	// Key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeyMsg(keyMsg)
	}

	// Viewport updates when detail pane is focused
	if m.ShowDetails && m.focus == FocusDetails {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleBackgroundMsg processes background/system messages (spinner, tick, refresh, window).
// Returns (model, cmd, handled). If handled is false, the message was not processed.
func (m *App) handleBackgroundMsg(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.refreshInFlight {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd, true
		}
		return m, nil, true

	case tickMsg:
		cmds := []tea.Cmd{scheduleTick(m.refreshInterval)}
		if m.autoRefresh {
			if cmd := m.checkDBForChanges(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...), true

	case startBackgroundCommentLoadMsg:
		return m, m.loadCommentsInBackground(), true

	case commentLoadedMsg:
		if m.applyLoadedComment(msg) {
			m.updateViewportContent()
		}
		return m, nil, true

	case commentBatchLoadedMsg:
		refreshedDetail := false
		for _, res := range msg.results {
			refreshedDetail = m.applyLoadedComment(res) || refreshedDetail
		}
		if refreshedDetail {
			m.updateViewportContent()
		}
		return m, nil, true

	case refreshCompleteMsg:
		m.refreshInFlight = false
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastErrorSource = errorSourceRefresh
			m.lastRefreshStats = ""
			if !m.errorShownOnce {
				m.showErrorToast = true
				m.errorToastStart = time.Now()
				m.errorShownOnce = true
				return m, scheduleErrorToastTick(), true
			}
			return m, nil, true
		}
		if m.lastErrorSource == errorSourceRefresh {
			m.lastError = ""
			m.lastErrorSource = errorSourceNone
			m.showErrorToast = false
		}
		m.errorShownOnce = false
		m.applyRefresh(msg.roots, msg.digest, msg.dbModTime)
		if modTime, err := m.latestDBModTime(); err == nil && !modTime.IsZero() {
			m.lastDBModTime = modTime
		}
		return m, scheduleBackgroundCommentLoad(), true

	case eventualRefreshMsg:
		if m.activeOverlay != OverlayCreate {
			return m, m.forceRefresh(), true
		}
		return m, nil, true

	case errorToastTickMsg:
		if !m.showErrorToast {
			return m, nil, true
		}
		if time.Since(m.errorToastStart) >= 10*time.Second {
			m.showErrorToast = false
			return m, nil, true
		}
		return m, scheduleErrorToastTick(), true

	case copyToastTickMsg:
		if !m.showCopyToast {
			return m, nil, true
		}
		if time.Since(m.copyToastStart) >= 5*time.Second {
			m.showCopyToast = false
			return m, nil, true
		}
		return m, scheduleCopyToastTick(), true

	case updateAvailableMsg:
		if msg.info != nil && msg.info.UpdateAvailable {
			m.updateInfo = msg.info
			m.updateToastVisible = true
			m.updateToastStart = time.Now()
			return m, scheduleUpdateToastTick(), true
		}
		return m, nil, true

	case updateToastTickMsg:
		if !m.updateToastVisible {
			return m, nil, true
		}
		if time.Since(m.updateToastStart) >= 10*time.Second {
			m.updateToastVisible = false
			return m, nil, true
		}
		return m, scheduleUpdateToastTick(), true

	case appUpdateCompleteMsg:
		m.updateInProgress = false
		m.updateToastVisible = false
		if msg.err != nil {
			// Show failure toast (ab-w1wp)
			m.updateFailureToastVisible = true
			m.updateFailureToastStart = time.Now()
			m.updateFailureError = msg.err.Error()
			m.updateFailureCommand = "Download from releases"
			m.updateError = msg.err.Error()
			return m, scheduleUpdateFailureToastTick(), true
		}
		// Show success toast (ab-w1wp)
		m.updateSuccessToastVisible = true
		m.updateSuccessToastStart = time.Now()
		if m.updateInfo != nil {
			m.updateSuccessVersion = m.updateInfo.LatestVersion.String()
		}
		m.updateError = ""
		// Clear updateInfo so footer stops showing update indicator (ab-lsn1)
		m.updateInfo = nil
		return m, scheduleUpdateSuccessToastTick(), true

	case updateSuccessToastTickMsg:
		if !m.updateSuccessToastVisible {
			return m, nil, true
		}
		if time.Since(m.updateSuccessToastStart) >= 5*time.Second {
			m.updateSuccessToastVisible = false
			return m, nil, true
		}
		return m, scheduleUpdateSuccessToastTick(), true

	case updateFailureToastTickMsg:
		if !m.updateFailureToastVisible {
			return m, nil, true
		}
		if time.Since(m.updateFailureToastStart) >= 10*time.Second {
			m.updateFailureToastVisible = false
			return m, nil, true
		}
		return m, scheduleUpdateFailureToastTick(), true

	case tea.WindowSizeMsg:
		// Always update dimensions immediately so the model is correct
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resizeLastEvent = time.Now()

		// Schedule debounce tick if not already pending
		if !m.resizePending {
			m.resizePending = true
			return m, scheduleResizeDebounceTick(), true
		}
		return m, nil, true

	case resizeDebounceTickMsg:
		if time.Since(m.resizeLastEvent) < resizeDebounceDelay {
			// More resize events came in recently; schedule another tick
			return m, scheduleResizeDebounceTick(), true
		}

		// Debounce period elapsed - do the expensive redraw
		m.resizePending = false

		rawViewportWidth := int(float64(m.width)*0.45) - 2
		maxViewportWidth := m.width - minTreeWidth - 4
		m.viewport.Width = clampDimension(rawViewportWidth, minViewportWidth, maxViewportWidth)

		rawViewportHeight := m.height - 5
		maxViewportHeight := m.height - 2
		m.viewport.Height = clampDimension(rawViewportHeight, minViewportHeight, maxViewportHeight)

		m.applyViewportTheme()
		m.updateViewportContent()

		if m.createOverlay != nil {
			m.createOverlay.SetSize(m.width, m.height)
		}
		if m.commentOverlay != nil {
			m.commentOverlay.SetSize(m.width, m.height)
		}

		return m, nil, true
	}

	return nil, nil, false
}

// Message types for status operations
type statusUpdateCompleteMsg struct {
	err error
}

type statusToastTickMsg struct{}
