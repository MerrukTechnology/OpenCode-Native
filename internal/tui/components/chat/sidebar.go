// Package chat provides UI components for rendering chat messages and managing
// the chat interface in the OpenCode TUI.
package chat

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/db"
	"github.com/MerrukTechnology/OpenCode-Native/internal/diff"
	"github.com/MerrukTechnology/OpenCode-Native/internal/history"
	"github.com/MerrukTechnology/OpenCode-Native/internal/pubsub"
	"github.com/MerrukTechnology/OpenCode-Native/internal/session"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// sidebarCmp displays the sidebar with project, session, and modified file information.
type sidebarCmp struct {
	width, height int
	session       session.Session
	sessions      session.Service
	history       history.Service
	modFiles      map[string]struct {
		additions int
		removals  int
	}
	childSessionIDs map[string]bool
	filesCh         <-chan pubsub.Event[history.File]
}

// waitForFileEvent returns a tea.Cmd that waits for the next file event from the history service.
func (m *sidebarCmp) waitForFileEvent() tea.Cmd {
	if m.filesCh == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-m.filesCh
		if !ok {
			return nil
		}
		return msg
	}
}

// Init initializes the sidebar component. It subscribes to file events from the history service and loads the initial modified files.
func (m *sidebarCmp) Init() tea.Cmd {
	if m.history != nil {
		ctx := context.Background()
		m.filesCh = m.history.SubscribeWithContext(ctx)

		m.modFiles = make(map[string]struct {
			additions int
			removals  int
		})

		m.loadModifiedFiles(ctx)

		return m.waitForFileEvent()
	}
	return nil
}

// Update handles messages for the sidebar component. It listens for session selection changes and file events to update the displayed information accordingly.
func (m *sidebarCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Also call SetSize for backward compatibility
		m.SetSize(msg.Width, msg.Height)
	case SessionSelectedMsg:
		if msg.ID != m.session.ID {
			m.session = msg
			ctx := context.Background()
			m.loadModifiedFiles(ctx)
		}
	case pubsub.Event[session.Session]:
		if msg.Type == pubsub.UpdatedEvent {
			if m.session.ID == msg.Payload.ID {
				m.session = msg.Payload
			}
		}
		if msg.Type == pubsub.CreatedEvent {
			if msg.Payload.RootSessionID == m.session.RootSessionID || msg.Payload.ParentSessionID == m.session.ID {
				if m.childSessionIDs == nil {
					m.childSessionIDs = make(map[string]bool)
					m.childSessionIDs[m.session.ID] = true
				}
				m.childSessionIDs[msg.Payload.ID] = true
			}
		}
	case pubsub.Event[history.File]:
		if m.isInSessionTree(msg.Payload.SessionID) {
			ctx := context.Background()
			m.processFileChanges(ctx, msg.Payload)
		}
		return m, m.waitForFileEvent()
	}
	return m, nil
}

// View renders the sidebar component. It displays the project information, session details, LSP configuration status, and modified files.
func (m *sidebarCmp) View() string {
	baseStyle := styles.BaseStyle()

	return baseStyle.
		Width(m.width).
		PaddingLeft(4).
		PaddingRight(2).
		Height(m.height - 1).
		MaxHeight(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				header(m.width),
				" ",
				m.projectSection(),
				" ",
				m.sessionSection(),
				" ",
				lspsConfigured(m.width),
				" ",
				m.modifiedFiles(),
			),
		)
}

// projectSection generates the project section of the sidebar. It retrieves the current project ID and formats it for display.
func (m *sidebarCmp) projectSection() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	cfg := config.Get()

	projectID := db.GetProjectID(cfg.WorkingDir)

	projectKey := baseStyle.
		Foreground(t.Primary()).
		Bold(true).
		Render("Project")

	projectValue := baseStyle.
		Foreground(t.Text()).
		Width(m.width - lipgloss.Width(projectKey)).
		Render(": " + projectID)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		projectKey,
		projectValue,
	)
}

