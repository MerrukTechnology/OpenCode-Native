// Package chat provides UI components for rendering chat messages and managing
// the chat interface in the OpenCode TUI.
package chat

import (
	"context"
	"fmt"

	"github.com/MerrukTechnology/OpenCode-Native/internal/app"
	"github.com/MerrukTechnology/OpenCode-Native/internal/message"
	"github.com/MerrukTechnology/OpenCode-Native/internal/pubsub"
	"github.com/MerrukTechnology/OpenCode-Native/internal/session"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/dialog"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// cacheItem is a cached rendering of chat messages at a specific width.
type cacheItem struct {
	width   int
	content []uiMessage
}

// messagesCmp is the main component for displaying chat messages.
type messagesCmp struct {
	app           *app.App
	width, height int
	viewport      viewport.Model
	session       session.Session
	messages      []message.Message
	uiMessages    []uiMessage
	currentMsgID  string
	cachedContent map[string]cacheItem
	spinner       spinner.Model
	rendering     bool
	attachments   viewport.Model
}
type renderFinishedMsg struct{}

// MessageKeys defines the key bindings for the messages component.
type MessageKeys struct {
	PageDown     key.Binding
	PageUp       key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
}

// MessageKeys is the set of key bindings for the messages component.
var messageKeys = MessageKeys{
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("f/pgdn", "page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("b/pgup", "page up"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "½ page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("ctrl+d", "ctrl+d"),
		key.WithHelp("ctrl+d", "½ page down"),
	),
}

// Init initializes the messages component, starting the viewport and spinner.
func (m *messagesCmp) Init() tea.Cmd {
	return tea.Batch(m.viewport.Init(), m.spinner.Tick)
}

// Update handles messages and user input for the messages component.
// It manages session changes, message updates, and pagination through the chat history.
func (m *messagesCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case dialog.ThemeChangedMsg:
		m.rerender()
	case SessionSelectedMsg:
		if msg.ID != m.session.ID {
			cmd := m.SetSession(msg)
			cmds = append(cmds, cmd)
		}
	case SessionClearedMsg:
		m.session = session.Session{}
		m.messages = make([]message.Message, 0)
		m.currentMsgID = ""
		m.rendering = false

	case tea.KeyMsg:
		if key.Matches(msg, messageKeys.PageUp) || key.Matches(msg, messageKeys.PageDown) ||
			key.Matches(msg, messageKeys.HalfPageUp) || key.Matches(msg, messageKeys.HalfPageDown) {
			u, cmd := m.viewport.Update(msg)
			m.viewport = u
			cmds = append(cmds, cmd)
		}

	case renderFinishedMsg:
		m.rendering = false
		m.viewport.GotoBottom()
	case pubsub.Event[session.Session]:
		if msg.Type == pubsub.UpdatedEvent && msg.Payload.ID == m.session.ID {
			m.session = msg.Payload
			if m.session.SummaryMessageID == m.currentMsgID {
				delete(m.cachedContent, m.currentMsgID)
				m.renderView()
			}
		}
	case pubsub.Event[message.Message]:
		needsRerender := false
		if msg.Type == pubsub.CreatedEvent {
			if msg.Payload.SessionID == m.session.ID {
				messageExists := false
				for _, v := range m.messages {
					if v.ID == msg.Payload.ID {
						messageExists = true
						break
					}
				}

				if !messageExists {
					if len(m.messages) > 0 {
						lastMsgID := m.messages[len(m.messages)-1].ID
						delete(m.cachedContent, lastMsgID)
					}

					m.messages = append(m.messages, msg.Payload)
					delete(m.cachedContent, m.currentMsgID)
					m.currentMsgID = msg.Payload.ID
					needsRerender = true
				}
			}
			// There are tool calls from the child task
			for _, v := range m.messages {
				for _, c := range v.ToolCalls() {
					if c.ID == msg.Payload.SessionID {
						delete(m.cachedContent, v.ID)
						needsRerender = true
					}
				}
			}
		} else if msg.Type == pubsub.UpdatedEvent && msg.Payload.SessionID == m.session.ID {
			for i, v := range m.messages {
				if v.ID == msg.Payload.ID {
					m.messages[i] = msg.Payload
					delete(m.cachedContent, msg.Payload.ID)
					needsRerender = true
					break
				}
			}
		}
		if needsRerender {
			m.renderView()
			if len(m.messages) > 0 {
				if (msg.Type == pubsub.CreatedEvent) ||
					(msg.Type == pubsub.UpdatedEvent && msg.Payload.ID == m.messages[len(m.messages)-1].ID) {
					m.viewport.GotoBottom()
				}
			}
		}
	}

	spinner, cmd := m.spinner.Update(msg)
	m.spinner = spinner
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// IsAgentWorking returns true if the agent is currently working on the session.
func (m *messagesCmp) IsAgentWorking() bool {
	return m.app.ActiveAgent().IsSessionBusy(m.session.ID)
}

// renderView renders the messages component view. It processes the messages, applies caching for performance, and updates the viewport content.
func (m *messagesCmp) renderView() {
	m.uiMessages = make([]uiMessage, 0)
	pos := 0
	baseStyle := styles.BaseStyle()

	if m.width == 0 {
		return
	}
	for inx, msg := range m.messages {
		switch msg.Role {
		case message.User:
			if cache, ok := m.cachedContent[msg.ID]; ok && cache.width == m.width {
				m.uiMessages = append(m.uiMessages, cache.content...)
				continue
			}
			userMsg := renderUserMessage(
				msg,
				msg.ID == m.currentMsgID,
				m.width,
				pos,
			)
			m.uiMessages = append(m.uiMessages, userMsg)
			m.cachedContent[msg.ID] = cacheItem{
				width:   m.width,
				content: []uiMessage{userMsg},
			}
			pos += userMsg.height + 1 // + 1 for spacing
		case message.Assistant:
			if cache, ok := m.cachedContent[msg.ID]; ok && cache.width == m.width {
				m.uiMessages = append(m.uiMessages, cache.content...)
				continue
			}
			isSummary := m.session.SummaryMessageID == msg.ID

			assistantMessages := renderAssistantMessage(
				msg,
				inx,
				m.messages,
				m.app.Messages,
				m.currentMsgID,
				isSummary,
				m.width,
				pos,
			)
			for _, msg := range assistantMessages {
				m.uiMessages = append(m.uiMessages, msg)
				pos += msg.height + 1 // + 1 for spacing
			}
			m.cachedContent[msg.ID] = cacheItem{
				width:   m.width,
				content: assistantMessages,
			}
		}
	}

	messages := make([]string, 0)
	for _, v := range m.uiMessages {
		messages = append(messages, lipgloss.JoinVertical(lipgloss.Left, v.content),
			baseStyle.
				Width(m.width).
				Render(
					"",
				),
		)
	}

	m.viewport.SetContent(
		baseStyle.
			Width(m.width).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Top,
					messages...,
				),
			),
	)
}

