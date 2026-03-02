package layout

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SplitPaneLayout is an interface for a split pane layout that divides
// the terminal into left/right panels and optionally a bottom panel.
// It implements tea.Model, Sizeable, and Bindings interfaces.
type SplitPaneLayout interface {
	tea.Model
	Sizeable
	Bindings
	// SetLeftPanel sets the left panel of the split pane.
	SetLeftPanel(panel Container) tea.Cmd
	// SetRightPanel sets the right panel of the split pane.
	SetRightPanel(panel Container) tea.Cmd
	// SetBottomPanel sets the bottom panel of the split pane.
	SetBottomPanel(panel Container) tea.Cmd

	// ClearLeftPanel removes the left panel.
	ClearLeftPanel() tea.Cmd
	// ClearRightPanel removes the right panel.
	ClearRightPanel() tea.Cmd
	// ClearBottomPanel removes the bottom panel.
	ClearBottomPanel() tea.Cmd
}

// splitPaneLayout is the internal implementation of SplitPaneLayout.
// It manages up to three panels: left, right (horizontal split),
// and bottom (vertical split).
type splitPaneLayout struct {
	width         int
	height        int
	ratio         float64
	verticalRatio float64

	rightPanel  Container
	leftPanel   Container
	bottomPanel Container
}

// SplitPaneOption is a functional option for configuring a SplitPaneLayout.
type SplitPaneOption func(*splitPaneLayout)

func (s *splitPaneLayout) Init() tea.Cmd {
	// Init initializes all child panels.
	var cmds []tea.Cmd

	if s.leftPanel != nil {
		cmds = append(cmds, s.leftPanel.Init())
	}

	if s.rightPanel != nil {
		cmds = append(cmds, s.rightPanel.Init())
	}

	if s.bottomPanel != nil {
		cmds = append(cmds, s.bottomPanel.Init())
	}

	return tea.Batch(cmds...)
}

