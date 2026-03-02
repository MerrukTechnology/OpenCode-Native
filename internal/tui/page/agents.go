package page

import (
	agentregistry "github.com/MerrukTechnology/OpenCode-Native/internal/agent"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/agents"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AgentsPage is the page ID for the agents page.
var AgentsPage PageID = "agents"

// agentsPage is the implementation of the agents page.
type agentsPage struct {
	width, height int
	table         layout.Container
	details       layout.Container
}

// Init implements tea.Model.
func (p *agentsPage) Init() tea.Cmd {
	return tea.Batch(
		p.table.Init(),
		p.details.Init(),
	)
}

// Update implements tea.Model.
func (p *agentsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		return p, p.SetSize(msg.Width, msg.Height)
	}

	tbl, cmd := p.table.Update(msg)
	cmds = append(cmds, cmd)
	p.table = tbl.(layout.Container)
	det, cmd := p.details.Update(msg)
	cmds = append(cmds, cmd)
	p.details = det.(layout.Container)

	return p, tea.Batch(cmds...)
}

// View implements tea.Model.
func (p *agentsPage) View() string {
	style := styles.BaseStyle().Width(p.width).Height(p.height)
	return style.Render(lipgloss.JoinVertical(lipgloss.Top,
		p.table.View(),
		p.details.View(),
	))
}

// BindingKeys implements Bindings.
func (p *agentsPage) BindingKeys() []key.Binding {
	return p.table.BindingKeys()
}

// GetSize implements Sizeable.
func (p *agentsPage) GetSize() (int, int) {
	return p.width, p.height
}

// SetSize implements Sizeable.
func (p *agentsPage) SetSize(width int, height int) tea.Cmd {
	p.width = width
	p.height = height
	return tea.Batch(
		p.table.SetSize(width, height/2),
		p.details.SetSize(width, height/2),
	)
}

// NewAgentsPage creates a new agents page.
func NewAgentsPage(registry agentregistry.Registry) tea.Model {
	return &agentsPage{
		table:   layout.NewContainer(agents.NewAgentsTable(registry), layout.WithBorderAll()),
		details: layout.NewContainer(agents.NewAgentsDetails(), layout.WithBorderAll()),
	}
}
