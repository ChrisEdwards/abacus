package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

//nolint:unused // used in tests, kept for future overlay implementations
func newCenteredOverlayLayer(content string, width, height int, topMargin, bottomMargin int) Layer {
	return LayerFunc(func() *Canvas {
		if content == "" {
			return nil
		}
		overlayWidth, overlayHeight := blockDimensions(content)
		if overlayWidth <= 0 || overlayHeight <= 0 {
			return nil
		}

		surface := NewSecondarySurface(overlayWidth, overlayHeight)
		surface.Draw(0, 0, content)
		x, y := centeredOffsets(width, height, overlayWidth, overlayHeight, topMargin, bottomMargin)

		surface.Canvas.SetOffset(x, y)
		return surface.Canvas
	})
}

func newToastLayer(content string, width, height int, mainBodyStart, mainBodyHeight int) Layer {
	return LayerFunc(func() *Canvas {
		if content == "" {
			return nil
		}
		toastWidth, toastHeight := blockDimensions(content)
		if toastWidth <= 0 || toastHeight <= 0 {
			return nil
		}

		surface := NewPrimarySurface(toastWidth, toastHeight)
		surface.Draw(0, 0, content)

		x := width - toastWidth - 2
		if x < 0 {
			x = 0
		}

		if mainBodyHeight <= 0 {
			mainBodyHeight = height
		}
		y := mainBodyStart + mainBodyHeight - toastHeight - 1
		if y < mainBodyStart {
			y = mainBodyStart
		}
		if y < 0 {
			y = 0
		}

		surface.Canvas.SetOffset(x, y)
		return surface.Canvas
	})
}

func blockDimensions(content string) (int, int) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	width := maxLineWidth(lines)
	if width <= 0 {
		width = lipgloss.Width(normalized)
	}
	if width <= 0 {
		width = 1
	}
	height := lipgloss.Height(normalized)
	if height <= 0 {
		height = len(lines)
	}
	if height <= 0 {
		height = 1
	}
	return width, height
}

func centeredOffsets(containerWidth, containerHeight, contentWidth, contentHeight, topMargin, bottomMargin int) (int, int) {
	if topMargin < 0 {
		topMargin = 0
	}
	if bottomMargin < 0 {
		bottomMargin = 0
	}

	usableHeight := containerHeight - topMargin - bottomMargin
	if usableHeight < contentHeight {
		usableHeight = contentHeight
	}

	y := topMargin
	if usableHeight > contentHeight {
		y = topMargin + (usableHeight-contentHeight)/2
	}
	maxY := containerHeight - bottomMargin - contentHeight
	if y > maxY {
		y = maxY
	}
	if y < topMargin {
		y = topMargin
	}
	if y < 0 {
		y = 0
	}

	x := (containerWidth - contentWidth) / 2
	if x < 0 {
		x = 0
	}

	return x, y
}
