// Package layout provides responsive layout calculations for the OpenCode TUI.
package layout

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Sidebar represents the left pane of the TUI application.
// It displays file tree, session list, or other navigation content.
type Sidebar struct {
	title      string
	items      []string
	selected   int
	dimensions PaneDimensions
	focused    bool
}

// NewSidebar creates a new Sidebar component with default values.
func NewSidebar() *Sidebar {
	return &Sidebar{
		title:    "Sessions",
		items:    []string{},
		selected: 0,
		focused:  false,
	}
}

// SetTitle sets the sidebar title.
func (s *Sidebar) SetTitle(title string) *Sidebar {
	s.title = title
	return s
}

// SetItems sets the sidebar items.
func (s *Sidebar) SetItems(items []string) *Sidebar {
	s.items = items
	return s
}

// SetSelected sets the selected item index.
func (s *Sidebar) SetSelected(index int) *Sidebar {
	s.selected = index
	return s
}

// SetDimensions sets the dimensions for the sidebar.
func (s *Sidebar) SetDimensions(width, height int) *Sidebar {
	s.dimensions = PaneDimensions{Width: width, Height: height}
	return s
}

// SetFocused sets the focused state of the sidebar.
func (s *Sidebar) SetFocused(focused bool) *Sidebar {
	s.focused = focused
	return s
}

// GetDimensions returns the current dimensions of the sidebar.
func (s *Sidebar) GetDimensions() PaneDimensions {
	return s.dimensions
}

// GetSelected returns the currently selected item index.
func (s *Sidebar) GetSelected() int {
	return s.selected
}

// Render renders the sidebar with the given content.
func (s *Sidebar) Render(content string) string {
	t := theme.CurrentTheme()
	borderStyle := GetBorderStyle(PaneSidebar, s.getEffectiveFocus())
	borderColor := GetBorderColor(PaneSidebar, s.getEffectiveFocus())

	// Create a temporary style to get frame sizes
	tempStyle := lipgloss.NewStyle().Border(borderStyle, true)
	hFrame := tempStyle.GetHorizontalFrameSize()
	vFrame := tempStyle.GetVerticalFrameSize()

	// Calculate inner dimensions accounting for borders
	innerWidth := s.dimensions.Width - hFrame
	innerHeight := s.dimensions.Height - vFrame

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	style := lipgloss.NewStyle().
		Border(borderStyle, true).
		BorderForeground(borderColor).
		Background(t.Background()).
		Foreground(t.Text()).
		Width(innerWidth).
		Height(innerHeight)

	return style.Render(content)
}

// RenderDefault renders the sidebar with default styling and content.
func (s *Sidebar) RenderDefault() string {
	t := theme.CurrentTheme()
	borderStyle := GetBorderStyle(PaneSidebar, s.getEffectiveFocus())
	borderColor := GetBorderColor(PaneSidebar, s.getEffectiveFocus())

	// Create a temporary style to get frame sizes
	tempStyle := lipgloss.NewStyle().Border(borderStyle, true)
	hFrame := tempStyle.GetHorizontalFrameSize()
	vFrame := tempStyle.GetVerticalFrameSize()

	// Calculate inner dimensions
	innerWidth := s.dimensions.Width - hFrame
	innerHeight := s.dimensions.Height - vFrame

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Create sidebar style
	sidebarStyle := lipgloss.NewStyle().
		Border(borderStyle, true).
		BorderForeground(borderColor).
		Background(t.Background()).
		Foreground(t.Text()).
		Width(innerWidth).
		Height(innerHeight)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Width(s.dimensions.Width - 4)

	// Item styles
	normalItemStyle := lipgloss.NewStyle().
		Foreground(t.Text())

	selectedItemStyle := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Background(t.BackgroundSecondary()).
		Bold(true)

	mutedItemStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted())

	// Build content
	var content []string

	// Add title
	content = append(content, titleStyle.Render(" "+s.title))
	content = append(content, "") // Empty line

	// Add items
	if len(s.items) == 0 {
		content = append(content, mutedItemStyle.Render("  No items"))
	} else {
		for i, item := range s.items {
			if i == s.selected {
				content = append(content, selectedItemStyle.Render(" > "+item))
			} else {
				content = append(content, normalItemStyle.Render("   "+item))
			}
		}
	}

	// Join content vertically
	joinedContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	// Render with style
	return sidebarStyle.Render(joinedContent)
}

// RenderItem renders a single sidebar item with the given index and content.
func (s *Sidebar) RenderItem(index int, itemContent string, isSelected bool) string {
	t := theme.CurrentTheme()

	if isSelected {
		return lipgloss.NewStyle().
			Foreground(t.Primary()).
			Background(t.BackgroundSecondary()).
			Bold(true).
			Render(" > " + itemContent)
	}

	return lipgloss.NewStyle().
		Foreground(t.Text()).
		Render("   " + itemContent)
}

// getEffectiveFocus returns the effective focus state for border styling.
func (s *Sidebar) getEffectiveFocus() PaneType {
	if s.focused {
		return PaneSidebar
	}
	return -1 // No pane focused
}
