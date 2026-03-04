// Package layout provides responsive layout calculations for the OpenCode TUI.
// This file contains helper functions for rendering with the layout engine.
package layout

import (
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// Renderer - Main rendering helper that combines layout engine with theming
// ============================================================================

// Renderer combines layout calculations with theme-aware rendering.
type Renderer struct {
	engine *LayoutEngine
	theme  theme.Theme
}

// NewRenderer creates a new renderer with the given layout engine.
func NewRenderer(engine *LayoutEngine) *Renderer {
	return &Renderer{
		engine: engine,
		theme:  theme.CurrentTheme(),
	}
}

// NewRendererWithTheme creates a new renderer with custom theme.
func NewRendererWithTheme(engine *LayoutEngine, t theme.Theme) *Renderer {
	return &Renderer{
		engine: engine,
		theme:  t,
	}
}

// RenderTopBar renders the top bar section.
func (r *Renderer) RenderTopBar(content string) string {
	dims := r.engine.GetDimensions()

	// If content fits, use it; otherwise truncate
	availableWidth := dims.TopBar.Width
	contentWidth := lipgloss.Width(content)

	if contentWidth > availableWidth {
		content = truncateString(content, availableWidth)
	}

	style := lipgloss.NewStyle().
		Width(dims.TopBar.Width).
		Height(dims.TopBar.Height).
		Background(r.theme.Background()).
		Foreground(r.theme.Text()).
		Padding(0, 1)

	return style.Render(content)
}

// RenderWarningBar renders the warning bar section.
func (r *Renderer) RenderWarningBar(content string) string {
	dims := r.engine.GetDimensions()

	// If content fits, use it; otherwise truncate
	availableWidth := dims.WarningBar.Width
	contentWidth := lipgloss.Width(content)

	if contentWidth > availableWidth {
		content = truncateString(content, availableWidth)
	}

	style := lipgloss.NewStyle().
		Width(dims.WarningBar.Width).
		Height(dims.WarningBar.Height).
		Background(r.theme.Warning()).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)

	return style.Render(content)
}

// RenderBottomBar renders the bottom bar (status bar) section.
func (r *Renderer) RenderBottomBar(content string) string {
	dims := r.engine.GetDimensions()

	// If content fits, use it; otherwise truncate
	availableWidth := dims.BottomBar.Width
	contentWidth := lipgloss.Width(content)

	if contentWidth > availableWidth {
		content = truncateString(content, availableWidth)
	}

	style := lipgloss.NewStyle().
		Width(dims.BottomBar.Width).
		Height(dims.BottomBar.Height).
		Background(r.theme.BackgroundSecondary()).
		Foreground(r.theme.Text()).
		Padding(0, 1)

	return style.Render(content)
}

// RenderSidebar renders the sidebar section with proper borders.
func (r *Renderer) RenderSidebar(content string) string {
	dims := r.engine.GetDimensions()

	// Get inner content dimensions
	innerWidth := dims.Sidebar.Width - 1 // Account for border
	innerHeight := dims.Sidebar.Height

	// Wrap content to fit
	wrapped := r.wrapToWidth(content, innerWidth)
	lines := strings.Split(wrapped, "\n")

	// Truncate if too many lines
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	content = strings.Join(lines, "\n")

	style := lipgloss.NewStyle().
		Width(dims.Sidebar.Width).
		Height(dims.Sidebar.Height).
		Background(r.theme.BackgroundSecondary()).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(r.theme.BorderDim())

	return style.Render(content)
}

// RenderChatArea renders the main chat area.
func (r *Renderer) RenderChatArea(content string) string {
	dims := r.engine.GetDimensions()

	// Use small padding
	hPad := r.engine.config.ContentPaddingHorizontal
	vPad := r.engine.config.ContentPaddingVertical

	innerWidth := dims.ChatArea.Width - (hPad * 2)
	innerHeight := dims.ChatArea.Height - (vPad * 2)

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	wrapped := r.wrapToWidth(content, innerWidth)
	lines := strings.Split(wrapped, "\n")

	// Handle scrolling - show only visible lines
	if len(lines) > innerHeight {
		lines = lines[len(lines)-innerHeight:]
	}

	content = strings.Join(lines, "\n")

	style := lipgloss.NewStyle().
		Width(dims.ChatArea.Width).
		Height(dims.ChatArea.Height).
		Background(r.theme.Background()).
		Foreground(r.theme.Text()).
		Padding(vPad, hPad)

	return style.Render(content)
}