// Update handles messages and updates child panels.
func (s *splitPaneLayout) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return s, s.SetSize(msg.Width, msg.Height)
	}

	if s.rightPanel != nil {
		u, cmd := s.rightPanel.Update(msg)
		s.rightPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if s.leftPanel != nil {
		u, cmd := s.leftPanel.Update(msg)
		s.leftPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if s.bottomPanel != nil {
		u, cmd := s.bottomPanel.Update(msg)
		s.bottomPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return s, tea.Batch(cmds...)
}

// View renders the split pane layout.
func (s *splitPaneLayout) View() string {
	var topSection string

	if s.leftPanel != nil && s.rightPanel != nil {
		leftView := s.leftPanel.View()
		rightView := s.rightPanel.View()
		topSection = lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	} else if s.leftPanel != nil {
		topSection = s.leftPanel.View()
	} else if s.rightPanel != nil {
		topSection = s.rightPanel.View()
	} else {
		topSection = ""
	}

	var finalView string

	if s.bottomPanel != nil && topSection != "" {
		bottomView := s.bottomPanel.View()
		finalView = lipgloss.JoinVertical(lipgloss.Left, topSection, bottomView)
	} else if s.bottomPanel != nil {
		finalView = s.bottomPanel.View()
	} else {
		finalView = topSection
	}

	if finalView != "" {
		t := theme.CurrentTheme()

		style := lipgloss.NewStyle().
			Width(s.width).
			Height(s.height).
			Background(t.Background())

		return style.Render(finalView)
	}

	return finalView
}

// SetSize sets the dimensions of the split pane and all child panels.
// It calculates the appropriate sizes for left/right and top/bottom panels
// based on the configured ratios.
func (s *splitPaneLayout) SetSize(width, height int) tea.Cmd {
	s.width = width
	s.height = height

	var topHeight, bottomHeight int
	if s.bottomPanel != nil {
		topHeight = int(float64(height) * s.verticalRatio)
		bottomHeight = height - topHeight
	} else {
		topHeight = height
		bottomHeight = 0
	}

	var leftWidth, rightWidth int
	if s.leftPanel != nil && s.rightPanel != nil {
		leftWidth = int(float64(width) * s.ratio)
		rightWidth = width - leftWidth
	} else if s.leftPanel != nil {
		leftWidth = width
		rightWidth = 0
	} else if s.rightPanel != nil {
		leftWidth = 0
		rightWidth = width
	}

	var cmds []tea.Cmd
	if s.leftPanel != nil {
		cmd := s.leftPanel.SetSize(leftWidth, topHeight)
		cmds = append(cmds, cmd)
	}

	if s.rightPanel != nil {
		cmd := s.rightPanel.SetSize(rightWidth, topHeight)
		cmds = append(cmds, cmd)
	}

	if s.bottomPanel != nil {
		cmd := s.bottomPanel.SetSize(width, bottomHeight)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

// GetSize returns the current dimensions of the split pane.
func (s *splitPaneLayout) GetSize() (int, int) {
	return s.width, s.height
}

// SetLeftPanel sets the left panel of the split pane.
func (s *splitPaneLayout) SetLeftPanel(panel Container) tea.Cmd {
	s.leftPanel = panel
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

// SetRightPanel sets the right panel of the split pane.
func (s *splitPaneLayout) SetRightPanel(panel Container) tea.Cmd {
	s.rightPanel = panel
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

// SetBottomPanel sets the bottom panel of the split pane.
func (s *splitPaneLayout) SetBottomPanel(panel Container) tea.Cmd {
	s.bottomPanel = panel
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

// ClearLeftPanel removes the left panel.
func (s *splitPaneLayout) ClearLeftPanel() tea.Cmd {
	s.leftPanel = nil
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

// ClearRightPanel removes the right panel.
func (s *splitPaneLayout) ClearRightPanel() tea.Cmd {
	s.rightPanel = nil
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

// ClearBottomPanel removes the bottom panel.
func (s *splitPaneLayout) ClearBottomPanel() tea.Cmd {
	s.bottomPanel = nil
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

// BindingKeys returns keyboard bindings from all child panels.
func (s *splitPaneLayout) BindingKeys() []key.Binding {
	keys := []key.Binding{}
	if s.leftPanel != nil {
		if b, ok := s.leftPanel.(Bindings); ok {
			keys = append(keys, b.BindingKeys()...)
		}
	}
	if s.rightPanel != nil {
		if b, ok := s.rightPanel.(Bindings); ok {
			keys = append(keys, b.BindingKeys()...)
		}
	}
	if s.bottomPanel != nil {
		if b, ok := s.bottomPanel.(Bindings); ok {
			keys = append(keys, b.BindingKeys()...)
		}
	}
	return keys
}

// NewSplitPane creates a new SplitPaneLayout with optional configuration.
// The default horizontal ratio is 0.7 (70% left, 30% right)
// and vertical ratio is 0.9 (90% top, 10% bottom).
func NewSplitPane(options ...SplitPaneOption) SplitPaneLayout {
	layout := &splitPaneLayout{
		ratio:         0.7,
		verticalRatio: 0.9, // Default 90% for top section, 10% for bottom
	}
	for _, option := range options {
		option(layout)
	}
	return layout
}

// WithLeftPanel sets the initial left panel for the split pane.
func WithLeftPanel(panel Container) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.leftPanel = panel
	}
}

// WithRightPanel sets the initial right panel for the split pane.
func WithRightPanel(panel Container) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.rightPanel = panel
	}
}

// WithRatio sets the horizontal split ratio.
// A ratio of 0.7 means 70% for left panel, 30% for right.
func WithRatio(ratio float64) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.ratio = ratio
	}
}

// WithBottomPanel sets the initial bottom panel for the split pane.
func WithBottomPanel(panel Container) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.bottomPanel = panel
	}
}

// WithVerticalRatio sets the vertical split ratio.
// A ratio of 0.9 means 90% for top section, 10% for bottom panel.
func WithVerticalRatio(ratio float64) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.verticalRatio = ratio
	}
}
