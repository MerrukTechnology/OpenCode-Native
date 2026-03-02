// page chat.go: This file defines the ChatPage, which is the main interface for users to interact with chat sessions,
// send messages, and view message history. It manages the layout of the chat interface, handles user input,
// and coordinates with the app's session and agent management to facilitate conversations.
package page

import (
	"context"
	"sort"
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/app"
	"github.com/MerrukTechnology/OpenCode-Native/internal/completions"
	"github.com/MerrukTechnology/OpenCode-Native/internal/message"
	"github.com/MerrukTechnology/OpenCode-Native/internal/session"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/chat"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/dialog"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/util"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChatPage is the main page for interacting with chat sessions, sending messages, and viewing message history.
var ChatPage PageID = "chat"

// chatPage is the main chat page implementation.
// It provides the chat interface with message display, input editor,
// and optional sidebar for session management.
type chatPage struct {
	app                         *app.App
	editor                      layout.Container
	messages                    layout.Container
	layout                      layout.SplitPaneLayout
	session                     session.Session
	completionDialog            dialog.CompletionDialog
	showCompletionDialog        bool
	commandCompletionDialog     dialog.CompletionDialog
	showCommandCompletionDialog bool
	commands                    []dialog.Command
}

// ChatKeyMap defines keyboard shortcuts for the chat page.
type ChatKeyMap struct {
	ShowCompletionDialog        key.Binding // Show completion dialog (@)
	ShowCommandCompletionDialog key.Binding
	NewSession                  key.Binding // Create new session (ctrl+n)
	Cancel                      key.Binding // Cancel current operation (esc)
}

// keyMap contains the default key bindings for the chat page.
var keyMap = ChatKeyMap{
	ShowCompletionDialog: key.NewBinding(
		key.WithKeys("@"),
		key.WithHelp("@", "Complete"),
	),
	ShowCommandCompletionDialog: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "Commands"),
	),
	NewSession: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "new session"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

// Init initializes the chat page and its components.
// It sets up the layout, initializes the completion dialog, and loads the session if one is selected.
func (p *chatPage) Init() tea.Cmd {
	cmds := []tea.Cmd{
		p.layout.Init(),
		p.completionDialog.Init(),
		p.commandCompletionDialog.Init(),
	}
	if p.session.ID != "" {
		cmds = append(cmds, p.setSidebar())
		cmds = append(cmds, util.CmdHandler(chat.SessionSelectedMsg(p.session)))
	}
	return tea.Batch(cmds...)
}

// findCommand searches for a command by its ID.
func (p *chatPage) findCommand(id string) (dialog.Command, bool) {
	for _, cmd := range p.commands {
		if cmd.ID == id {
			return cmd, true
		}
	}
	return dialog.Command{}, false
}

// Update handles messages for the chat page.
// It processes various messages such as window size changes, completion dialog actions, and command execution.
func (p *chatPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd := p.layout.SetSize(msg.Width, msg.Height)
		cmds = append(cmds, cmd)
	case dialog.CompletionDialogCloseMsg:
		if msg.ProviderID == completions.CommandCompletionProviderID {
			p.showCommandCompletionDialog = false
		} else {
			p.showCompletionDialog = false
		}
	case dialog.CompletionSelectedMsg:
		if msg.ProviderID == completions.CommandCompletionProviderID {
			p.showCommandCompletionDialog = false
			// Remove the /query text from the editor
			cmds = append(cmds, util.CmdHandler(dialog.CompletionRemoveTextMsg{
				SearchString: msg.SearchString,
			}))
			// Execute the selected command
			if cmd, ok := p.findCommand(msg.CompletionValue); ok {
				cmds = append(cmds, util.CmdHandler(dialog.CommandSelectedMsg{Command: cmd}))
			}
			return p, tea.Batch(cmds...)
		}
	case chat.SendMsg:
		cmd := p.sendMessage(msg.Text, msg.Attachments)
		if cmd != nil {
			return p, cmd
		}
	case dialog.CommandRunCustomMsg:
		if p.app.ActiveAgent().IsBusy() {
			return p, util.ReportWarn("Agent is busy, please wait before executing a command...")
		}

		content := msg.Content
		if msg.Args != nil {
			// Sort keys by length (longest first) to avoid partial replacements
			keys := make([]string, 0, len(msg.Args))
			for name := range msg.Args {
				keys = append(keys, name)
			}
			sort.Slice(keys, func(i, j int) bool {
				return len(keys[i]) > len(keys[j])
			})
			for _, name := range keys {
				value := msg.Args[name]
				placeholder := "$" + name
				content = strings.ReplaceAll(content, placeholder, value)
			}
		}

		cmd := p.sendMessage(content, nil)
		if cmd != nil {
			return p, cmd
		}
	case chat.SessionClearedMsg:
		p.session = session.Session{}
		cmds = append(cmds, p.clearSidebar())
	case chat.SessionSelectedMsg:
		if p.session.ID == "" {
			cmd := p.setSidebar()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		p.session = msg
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyMap.ShowCompletionDialog):
			p.showCompletionDialog = true
		case key.Matches(msg, keyMap.ShowCommandCompletionDialog):
			p.showCommandCompletionDialog = true
		case key.Matches(msg, keyMap.NewSession):
			p.session = session.Session{}
			return p, tea.Batch(
				p.clearSidebar(),
				util.CmdHandler(chat.SessionClearedMsg{}),
			)
		case key.Matches(msg, keyMap.Cancel):
			if p.session.ID != "" {
				p.app.ActiveAgent().Cancel(p.session.ID)
				return p, nil
			}
		}
	}

	// Route to command completion dialog if active
	if p.showCommandCompletionDialog {
		context, contextCmd := p.commandCompletionDialog.Update(msg)
		p.commandCompletionDialog = context.(dialog.CompletionDialog)
		cmds = append(cmds, contextCmd)

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" || keyMsg.String() == "tab" || keyMsg.String() == "shift+tab" {
				return p, tea.Batch(cmds...)
			}
		}
	}

	// Route to file completion dialog if active
	if p.showCompletionDialog {
		context, contextCmd := p.completionDialog.Update(msg)
		p.completionDialog = context.(dialog.CompletionDialog)
		cmds = append(cmds, contextCmd)

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" || keyMsg.String() == "tab" || keyMsg.String() == "shift+tab" {
				return p, tea.Batch(cmds...)
			}
		}
	}

	u, cmd := p.layout.Update(msg)
	cmds = append(cmds, cmd)
	p.layout = u.(layout.SplitPaneLayout)

	return p, tea.Batch(cmds...)
}

