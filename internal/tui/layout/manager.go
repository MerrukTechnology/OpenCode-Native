// Package layout provides responsive layout calculations for the OpenCode TUI.
// It offers a LayoutManager for calculating 3-pane layouts with header, sidebar,
// and footer sections, along with component implementations for each pane.
package layout

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// PaneDimensions holds the width and height of a layout pane.
type PaneDimensions struct {
	Width  int
	Height int
}

// LayoutManager calculates and manages the dimensions of a 3-pane TUI layout.
// It divides the terminal window into:
//   - Header: fixed 3 lines at the top
//   - Footer: fixed 1 line at the bottom
//   - Sidebar: 20-25% of remaining width on the left
//   - Main: remaining space for the primary content
type LayoutManager struct {
	width  int
	height int

	header  PaneDimensions
	footer  PaneDimensions
	sidebar PaneDimensions
	main    PaneDimensions

	// Focus state for panes
	focusedPane PaneType
}

// PaneType represents the different panes in the layout.
type PaneType int

const (
	PaneHeader PaneType = iota
	PaneSidebar
	PaneMain
	PaneFooter
)

// NewLayoutManager creates a new LayoutManager with the given dimensions.
func NewLayoutManager(width, height int) *LayoutManager {
	m := &LayoutManager{
		focusedPane: PaneMain,
	}
	m.Calculate(width, height)
	return m
}

// Calculate recalculates all pane dimensions based on the given window size.
func (m *LayoutManager) Calculate(width, height int) {
	m.width = width
	m.height = height

	// Header: 3 lines fixed height
	m.header = PaneDimensions{
		Width:  width,
		Height: 3,
	}

	// Footer: 1 line fixed height
	m.footer = PaneDimensions{
		Width:  width,
		Height: 1,
	}

	// Calculate available space for sidebar and main content
	// Subtract header (3) and footer (1) = 4 lines
	availableHeight := height - m.header.Height - m.footer.Height
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Sidebar: 20-25% of remaining width (use 22.5% as midpoint)
	// Minimum sidebar width of 20 characters
	sidebarWidth := int(float64(width) * 0.225)
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}
	// Maximum sidebar width of 40 characters or 35% of width
	maxSidebarWidth := int(float64(width) * 0.35)
	if maxSidebarWidth > 40 {
		maxSidebarWidth = 40
	}
	if sidebarWidth > maxSidebarWidth {
		sidebarWidth = maxSidebarWidth
	}

	m.sidebar = PaneDimensions{
		Width:  sidebarWidth,
		Height: availableHeight,
	}

	// Main content: remaining width
	m.main = PaneDimensions{
		Width:  width - sidebarWidth,
		Height: availableHeight,
	}
}

// GetHeaderDimensions returns the dimensions for the header pane.
func (m *LayoutManager) GetHeaderDimensions() PaneDimensions {
	return m.header
}

// GetFooterDimensions returns the dimensions for the footer pane.
func (m *LayoutManager) GetFooterDimensions() PaneDimensions {
	return m.footer
}

// GetSidebarDimensions returns the dimensions for the sidebar pane.
func (m *LayoutManager) GetSidebarDimensions() PaneDimensions {
	return m.sidebar
}

// GetMainDimensions returns the dimensions for the main content pane.
func (m *LayoutManager) GetMainDimensions() PaneDimensions {
	return m.main
}

// GetWidth returns the current total width.
func (m *LayoutManager) GetWidth() int {
	return m.width
}

// GetHeight returns the current total height.
func (m *LayoutManager) GetHeight() int {
	return m.height
}

// SetFocusedPane sets which pane currently has focus.
func (m *LayoutManager) SetFocusedPane(pane PaneType) {
	m.focusedPane = pane
}

// GetFocusedPane returns the currently focused pane.
func (m *LayoutManager) GetFocusedPane() PaneType {
	return m.focusedPane
}

