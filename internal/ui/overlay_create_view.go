package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// header returns the overlay header text based on mode.
func (m *CreateOverlay) header() string {
	if m.isEditMode() {
		return fmt.Sprintf("EDIT: %s", m.editingBead.ID)
	}
	return "NEW BEAD"
}

// submitFooterText returns the action verb for the footer (Create/Save).
func (m *CreateOverlay) submitFooterText() string {
	if m.isEditMode() {
		return "Save"
	}
	return "Create"
}

// Styles for the create overlay

func styleCreateLabel() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(currentThemeWrapper().TextMuted()).
		MarginRight(1)
}

func styleCreatePill() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(currentThemeWrapper().TextMuted())
}

func styleCreatePillSelected() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(currentThemeWrapper().Primary())
}

func styleCreatePillFocused() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(currentThemeWrapper().Success()).
		Background(currentThemeWrapper().BorderDim())
}

// styleCreateInput returns a bordered input style.
// width is the desired VISUAL width; the style accounts for border (+2) internally.
func styleCreateInput(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(currentThemeWrapper().BorderDim()).
		Padding(0, 1).
		Width(width - 2) // Border adds 2 chars outside Width
}

// styleCreateInputFocused returns a bordered input style with focus highlight.
// width is the desired VISUAL width; the style accounts for border (+2) internally.
func styleCreateInputFocused(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(currentThemeWrapper().Success()).
		Padding(0, 1).
		Width(width - 2) // Border adds 2 chars outside Width
}

// styleCreateInputError returns a bordered input style with error highlight.
// width is the desired VISUAL width; the style accounts for border (+2) internally.
// (spec Section 4.4 - red border flash)
func styleCreateInputError(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(currentThemeWrapper().Error()).
		Padding(0, 1).
		Width(width - 2) // Border adds 2 chars outside Width
}

// Dimmed style for modal depth effect (spec Section 2.4)
func styleCreateDimmed() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(currentThemeWrapper().BorderNormal())
}

// View implements tea.Model - 5-zone HUD layout per spec Section 3.
// Uses the unified OverlayBuilder framework for consistent sizing and layout.
func (m *CreateOverlay) View() string {
	// Use responsive width via OverlayBuilder
	dialogWidth := m.calcDialogWidth()
	ob := NewOverlayBuilderWithWidth(dialogWidth)
	contentWidth := ob.ContentWidth()

	// Check if parent search is active for dimming effect (spec Section 2.4)
	parentSearchActive := m.parentCombo.IsDropdownOpen()

	// Parent combo box (always editable; placeholder shows root state)
	if m.parentCombo.Value() == "" {
		m.parentCombo = m.parentCombo.WithPlaceholder("No Parent (Root Item)")
	}

	// Header
	ob.Line(styleOverlayTitle().Render(m.header()))
	ob.Line(ob.Divider())
	ob.BlankLine()

	// Zone 1: Parent (anchor at top) - never dimmed
	parentLabel := m.renderSectionLabel("PARENT", m.focus == FocusParent, false)
	hint := "Shift+Tab"
	hintStyle := lipgloss.NewStyle().Foreground(currentThemeWrapper().TextMuted())
	parentLabelWidth := lipgloss.Width(parentLabel)
	padding := contentWidth - parentLabelWidth - lipgloss.Width(hint)
	if padding < 1 {
		padding = 1
	}
	ob.Line(parentLabel + hintStyle.Render(strings.Repeat(" ", padding)+hint))
	ob.Line(m.parentCombo.View())
	ob.BlankLine()

	// Zone 2: Title (hero element) - dimmed when parent search active
	ob.Line(m.renderSectionLabel("TITLE", m.focus == FocusTitle, parentSearchActive))
	titleView := m.renderTitleInput(contentWidth, parentSearchActive)
	ob.Line(titleView)
	ob.BlankLine()

	// Zone 2b: Description - dimmed when parent search active
	ob.Line(m.renderSectionLabel("DESCRIPTION", m.focus == FocusDescription, parentSearchActive))
	descView := m.renderDescInput(contentWidth, parentSearchActive)
	ob.Line(descView)
	ob.BlankLine()

	// Zone 3: Type and Priority - dimmed when parent search active
	var propsGrid string
	if m.isEditMode() {
		propsGrid = m.renderPriorityRow()
	} else {
		propsGrid = lipgloss.JoinVertical(lipgloss.Left,
			m.renderTypeRow(),
			"",
			m.renderPriorityRow(),
		)
	}
	if parentSearchActive {
		propsGrid = styleCreateDimmed().Render(propsGrid)
	}
	ob.Line(propsGrid)
	ob.BlankLine()

	// Zone 4: Labels - dimmed when parent search active
	ob.Line(m.renderSectionLabel("LABELS", m.focus == FocusLabels, parentSearchActive))
	labelsView := m.labelsCombo.View()
	if parentSearchActive {
		labelsView = styleCreateDimmed().Render(labelsView)
	}
	ob.Line(labelsView)
	ob.BlankLine()

	// Zone 5: Assignee - dimmed when parent search active
	ob.Line(m.renderSectionLabel("ASSIGNEE", m.focus == FocusAssignee, parentSearchActive))
	assigneeView := m.assigneeCombo.View()
	if parentSearchActive {
		assigneeView = styleCreateDimmed().Render(assigneeView)
	}
	ob.Line(assigneeView)
	ob.BlankLine()

	// Footer - show "Creating bead..." when submitting, otherwise show hints
	if m.isCreating {
		ob.FooterText("Creating bead...")
	} else {
		ob.Footer(m.footerHints())
	}

	return ob.Build()
}

