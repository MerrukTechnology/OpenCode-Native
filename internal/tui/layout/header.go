// Package layout provides responsive layout calculations for the OpenCode TUI.
package layout

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Header represents the top bar of the TUI application.
// It displays application title, mode indicators, and status information.
type Header struct {
	title      string
	mode       string
	dimensions Rect
	focused    bool
}

// NewHeader creates a new Header component with default values.
func NewHeader() *Header {
	return &Header{
		title:   "OpenCode Pro",
		mode:    "Chat",
		focused: false,
	}
}

// SetTitle sets the header title.
func (h *Header) SetTitle(title string) *Header {
	h.title = title
	return h
}

// SetMode sets the current mode displayed in the header.
func (h *Header) SetMode(mode string) *Header {
	h.mode = mode
	return h
}

// SetDimensions sets the dimensions for the header.
func (h *Header) SetDimensions(dims Rect) *Header {
	h.dimensions = dims
	return h
}

// SetFocused sets the focused state of the header.
func (h *Header) SetFocused(focused bool) *Header {
	h.focused = focused
	return h
}

// GetDimensions returns the current dimensions of the header.
func (h *Header) GetDimensions() Rect {
	return h.dimensions
}

// Render renders the header with the given content.
func (h *Header) Render(content string) string {
	style := theme.CurrentStyles.TopBar

	borderStyle := GetBorderStyle(PaneHeader, h.getEffectiveFocus())
	borderColor := GetBorderColor(PaneHeader, h.getEffectiveFocus())

	// Subtract the frame (borders + padding) from the total allocated space
	innerWidth := h.dimensions.Width - style.GetHorizontalFrameSize()
	innerHeight := h.dimensions.Height - style.GetVerticalFrameSize()

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	return style.Width(innerWidth).Height(innerHeight).Border(borderStyle, true).BorderForeground(borderColor).Render(content)
}

// RenderDefault renders the header with default styling and content.
func (h *Header) RenderDefault() string {
	style := theme.CurrentStyles.TopBar
	t := theme.CurrentTheme()

	borderStyle := GetBorderStyle(PaneHeader, h.getEffectiveFocus())
	borderColor := GetBorderColor(PaneHeader, h.getEffectiveFocus())

	// Calculate inner dimensions
	innerWidth := h.dimensions.Width - style.GetHorizontalFrameSize()
	innerHeight := h.dimensions.Height - style.GetVerticalFrameSize()

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Create content with title and mode
	titleStyle := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true)

	modeStyle := lipgloss.NewStyle().
		Foreground(t.Info())

	separatorStyle := lipgloss.NewStyle().
		Foreground(t.BorderDim())

	content := titleStyle.Render(" "+h.title) +
		separatorStyle.Render(" │ ") +
		modeStyle.Render(h.mode)

	// Center the content horizontally inside the inner boundary
	content = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(content)

	return style.Border(borderStyle, true).Width(innerWidth).Height(innerHeight).Align(lipgloss.Center).BorderForeground(borderColor).Background(t.BackgroundSecondary()).Foreground(t.Text()).Render(content)
}

// getEffectiveFocus returns the effective focus state for border styling.
func (h *Header) getEffectiveFocus() PaneType {
	if h.focused {
		return PaneHeader
	}
	return -1 // No pane focused
}
