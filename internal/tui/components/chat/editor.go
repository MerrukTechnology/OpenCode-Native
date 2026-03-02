// Package chat provides UI components for rendering chat messages and managing
// the chat interface in the OpenCode TUI.
package chat

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"unicode"

	"github.com/MerrukTechnology/OpenCode-Native/internal/app"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"github.com/MerrukTechnology/OpenCode-Native/internal/message"
	"github.com/MerrukTechnology/OpenCode-Native/internal/session"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/dialog"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// editorCmp is the main editor component for composing chat messages.
type editorCmp struct {
	width       int
	height      int
	app         *app.App
	session     session.Session
	textarea    textarea.Model
	attachments []message.Attachment
	deleteMode  bool
}

// EditorKeyMaps defines key bindings for the message editor.
type EditorKeyMaps struct {
	// Send binds to keys that send the message (Enter, Ctrl+S)
	Send key.Binding
	// OpenEditor binds to keys that open external editor (Ctrl+E)
	OpenEditor key.Binding
}

// DeleteAttachmentKeyMaps defines key bindings for managing attachments.
type DeleteAttachmentKeyMaps struct {
	// AttachmentDeleteMode enters delete mode to remove attachments
	AttachmentDeleteMode key.Binding
	// Escape cancels delete mode
	Escape key.Binding
	// DeleteAllAttachments removes all attachments
	DeleteAllAttachments key.Binding
}

var editorMaps = EditorKeyMaps{
	Send: key.NewBinding(
		key.WithKeys("enter", "ctrl+s"),
		key.WithHelp("enter", "send message"),
	),
	OpenEditor: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "open editor"),
	),
}

var DeleteKeyMaps = DeleteAttachmentKeyMaps{
	AttachmentDeleteMode: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r+{i}", "delete attachment at index i"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel delete mode"),
	),
	DeleteAllAttachments: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("ctrl+r+r", "delete all attchments"),
	),
}

// maxAttachments is the maximum number of file attachments allowed per message.
const maxAttachments = 5

// openEditor opens the system's default editor for composing messages.
func (m *editorCmp) openEditor() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	tmpfile, err := os.CreateTemp("", "msg_*.md")
	if err != nil {
		return util.ReportError(err)
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
		if len(content) == 0 {
			return util.ReportWarn("Message is empty")
		}
		os.Remove(tmpfile.Name())
		attachments := m.attachments
		m.attachments = nil
		return SendMsg{
			Text:        string(content),
			Attachments: attachments,
		}
	})
}

// Init initializes the editor component and returns a command to start blinking the cursor.
func (m *editorCmp) Init() tea.Cmd {
	return textarea.Blink
}

// send sends the current message if the agent is not busy. It retrieves the text from the textarea, resets it, and sends a SendMsg command with the text and attachments.
func (m *editorCmp) send() tea.Cmd {
	if m.app.ActiveAgent().IsSessionBusy(m.session.ID) {
		return util.ReportWarn("Agent is working, please wait...")
	}

	value := m.textarea.Value()
	m.textarea.Reset()
	attachments := m.attachments

	m.attachments = nil
	if value == "" {
		return nil
	}
	return tea.Batch(
		util.CmdHandler(SendMsg{
			Text:        value,
			Attachments: attachments,
		}),
	)
}