// View renders the messages component. It displays the initial screen if there are no messages, otherwise it shows the messages in the viewport.
func (m *messagesCmp) View() string {
	baseStyle := styles.BaseStyle()

	if m.rendering {
		return baseStyle.
			Width(m.width).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Top,
					"Loading...",
					m.working(),
					m.help(),
				),
			)
	}
	if len(m.messages) == 0 {
		content := baseStyle.
			Width(m.width).
			Height(m.height - 1).
			Render(
				m.initialScreen(),
			)

		return baseStyle.
			Width(m.width).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Top,
					content,
					"",
					m.help(),
				),
			)
	}

	return baseStyle.
		Width(m.width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				m.viewport.View(),
				m.working(),
				m.help(),
			),
		)
}

// hasToolsWithoutResponse checks if there are any tool calls that have not been responded to.
func hasToolsWithoutResponse(messages []message.Message) bool {
	toolCalls := make([]message.ToolCall, 0)
	toolResults := make([]message.ToolResult, 0)
	for _, m := range messages {
		toolCalls = append(toolCalls, m.ToolCalls()...)
		toolResults = append(toolResults, m.ToolResults()...)
	}

	for _, v := range toolCalls {
		found := false
		for _, r := range toolResults {
			if v.ID == r.ToolCallID {
				found = true
				break
			}
		}
		if !found && v.Finished {
			return true
		}
	}
	return false
}

// hasUnfinishedToolCalls checks if there are any tool calls that have not been finished.
func hasUnfinishedToolCalls(messages []message.Message) bool {
	toolCalls := make([]message.ToolCall, 0)
	for _, m := range messages {
		toolCalls = append(toolCalls, m.ToolCalls()...)
	}
	for _, v := range toolCalls {
		if !v.Finished {
			return true
		}
	}
	return false
}

// working generates the working string.
func (m *messagesCmp) working() string {
	text := ""
	if m.IsAgentWorking() && len(m.messages) > 0 {
		t := theme.CurrentTheme()
		baseStyle := styles.BaseStyle()

		task := "Thinking"
		lastMessage := m.messages[len(m.messages)-1]
		if hasToolsWithoutResponse(m.messages) {
			task = "Waiting for tool"
		} else if hasUnfinishedToolCalls(m.messages) {
			task = "Building tool call"
		} else if !lastMessage.IsFinished() {
			task = "Generating"
		}
		if task != "" {
			text += baseStyle.
				Width(m.width).
				Foreground(t.Primary()).
				Bold(true).
				Render(fmt.Sprintf("%s %s ", m.spinner.View(), task))
		}
	}
	return text
}

