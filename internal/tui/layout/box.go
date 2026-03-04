// Package layout provides responsive layout calculations for the OpenCode TUI.
// This file contains box model definitions and styling helpers.
//
// The Box model represents a rectangular region with styling, supporting:
//   - Dimension specification (width, height)
//   - Lipgloss styling (background, foreground, borders)
//   - Content rendering
//   - Functional options for configuration
package layout

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// Box Model - Represents a styled container with box properties
// ============================================================================

// Box represents a rectangular region with styling for the TUI.
// It combines dimension (Rect) with lipgloss styling and content.
type Box struct {
	Rect
	Style   lipgloss.Style
	Content string
}

// BoxOption configures a Box using the functional options pattern.
type BoxOption func(*Box)

// NewBox creates a new Box with the given dimensions and options.
// Use functional options to configure the box with custom styling and content.
//
// Example:
//
//	box := NewBox(80, 24,
//		WithBoxStyle(lipgloss.NewStyle().Background(lipgloss.Color("#222222"))),
//		WithBoxContent("Hello, World!"),
//	)
func NewBox(width, height int, opts ...BoxOption) *Box {
	b := &Box{
		Rect: Rect{
			Width:  width,
			Height: height,
		},
		Style: lipgloss.NewStyle(),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// WithBoxStyle sets the lipgloss style for the box.
func WithBoxStyle(style lipgloss.Style) BoxOption {
	return func(b *Box) {
		b.Style = style
	}
}

// WithBoxContent sets the content for the box.
func WithBoxContent(content string) BoxOption {
	return func(b *Box) {
		b.Content = content
	}
}

// WithBoxBackground sets the background color.
func WithBoxBackground(color lipgloss.Color) BoxOption {
	return func(b *Box) {
		b.Style = b.Style.Background(color)
	}
}

// WithBoxForeground sets the foreground (text) color.
func WithBoxForeground(color lipgloss.Color) BoxOption {
	return func(b *Box) {
		b.Style = b.Style.Foreground(color)
	}
}

// WithBoxPadding sets padding on all sides.
func WithBoxPadding(padding int) BoxOption {
	return func(b *Box) {
		b.Style = b.Style.Padding(padding)
	}
}

// WithBoxPaddingXY sets horizontal and vertical padding.
func WithBoxPaddingXY(horizontal, vertical int) BoxOption {
	return func(b *Box) {
		b.Style = b.Style.Padding(vertical, horizontal)
	}
}

// WithBoxMargin sets margin on all sides.
func WithBoxMargin(margin int) BoxOption {
	return func(b *Box) {
		b.Style = b.Style.Margin(margin)
	}
}

// WithBoxBorder sets border on all sides.
func WithBoxBorder(border lipgloss.Border) BoxOption {
	return func(b *Box) {
		b.Style = b.Style.Border(border)
	}
}

// WithBoxBorderForeground sets border color.
func WithBoxBorderForeground(color lipgloss.Color) BoxOption {
	return func(b *Box) {
		b.Style = b.Style.BorderForeground(color)
	}
}

// Render renders the box content with its style.
// It properly accounts for borders and padding to prevent layout wrapping.
func (b *Box) Render(content string) string {
	// Get the frame size (borders) to calculate inner dimensions
	hFrame := b.Style.GetHorizontalFrameSize()
	vFrame := b.Style.GetVerticalFrameSize()

	// Calculate inner dimensions available for content
	innerWidth := b.Width - hFrame
	innerHeight := b.Height - vFrame

	// Ensure we don't have negative dimensions
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	return b.Style.Width(innerWidth).Height(innerHeight).Render(content)
}

// RenderWithFullDimensions renders content using the FULL allocated dimensions.
// Use this when you need the outer dimensions (including borders).
func (b *Box) RenderWithFullDimensions(content string) string {
	return b.Style.Width(b.Width).Height(b.Height).Render(content)
}

// GetInnerDimensions returns the inner content dimensions accounting for borders/padding.
func (b *Box) GetInnerDimensions() (width, height int) {
	hFrame := b.Style.GetHorizontalFrameSize()
	vFrame := b.Style.GetVerticalFrameSize()

	width = b.Width - hFrame
	height = b.Height - vFrame

	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	return width, height
}

// ============================================================================
// Panel Box - Pre-configured panel boxes for common UI sections
// ============================================================================

// PanelBox creates a panel-style box using the current theme.
func PanelBox(rect Rect) *Box {
	t := theme.CurrentTheme()
	return &Box{
		Rect: rect,
		Style: lipgloss.NewStyle().
			Background(t.Background()).
			Foreground(t.Text()).
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.BorderNormal()),
	}
}

// SidebarBox creates a sidebar-style box.
func SidebarBox(rect Rect) *Box {
	t := theme.CurrentTheme()
	return &Box{
		Rect: rect,
		Style: lipgloss.NewStyle().
			Background(t.Background()).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(t.BorderNormal()),
	}
}

// ChatBox creates a chat area box with proper padding.
func ChatBox(rect Rect) *Box {
	t := theme.CurrentTheme()

	return &Box{
		Rect: rect,
		Style: lipgloss.NewStyle().
			Background(t.Background()).
			Padding(24, 32), // From HTML design: 24px 32px
	}
}

// InputBox creates an input area box.
func InputBox(rect Rect) *Box {
	t := theme.CurrentTheme()
	return &Box{
		Rect: rect,
		Style: lipgloss.NewStyle().
			Background(t.Background()).
			Border(lipgloss.NormalBorder()),
	}
}

// StatusBox creates a status bar box.
func StatusBox(rect Rect) *Box {
	t := theme.CurrentTheme()
	return &Box{
		Rect: rect,
		Style: lipgloss.NewStyle().
			Background(t.BackgroundSecondary()),
	}
}

// WarningBox creates a warning bar box with orange background.
func WarningBox(rect Rect) *Box {
	return &Box{
		Rect: rect,
		Style: lipgloss.NewStyle().
			Background(lipgloss.Color("#ff9800")).
			Foreground(lipgloss.Color("#000000")).
			Height(rect.Height).
			Padding(0, 1),
	}
}

// ============================================================================
// Flex Layout - Horizontal and vertical joining with proper spacing
// ============================================================================

// FlexJoinHorizontal joins boxes horizontally with optional spacing.
func FlexJoinHorizontal(align lipgloss.Position, boxes ...*Box) string {
	views := make([]string, len(boxes))
	for i, box := range boxes {
		views[i] = box.Render(box.Content)
	}
	return lipgloss.JoinHorizontal(align, views...)
}

// FlexJoinVertical joins boxes vertically with optional spacing.
func FlexJoinVertical(align lipgloss.Position, boxes ...*Box) string {
	views := make([]string, len(boxes))
	for i, box := range boxes {
		views[i] = box.Render(box.Content)
	}
	return lipgloss.JoinVertical(align, views...)
}

// ============================================================================
// Place Content - Position content within a box
// ============================================================================

// PlaceContent positions content within a box using lipgloss placement.
func PlaceContent(box *Box, content string, hAlign, vAlign lipgloss.Position) string {
	// Get the inner dimensions accounting for padding/borders
	innerWidth := box.Width
	innerHeight := box.Height

	// Render content first to get its dimensions
	renderedContent := content
	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)

	// If content fits, use Place() for positioning
	if contentWidth <= innerWidth && contentHeight <= innerHeight {
		return lipgloss.Place(innerWidth, innerHeight, vAlign, hAlign, renderedContent)
	}

	// Content exceeds box - render with box style
	return box.Render(content)
}

