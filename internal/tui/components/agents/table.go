package agents

import (
	"sort"
	"strings"

	agentregistry "github.com/MerrukTechnology/OpenCode-Native/internal/agent"
	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/util"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// TableComponent defines the interface for a table component in the TUI. It includes methods for tea.Model, layout.Sizeable, and layout.Bindings.
type TableComponent interface {
	tea.Model
	layout.Sizeable
	layout.Bindings
}

// tableCmp is a concrete implementation of TableComponent. It manages the display and interaction of a table of agents.
type tableCmp struct {
	table    table.Model
	agents   []agentregistry.AgentInfo
	registry agentregistry.Registry
}

// selectedAgentMsg is a message type used to indicate that an agent has been selected. It carries the information of the selected agent and is sent when the user selects a different row in the table. The details component listens for this message to update the displayed information about the selected agent.
type selectedAgentMsg agentregistry.AgentInfo

// Init initializes the table component by loading the agents from the registry and setting the initial rows of the table. If there are agents available, it sends a selectedAgentMsg for the first agent to display its details by default.
func (t *tableCmp) Init() tea.Cmd {
	t.loadAgents()
	t.setRows()
	if len(t.agents) > 0 {
		return util.CmdHandler(selectedAgentMsg(t.agents[0]))
	}
	return nil
}

// loadAgents loads the agents from the registry and combines the primary and subagents into a single list. It retrieves the agents using the ListByMode method of the registry for both AgentModeAgent and AgentModeSubagent, and appends them to the t.agents slice.
func (t *tableCmp) loadAgents() {
	primary := t.registry.ListByMode(config.AgentModeAgent)
	sub := t.registry.ListByMode(config.AgentModeSubagent)
	t.agents = make([]agentregistry.AgentInfo, 0, len(primary)+len(sub))
	t.agents = append(t.agents, primary...)
	t.agents = append(t.agents, sub...)
}

// findAgent searches for an agent by its ID in the list of agents. It iterates through the t.agents slice and returns the agent if found, along with a boolean indicating success. If no agent with the given ID is found, it returns an empty AgentInfo and false.
func (t *tableCmp) findAgent(id string) (agentregistry.AgentInfo, bool) {
	for _, a := range t.agents {
		if a.ID == id {
			return a, true
		}
	}
	return agentregistry.AgentInfo{}, false
}

// Update handles messages for the table component. It updates the table and checks if the selected row has changed. If the selected row has changed, it sends a selectedAgentMsg for the newly selected agent. It first stores the previously selected row, then updates the table with the incoming message. After updating, it checks the currently selected row and compares it with the previous selection. If there is a change in selection, it finds the corresponding agent and sends a selectedAgentMsg to update the details view.
func (t *tableCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	prevSelectedRow := t.table.SelectedRow()
	tbl, cmd := t.table.Update(msg)
	cmds = append(cmds, cmd)
	t.table = tbl

	selectedRow := t.table.SelectedRow()
	if selectedRow != nil {
		if prevSelectedRow == nil || selectedRow[0] != prevSelectedRow[0] {
			if a, ok := t.findAgent(selectedRow[0]); ok {
				cmds = append(cmds, util.CmdHandler(selectedAgentMsg(a)))
			}
		}
	}
	return t, tea.Batch(cmds...)
}

// View renders the table component. It applies the current theme's primary color to the selected row and returns the rendered table view. It retrieves the current theme and applies it to the selected row style of the table. Then it renders the table view and forces the background color to match the theme's background using a utility function.
func (t *tableCmp) View() string {
	th := theme.CurrentTheme()
	defaultStyles := table.DefaultStyles()
	defaultStyles.Selected = defaultStyles.Selected.Foreground(th.Primary())
	t.table.SetStyles(defaultStyles)
	return styles.ForceReplaceBackgroundWithLipgloss(t.table.View(), th.Background())
}

// GetSize returns the current width and height of the table component.
func (t *tableCmp) GetSize() (int, int) {
	return t.table.Width(), t.table.Height()
}

// SetSize updates the width and height of the table component and adjusts the column widths proportionally. It sets the width and height of the table, then calculates the column widths based on predefined percentages of the total width. The first column (ID) gets 15%, the second column (Type) gets 10%, the third column (Name) gets 20%, the fourth column (Model) gets 20%, and the fifth column (Tools) gets 35%. It then updates the columns of the table with the new widths.
func (t *tableCmp) SetSize(width int, height int) tea.Cmd {
	t.table.SetWidth(width)
	t.table.SetHeight(height)
	columns := t.table.Columns()
	if len(columns) > 0 {
		colWidths := []int{
			width * 15 / 100,
			width * 10 / 100,
			width * 20 / 100,
			width * 20 / 100,
			width * 35 / 100,
		}
		for i, col := range columns {
			if i < len(colWidths) {
				col.Width = colWidths[i] - 2
			}
			columns[i] = col
		}
		t.table.SetColumns(columns)
	}
	return nil
}

// BindingKeys returns the key bindings for the table component. It converts the KeyMap of the table into a slice of key.Binding using a utility function from the layout package.
func (t *tableCmp) BindingKeys() []key.Binding {
	return layout.KeyMapToSlice(t.table.KeyMap)
}

// setRows updates the rows of the table with the current list of agents. It creates a new slice of table.Row for each agent, extracting the relevant fields (ID, Mode, Name, Model, and Tools) and formatting the Tools field using the formatTools function. It then sets the rows of the table to the new slice of rows.
func (t *tableCmp) setRows() {
	rows := make([]table.Row, 0, len(t.agents))
	for _, a := range t.agents {
		model := a.Model
		if model == "" {
			model = "default"
		}
		rows = append(rows, table.Row{
			a.ID,
			string(a.Mode),
			a.Name,
			model,
			formatTools(a.Tools),
		})
	}
	t.table.SetRows(rows)
}

// formatTools formats the tools map into a string representation. If the tools map is empty, it returns "default". If all tools are enabled, it returns "all enabled". If all tools are disabled, it returns "none". Otherwise, it returns a string listing the disabled tools.
func formatTools(tools map[string]bool) string {
	if len(tools) == 0 {
		return "default"
	}
	disabled := make([]string, 0)
	for name, enabled := range tools {
		if !enabled {
			disabled = append(disabled, name)
		}
	}
	sort.Strings(disabled)
	if len(disabled) == 0 {
		return "all enabled"
	}
	if len(disabled) == 1 && disabled[0] == "*" {
		return "none"
	}
	return "disabled: " + strings.Join(disabled, ", ")
}

// NewAgentsTable creates a new instance of the table component. It initializes the table with the specified columns and sets the initial focus on the table. It returns the initialized table component.
func NewAgentsTable(registry agentregistry.Registry) TableComponent {
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Type", Width: 8},
		{Title: "Name", Width: 15},
		{Title: "Model", Width: 15},
		{Title: "Tools", Width: 20},
	}

	tableModel := table.New(
		table.WithColumns(columns),
	)
	tableModel.KeyMap.PageUp.SetEnabled(false)
	tableModel.KeyMap.PageDown.SetEnabled(false)
	tableModel.KeyMap.HalfPageUp.SetEnabled(false)
	tableModel.KeyMap.HalfPageDown.SetEnabled(false)
	tableModel.Focus()
	return &tableCmp{
		table:    tableModel,
		registry: registry,
	}
}
