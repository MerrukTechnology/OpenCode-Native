// Package layout provides responsive layout calculations for the OpenCode TUI.
// It offers utilities for dimension calculation, split panes, and container wrappers
// that automatically adapt to terminal window size changes.
//
// The package provides:
//   - LayoutEngine: Responsive dimension calculator for all UI sections
//   - SplitPaneLayout: Horizontal/vertical split container layout
//   - Container: Wrapper with padding and borders
//   - Sizeable, Bindings, Focusable: Core interfaces for tea.Model components
package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// Layout Constants - Default dimensions based on HTML design
// ============================================================================
// These are now in TERMINAL CELLS (Rows/Cols), not pixels!
const (
	// CharacterCellWidth is the typical width of a character cell in pixels.
	// This is used to convert pixel dimensions to cell dimensions.
	CharacterCellWidth = 8
	// CharacterCellHeight is the typical height of a character cell in pixels.
	// This is used to convert pixel dimensions to cell dimensions.
	CharacterCellHeight = 16

	// DefaultTopBarHeight is the default height of the top bar.
	DefaultTopBarHeight = 1
	// DefaultWarningBarHeight is the default height of the warning bar.
	DefaultWarningBarHeight = 1
	// DefaultBottomBarHeight is the default height of the bottom bar.
	DefaultBottomBarHeight = 1
	// DefaultSidebarMinWidth is the minimum width of the sidebar.
	DefaultSidebarMinWidth = 20
	// DefaultSidebarMaxWidth is the maximum width of the sidebar.
	DefaultSidebarMaxWidth = 45
	// DefaultSidebarRatio is the default ratio of sidebar width to terminal width (25%).
	DefaultSidebarRatio = 0.25
	// DefaultInputMinHeight is the minimum height of the input area.
	DefaultInputMinHeight = 3
	// DefaultInputMaxHeight is the maximum height of the input area.
	DefaultInputMaxHeight = 15
	// DefaultMinTerminalWidth is the minimum terminal width supported.
	DefaultMinTerminalWidth = 60
	// DefaultMinTerminalHeight is the minimum terminal height supported.
	DefaultMinTerminalHeight = 15
)

// ============================================================================
// LayoutEngine - Main layout calculator for responsive TUI
// ============================================================================

// LayoutEngine calculates and manages responsive dimensions for all UI sections.
// It provides methods to compute layout based on terminal size and configuration.
type LayoutEngine struct {
	// Terminal dimensions in cells (used for layout calculations)
	terminalWidth  int
	terminalHeight int

	// Original pixel dimensions (for accurate conversion)
	pixelWidth  int
	pixelHeight int

	// Character cell size (determined dynamically from first resize)
	cellWidth  int
	cellHeight int

	// Configuration
	config LayoutConfig

	// Calculated dimensions (cached)
	dimensions Dimensions
}

// LayoutConfig contains configuration options for the layout engine.
// All values are used to customize the layout calculation.
type LayoutConfig struct {
	// Sidebar configuration
	SidebarWidth    int     // Fixed width, 0 = use ratio
	SidebarRatio    float64 // Percentage of total width (0.0-1.0)
	SidebarMinWidth int     // Minimum sidebar width
	SidebarMaxWidth int     // Maximum sidebar width

	// Fixed heights (0 = auto-calculate)
	TopBarHeight     int
	WarningBarHeight int
	BottomBarHeight  int

	// Input configuration
	InputMinHeight int
	InputMaxHeight int
	InputRatio     float64 // Ratio of remaining space for input

	// Padding
	ContentPaddingHorizontal int
	ContentPaddingVertical   int
}

// Dimensions holds the calculated dimensions for all UI sections.
// Each field represents a rectangular region with position and size.
type Dimensions struct {
	// Terminal
	TerminalWidth  int
	TerminalHeight int

	// Main sections
	TopBar      Rect
	WarningBar  Rect
	BottomBar   Rect
	ContentArea Rect // Combined sidebar + workspace

	// Content area splits
	Sidebar   Rect // Left panel (~320px or 25%)
	Workspace Rect // Right panel (remaining space)

	// Workspace splits
	ChatArea  Rect // Scrollable chat
	InputArea Rect // Input box
}