// RenderInputArea renders the input area.
func (r *Renderer) RenderInputArea(content string) string {
	dims := r.engine.GetDimensions()
	hPad := r.engine.config.ContentPaddingHorizontal

	innerWidth := dims.InputArea.Width - (hPad * 2)
	innerHeight := dims.InputArea.Height - 2

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	wrapped := r.wrapToWidth(content, innerWidth)
	lines := strings.Split(wrapped, "\n")

	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	content = strings.Join(lines, "\n")

	style := lipgloss.NewStyle().
		Width(dims.InputArea.Width).
		Height(dims.InputArea.Height).
		Background(r.theme.BackgroundSecondary()).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(r.theme.BorderNormal()).
		Padding(0, hPad)

	return style.Render(content)
}

// RenderMainLayout renders the complete main layout with all sections.
func (r *Renderer) RenderMainLayout(
	topBarContent,
	warningContent,
	sidebarContent,
	chatContent,
	inputContent,
	bottomBarContent string,
) string {
	// Render each section
	topBar := r.RenderTopBar(topBarContent)
	warningBar := ""
	if warningContent != "" {
		warningBar = r.RenderWarningBar(warningContent)
	}

	// Sidebar and workspace (horizontal split)
	sidebar := r.RenderSidebar(sidebarContent)
	workspaceContent := r.RenderChatArea(chatContent) + "\n" + r.RenderInputArea(inputContent)

	// Join sidebar and workspace horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, workspaceContent)

	// Join all sections vertically
	var sections []string
	sections = append(sections, topBar)
	if warningBar != "" {
		sections = append(sections, warningBar)
	}
	sections = append(sections, mainContent)
	sections = append(sections, r.RenderBottomBar(bottomBarContent))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// ============================================================================
// Utility Rendering Functions
// ============================================================================

// wrapToWidth wraps text to fit within the given width.
func (r *Renderer) wrapToWidth(content string, width int) string {
	if width < 1 {
		return content
	}

	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		if lipgloss.Width(line) <= width {
			result = append(result, line)
			continue
		}

		// Wrap long lines
		result = append(result, r.wrapLine(line, width)...)
	}

	return strings.Join(result, "\n")
}

