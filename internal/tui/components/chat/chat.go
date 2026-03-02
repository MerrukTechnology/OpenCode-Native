// Package chat provides UI components for rendering chat messages and managing
// the chat interface in the OpenCode TUI. It includes components for viewing
// messages, editing input, displaying diffs, and navigating the chat sidebar.
package chat

import (
	"fmt"
	"sort"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/lsp/install"
	"github.com/MerrukTechnology/OpenCode-Native/internal/message"
	"github.com/MerrukTechnology/OpenCode-Native/internal/session"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/MerrukTechnology/OpenCode-Native/internal/version"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// SendMsg is a message sent when the user submits a message in the chat.
type SendMsg struct {
	Text        string
	Attachments []message.Attachment
}

// SessionSelectedMsg is a message sent when a session is selected.
type SessionSelectedMsg = session.Session

// SessionClearedMsg is a message sent when the current session is cleared.
type SessionClearedMsg struct{}

// EditorFocusMsg is a message sent to focus or unfocus the editor.
type EditorFocusMsg bool

// header generates the header string for the chat page, including the logo and current working directory.
func header(width int) string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		logo(width),
		"",
		cwd(width),
	)
}

// lspsConfigured generates a string listing the configured LSP servers.
func lspsConfigured(width int) string {
	cfg := config.Get()
	servers := install.ResolveServers(cfg)

	title := "LSP"
	title = ansi.Truncate(title, width, "…")

	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	lsps := baseStyle.
		Width(width).
		Foreground(t.Primary()).
		Bold(true).
		Render(title)

	// Get LSP names and sort them for consistent ordering
	var lspNames []string
	for name := range servers {
		lspNames = append(lspNames, name)
	}
	sort.Strings(lspNames)

	var lspViews []string
	for _, name := range lspNames {
		server := servers[name]
		lspName := baseStyle.
			Foreground(t.Text()).
			Render("• " + name)

		cmd := ""
		if len(server.Command) > 0 {
			cmd = server.Command[0]
		}
		cmd = ansi.Truncate(cmd, width-lipgloss.Width(lspName)-3, "…")

		lspPath := baseStyle.
			Foreground(t.TextMuted()).
			Render(fmt.Sprintf(" (%s)", cmd))

		lspViews = append(lspViews,
			baseStyle.
				Width(width).
				Render(
					lipgloss.JoinHorizontal(
						lipgloss.Left,
						lspName,
						lspPath,
					),
				),
		)
	}

	return baseStyle.
		Width(width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lsps,
				lipgloss.JoinVertical(
					lipgloss.Left,
					lspViews...,
				),
			),
		)
}

// logo generates the logo string for the chat page.
func logo(width int) string {
	logo := fmt.Sprintf("%s %s", styles.OpenCodeIcon, "OpenCode")
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	versionText := baseStyle.
		Foreground(t.TextMuted()).
		Render(version.Version)

	return baseStyle.
		Bold(true).
		Width(width).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				logo,
				" ",
				versionText,
			),
		)
}

// cwd generates the current working directory string.
func cwd(width int) string {
	cwd := "cwd: " + config.WorkingDirectory()
	t := theme.CurrentTheme()

	return styles.BaseStyle().
		Foreground(t.TextMuted()).
		Width(width).
		Render(cwd)
}
