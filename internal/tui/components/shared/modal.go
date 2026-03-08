// Package shared provides common UI components for the OpenCode TUI,
// including modals, spinners, and icon constants.
package shared

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalRenderer provides methods to render modal dialogs.
type ModalRenderer struct{}

// NewModalRenderer creates a new ModalRenderer instance.
func NewModalRenderer() *ModalRenderer {
	return &ModalRenderer{}
}

// ModalOption is a functional option for configuring modal styles.
type ModalOption func(*modalConfig)

type modalConfig struct {
	width     int
	height    int
	title     string
	focused   bool
	showClose bool
}

// WithModalWidth sets the width for the modal.
func WithModalWidth(width int) ModalOption {
	return func(c *modalConfig) {
		c.width = width
	}
}

// WithModalHeight sets the height for the modal.
func WithModalHeight(height int) ModalOption {
	return func(c *modalConfig) {
		c.height = height
	}
}

// WithModalTitle sets the title for the modal.
func WithModalTitle(title string) ModalOption {
	return func(c *modalConfig) {
		c.title = title
	}
}

// WithModalFocused sets the focused state for the modal.
func WithModalFocused(focused bool) ModalOption {
	return func(c *modalConfig) {
		c.focused = focused
	}
}

// WithModalCloseButton shows or hides the close button.
func WithModalCloseButton(show bool) ModalOption {
	return func(c *modalConfig) {
		c.showClose = show
	}
}