// Rect represents a rectangular region with position and size.
type Rect struct {
	X      int // Horizontal position (column)
	Y      int // Vertical position (row)
	Width  int // Width in characters
	Height int // Height in rows
}

// NewLayoutEngine creates a new layout engine with default configuration.
func NewLayoutEngine() *LayoutEngine {
	return &LayoutEngine{
		config: DefaultLayoutConfig(),
	}
}

// DefaultLayoutConfig returns the default layout configuration.
// This configuration uses sensible defaults for a standard chat interface.
func DefaultLayoutConfig() LayoutConfig {
	return LayoutConfig{
		SidebarWidth:    0, // 0 = use ratio
		SidebarRatio:    DefaultSidebarRatio,
		SidebarMinWidth: DefaultSidebarMinWidth,
		SidebarMaxWidth: DefaultSidebarMaxWidth,

		TopBarHeight:     DefaultTopBarHeight,
		WarningBarHeight: 0, // 0 by default, dynamically set to DefaultWarningBarHeight when needed
		BottomBarHeight:  DefaultBottomBarHeight,

		InputMinHeight: DefaultInputMinHeight,
		InputMaxHeight: DefaultInputMaxHeight,
		InputRatio:     0.20, // 20% for input, 80% for chat

		// Padding in cells (1-2 characters), NOT pixels (32px)!
		ContentPaddingHorizontal: 2,
		ContentPaddingVertical:   1,
	}
}

// WithConfig sets the layout configuration and returns the engine for chaining.
func (e *LayoutEngine) WithConfig(config LayoutConfig) *LayoutEngine {
	e.config = config
	return e
}

// Calculate computes all layout dimensions based on terminal size (in cells).
// This should be called with cell dimensions from the terminal.
func (e *LayoutEngine) Calculate(width, height int) Dimensions {
	e.terminalWidth = width
	e.terminalHeight = height

	// Ensure minimum terminal size so math never results in negative values
	e.terminalWidth = max(e.terminalWidth, DefaultMinTerminalWidth)
	e.terminalHeight = max(e.terminalHeight, DefaultMinTerminalHeight)

	dims := e.calculateDimensions()
	e.dimensions = dims
	return dims
}

// CalculateFromPixels computes layout from pixel dimensions.
// It stores the original pixel values and converts to cells using
// the stored cell size, or estimates it if not yet determined.
func (e *LayoutEngine) CalculateFromPixels(pixelWidth, pixelHeight int) Dimensions {
	// Store original pixel dimensions
	e.pixelWidth = pixelWidth
	e.pixelHeight = pixelHeight

	// Always recalculate cell size from current measurements to ensure
	// accurate conversion even when pixel values change (e.g., when dialogs/panels
	// open and close). The cell size can vary slightly between terminal sessions.
	e.terminalWidth = pixelWidth / CharacterCellWidth
	e.terminalHeight = pixelHeight / CharacterCellHeight

	// Determine cell size from current measurement
	if e.terminalWidth > 0 && e.terminalHeight > 0 {
		e.cellWidth = pixelWidth / e.terminalWidth
		e.cellHeight = pixelHeight / e.terminalHeight
		// Sanity check: cells should be in reasonable range
		if e.cellWidth < 4 || e.cellWidth > 20 {
			e.cellWidth = CharacterCellWidth
		}
		if e.cellHeight < 8 || e.cellHeight > 30 {
			e.cellHeight = CharacterCellHeight
		}
	}

	// Ensure minimum terminal size
	e.terminalWidth = max(e.terminalWidth, DefaultMinTerminalWidth)
	e.terminalHeight = max(e.terminalHeight, DefaultMinTerminalHeight)

	dims := e.calculateDimensions()
	e.dimensions = dims
	return dims
}