// help generates the help string.
func (m *messagesCmp) help() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	text := ""

	if m.app.ActiveAgent().IsBusy() {
		text += lipgloss.JoinHorizontal(
			lipgloss.Left,
			baseStyle.Foreground(t.TextMuted()).Bold(true).Render("press "),
			baseStyle.Foreground(t.Text()).Bold(true).Render("esc"),
			baseStyle.Foreground(t.TextMuted()).Bold(true).Render(" to exit cancel"),
		)
	} else {
		text += lipgloss.JoinHorizontal(
			lipgloss.Left,
			baseStyle.Foreground(t.TextMuted()).Bold(true).Render("press "),
			baseStyle.Foreground(t.Text()).Bold(true).Render("enter"),
			baseStyle.Foreground(t.TextMuted()).Bold(true).Render(" to send,"),
			baseStyle.Foreground(t.TextMuted()).Bold(true).Render(" type"),
			baseStyle.Foreground(t.Text()).Bold(true).Render(" \\"),
			baseStyle.Foreground(t.TextMuted()).Bold(true).Render(" to add a new line,"),
			baseStyle.Foreground(t.TextMuted()).Bold(true).Render(" type"),
			baseStyle.Foreground(t.Text()).Bold(true).Render(" /"),
			baseStyle.Foreground(t.TextMuted()).Bold(true).Render(" to command"),
		)
	}
	return baseStyle.
		Width(m.width).
		Render(text)
}

// initialScreen generates the initial screen string.
func (m *messagesCmp) initialScreen() string {
	baseStyle := styles.BaseStyle()

	return baseStyle.Width(m.width).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			header(m.width),
			"",
			lspsConfigured(m.width),
		),
	)
}

// rerender clears the cached content and re-renders the view.
func (m *messagesCmp) rerender() {
	for _, msg := range m.messages {
		delete(m.cachedContent, msg.ID)
	}
	m.renderView()
}

// SetSize sets the size of the messages component and updates the viewport and attachments dimensions accordingly. It also triggers a re-render if the size has changed.
func (m *messagesCmp) SetSize(width, height int) tea.Cmd {
	if m.width == width && m.height == height {
		return nil
	}
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height - 2
	m.attachments.Width = width + 40
	m.attachments.Height = 3
	m.rerender()
	return nil
}

// GetSize returns the current width and height of the messages component.
func (m *messagesCmp) GetSize() (int, int) {
	return m.width, m.height
}

// SetSession sets the current session and updates the messages list. It also triggers a re-render if the session has changed.
func (m *messagesCmp) SetSession(session session.Session) tea.Cmd {
	if m.session.ID == session.ID {
		return nil
	}
	m.session = session
	messages, err := m.app.Messages.List(context.Background(), session.ID)
	if err != nil {
		return util.ReportError(err)
	}
	m.messages = messages
	if len(m.messages) > 0 {
		m.currentMsgID = m.messages[len(m.messages)-1].ID
	}
	delete(m.cachedContent, m.currentMsgID)
	m.rendering = true
	return func() tea.Msg {
		m.renderView()
		return renderFinishedMsg{}
	}
}

// BindingKeys returns the key bindings for the messages component.
func (m *messagesCmp) BindingKeys() []key.Binding {
	return []key.Binding{
		m.viewport.KeyMap.PageDown,
		m.viewport.KeyMap.PageUp,
		m.viewport.KeyMap.HalfPageUp,
		m.viewport.KeyMap.HalfPageDown,
	}
}

// NewMessagesCmp creates a new instance of the messages component. It initializes the spinner and viewport, sets up key bindings for pagination, and returns the component ready for use in the TUI.
func NewMessagesCmp(app *app.App) tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Points
	vp := viewport.New(0, 0)
	attachmets := viewport.New(0, 0)
	vp.KeyMap.PageUp = messageKeys.PageUp
	vp.KeyMap.PageDown = messageKeys.PageDown
	vp.KeyMap.HalfPageUp = messageKeys.HalfPageUp
	vp.KeyMap.HalfPageDown = messageKeys.HalfPageDown
	return &messagesCmp{
		app:           app,
		cachedContent: make(map[string]cacheItem),
		viewport:      vp,
		spinner:       s,
		attachments:   attachmets,
	}
}
