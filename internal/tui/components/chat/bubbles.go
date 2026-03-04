// Package chat provides UI components for rendering chat messages and managing
// the chat interface in the OpenCode TUI.
//
// This subpackage contains bubble components for distinct user and assistant messages.
package chat

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// BubbleOption is a functional option for configuring bubble styles.
type BubbleOption func(*bubbleConfig)

type bubbleConfig struct {
	width    int
	focused  bool
	maxWidth int
	padding  int
}

// WithBubbleWidth sets the width for the bubble.
func WithBubbleWidth(width int) BubbleOption {
	return func(c *bubbleConfig) {
		c.width = width
	}
}

// WithBubbleFocused sets the focused state for the bubble.
func WithBubbleFocused(focused bool) BubbleOption {
	return func(c *bubbleConfig) {
		c.focused = focused
	}
}

// WithBubbleMaxWidth sets the maximum width for the bubble.
func WithBubbleMaxWidth(maxWidth int) BubbleOption {
	return func(c *bubbleConfig) {
		c.maxWidth = maxWidth
	}
}

// WithBubblePadding sets the padding for the bubble.
func WithBubblePadding(padding int) BubbleOption {
	return func(c *bubbleConfig) {
		c.padding = padding
	}
}

// BubbleRenderer provides methods to render chat bubbles with different styles.
type BubbleRenderer struct{}

// NewBubbleRenderer creates a new BubbleRenderer instance.
func NewBubbleRenderer() *BubbleRenderer {
	return &BubbleRenderer{}
}

// UserBubble renders a message bubble for the user (typically on the right side).
// It uses the Primary color from the theme and the rounded border style.
func (r *BubbleRenderer) UserBubble(content string, opts ...BubbleOption) string {
	cfg := bubbleConfig{
		width:   60,
		focused: false,
		padding: 1,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	t := theme.CurrentTheme()
	borderStyle := t.GetBorderStyle()

	// Determine border based on focus state
	border := borderStyle.Normal
	borderColor := t.BorderNormal()

	if cfg.focused {
		border = borderStyle.Focused
		borderColor = t.BorderFocused()
	}

	style := lipgloss.NewStyle().
		Width(cfg.width).
		MaxWidth(cfg.maxWidth).
		Padding(cfg.padding).
		Background(t.Primary()).
		Foreground(t.Text()).
		Border(border).
		BorderForeground(borderColor).
		MarginLeft(2).
		MarginRight(0)

	return style.Render(content)
}

// AssistantBubble renders a message bubble for the AI assistant (typically on the left side).
// It uses the BackgroundSecondary color from the theme and rounded borders.
func (r *BubbleRenderer) AssistantBubble(content string, opts ...BubbleOption) string {
	cfg := bubbleConfig{
		width:   60,
		focused: false,
		padding: 1,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	t := theme.CurrentTheme()
	borderStyle := t.GetBorderStyle()

	// Determine border based on focus state
	border := borderStyle.Normal
	borderColor := t.BorderNormal()

	if cfg.focused {
		border = borderStyle.Focused
		borderColor = t.BorderFocused()
	}

	style := lipgloss.NewStyle().
		Width(cfg.width).
		MaxWidth(cfg.maxWidth).
		Padding(cfg.padding).
		Background(t.BackgroundSecondary()).
		Foreground(t.Text()).
		Border(border).
		BorderForeground(borderColor).
		MarginLeft(0).
		MarginRight(2)

	return style.Render(content)
}

// SystemBubble renders a message bubble for system messages.
// It uses the BackgroundDarker color with a dim border.
func (r *BubbleRenderer) SystemBubble(content string, opts ...BubbleOption) string {
	cfg := bubbleConfig{
		width:   60,
		focused: false,
		padding: 1,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	t := theme.CurrentTheme()
	borderStyle := t.GetBorderStyle()

	border := borderStyle.Dim
	borderColor := t.BorderDim()

	if cfg.focused {
		border = borderStyle.Focused
		borderColor = t.BorderFocused()
	}

	style := lipgloss.NewStyle().
		Width(cfg.width).
		MaxWidth(cfg.maxWidth).
		Padding(cfg.padding).
		Background(t.BackgroundDarker()).
		Foreground(t.TextMuted()).
		Border(border).
		BorderForeground(borderColor).
		MarginLeft(10).
		MarginRight(10)

	return style.Render(content)
}

// ThinkingBubble renders a thinking/loading bubble for the assistant.
// This is shown while the assistant is generating a response.
func (r *BubbleRenderer) ThinkingBubble(opts ...BubbleOption) string {
	cfg := bubbleConfig{
		width:   60,
		focused: false,
		padding: 1,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	t := theme.CurrentTheme()
	borderStyle := t.GetBorderStyle()

	style := lipgloss.NewStyle().
		Width(cfg.width).
		MaxWidth(cfg.maxWidth).
		Padding(cfg.padding).
		Background(t.BackgroundSecondary()).
		Foreground(t.TextMuted()).
		Border(borderStyle.Normal).
		BorderForeground(t.BorderDim()).
		Italic(true)

	return style.Render("Thinking...")
}

// BubbleRendererInstance is a global instance for convenient bubble rendering.
var BubbleRendererInstance = NewBubbleRenderer()
