// Package input provides the Input Box component with focus state handling.
// This component wraps a bubble tea textarea in a bordered container with
// dynamic border colors based on focus state.
package input

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// Input Mode - Represents the current input mode (INSERT, VISUAL, COMMAND, etc.)
// ============================================================================

// InputMode represents the current editing mode.
type InputMode string

const (
	ModeInsert  InputMode = "INSERT"
	ModeVisual  InputMode = "VISUAL"
	ModeCommand InputMode = "COMMAND"
	ModeSearch  InputMode = "SEARCH"
	ModeNormal  InputMode = "NORMAL"
)

// String returns the string representation of the InputMode.
func (m InputMode) String() string {
	return string(m)
}

// ============================================================================
// InputBox - Main input component with bordered container
// ============================================================================

// InputBox is a Bubble Tea model that wraps a textarea in a bordered container
// with focus state handling.
type InputBox struct {
	textarea     textarea.Model
	focused      bool
	mode         InputMode
	width        int
	height       int
	showHints    bool
	mentionHints []string
}

// InputBoxKeyMaps defines key bindings for the InputBox.
type InputBoxKeyMaps struct {
	Send      key.Binding
	Multiline key.Binding
	Clear     key.Binding
	Escape    key.Binding
}

// DefaultInputBoxKeyMaps returns the default key bindings.
func DefaultInputBoxKeyMaps() InputBoxKeyMaps {
	return InputBoxKeyMaps{
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send message"),
		),
		Multiline: key.NewBinding(
			key.WithKeys("shift+enter"),
			key.WithHelp("shift+enter", "new line"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "clear input"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "blur input"),
		),
	}
}

var defaultKeyMaps = DefaultInputBoxKeyMaps()

// ============================================================================
// NewInputBox - Constructor
// ============================================================================

// NewInputBox creates a new InputBox component.
func NewInputBox() *InputBox {
	ta := textarea.New()
	ta.Placeholder = "Vibe your instructions here..."
	ta.ShowLineNumbers = false
	ta.CharLimit = -1

	// Set initial styling using the theme system
	t := theme.CurrentTheme()
	bgColor := t.Background()
	textColor := t.Text()
	textMutedColor := t.TextMuted()

	ta.BlurredStyle = textarea.Style{
		Base:        styles.BaseStyle().Background(bgColor).Foreground(textColor),
		CursorLine:  styles.BaseStyle().Background(bgColor),
		Placeholder: styles.BaseStyle().Background(bgColor).Foreground(textMutedColor),
		Text:        styles.BaseStyle().Background(bgColor).Foreground(textColor),
	}
	ta.FocusedStyle = ta.BlurredStyle

	return &InputBox{
		textarea:  ta,
		focused:   true,
		mode:      ModeInsert,
		showHints: true,
		mentionHints: []string{
			"Type @ to mention files, agents, or tools",
		},
	}
}

// ============================================================================
// Bubble Tea Model Interface Implementation
// ============================================================================

// Init initializes the InputBox component.
func (m *InputBox) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the InputBox component.
func (m *InputBox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle custom key bindings
		if m.focused {
			if key.Matches(msg, defaultKeyMaps.Clear) {
				m.textarea.Reset()
				return m, nil
			}
			if key.Matches(msg, defaultKeyMaps.Escape) {
				m.focused = false
				m.textarea.Blur()
				return m, nil
			}
		}

	case tea.FocusMsg:
		m.focused = true
		return m, m.textarea.Focus()

	case tea.BlurMsg:
		m.focused = false
	}

	// Pass messages to the textarea
	m.textarea, cmd = m.textarea.Update(msg)

	// Update focused state based on textarea
	m.focused = m.textarea.Focused()

	return m, cmd
}

// View renders the InputBox component.
func (m *InputBox) View() string {
	t := theme.CurrentTheme()

	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	// Calculate base style to measure exact frame sizes
	var borderColor lipgloss.AdaptiveColor
	if m.focused {
		borderColor = t.BorderFocused()
	} else {
		borderColor = t.BorderDim()
	}
	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor)

	hFrame := borderStyle.GetHorizontalFrameSize()
	vFrame := borderStyle.GetVerticalFrameSize()

	// Reserve 1 line for the hints bar
	hintsHeight := 1
	boxHeight := m.height - hintsHeight
	if boxHeight < 3 {
		boxHeight = 3
	}

	// Calculate exact inner dimensions for textarea
	taWidth := m.width - hFrame
	if taWidth < 1 {
		taWidth = 1
	}

	// -1 for the title ("╭─ ✍️ PROMPT...")
	taHeight := boxHeight - vFrame - 1
	if taHeight < 1 {
		taHeight = 1
	}

	m.textarea.SetWidth(taWidth)
	m.textarea.SetHeight(taHeight)

	textareaContent := m.textarea.View()

	// Pass the exact inner dimensions to renderBorder
	borderRender := m.renderBorder(t, taWidth, taHeight, textareaContent)
	hintsBar := m.renderHintsBar(t)

	return lipgloss.JoinVertical(lipgloss.Top, borderRender, hintsBar)
}