// Update handles incoming messages and updates the editor component accordingly. It processes theme changes, completion selections, session changes, attachment additions, and key presses for sending messages and managing attachments.
func (m *editorCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case dialog.ThemeChangedMsg:
		m.textarea = CreateTextArea(&m.textarea)
	case dialog.CompletionSelectedMsg:
		existingValue := m.textarea.Value()
		modifiedValue := strings.Replace(existingValue, msg.SearchString, msg.CompletionValue, 1)
		m.textarea.SetValue(modifiedValue)
		return m, nil
	case dialog.CompletionRemoveTextMsg:
		existingValue := m.textarea.Value()
		modifiedValue := strings.Replace(existingValue, msg.SearchString, "", 1)
		m.textarea.SetValue(modifiedValue)
		return m, nil
	case SessionClearedMsg:
		m.session = session.Session{}
		return m, nil
	case SessionSelectedMsg:
		if msg.ID != m.session.ID {
			m.session = msg
		}
		return m, nil
	case dialog.AttachmentAddedMsg:
		if len(m.attachments) >= maxAttachments {
			logging.ErrorPersist(fmt.Sprintf("cannot add more than %d images", maxAttachments))
			return m, cmd
		}
		m.attachments = append(m.attachments, msg.Attachment)
	case tea.KeyMsg:
		if key.Matches(msg, DeleteKeyMaps.AttachmentDeleteMode) {
			m.deleteMode = true
			return m, nil
		}
		if key.Matches(msg, DeleteKeyMaps.DeleteAllAttachments) && m.deleteMode {
			m.deleteMode = false
			m.attachments = nil
			return m, nil
		}
		if m.deleteMode && len(msg.Runes) > 0 && unicode.IsDigit(msg.Runes[0]) {
			num := int(msg.Runes[0] - '0')
			m.deleteMode = false
			if num < 10 && len(m.attachments) > num {
				if num == 0 {
					m.attachments = m.attachments[num+1:]
				} else {
					m.attachments = slices.Delete(m.attachments, num, num+1)
				}
				return m, nil
			}
		}
		if key.Matches(msg, messageKeys.PageUp) || key.Matches(msg, messageKeys.PageDown) ||
			key.Matches(msg, messageKeys.HalfPageUp) || key.Matches(msg, messageKeys.HalfPageDown) {
			return m, nil
		}
		if key.Matches(msg, editorMaps.OpenEditor) {
			if m.app.ActiveAgent().IsSessionBusy(m.session.ID) {
				return m, util.ReportWarn("Agent is working, please wait...")
			}
			return m, m.openEditor()
		}
		if key.Matches(msg, DeleteKeyMaps.Escape) {
			m.deleteMode = false
			return m, nil
		}
		// Hanlde Enter key
		if m.textarea.Focused() && key.Matches(msg, editorMaps.Send) {
			value := m.textarea.Value()
			if len(value) > 0 && value[len(value)-1] == '\\' {
				// If the last character is a backslash, remove it and add a newline
				m.textarea.SetValue(value[:len(value)-1] + "\n")
				return m, nil
			} else {
				// Otherwise, send the message
				return m, m.send()
			}
		}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the editor component, displaying the text area and any attachments. It styles the prompt and attachments using the current theme and lipgloss for layout.
func (m *editorCmp) View() string {
	t := theme.CurrentTheme()

	// Style the prompt with theme colors
	style := lipgloss.NewStyle().
		Padding(0, 0, 0, 1).
		Bold(true).
		Foreground(t.Primary())

	if len(m.attachments) == 0 {
		return lipgloss.JoinHorizontal(lipgloss.Top, style.Render(">"), m.textarea.View())
	}
	m.textarea.SetHeight(m.height - 1)
	return lipgloss.JoinVertical(lipgloss.Top,
		m.attachmentsContent(),
		lipgloss.JoinHorizontal(lipgloss.Top, style.Render(">"),
			m.textarea.View()),
	)
}

// SetSize sets the dimensions of the editor component.
// It adjusts the width and height of the text area to fit within the given dimensions.
func (m *editorCmp) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	m.textarea.SetWidth(width - 3) // account for the prompt and padding right
	m.textarea.SetHeight(height)
	m.textarea.SetWidth(width)
	return nil
}

// GetSize returns the current dimensions of the editor component.
// It returns the width and height of the text area.
func (m *editorCmp) GetSize() (int, int) {
	return m.textarea.Width(), m.textarea.Height()
}

// attachmentsContent returns the styled display of attached files.
// It formats each attachment's filename and applies styling based on the current theme.
func (m *editorCmp) attachmentsContent() string {
	var styledAttachments []string
	t := theme.CurrentTheme()
	attachmentStyles := styles.BaseStyle().
		MarginLeft(1).
		Background(t.TextMuted()).
		Foreground(t.Text())
	for i, attachment := range m.attachments {
		var filename string
		if len(attachment.FileName) > 10 {
			filename = fmt.Sprintf(" %s %s...", styles.DocumentIcon, attachment.FileName[0:7])
		} else {
			filename = fmt.Sprintf(" %s %s", styles.DocumentIcon, attachment.FileName)
		}
		if m.deleteMode {
			filename = fmt.Sprintf("%d%s", i, filename)
		}
		styledAttachments = append(styledAttachments, attachmentStyles.Render(filename))
	}
	content := lipgloss.JoinHorizontal(lipgloss.Left, styledAttachments...)
	return content
}

// BindingKeys returns the key bindings for the editor component.
func (m *editorCmp) BindingKeys() []key.Binding {
	bindings := []key.Binding{}
	bindings = append(bindings, layout.KeyMapToSlice(editorMaps)...)
	bindings = append(bindings, layout.KeyMapToSlice(DeleteKeyMaps)...)
	return bindings
}

// CreateTextArea creates a new text area component with the given theme and existing text area.
// It sets the text area's properties such as prompt, line numbers, and character limit.
// It also applies the theme's styles to the text area.
func CreateTextArea(existing *textarea.Model) textarea.Model {
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

// NewEditorCmp creates a new instance of the editor component.
func NewEditorCmp(app *app.App) tea.Model {
	ta := CreateTextArea(nil)
	return &editorCmp{
		app:      app,
		textarea: ta,
	}
}