// renderSectionLabel renders a section label with appropriate styling.
func (m *CreateOverlay) renderSectionLabel(label string, focused, dimmed bool) string {
	if dimmed {
		return styleCreateDimmed().Render(label)
	}
	if focused {
		return styleHelpSectionHeader().Render(label)
	}
	return styleCreateLabel().Render(label)
}

// renderTitleInput renders the title textarea with appropriate styling.
func (m *CreateOverlay) renderTitleInput(contentWidth int, dimmed bool) string {
	if dimmed {
		return styleCreateDimmed().Render(m.titleInput.View())
	}
	style := styleCreateInput(contentWidth)
	if m.focus == FocusTitle {
		style = styleCreateInputFocused(contentWidth)
	}
	if m.titleValidationError {
		style = styleCreateInputError(contentWidth)
	}
	return style.Render(m.titleInput.View())
}

// renderDescInput renders the description textarea with appropriate styling.
func (m *CreateOverlay) renderDescInput(contentWidth int, dimmed bool) string {
	if dimmed {
		return styleCreateDimmed().Render(m.descriptionInput.View())
	}
	style := styleCreateInput(contentWidth)
	if m.focus == FocusDescription {
		style = styleCreateInputFocused(contentWidth)
	}
	return style.Render(m.descriptionInput.View())
}

// footerHints returns the footer hints based on current state.
func (m *CreateOverlay) footerHints() []footerHint {
	if m.parentCombo.IsDropdownOpen() || m.labelsCombo.IsDropdownOpen() || m.assigneeCombo.IsDropdownOpen() {
		return []footerHint{
			{"⏎", "Select"},
			{"esc", "Revert"},
		}
	}
	if m.focus == FocusParent || m.focus == FocusLabels || m.focus == FocusAssignee {
		return []footerHint{
			{"↓", "Browse"},
			{"⏎", m.submitFooterText()},
			{"Tab", "Next"},
			{"esc", "Cancel"},
		}
	}
	// Description field: Enter inserts newlines, Ctrl+S submits
	if m.focus == FocusDescription {
		return []footerHint{
			{"^s", m.submitFooterText()},
			{"Tab", "Next"},
			{"esc", "Cancel"},
		}
	}
	// Default state
	return []footerHint{
		{"⏎", m.submitFooterText()},
		{"Tab", "Next"},
		{"esc", "Cancel"},
	}
}

// Layer returns a centered layer for the create overlay.
// Uses the shared BaseOverlayLayer to eliminate boilerplate.
func (m *CreateOverlay) Layer(width, height, topMargin, bottomMargin int) Layer {
	return BaseOverlayLayer(m.View, width, height, topMargin, bottomMargin)
}

