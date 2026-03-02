// agent details component, shows information about the selected agent, such as name, description, model, tools, etc.
// It listens for selectedAgentMsg to update the displayed information when a new agent is selected from the list.
package agents

import (
	"fmt"
	"sort"
	"strings"

	agentregistry "github.com/MerrukTechnology/OpenCode-Native/internal/agent"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailComponent is the interface for the agent details component.
type DetailComponent interface {
	tea.Model
	layout.Sizeable
	layout.Bindings
}

// detailCmp is the implementation of DetailComponent that displays information about the selected agent.
type detailCmp struct {
	width, height int
	current       agentregistry.AgentInfo
	viewport      viewport.Model
}

// Init initializes the detail component.
func (d *detailCmp) Init() tea.Cmd {
	return nil
}

// Update updates the detail component. It listens for selectedAgentMsg to update the displayed information when a new agent is selected from the list.
func (d *detailCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case selectedAgentMsg:
		d.current = agentregistry.AgentInfo(msg)
		d.updateContent()
	}

	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

// updateContent updates the content of the viewport with the information about the selected agent. It formats the information using lipgloss styles and handles wrapping for long descriptions.
// It displays the agent's name, type, description, model, tools, permissions, and location in a structured format.
func (d *detailCmp) updateContent() {
	var content strings.Builder
	t := theme.CurrentTheme()

	labelStyle := lipgloss.NewStyle().Foreground(t.Primary()).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(t.Text())
	mutedStyle := lipgloss.NewStyle().Foreground(t.TextMuted())

	availableWidth := d.width - 4
	if availableWidth < 1 {
		availableWidth = 1
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(t.TextEmphasized()).
		Render(fmt.Sprintf("%s (%s)", d.current.Name, d.current.ID))
	content.WriteString(header)
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Type:"))
	content.WriteString(" ")
	content.WriteString(valueStyle.Render(string(d.current.Mode)))
	content.WriteString("\n")

	if d.current.Description != "" {
		content.WriteString(labelStyle.Render("Description:"))
		content.WriteString("\n")
		wrapped := lipgloss.NewStyle().Width(availableWidth).Padding(0, 2).
			Render(valueStyle.Render(d.current.Description))
		content.WriteString(wrapped)
		content.WriteString("\n")
	}

	content.WriteString(labelStyle.Render("Model:"))
	content.WriteString(" ")
	model := d.current.Model
	if model == "" {
		model = "default"
	}
	content.WriteString(valueStyle.Render(model))
	content.WriteString("\n")

	if len(d.current.Tools) > 0 {
		content.WriteString(labelStyle.Render("Tools:"))
		content.WriteString("\n")
		names := make([]string, 0, len(d.current.Tools))
		for name := range d.current.Tools {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			enabled := d.current.Tools[name]
			status := mutedStyle.Render("disabled")
			if enabled {
				status = lipgloss.NewStyle().Foreground(t.Success()).Render("enabled")
			}
			line := fmt.Sprintf("  %s: %s", valueStyle.Render(name), status)
			content.WriteString(line)
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	if len(d.current.Permission) > 0 {
		content.WriteString(labelStyle.Render("Permissions:"))
		content.WriteString("\n")
		for key, val := range d.current.Permission {
			line := fmt.Sprintf("  %s: %v", valueStyle.Render(key), val)
			content.WriteString(line)
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	if d.current.Location != "" {
		content.WriteString(labelStyle.Render("Location:"))
		content.WriteString(" ")
		content.WriteString(mutedStyle.Render(d.current.Location))
		content.WriteString("\n")
	}

	d.viewport.SetContent(content.String())
}

// View renders the detail component. It returns the content of the viewport as a string. It applies a style to replace the background with the theme's background color to ensure it blends seamlessly with the overall UI.
func (d *detailCmp) View() string {
	t := theme.CurrentTheme()
	return styles.ForceReplaceBackgroundWithLipgloss(d.viewport.View(), t.Background())
}

// GetSize returns the current width and height of the detail component.
func (d *detailCmp) GetSize() (int, int) {
	return d.width, d.height
}

// SetSize updates the width and height of the detail component and its viewport. It also triggers an update of the content to reflect the new dimensions.
func (d *detailCmp) SetSize(width int, height int) tea.Cmd {
	d.width = width
	d.height = height
	d.viewport.Width = width
	d.viewport.Height = height
	d.updateContent()
	return nil
}

// BindingKeys returns the key bindings for the detail component.
func (d *detailCmp) BindingKeys() []key.Binding {
	return layout.KeyMapToSlice(d.viewport.KeyMap)
}

// NewAgentsDetails creates a new instance of the detail component. It initializes the viewport and returns the component ready for use in the TUI.
func NewAgentsDetails() DetailComponent {
	return &detailCmp{
		viewport: viewport.New(0, 0),
	}
}