// Render renders a modal with the given content.
func (r *ModalRenderer) Render(content string, opts ...ModalOption) string {
	cfg := modalConfig{
		width:     70, // Larger default width
		height:    18, // Larger default height
		title:     "",
		focused:   false,
		showClose: true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	t := theme.CurrentTheme()

	// Determine border color based on focus state
	borderColor := t.BorderNormal()

	if cfg.focused {
		borderColor = t.BorderFocused()
	}

	// Create title style with better padding
	titleStyle := lipgloss.NewStyle().
		Foreground(t.TextEmphasized()).
		Bold(true).
		Padding(0, 1, 0, 0)

	// Create content style - use left alignment for better readability
	contentStyle := lipgloss.NewStyle().
		Width(cfg.width - 4).
		Height(cfg.height - 4).
		Foreground(t.Text())

	// Build the modal with rounded border
	modalStyle := lipgloss.NewStyle().
		Width(cfg.width).
		Height(cfg.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(t.Background()).
		Foreground(t.Text()).
		Padding(1, 2)

	var titleBar string
	if cfg.title != "" {
		titleBar = titleStyle.Render(cfg.title)
		if cfg.showClose {
			// Add close button aligned to right
			titleBar = lipgloss.JoinHorizontal(
				lipgloss.Right,
				titleBar,
				titleStyle.Render("✕"),
			)
		}
	}

	// Join title and content
	var body string
	if cfg.title != "" {
		body = lipgloss.JoinVertical(
			lipgloss.Center,
			titleBar,
			contentStyle.Render(content),
		)
	} else {
		body = contentStyle.Render(content)
	}

	return modalStyle.Render(body)
}

// ConfirmModal renders a confirmation modal with OK and Cancel buttons.
func (r *ModalRenderer) ConfirmModal(message string, opts ...ModalOption) string {
	cfg := modalConfig{
		width:   40,
		height:  8,
		focused: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	t := theme.CurrentTheme()

	// Button styles
	okStyle := lipgloss.NewStyle().
		Foreground(t.Success()).
		Bold(true).
		Padding(0, 2)

	cancelStyle := lipgloss.NewStyle().
		Foreground(t.Error()).
		Padding(0, 2)

	buttons := okStyle.Render("[ OK ]") + "  " + cancelStyle.Render("[ Cancel ]")

	content := message + "\n\n" + buttons

	return r.Render(content, opts...)
}

// InputModal renders a modal with an input field.
func (r *ModalRenderer) InputModal(message, placeholder string, opts ...ModalOption) string {
	cfg := modalConfig{
		width:   50,
		height:  10,
		focused: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	t := theme.CurrentTheme()

	inputStyle := lipgloss.NewStyle().
		Width(cfg.width - 4).
		Foreground(t.Text()).
		Background(t.BackgroundSecondary()).
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.BorderNormal())

	inputField := inputStyle.Render(placeholder)

	content := message + "\n\n" + inputField

	return r.Render(content, opts...)
}

// ModalRendererInstance is a global instance for convenient modal rendering.
var ModalRendererInstance = NewModalRenderer()

// ModalMsg is a message type for modal events.
type ModalMsg struct {
	Action  string
	Value   string
	Confirm bool
}

// ModalModel implements the tea.Model interface for modal dialogs.
type ModalModel struct {
	Title     string
	Content   string
	Width     int
	Height    int
	Focused   bool
	ShowClose bool
	Buttons   []string
	Selected  int
	result    chan<- ModalMsg
}

// NewModalModel creates a new ModalModel with the given options.
func NewModalModel(title, content string, opts ...ModalOption) ModalModel {
	cfg := modalConfig{
		width:     50,
		height:    10,
		title:     title,
		focused:   true,
		showClose: true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return ModalModel{
		Title:     cfg.title,
		Content:   content,
		Width:     cfg.width,
		Height:    cfg.height,
		Focused:   cfg.focused,
		ShowClose: cfg.showClose,
		Buttons:   []string{"OK", "Cancel"},
		Selected:  0,
	}
}

// Init implements tea.Model Init method.
func (m ModalModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model Update method.
func (m ModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "up":
			if m.Selected > 0 {
				m.Selected--
			}
		case "right", "down":
			if m.Selected < len(m.Buttons)-1 {
				m.Selected++
			}
		case "enter":
			result := ModalMsg{
				Confirm: m.Selected == 0,
				Action:  "submit",
			}
			if m.result != nil {
				m.result <- result
			}
		case "esc":
			result := ModalMsg{
				Confirm: false,
				Action:  "close",
			}
			if m.result != nil {
				m.result <- result
			}
		}
	case tea.WindowSizeMsg:
		// Center the modal based on window size
		m.Width = min(msg.Width-4, m.Width)
		m.Height = min(msg.Height-4, m.Height)
	}
	return m, nil
}

// View implements tea.Model View method.
func (m ModalModel) View() string {
	t := theme.CurrentTheme()
	borderStyle := t.GetBorderStyle()

	border := borderStyle.Normal
	borderColor := t.BorderNormal()

	if m.Focused {
		border = borderStyle.Focused
		borderColor = t.BorderFocused()
	}

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(t.TextEmphasized()).
		Bold(true).
		Padding(0, 1)

	// Content style
	contentStyle := lipgloss.NewStyle().
		Width(m.Width-2).
		Height(m.Height-6).
		Foreground(t.Text()).
		Align(lipgloss.Center, lipgloss.Center)

	// Button styles
	okStyle := lipgloss.NewStyle().
		Foreground(t.Success()).
		Bold(true).
		Padding(0, 2)

	cancelStyle := lipgloss.NewStyle().
		Foreground(t.Error()).
		Padding(0, 2)

	// Build button row
	var buttons []string
	for i, btn := range m.Buttons {
		if i == m.Selected {
			if i == 0 {
				buttons = append(buttons, okStyle.Render("[ "+btn+" ]"))
			} else {
				buttons = append(buttons, cancelStyle.Render("[ "+btn+" ]"))
			}
		} else {
			buttons = append(buttons, lipgloss.NewStyle().
				Foreground(t.TextMuted()).
				Padding(0, 2).
				Render(" "+btn+" "))
		}
	}
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)

	// Build title bar
	var titleBar string
	if m.Title != "" {
		titleBar = titleStyle.Render(m.Title)
		if m.ShowClose {
			titleBar += " " + titleStyle.Render(IconClose)
		}
	}

	// Modal style
	modalStyle := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		Border(border).
		BorderForeground(borderColor).
		Background(t.Background()).
		Foreground(t.Text())

	// Assemble the modal
	body := lipgloss.JoinVertical(
		lipgloss.Center,
		titleBar,
		contentStyle.Render(m.Content),
		"",
		buttonRow,
	)

	return modalStyle.Render(body)
}

// SetResultChannel sets the channel to receive modal results.
func (m *ModalModel) SetResultChannel(ch chan<- ModalMsg) {
	m.result = ch
}