// ============================================================================
// Adaptive Width/Height - Calculate dimensions that adapt to content
// ============================================================================

// MeasureContent returns the width and height needed to render content.
func MeasureContent(content string) (width, height int) {
	return lipgloss.Width(content), lipgloss.Height(content)
}

// AdaptiveWidth returns a width that fits content with optional padding.
func AdaptiveWidth(content string, padding int) int {
	width := lipgloss.Width(content)
	if width > 0 {
		return width + padding*2
	}
	return width
}

// AdaptiveHeight returns a height that fits content with optional padding.
func AdaptiveHeight(content string, padding int) int {
	height := lipgloss.Height(content)
	if height > 0 {
		return height + padding*2
	}
	return height
}

// ============================================================================
// Responsive Helpers
// ============================================================================

// ResponsiveValue returns value if terminal is large enough, otherwise returns fallback.
func ResponsiveValue(terminalSize, threshold int, value, fallback int) int {
	if terminalSize >= threshold {
		return value
	}
	return fallback
}

// ResponsiveWidth returns responsive width based on terminal width.
func ResponsiveWidth(terminalWidth int, thresholds []int, values []int) int {
	for i, threshold := range thresholds {
		if terminalWidth >= threshold && i < len(values) {
			return values[i]
		}
	}
	// Return last value or minimum
	if len(values) > 0 {
		return values[len(values)-1]
	}
	return 0
}

// ============================================================================
// Layout Builders - High-level layout construction
// ============================================================================

// MainLayout holds the lipgloss styles for all main layout sections.
type MainLayout struct {
	TopBar  lipgloss.Style
	Sidebar lipgloss.Style
	Content lipgloss.Style
	Status  lipgloss.Style
}

// BuildMainLayout creates the complete main layout structure with all styles.
func BuildMainLayout(dims *Dimensions, theme theme.Theme) MainLayout {
	// Build the main layout structure
	// Top Bar
	topBar := lipgloss.NewStyle().
		Width(dims.TopBar.Width).
		Height(dims.TopBar.Height).
		Background(theme.Background())

	// Sidebar
	sidebar := lipgloss.NewStyle().
		Width(dims.Sidebar.Width).
		Height(dims.Sidebar.Height).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.BorderNormal())

	// Content (workspace area)
	content := lipgloss.NewStyle().
		Width(dims.Workspace.Width).
		Height(dims.Workspace.Height).
		Background(theme.Background())

	// Status (bottom bar)
	status := lipgloss.NewStyle().
		Width(dims.BottomBar.Width).
		Height(dims.BottomBar.Height).
		Background(theme.BackgroundSecondary())

	return MainLayout{
		TopBar:  topBar,
		Sidebar: sidebar,
		Content: content,
		Status:  status,
	}
}
