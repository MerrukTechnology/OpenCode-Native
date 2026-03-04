// Package chat provides UI components for rendering chat messages and diffs.
// This file contains the main chat viewer with border container.
package chat

import (
	"fmt"
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/shared"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// Chat Viewer - Main container for chat messages with border
// ============================================================================

// ChatViewerConfig contains configuration options for the ChatViewer.
// It controls appearance settings like maximum width, borders, titles, and agent status.
type ChatViewerConfig struct {
	// MaxWidth is the maximum width of the chat viewer in characters (default: 900)
	MaxWidth int

	// ShowBorder controls whether to show the border around the viewer (default: true)
	ShowBorder bool

	// Title is the title to display in the border (default: "Chat")
	Title string

	// AgentName is the name of the active agent displayed in the border (default: "Coder")
	AgentName string

	// AgentActive indicates if the agent is currently active, displayed in the border status
	AgentActive bool

	// ShowToolBadges indicates whether to show tool badges for tool calls
	ShowToolBadges bool
}

// DefaultChatViewerConfig returns a ChatViewerConfig with default values.
func DefaultChatViewerConfig() ChatViewerConfig {
	return ChatViewerConfig{
		MaxWidth:       900,
		ShowBorder:     true,
		Title:          "Chat",
		AgentName:      "Coder",
		AgentActive:    true,
		ShowToolBadges: true,
	}
}

// ChatViewer renders the main chat container with messages and border.
type ChatViewer struct {
	config     ChatViewerConfig
	theme      theme.Theme
	dimensions layout.Dimensions
	diffViewer *DiffViewer
}

// NewChatViewer creates a new chat viewer with the given configuration.
func NewChatViewer(config ChatViewerConfig) *ChatViewer {
	if config.MaxWidth == 0 {
		config.MaxWidth = DefaultChatViewerConfig().MaxWidth
	}

	cv := &ChatViewer{
		config:     config,
		theme:      theme.CurrentTheme(),
		diffViewer: NewDiffViewer(config.MaxWidth - 4), // Account for padding/border
	}

	return cv
}

// SetTheme sets the theme for the chat viewer.
func (cv *ChatViewer) SetTheme(t theme.Theme) {
	cv.theme = t
	cv.diffViewer.SetTheme(t)
}

// SetDimensions sets the layout dimensions for the chat viewer.
func (cv *ChatViewer) SetDimensions(dims layout.Dimensions) {
	cv.dimensions = dims
}

// Render renders the chat viewer and returns the formatted string.
func (cv *ChatViewer) Render(messages []string) string {
	if cv.theme == nil {
		cv.theme = theme.CurrentTheme()
	}

	var builder strings.Builder

	// Render border with title
	if cv.config.ShowBorder {
		builder.WriteString(cv.renderBorderTitle())
	}

	// Render messages
	for i, msg := range messages {
		builder.WriteString(msg)
		if i < len(messages)-1 {
			// Add spacing between messages (24px equivalent)
			builder.WriteString("\n\n")
		}
	}

	return builder.String()
}

// renderBorderTitle renders the border with the agent title.
func (cv *ChatViewer) renderBorderTitle() string {
	// Build the title string
	var title strings.Builder

	// Agent icon and name
	agentIcon := "🤖"
	agentStatus := "Active"
	if !cv.config.AgentActive {
		agentStatus = "Idle"
	}

	title.WriteString("╭─ ")
	title.WriteString(agentIcon)
	title.WriteString(" ")
	title.WriteString(cv.config.AgentName)
	title.WriteString(" (")
	title.WriteString(agentStatus)
	title.WriteString(") ")
	title.WriteString(" ─╮")

	// Create border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cv.theme.Colors().BorderFocus).
		Padding(0, 1)

	// Safely calculate target width
	targetWidth := cv.dimensions.ChatArea.Width
	if targetWidth == 0 && cv.dimensions.Workspace.Width > 0 {
		targetWidth = cv.dimensions.Workspace.Width
	} else if targetWidth == 0 {
		targetWidth = cv.config.MaxWidth
	}

	// Subtract the frame size so Lipgloss doesn't overflow!
	if targetWidth > 0 {
		innerW := targetWidth - borderStyle.GetHorizontalFrameSize()
		if innerW < 1 {
			innerW = 1
		}
		borderStyle = borderStyle.Width(innerW)
	}

	return borderStyle.Render(title.String()) + "\n"
}

// calculateAvailableWidth calculates the available width for message content.
func (cv *ChatViewer) calculateAvailableWidth() int {
	width := cv.config.MaxWidth

	if cv.dimensions.ChatArea.Width > 0 {
		width = cv.dimensions.ChatArea.Width
	} else if cv.dimensions.Workspace.Width > 0 && cv.dimensions.Workspace.Width < width {
		width = cv.dimensions.Workspace.Width
	}

	if cv.config.ShowBorder {
		// Use native measurement to subtract exactly the right amount
		dummy := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
		width -= dummy.GetHorizontalFrameSize()
	}

	if width < 1 {
		width = 1
	}
	return width
}

// RenderWithDiff renders the chat viewer including a diff viewer.
func (cv *ChatViewer) RenderWithDiff(messages []string, diffContent string, filePath string) string {
	if cv.theme == nil {
		cv.theme = theme.CurrentTheme()
	}

	var builder strings.Builder

	// Render border with title
	if cv.config.ShowBorder {
		builder.WriteString(cv.renderBorderTitle())
	}

	// Render messages
	for i, msg := range messages {
		builder.WriteString(msg)
		if i < len(messages)-1 {
			builder.WriteString("\n\n")
		}
	}

	// Add diff viewer if content is provided
	if diffContent != "" {
		builder.WriteString("\n\n")
		builder.WriteString(cv.renderDiffViewer(diffContent, filePath))
	}

	return builder.String()
}

// renderDiffViewer renders the diff viewer with file header.
func (cv *ChatViewer) renderDiffViewer(diffContent, filePath string) string {
	// Create bordered container for diff
	diffBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(cv.theme.Colors().BorderDim).
		Padding(0, 1)

	availableWidth := cv.calculateAvailableWidth()
	if availableWidth > 0 {
		innerW := availableWidth - diffBorderStyle.GetHorizontalFrameSize()
		if innerW < 1 {
			innerW = 1
		}
		diffBorderStyle = diffBorderStyle.Width(innerW)
	}

	// Render file header
	fileHeader := cv.diffViewer.RenderFileHeader(filePath)

	// Render diff content
	diffContentRendered := cv.diffViewer.Render(diffContent)

	// Combine file header and diff
	diffContentFull := fileHeader + diffContentRendered

	return diffBorderStyle.Render(diffContentFull)
}

// RenderToolBadge renders a tool badge with success/failure indicator.
func (cv *ChatViewer) RenderToolBadge(toolName string, success bool) string {
	if cv.theme == nil {
		cv.theme = theme.CurrentTheme()
	}

	// Style for the checkmark
	checkStyle := lipgloss.NewStyle().
		Foreground(cv.theme.Colors().AccentGreen)
	checkMark := shared.IconCheck

	if !success {
		checkStyle = checkStyle.Foreground(cv.theme.Colors().AccentRed)
		checkMark = shared.IconErrorAlt
	}

	// Tool badge style
	badgeStyle := lipgloss.NewStyle().
		Background(cv.theme.Colors().BgSurface).
		Foreground(cv.theme.Colors().TextMain).
		Padding(0, 1).
		MarginRight(1)

	badge := badgeStyle.Render(toolName)
	check := checkStyle.Render(checkMark)

	return check + " " + badge
}

// RenderToolBadges renders multiple tool badges.
func (cv *ChatViewer) RenderToolBadges(toolNames []string, successes []bool) string {
	if len(toolNames) == 0 {
		return ""
	}

	var builder strings.Builder

	for i, toolName := range toolNames {
		success := true
		if i < len(successes) {
			success = successes[i]
		}
		builder.WriteString(cv.RenderToolBadge(toolName, success))
	}

	return builder.String()
}

// RenderMessageBubble renders a single message bubble.
func (cv *ChatViewer) RenderMessageBubble(role, content string) string {
	if cv.theme == nil {
		cv.theme = theme.CurrentTheme()
	}

	var bubbleStyle lipgloss.Style
	var icon string

	switch role {
	case "user":
		bubbleStyle = cv.theme.Chat().User
		icon = "👤"
	case "assistant":
		bubbleStyle = cv.theme.Chat().Assistant
		icon = "▶"
	default:
		bubbleStyle = cv.theme.Chat().Assistant
		icon = "●"
	}

	// Role label style
	roleLabel := cv.theme.Chat().Role

	// Render role label with icon
	roleText := fmt.Sprintf("%s %s:", icon, role)
	roleStr := roleLabel.Render(roleText)

	// Ensure message doesn't overflow horizontally
	availWidth := cv.calculateAvailableWidth()
	contentRendered := bubbleStyle.Width(availWidth).Render(content)

	return roleStr + "\n" + contentRendered
}

// GetDimensions returns the current dimensions.
func (cv *ChatViewer) GetDimensions() layout.Dimensions {
	return cv.dimensions
}

// GetMaxWidth returns the maximum width.
func (cv *ChatViewer) GetMaxWidth() int {
	return cv.config.MaxWidth
}

// SetMaxWidth sets the maximum width.
func (cv *ChatViewer) SetMaxWidth(width int) {
	cv.config.MaxWidth = width
	if cv.diffViewer != nil {
		// Recreate diff viewer with new width
		cv.diffViewer = NewDiffViewer(width - 4)
		cv.diffViewer.SetTheme(cv.theme)
	}
}
