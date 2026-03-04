// Package layout provides responsive layout calculations for the OpenCode TUI.
package layout

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Footer represents the bottom bar of the TUI application.
// It displays status information, keyboard shortcuts, and contextual hints.
type Footer struct {
	status     string
	hint       string
	dimensions Rect
	focused    bool
}

// NewFooter creates a new Footer component with default values.
func NewFooter() *Footer {
	return &Footer{
		status:  "Ready",
		hint:    "Ctrl+H: Help",
		focused: false,
	}
}

// SetStatus sets the status message displayed in the footer.
func (f *Footer) SetStatus(status string) *Footer {
	f.status = status
	return f
}

// SetHint sets the hint message displayed in the footer.
func (f *Footer) SetHint(hint string) *Footer {
	f.hint = hint
	return f
}

// SetDimensions sets the dimensions for the footer.
func (f *Footer) SetDimensions(dims Rect) *Footer {
	f.dimensions = dims
	return f
}

// SetFocused sets the focused state of the footer.
func (f *Footer) SetFocused(focused bool) *Footer {
	f.focused = focused
	return f
}

// GetDimensions returns the current dimensions of the footer.
func (f *Footer) GetDimensions() Rect {
	return f.dimensions
}

// Render renders the footer with the given content.
func (f *Footer) Render(content string) string {
	t := theme.CurrentTheme()
	borderStyle := GetBorderStyle(PaneFooter, f.getEffectiveFocus())
	borderColor := GetBorderColor(PaneFooter, f.getEffectiveFocus())

	// Create a temporary style to get frame sizes
	tempStyle := lipgloss.NewStyle().Border(borderStyle, true)
	hFrame := tempStyle.GetHorizontalFrameSize()
	vFrame := tempStyle.GetVerticalFrameSize()

	// Calculate inner dimensions
	innerWidth := f.dimensions.Width - hFrame
	innerHeight := f.dimensions.Height - vFrame

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	style := lipgloss.NewStyle().
		Border(borderStyle, true).
		BorderForeground(borderColor).
		Background(t.BackgroundSecondary()).
		Foreground(t.Text()).
		Width(innerWidth).
		Height(innerHeight)

	return style.Render(content)
}

// RenderDefault renders the footer with default styling and content.
func (f *Footer) RenderDefault() string {
	t := theme.CurrentTheme()
	borderStyle := GetBorderStyle(PaneFooter, f.getEffectiveFocus())
	borderColor := GetBorderColor(PaneFooter, f.getEffectiveFocus())

	// Create a temporary style to get frame sizes
	tempStyle := lipgloss.NewStyle().Border(borderStyle, true)
	hFrame := tempStyle.GetHorizontalFrameSize()
	vFrame := tempStyle.GetVerticalFrameSize()

	// Calculate inner dimensions
	innerWidth := f.dimensions.Width - hFrame
	innerHeight := f.dimensions.Height - vFrame

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Create footer style
	footerStyle := lipgloss.NewStyle().
		Border(borderStyle, true).
		BorderForeground(borderColor).
		Background(t.BackgroundSecondary()).
		Foreground(t.Text()).
		Width(innerWidth).
		Height(innerHeight)

	// Create content with status and hint
	statusStyle := lipgloss.NewStyle().
		Foreground(t.Success())

	hintStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted())

	separatorStyle := lipgloss.NewStyle().
		Foreground(t.BorderDim())

	// Build content - left align status, right align hint
	statusContent := statusStyle.Render(f.status)
	hintContent := hintStyle.Render(f.hint)

	// Calculate available space for centering
	availableWidth := f.dimensions.Width - 4
	statusWidth := lipgloss.Width(statusContent)
	hintWidth := lipgloss.Width(hintContent)
	sepWidth := lipgloss.Width(separatorStyle.Render(" │ "))

	totalContentWidth := statusWidth + sepWidth + hintWidth
	var content string

	if totalContentWidth < availableWidth {
		// Pad with spaces between status and hint
		padding := availableWidth - totalContentWidth
		// Use approximately 60% of padding on left, 40% on right
		leftPadding := (padding * 3) / 5

		// Create left-padded status content
		leftPadded := lipgloss.NewStyle().
			Width(statusWidth + leftPadding).
			Align(lipgloss.Left).
			Render(statusContent)

		content = leftPadded +
			separatorStyle.Render(" │ ") +
			hintContent
		content = lipgloss.NewStyle().
			Width(availableWidth).
			Align(lipgloss.Left).
			Render(content)
	} else {
		content = statusContent + separatorStyle.Render(" │ ") + hintContent
	}

	return footerStyle.Render(content)
}

// getEffectiveFocus returns the effective focus state for border styling.
func (f *Footer) getEffectiveFocus() PaneType {
	if f.focused {
		return PaneFooter
	}
	return -1 // No pane focused
}