// sessionSection generates the session section of the sidebar. It displays the session title and provider information.
// It determines the session provider type (local or remote) and formats the display accordingly.
func (m *sidebarCmp) sessionSection() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	cfg := config.Get()

	providerInfo := "local"
	if cfg.SessionProvider.Type == config.ProviderMySQL {
		if cfg.SessionProvider.MySQL.Host != "" {
			providerInfo = fmt.Sprintf("remote (%s)", cfg.SessionProvider.MySQL.Host)
		} else if cfg.SessionProvider.MySQL.DSN != "" {
			dsn := cfg.SessionProvider.MySQL.DSN
			if idx := strings.Index(dsn, "@tcp("); idx != -1 {
				hostPart := dsn[idx+5:]
				if endIdx := strings.Index(hostPart, ")"); endIdx != -1 {
					host := hostPart[:endIdx]
					if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
						host = host[:colonIdx]
					}
					providerInfo = fmt.Sprintf("remote (%s)", host)
				} else {
					providerInfo = "remote"
				}
			} else {
				providerInfo = "remote"
			}
		} else {
			providerInfo = "remote"
		}
	}

	sessionKey := baseStyle.
		Foreground(t.Primary()).
		Bold(true).
		Render("Session")

	provider := baseStyle.
		Foreground(t.TextMuted()).
		Render(fmt.Sprintf(" [%s]", providerInfo))

	sessionValue := baseStyle.
		Foreground(t.Text()).
		Render(": " + m.session.Title)

	sessionView := baseStyle.
		Width(m.width - lipgloss.Width(sessionKey)).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				sessionValue,
				provider,
			),
		)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		sessionKey,
		sessionView,
	)
}

// modifiedFile generates a string representation of a modified file with its addition and removal statistics. It formats the file path and the stats for additions and removals, applying appropriate colors and spacing.
func (m *sidebarCmp) modifiedFile(filePath string, additions, removals int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	stats := ""
	if additions > 0 && removals > 0 {
		additionsStr := baseStyle.
			Foreground(t.Success()).
			PaddingLeft(1).
			Render(fmt.Sprintf("+%d", additions))

		removalsStr := baseStyle.
			Foreground(t.Error()).
			PaddingLeft(1).
			Render(fmt.Sprintf("-%d", removals))

		content := lipgloss.JoinHorizontal(lipgloss.Left, additionsStr, removalsStr)
		stats = baseStyle.Width(lipgloss.Width(content)).Render(content)
	} else if additions > 0 {
		additionsStr := " " + baseStyle.
			PaddingLeft(1).
			Foreground(t.Success()).
			Render(fmt.Sprintf("+%d", additions))
		stats = baseStyle.Width(lipgloss.Width(additionsStr)).Render(additionsStr)
	} else if removals > 0 {
		removalsStr := " " + baseStyle.
			PaddingLeft(1).
			Foreground(t.Error()).
			Render(fmt.Sprintf("-%d", removals))
		stats = baseStyle.Width(lipgloss.Width(removalsStr)).Render(removalsStr)
	}

	filePathStr := baseStyle.Render(filePath)

	return baseStyle.
		Width(m.width).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				filePathStr,
				stats,
			),
		)
}

// modifiedFiles generates a string representation of all modified files in the sidebar. It lists each file with its addition and removal statistics, sorted alphabetically. If there are no modified files, it displays a message indicating that there are no modified files.
func (m *sidebarCmp) modifiedFiles() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	modifiedFiles := baseStyle.
		Width(m.width).
		Foreground(t.Primary()).
		Bold(true).
		Render("Modified Files:")

	if len(m.modFiles) == 0 {
		message := "No modified files"
		remainingWidth := m.width - lipgloss.Width(message)
		if remainingWidth > 0 {
			message += strings.Repeat(" ", remainingWidth)
		}
		return baseStyle.
			Width(m.width).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Top,
					modifiedFiles,
					baseStyle.Foreground(t.TextMuted()).Render(message),
				),
			)
	}

	var paths []string
	for path := range m.modFiles {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var fileViews []string
	for _, path := range paths {
		stats := m.modFiles[path]
		fileViews = append(fileViews, m.modifiedFile(path, stats.additions, stats.removals))
	}

	return baseStyle.
		Width(m.width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				modifiedFiles,
				lipgloss.JoinVertical(
					lipgloss.Left,
					fileViews...,
				),
			),
		)
}

// SetSize sets the dimensions of the sidebar.
func (m *sidebarCmp) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize returns the current dimensions of the sidebar.
func (m *sidebarCmp) GetSize() (int, int) {
	return m.width, m.height
}

// NewSidebarCmp creates a new sidebar component.
func NewSidebarCmp(s session.Session, sessions session.Service, history history.Service) tea.Model {
	return &sidebarCmp{
		session:  s,
		sessions: sessions,
		history:  history,
	}
}