// PaneAppearance holds border and color styling for a pane.
type PaneAppearance struct {
	Border lipgloss.Border
	Color  lipgloss.AdaptiveColor
}

// getPaneAppearance returns the border and color for a pane based on focus state.
func getPaneAppearance(pane PaneType, focused PaneType) PaneAppearance {
	t := theme.CurrentTheme()
	isFocused := pane == focused

	return PaneAppearance{
		Border: map[bool]lipgloss.Border{true: t.GetBorderStyle().Focused, false: t.GetBorderStyle().Normal}[isFocused],
		Color:  map[bool]lipgloss.AdaptiveColor{true: t.BorderFocused(), false: t.BorderNormal()}[isFocused],
	}
}

// calculateInnerWidth calculates the inner width accounting for borders.
func calculateInnerWidth(width, frame int) int {
	inner := width - frame
	if inner < 1 {
		return 1
	}
	return inner
}

// renderPane renders a content string with the appropriate style for a given pane.
func (m *LayoutManager) renderPane(pane PaneType, dims PaneDimensions, content string) string {
	appearance := getPaneAppearance(pane, m.focusedPane)
	frame := lipgloss.NewStyle().Border(appearance.Border, true).GetHorizontalFrameSize()
	innerWidth := calculateInnerWidth(dims.Width, frame)

	return lipgloss.NewStyle().
		Border(appearance.Border, true).
		BorderForeground(appearance.Color).
		Width(innerWidth).
		Height(dims.Height).
		Render(content)
}

// GetBorderStyle returns the appropriate lipgloss.Border for a pane
// based on whether it is focused or unfocused.
func GetBorderStyle(pane PaneType, focused PaneType) lipgloss.Border {
	return getPaneAppearance(pane, focused).Border
}

// GetBorderColor returns the appropriate border color for a pane
// based on whether it is focused or unfocused.
func GetBorderColor(pane PaneType, focused PaneType) lipgloss.AdaptiveColor {
	return getPaneAppearance(pane, focused).Color
}

// GetPaneStyle returns a lipgloss.Style configured for a specific pane
// with appropriate borders and colors.
func GetPaneStyle(pane PaneType, focused PaneType) lipgloss.Style {
	t := theme.CurrentTheme()
	border := GetBorderStyle(pane, focused)
	borderClr := GetBorderColor(pane, focused)

	return lipgloss.NewStyle().
		Border(border, true).
		BorderForeground(borderClr).
		Background(t.Background()).
		Foreground(t.Text())
}

// JoinHorizontal joins multiple content blocks horizontally with proper spacing.
func JoinHorizontal(blocks ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
}

// JoinVertical joins multiple content blocks vertically with proper spacing.
func JoinVertical(blocks ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

// Render assembles and renders the complete layout with all panes.
// It takes content strings for each pane and returns the complete UI.
// CRITICAL: This properly calculates inner dimensions accounting for borders
// to prevent layout wrapping issues.
func (m *LayoutManager) Render(headerContent, sidebarContent, mainContent, footerContent string) string {
	t := theme.CurrentTheme()

	// Use helper to render each pane - handles style, frame, and dimensions
	headerRendered := m.renderPane(PaneHeader, m.header, headerContent)
	sidebarRendered := m.renderPane(PaneSidebar, m.sidebar, sidebarContent)
	mainRendered := m.renderPane(PaneMain, m.main, mainContent)
	footerRendered := m.renderPane(PaneFooter, m.footer, footerContent)

	// Join sidebar and main content horizontally (side by side)
	contentArea := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarRendered,
		mainRendered,
	)

	// Join header, content area, and footer vertically
	fullLayout := lipgloss.JoinVertical(
		lipgloss.Left,
		headerRendered,
		contentArea,
		footerRendered,
	)

	// Apply overall background
	return lipgloss.NewStyle().
		Background(t.Background()).
		Render(fullLayout)
}