// wrapLine wraps a single line to fit within the width.
func (r *Renderer) wrapLine(line string, width int) []string {
	if width < 1 {
		return []string{line}
	}

	var lines []string
	currentLine := ""

	for _, word := range strings.Fields(line) {
		wordWidth := lipgloss.Width(word)

		if lipgloss.Width(currentLine)+wordWidth+1 <= width {
			if currentLine != "" {
				currentLine += " " + word
			} else {
				currentLine = word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			// Handle words longer than width
			if wordWidth > width {
				// Split long words
				for wordWidth > width {
					lines = append(lines, word[:width])
					word = word[width:]
					wordWidth = lipgloss.Width(word)
				}
				currentLine = word
			} else {
				currentLine = word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// truncateString truncates a string to fit within the given width.
func truncateString(s string, maxWidth int) string {
	if maxWidth < 1 {
		return ""
	}

	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Find the longest prefix that fits
	for i := len(s); i > 0; i-- {
		if lipgloss.Width(s[:i]) <= maxWidth {
			return s[:i]
		}
	}

	return ""
}

// ============================================================================
// Powerline-Style Status Bar Rendering
// ============================================================================

// RenderStatusSegment renders a single status segment with Powerline style.
func (r *Renderer) RenderStatusSegment(text string, isActive bool) string {
	if isActive {
		style := lipgloss.NewStyle().
			Background(r.theme.Primary()).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)
		return style.Render(text)
	}

	style := lipgloss.NewStyle().
		Background(r.theme.BackgroundDarker()).
		Foreground(r.theme.Text()).
		Padding(0, 1)
	return style.Render(text)
}

// RenderStatusBar renders a complete status bar with segments.
func (r *Renderer) RenderStatusBar(segments []string, activeIndex int) string {
	var styledSegments []string

	for i, segment := range segments {
		styledSegments = append(styledSegments, r.RenderStatusSegment(segment, i == activeIndex))
	}

	// Join with chevron separators
	separatorStyle := lipgloss.NewStyle().Foreground(r.theme.BorderDim())
	separator := separatorStyle.Render(" ") // ChevronRight
	result := strings.Join(styledSegments, separator)

	// Wrap with base style
	style := lipgloss.NewStyle().
		Height(r.engine.GetDimensions().BottomBar.Height)

	return style.Render(result)
}

// ============================================================================
// Diff View Rendering
// ============================================================================

// RenderDiffLine renders a single diff line with appropriate styling.
func (r *Renderer) RenderDiffLine(lineType, content string) string {
	var style lipgloss.Style

	switch lineType {
	case "+":
		style = lipgloss.NewStyle().
			Background(r.theme.DiffAddedBg()).
			Foreground(r.theme.DiffAdded()).
			Padding(0, 1)
	case "-":
		style = lipgloss.NewStyle().
			Background(r.theme.DiffRemovedBg()).
			Foreground(r.theme.DiffRemoved()).
			Padding(0, 1)
	case "@":
		style = lipgloss.NewStyle().
			Background(r.theme.BackgroundDarker()).
			Foreground(r.theme.DiffHunkHeader()).
			Bold(true).
			Padding(0, 1)
	default:
		style = lipgloss.NewStyle().
			Background(r.theme.Background()).
			Foreground(r.theme.DiffContext()).
			Padding(0, 1)
	}

	return style.Render(content)
}

// RenderDiffBlock renders a complete diff block.
func (r *Renderer) RenderDiffBlock(lines []string) string {
	dims := r.engine.GetDimensions()
	maxLines := dims.ChatArea.Height - 4

	var renderedLines []string
	for i, line := range lines {
		if i >= maxLines {
			break
		}

		var lineType string
		switch {
		case len(line) > 0 && line[0] == '+':
			lineType = "+"
		case len(line) > 0 && line[0] == '-':
			lineType = "-"
		case len(line) > 1 && line[0:2] == "@@":
			lineType = "@"
		default:
			lineType = " "
		}

		renderedLines = append(renderedLines, r.RenderDiffLine(lineType, line))
	}

	return lipgloss.JoinVertical(lipgloss.Left, renderedLines...)
}

// ============================================================================
// Chat Message Rendering
// ============================================================================

// RenderChatMessage renders a chat message with role and content.
func (r *Renderer) RenderChatMessage(role, content, timestamp string) string {
	dims := r.engine.GetDimensions()
	maxWidth := dims.ChatArea.Width - 64 // Account for padding

	// Wrap content
	wrappedContent := r.wrapToWidth(content, maxWidth)

	// Build the message
	var roleStyle lipgloss.Style
	if role == "user" {
		roleStyle = lipgloss.NewStyle().
			Foreground(r.theme.Primary()).
			Bold(true)
	} else {
		roleStyle = lipgloss.NewStyle().
			Foreground(r.theme.Accent()).
			Bold(true)
	}

	roleLabel := roleStyle.Render(role + ":")

	timestampStyle := lipgloss.NewStyle().
		Foreground(r.theme.TextMuted())
	timestampLabel := timestampStyle.Render(timestamp)

	contentStyle := lipgloss.NewStyle().
		Background(r.theme.BackgroundSecondary()).
		Foreground(r.theme.Text()).
		Padding(1, 2)

	messageContent := roleLabel + " " + timestampLabel + "\n" + contentStyle.Render(wrappedContent)

	return messageContent
}

// RenderToolBadge renders a tool name as a badge.
func (r *Renderer) RenderToolBadge(toolName string) string {
	style := lipgloss.NewStyle().
		Background(r.theme.Info()).
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 1).
		MarginRight(1)

	return style.Render(toolName)
}