// calculateDimensions performs the actual modular dimension calculations.
func (e *LayoutEngine) calculateDimensions() Dimensions {
	dims := Dimensions{
		TerminalWidth:  e.terminalWidth,
		TerminalHeight: e.terminalHeight,
	}

	// Calculate fixed vertical sections
	dims.TopBar = e.calcTopBar()
	dims.WarningBar = e.calcWarningBar()
	dims.BottomBar = e.calcBottomBar()

	// Calculate content area (remaining space after fixed sections)
	contentHeight := e.terminalHeight - dims.TopBar.Height - dims.WarningBar.Height - dims.BottomBar.Height
	contentHeight = Max(contentHeight, 1)

	dims.ContentArea = Rect{
		X:      0,
		Y:      dims.TopBar.Height + dims.WarningBar.Height,
		Width:  e.terminalWidth,
		Height: contentHeight,
	}

	// Calculate sidebar and workspace
	dims.Sidebar = e.calcSidebar()
	dims.Workspace = e.calcWorkspace(dims.Sidebar.Width)

	// Calculate workspace splits (chat + input)
	dims.ChatArea, dims.InputArea = e.calcWorkspaceSplits(
		dims.Workspace.Width,
		dims.Workspace.Height,
	)

	return dims
}

// calcTopBar calculates the top bar dimensions.
func (e *LayoutEngine) calcTopBar() Rect {
	height := e.config.TopBarHeight
	if height <= 0 {
		height = DefaultTopBarHeight
	}
	height = Min(height, e.terminalHeight/5) // Don't take more than 20% of screen

	return Rect{X: 0, Y: 0, Width: e.terminalWidth, Height: height}
}

// calcWarningBar calculates the warning bar dimensions.
func (e *LayoutEngine) calcWarningBar() Rect {
	height := e.config.WarningBarHeight
	if height < 0 {
		height = 0
	}
	height = Min(height, e.terminalHeight/6)

	return Rect{X: 0, Y: e.calcTopBar().Height, Width: e.terminalWidth, Height: height}
}

// calcBottomBar calculates the bottom bar dimensions.
func (e *LayoutEngine) calcBottomBar() Rect {
	height := e.config.BottomBarHeight
	if height <= 0 {
		height = DefaultBottomBarHeight
	}
	height = Min(height, e.terminalHeight/5)

	return Rect{X: 0, Y: e.terminalHeight - height, Width: e.terminalWidth, Height: height}
}

// calcSidebar calculates the sidebar dimensions.
func (e *LayoutEngine) calcSidebar() Rect {
	var width int

	// Use fixed width if configured
	if e.config.SidebarWidth > 0 {
		width = e.config.SidebarWidth
	} else {
		// Calculate based on ratio
		width = int(float64(e.terminalWidth) * e.config.SidebarRatio)
	}
	// Apply min/max constraints
	width = Max(width, e.config.SidebarMinWidth)
	width = Min(width, e.config.SidebarMaxWidth)
	width = Min(width, e.terminalWidth/2) // Never more than 50%

	topBarHeight := e.calcTopBar().Height
	warningBarHeight := e.calcWarningBar().Height

	return Rect{
		X:      0,
		Y:      topBarHeight + warningBarHeight,
		Width:  width,
		Height: e.terminalHeight - topBarHeight - warningBarHeight - e.calcBottomBar().Height,
	}
}

// calcWorkspace calculates the workspace (right panel) dimensions.
func (e *LayoutEngine) calcWorkspace(sidebarWidth int) Rect {
	topBarHeight := e.calcTopBar().Height
	warningBarHeight := e.calcWarningBar().Height
	bottomBarHeight := e.calcBottomBar().Height

	width := e.terminalWidth - sidebarWidth
	height := e.terminalHeight - topBarHeight - warningBarHeight - bottomBarHeight

	return Rect{
		X:      sidebarWidth,
		Y:      topBarHeight + warningBarHeight,
		Width:  width,
		Height: height,
	}
}

