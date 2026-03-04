// Package editor provides UI components for code viewing and editing in the TUI.
// This file contains the TextArea component for text input with theme support.
package editor

import (
	"os"
	"os/exec"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// TextArea is a component for text input with theme support.
type TextArea struct {
	textarea textarea.Model
	rect     layout.Rect // Uses Rect from layout package
	focused  bool
}

// TextAreaKeyMaps defines key bindings for the text area.
type TextAreaKeyMaps struct {
	// Send binds to keys that submit the text (Enter, Ctrl+S)
	Send key.Binding
	// OpenExternalEditor binds to keys that open external editor (Ctrl+E)
	OpenExternalEditor key.Binding
}

// DefaultTextAreaKeyMaps returns the default key bindings for text area.
var DefaultTextAreaKeyMaps = TextAreaKeyMaps{
	Send: key.NewBinding(
		key.WithKeys("enter", "ctrl+s"),
		key.WithHelp("enter", "submit"),
	),
	OpenExternalEditor: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "open editor"),
	),
}

// NewTextArea creates a new text area component.
func NewTextArea() *TextArea {
	ta := createTextArea(nil)
	return &TextArea{
		textarea: ta,
		rect:     layout.Rect{Width: 80, Height: 3},
		focused:  true,
	}
}

// NewTextAreaWithDimensions creates a new text area with given dimensions.
func NewTextAreaWithDimensions(width, height int) *TextArea {
	ta := createTextArea(nil)
	ta.SetWidth(width)
	ta.SetHeight(height)
	return &TextArea{
		textarea: ta,
		rect:     layout.Rect{Width: width, Height: height},
		focused:  true,
	}
}

// SetRect sets the dimensions of the text area using a Rect.
func (t *TextArea) SetRect(rect layout.Rect) {
	t.rect = rect
	t.textarea.SetWidth(rect.Width)
	t.textarea.SetHeight(rect.Height)
}

// createTextArea creates a textarea with theme styling.
func createTextArea(existing *textarea.Model) textarea.Model {
	t := theme.CurrentTheme()
	bgColor := t.Background()
	textColor := t.Text()
	textMutedColor := t.TextMuted()

	ta := textarea.New()
	ta.BlurredStyle.Base = styles.BaseStyle().Background(bgColor).Foreground(textColor)
	ta.BlurredStyle.CursorLine = styles.BaseStyle().Background(bgColor)
	ta.BlurredStyle.Placeholder = styles.BaseStyle().Background(bgColor).Foreground(textMutedColor)
	ta.BlurredStyle.Text = styles.BaseStyle().Background(bgColor).Foreground(textColor)
	ta.FocusedStyle.Base = styles.BaseStyle().Background(bgColor).Foreground(textColor)
	ta.FocusedStyle.CursorLine = styles.BaseStyle().Background(bgColor)
	ta.FocusedStyle.Placeholder = styles.BaseStyle().Background(bgColor).Foreground(textMutedColor)
	ta.FocusedStyle.Text = styles.BaseStyle().Background(bgColor).Foreground(textColor)

	ta.Prompt = " "
	ta.ShowLineNumbers = false
	ta.CharLimit = -1

	if existing != nil {
		ta.SetValue(existing.Value())
		ta.SetWidth(existing.Width())
		ta.SetHeight(existing.Height())
	}

	ta.Focus()
	return ta
}

// Init initializes the text area (starts cursor blinking).
func (t *TextArea) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the text area.
func (t *TextArea) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		// Handle key bindings
		if key.Matches(msg, DefaultTextAreaKeyMaps.OpenExternalEditor) {
			return t, t.openExternalEditor()
		}
		// Handle Enter for newline with backslash
		if t.textarea.Focused() && key.Matches(msg, DefaultTextAreaKeyMaps.Send) {
			value := t.textarea.Value()
			if len(value) > 0 && value[len(value)-1] == '\\' {
				// If the last character is a backslash, remove it and add a newline
				t.textarea.SetValue(value[:len(value)-1] + "\n")
				return t, nil
			}
		}
	}
	t.textarea, cmd = t.textarea.Update(msg)
	return t, cmd
}

// View renders the text area.
func (t *TextArea) View() string {
	return t.textarea.View()
}

// SetSize sets the dimensions of the text area.
func (t *TextArea) SetSize(width, height int) {
	t.rect.Width = width
	t.rect.Height = height
	t.textarea.SetWidth(width)
	t.textarea.SetHeight(height)
}

// SetFocused sets the focused state.
func (t *TextArea) SetFocused(focused bool) {
	t.focused = focused
	if focused {
		t.textarea.Focus()
	} else {
		t.textarea.Blur()
	}
}

// Focused returns whether the text area is focused.
func (t *TextArea) Focused() bool {
	return t.focused
}

// Value returns the current text value.
func (t *TextArea) Value() string {
	return t.textarea.Value()
}

// SetValue sets the text value.
func (t *TextArea) SetValue(value string) {
	t.textarea.SetValue(value)
}

// Reset clears the text area.
func (t *TextArea) Reset() {
	t.textarea.Reset()
}

// GetKeyBindings returns the key bindings for this component.
func (t *TextArea) GetKeyBindings() []key.Binding {
	return []key.Binding{
		DefaultTextAreaKeyMaps.Send,
		DefaultTextAreaKeyMaps.OpenExternalEditor,
	}
}

// openExternalEditor opens the system's default editor for text input.
func (t *TextArea) openExternalEditor() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	tmpfile, err := os.CreateTemp("", "editor_*.txt")
	if err != nil {
		return util.ReportError(err)
	}
	// Write current content to temp file
	if content := t.textarea.Value(); content != "" {
		tmpfile.WriteString(content) //nolint:errcheck
	}
	tmpfile.Close()

	c := exec.Command(editor, tmpfile.Name()) //nolint:gosec
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return util.ReportError(err)
		}
		content, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			return util.ReportError(err)
		}
		os.Remove(tmpfile.Name())
		t.textarea.SetValue(string(content))
		return nil
	})
}

// TextAreaWithContent creates a text area with initial content.
func TextAreaWithContent(content string) *TextArea {
	ta := NewTextArea()
	ta.SetValue(content)
	return ta
}