// loadModifiedFiles loads the modified files for the current session. It retrieves the latest files from the session tree and compares them with their initial versions to determine the number of additions and removals for each modified file. The results are stored in the modFiles map for display in the sidebar.
func (m *sidebarCmp) loadModifiedFiles(ctx context.Context) {
	if m.history == nil || m.session.ID == "" {
		return
	}

	rootSessionID := m.session.RootSessionID
	if rootSessionID == "" {
		rootSessionID = m.session.ID
	}

	m.buildChildSessionCache(ctx, rootSessionID)

	latestFiles, err := m.history.ListLatestSessionTreeFiles(ctx, rootSessionID)
	if err != nil || len(latestFiles) == 0 {
		latestFiles, err = m.history.ListLatestSessionFiles(ctx, m.session.ID)
		if err != nil {
			return
		}
	}

	allFiles, err := m.history.ListBySessionTree(ctx, rootSessionID)
	if err != nil || len(allFiles) == 0 {
		allFiles, err = m.history.ListBySession(ctx, m.session.ID)
		if err != nil {
			return
		}
	}

	m.modFiles = make(map[string]struct {
		additions int
		removals  int
	})

	for _, file := range latestFiles {
		if file.Version == history.InitialVersion {
			continue
		}

		var initialVersion history.File
		for _, v := range allFiles {
			if v.Path == file.Path && v.Version == history.InitialVersion {
				initialVersion = v
				break
			}
		}

		if initialVersion.ID == "" {
			continue
		}
		if initialVersion.Content == file.Content {
			continue
		}

		_, additions, removals := diff.GenerateDiff(initialVersion.Content, file.Content, file.Path)

		if additions > 0 || removals > 0 {
			displayPath := getDisplayPath(file.Path)
			m.modFiles[displayPath] = struct {
				additions int
				removals  int
			}{
				additions: additions,
				removals:  removals,
			}
		}
	}
}

// buildChildSessionCache builds a cache of child session IDs for the given root session ID. This cache is used to efficiently determine if a file event belongs to the current session tree when processing file changes.
func (m *sidebarCmp) buildChildSessionCache(ctx context.Context, rootSessionID string) {
	m.childSessionIDs = make(map[string]bool)
	m.childSessionIDs[m.session.ID] = true

	if m.sessions == nil {
		return
	}

	children, err := m.sessions.ListChildren(ctx, rootSessionID)
	if err != nil {
		return
	}
	for _, child := range children {
		m.childSessionIDs[child.ID] = true
	}
}

// isInSessionTree checks if the given session ID is part of the current session tree. It uses the childSessionIDs cache to determine if the session ID belongs to the current session or any of its child sessions.
func (m *sidebarCmp) isInSessionTree(sessionID string) bool {
	if m.childSessionIDs == nil {
		return sessionID == m.session.ID
	}
	return m.childSessionIDs[sessionID]
}

// processFileChanges processes file changes and updates the modified files cache. It compares the content of the file with its initial version to determine if there are any additions or removals.
func (m *sidebarCmp) processFileChanges(ctx context.Context, file history.File) {
	if file.Version == history.InitialVersion {
		return
	}

	initialVersion, err := m.findInitialVersion(ctx, file.Path)
	if err != nil || initialVersion.ID == "" {
		return
	}

	if initialVersion.Content == file.Content {
		displayPath := getDisplayPath(file.Path)
		delete(m.modFiles, displayPath)
		return
	}

	_, additions, removals := diff.GenerateDiff(initialVersion.Content, file.Content, file.Path)

	if additions > 0 || removals > 0 {
		displayPath := getDisplayPath(file.Path)
		m.modFiles[displayPath] = struct {
			additions int
			removals  int
		}{
			additions: additions,
			removals:  removals,
		}
	} else {
		displayPath := getDisplayPath(file.Path)
		delete(m.modFiles, displayPath)
	}
}

// findInitialVersion finds the initial version of a file in the session tree. It first tries to find the initial version in the session tree, and if not found, it tries to find it in the current session.
func (m *sidebarCmp) findInitialVersion(ctx context.Context, path string) (history.File, error) {
	rootSessionID := m.session.RootSessionID
	if rootSessionID == "" {
		rootSessionID = m.session.ID
	}

	fileVersions, err := m.history.ListBySessionTree(ctx, rootSessionID)
	if err != nil || len(fileVersions) == 0 {
		fileVersions, err = m.history.ListBySession(ctx, m.session.ID)
		if err != nil {
			return history.File{}, err
		}
	}

	for _, v := range fileVersions {
		if v.Path == path && v.Version == history.InitialVersion {
			return v, nil
		}
	}

	return history.File{}, errors.New("initial version not found")
}

// getDisplayPath returns the display path for a given file path. It removes the working directory prefix from the path.
func getDisplayPath(path string) string {
	workingDir := config.WorkingDirectory()
	displayPath := strings.TrimPrefix(path, workingDir)
	return strings.TrimPrefix(displayPath, "/")
}