// calcWorkspaceSplits calculates chat area and input area within the workspace.
func (e *LayoutEngine) calcWorkspaceSplits(width, height int) (chat, input Rect) {
	// Calculate input height
	inputHeight := int(float64(height) * e.config.InputRatio)
	inputHeight = Max(inputHeight, e.config.InputMinHeight)
	inputHeight = Min(inputHeight, e.config.InputMaxHeight)
	inputHeight = Min(inputHeight, height/2) // Never more than 50%

	// Chat takes remaining space
	chatHeight := height - inputHeight

	chat = Rect{X: 0, Y: 0, Width: width, Height: chatHeight}
	input = Rect{X: 0, Y: chatHeight, Width: width, Height: inputHeight}

	return chat, input
}

// GetDimensions returns the cached dimensions.
// Returns zero dimensions if Calculate() hasn't been called.
func (e *LayoutEngine) GetDimensions() Dimensions {
	return e.dimensions
}

// GetPixelDimensions returns the original pixel dimensions.
// Returns (0, 0) if CalculateFromPixels() hasn't been called.
func (e *LayoutEngine) GetPixelDimensions() (int, int) {
	return e.pixelWidth, e.pixelHeight
}

// GetTerminalDimensions returns the terminal dimensions in cells.
func (e *LayoutEngine) GetTerminalDimensions() (int, int) {
	return e.terminalWidth, e.terminalHeight
}

// GetSidebarStyle returns a lipgloss style for the sidebar with proper dimensions.
func (e *LayoutEngine) GetSidebarStyle() lipgloss.Style {
	dims := e.dimensions
	return lipgloss.NewStyle().
		Width(dims.Sidebar.Width).
		Height(dims.Sidebar.Height).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#333333"))
}

// GetWorkspaceStyle returns a lipgloss style for the workspace.
func (e *LayoutEngine) GetWorkspaceStyle() lipgloss.Style {
	dims := e.dimensions
	return lipgloss.NewStyle().
		Width(dims.Workspace.Width).
		Height(dims.Workspace.Height)
}

// GetChatStyle returns a lipgloss style for the chat area with padding.
func (e *LayoutEngine) GetChatStyle() lipgloss.Style {
	dims := e.dimensions
	return lipgloss.NewStyle().
		Width(dims.ChatArea.Width).
		Height(dims.ChatArea.Height).
		Padding(e.config.ContentPaddingVertical, e.config.ContentPaddingHorizontal)
}

// GetInputStyle returns a lipgloss style for the input area.
func (e *LayoutEngine) GetInputStyle() lipgloss.Style {
	dims := e.dimensions
	return lipgloss.NewStyle().
		Width(dims.InputArea.Width).
		Height(dims.InputArea.Height).
		Padding(0, e.config.ContentPaddingHorizontal)
}

// ============================================================================
// Utility Functions
// ============================================================================

// Width returns the printable width of a string (ignoring ANSI escape codes).
func Width(s string) int {
	return lipgloss.Width(s)
}

// Height returns the number of lines in a string when rendered.
func Height(s string) int {
	return lipgloss.Height(s)
}

// Clamp ensures value is within min/max bounds.
// If value is less than min, returns min.
// If value is greater than max, returns max.
// Otherwise returns value unchanged.
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Max returns the larger of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the smaller of two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PixelsToCells converts pixel dimensions to terminal cell dimensions.
// This is necessary because tea.WindowSizeMsg provides pixel dimensions,
// but the layout engine works with terminal cell dimensions.
// Typical character cell size is 8x16 pixels (width x height).
//
// Note: This function uses conservative estimates. For more accurate conversion,
// use LayoutEngine.CalculateFromPixels() which can determine actual cell size.
func PixelsToCells(pixelWidth, pixelHeight int) (cellsWidth, cellsHeight int) {
	// Handle edge cases - return defaults for invalid input
	if pixelWidth <= 0 || pixelHeight <= 0 {
		return DefaultMinTerminalWidth, DefaultMinTerminalHeight
	}

	cellsWidth = pixelWidth / CharacterCellWidth
	cellsHeight = pixelHeight / CharacterCellHeight

	// Ensure we always have at least minimum dimensions
	cellsWidth = Max(cellsWidth, DefaultMinTerminalWidth)
	cellsHeight = Max(cellsHeight, DefaultMinTerminalHeight)

	return cellsWidth, cellsHeight
}