// setSidebar creates and sets the sidebar panel for session management.
// It creates a new sidebar component with the current session and app sessions,
// then sets it as the right panel of the layout.
func (p *chatPage) setSidebar() tea.Cmd {
	sidebarContainer := layout.NewContainer(
		chat.NewSidebarCmp(p.session, p.app.Sessions, p.app.History),
		layout.WithPadding(1, 1, 1, 1),
	)
	return tea.Batch(p.layout.SetRightPanel(sidebarContainer), sidebarContainer.Init())
}

// clearSidebar removes the sidebar panel.
func (p *chatPage) clearSidebar() tea.Cmd {
	return p.layout.ClearRightPanel()
}

// sendMessage sends a message to the active agent for processing.
// If no session is active, it creates a new session and updates the sidebar.
func (p *chatPage) sendMessage(text string, attachments []message.Attachment) tea.Cmd {
	var cmds []tea.Cmd
	if p.session.ID == "" {
		var sess session.Session
		var err error
		if p.app.InitialSessionID != "" {
			sess, err = p.app.Sessions.CreateWithID(context.Background(), p.app.InitialSessionID, "New Session")
			p.app.InitialSessionID = ""
		} else {
			sess, err = p.app.Sessions.Create(context.Background(), "New Session")
		}
		if err != nil {
			return util.ReportError(err)
		}

		p.session = sess
		cmd := p.setSidebar()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, util.CmdHandler(chat.SessionSelectedMsg(sess)))
	}

	_, err := p.app.ActiveAgent().Run(context.Background(), p.session.ID, text, attachments...)
	if err != nil {
		return util.ReportError(err)
	}
	return tea.Batch(cmds...)
}

// SetSize sets the dimensions of the chat page layout.
func (p *chatPage) SetSize(width, height int) tea.Cmd {
	return p.layout.SetSize(width, height)
}

// GetSize returns the current dimensions of the chat page.
func (p *chatPage) GetSize() (int, int) {
	return p.layout.GetSize()
}

// View renders the chat page.
func (p *chatPage) View() string {
	layoutView := p.layout.View()

	activeDialog := p.activeCompletionDialog()
	if activeDialog != nil {
		_, layoutHeight := p.layout.GetSize()
		editorWidth, editorHeight := p.editor.GetSize()

		activeDialog.SetWidth(editorWidth)
		overlay := activeDialog.View()

		layoutView = layout.PlaceOverlay(
			0,
			layoutHeight-editorHeight-lipgloss.Height(overlay),
			overlay,
			layoutView,
			false,
		)
	}

	return layoutView
}

// activeCompletionDialog returns the active completion dialog if any.
func (p *chatPage) activeCompletionDialog() dialog.CompletionDialog {
	if p.showCommandCompletionDialog {
		return p.commandCompletionDialog
	}
	if p.showCompletionDialog {
		return p.completionDialog
	}
	return nil
}

// HasActiveOverlay returns true if the chat page has an active overlay.
func (p *chatPage) HasActiveOverlay() bool {
	return p.showCompletionDialog || p.showCommandCompletionDialog
}

// BindingKeys returns keyboard bindings for the chat page.
func (p *chatPage) BindingKeys() []key.Binding {
	bindings := layout.KeyMapToSlice(keyMap)
	bindings = append(bindings, p.messages.BindingKeys()...)
	bindings = append(bindings, p.editor.BindingKeys()...)
	return bindings
}

// NewChatPage creates a new chat page with the given application.
func NewChatPage(app *app.App, commands []dialog.Command) tea.Model {
	cg := completions.NewFileAndFolderContextGroup()
	completionDialog := dialog.NewCompletionDialogCmp(cg)

	cmdProvider := completions.NewCommandCompletionProvider(commands)
	commandCompletionDialog := dialog.NewCompletionDialogCmp(cmdProvider)

	messagesContainer := layout.NewContainer(
		chat.NewMessagesCmp(app),
		layout.WithPadding(1, 1, 0, 1),
	)
	editorContainer := layout.NewContainer(
		chat.NewEditorCmp(app),
		layout.WithBorder(true, false, false, false),
	)

	var sess session.Session
	if app.InitialSession != nil {
		sess = *app.InitialSession
	}

	return &chatPage{
		app:                     app,
		editor:                  editorContainer,
		messages:                messagesContainer,
		session:                 sess,
		completionDialog:        completionDialog,
		commandCompletionDialog: commandCompletionDialog,
		commands:                commands,
		layout: layout.NewSplitPane(
			layout.WithLeftPanel(messagesContainer),
			layout.WithBottomPanel(editorContainer),
		),
	}
}