// ============================================================================
// Helper Methods
// ============================================================================

// renderBorder renders the border container with the title.
func (m *InputBox) renderBorder(t theme.Theme, innerWidth, innerHeight int, content string) string {
	var borderColor lipgloss.AdaptiveColor
	if m.focused {
		borderColor = t.BorderFocused()
	} else {
		borderColor = t.BorderDim()
	}

	title := m.buildTitle(t)

	// We apply Width/Height to the INNER space so it doesn't overflow
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(innerHeight + 1) // +1 for the title line

	// Render content with title
	return borderStyle.Render(title + "\n" + content)
}

// buildTitle builds the border title string based on the current mode.
func (m *InputBox) buildTitle(t theme.Theme) string {
	// Style the title based on focus state
	titleStyle := lipgloss.NewStyle()
	if m.focused {
		titleStyle = titleStyle.Foreground(t.Accent())
	} else {
		titleStyle = titleStyle.Foreground(t.BorderDim())
	}

	// Build the title string
	modeStr := m.mode.String()
	return titleStyle.Render("╭─ ✍️ PROMPT (" + modeStr + ") ─╮")
}

// renderHintsBar renders the input tools bar with hints.
func (m *InputBox) renderHintsBar(t theme.Theme) string {
	// Create hint text style
	hintStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Padding(0, 1)

	// Style for mention hints (accent color when active)
	mentionStyle := lipgloss.NewStyle().
		Foreground(t.Accent())

	// Build the hints string
	var hints []string
	for _, hint := range m.mentionHints {
		// Highlight the @ symbol
		if m.showHints && len(hint) > 0 {
			// Replace @ with styled version
			hint = mentionStyle.Render("@") + hint[1:]
		}
		hints = append(hints, hintStyle.Render(hint))
	}

	// Join hints with separator
	hintsBar := lipgloss.JoinHorizontal(lipgloss.Left, hints...)

	// Wrap in a container style
	containerStyle := lipgloss.NewStyle().
		Background(t.BackgroundSecondary()).
		Foreground(t.TextMuted())

	return containerStyle.Render(hintsBar)
}

// ============================================================================
// Public API
// ============================================================================

// SetSize sets the dimensions of the InputBox.
func (m *InputBox) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSize returns the current dimensions of the InputBox.
func (m *InputBox) GetSize() (int, int) {
	return m.textarea.Width(), m.textarea.Height()
}

// Value returns the current value of the input.
func (m *InputBox) Value() string {
	return m.textarea.Value()
}

// SetValue sets the value of the input.
func (m *InputBox) SetValue(value string) {
	m.textarea.SetValue(value)
}

// Focus focuses the input.
func (m *InputBox) Focus() {
	m.focused = true
}

// Blur removes focus from the input.
func (m *InputBox) Blur() {
	m.focused = false
}

// Focused returns whether the input is currently focused.
func (m *InputBox) Focused() bool {
	return m.focused
}

// SetMode sets the input mode.
func (m *InputBox) SetMode(mode InputMode) {
	m.mode = mode
}

// GetMode returns the current input mode.
func (m *InputBox) GetMode() InputMode {
	return m.mode
}

// SetShowHints controls whether to show the hints bar.
func (m *InputBox) SetShowHints(show bool) {
	m.showHints = show
}

// BindingKeys returns the key bindings for the InputBox.
func (m *InputBox) BindingKeys() []key.Binding {
	return []key.Binding{
		defaultKeyMaps.Send,
		defaultKeyMaps.Multiline,
		defaultKeyMaps.Clear,
		defaultKeyMaps.Escape,
	}
}

// Reset clears the input value.
func (m *InputBox) Reset() {
	m.textarea.Reset()
}

// TextArea returns the underlying textarea model for external access.
func (m *InputBox) TextArea() textarea.Model {
	return m.textarea
}

// ============================================================================
// Factory Functions
// ============================================================================

// New creates a new InputBox model (for Bubble Tea compatibility).
func New() tea.Model {
	return NewInputBox()
}