func (m *CreateOverlay) renderTypeRow() string {
	var b strings.Builder

	// Row header
	headerStyle := lipgloss.NewStyle().Foreground(currentThemeWrapper().TextMuted())
	if m.focus == FocusType {
		headerStyle = lipgloss.NewStyle().Foreground(currentThemeWrapper().Secondary()).Bold(true)
	}
	// Add flash animation when type was auto-inferred (spec Section 5)
	if m.typeInferenceActive {
		headerStyle = lipgloss.NewStyle().Foreground(currentThemeWrapper().Warning()).Bold(true)
	}
	b.WriteString(headerStyle.Render("TYPE"))
	b.WriteString("\n")

	// Options rendered horizontally
	var options []string
	for i, label := range typeLabels {
		style := styleCreatePill()
		underlineHotkey := false

		if i == m.typeIndex {
			// Selected option - use parentheses format per spec mockup
			if m.focus == FocusType {
				style = styleCreatePillFocused()
				underlineHotkey = true
			} else {
				style = styleCreatePillSelected()
			}
			options = append(options, renderHorizontalOption(style, label, true, underlineHotkey))
		} else {
			if m.focus == FocusType {
				underlineHotkey = true
			}
			options = append(options, renderHorizontalOption(style, label, false, underlineHotkey))
		}
	}
	b.WriteString(strings.Join(options, "   "))

	return b.String()
}

func (m *CreateOverlay) renderPriorityRow() string {
	var b strings.Builder

	// Row header
	headerStyle := lipgloss.NewStyle().Foreground(currentThemeWrapper().TextMuted())
	if m.focus == FocusPriority {
		headerStyle = lipgloss.NewStyle().Foreground(currentThemeWrapper().Secondary()).Bold(true)
	}
	b.WriteString(headerStyle.Render("PRIORITY"))
	b.WriteString("\n")

	// Options rendered horizontally
	var options []string
	for i, label := range priorityLabels {
		style := styleCreatePill()
		underlineHotkey := false

		if i == m.priorityIndex {
			// Selected option - use parentheses format per spec mockup
			if m.focus == FocusPriority {
				style = styleCreatePillFocused()
				underlineHotkey = true
			} else {
				style = styleCreatePillSelected()
			}
			options = append(options, renderHorizontalOption(style, label, true, underlineHotkey))
		} else {
			if m.focus == FocusPriority {
				underlineHotkey = true
			}
			options = append(options, renderHorizontalOption(style, label, false, underlineHotkey))
		}
	}
	b.WriteString(strings.Join(options, "   "))

	return b.String()
}

// renderHorizontalOption renders a single option for horizontal Type/Priority rows.
// Selected items use parentheses format: (Task), unselected: Task
// When underline is true, the first letter is underlined (hotkey hint).
func renderHorizontalOption(style lipgloss.Style, label string, selected bool, underline bool) string {
	if label == "" {
		return ""
	}

	runes := []rune(label)
	innerStyle := style.Padding(0)
	underlineStyle := innerStyle.Underline(true)

	var content strings.Builder

	if selected {
		// When underline=true (focused), style parentheses with innerStyle
		// so they match the styled label text and don't get interrupted
		// by the ANSI escape sequences from the inner styled content.
		if underline {
			content.WriteString(innerStyle.Render("("))
		} else {
			content.WriteString("(")
		}
	}

	if underline {
		content.WriteString(underlineStyle.Render(string(runes[0])))
		if len(runes) > 1 {
			content.WriteString(innerStyle.Render(string(runes[1:])))
		}
	} else {
		content.WriteString(label)
	}

	if selected {
		if underline {
			content.WriteString(innerStyle.Render(")"))
		} else {
			content.WriteString(")")
		}
	}

	return style.Render(content.String())
}

// calcDialogWidth returns the responsive dialog width based on terminal width.
// Formula: min(120, max(44, int(0.7 * termWidth))) per ab-11wd spec.
func (m *CreateOverlay) calcDialogWidth() int {
	if m.termWidth == 0 {
		return 44 // Fallback to current default
	}
	width := int(float64(m.termWidth) * 0.7)
	if width < 44 {
		width = 44
	}
	if width > 120 {
		width = 120
	}
	return width
}

// SetSize updates the overlay dimensions based on terminal size.
// Called when tea.WindowSizeMsg is received.
func (m *CreateOverlay) SetSize(width, height int) {
	m.termWidth = width
	dialogWidth := m.calcDialogWidth()
	// Use OverlayContentWidth for consistency with the unified overlay framework
	contentWidth := OverlayContentWidth(dialogWidth)

	// Update text areas to fill the content area inside overlay padding
	// Textareas have their own border+padding, so subtract 4 more (2 border + 2 padding)
	m.titleInput.SetWidth(contentWidth - 4)
	m.descriptionInput.SetWidth(contentWidth - 4)

	// Update combo boxes - they should fit within contentWidth
	m.parentCombo = m.parentCombo.WithWidth(contentWidth)
	m.labelsCombo = m.labelsCombo.WithWidth(contentWidth)
	m.assigneeCombo = m.assigneeCombo.WithWidth(contentWidth)
}
