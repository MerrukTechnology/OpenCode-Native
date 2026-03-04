// Package shared provides common UI components for the OpenCode TUI,
// including modals, spinners, and icon constants.
package shared

import (
	"sync"
	"time"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerFrame represents a single frame in the spinner animation.
type SpinnerFrame string

// Spinner frames using Unicode characters for smooth animation.
var (
	// SpinnerFrames holds the animation frames for a spinner.
	SpinnerFrames = []SpinnerFrame{
		"⠋",
		"⠙",
		"⠹",
		"⠸",
		"⠼",
		"⠴",
		"⠦",
		"⠧",
		"⠇",
		"⠏",
	}

	// DotsSpinnerFrames is an alternative dots-style spinner.
	DotsSpinnerFrames = []SpinnerFrame{
		"⠁",
		"⠂",
		"⠄",
		"⡀",
		"⠈",
		"⠐",
		"⠠",
		"⠰",
		"⠸",
		"⠧",
	}

	// LineSpinnerFrames is a simple line spinner.
	LineSpinnerFrames = []SpinnerFrame{
		"|",
		"/",
		"−",
		"\\",
	}
)

// SpinnerRenderer provides methods to render loading spinners.
type SpinnerRenderer struct {
	mu       sync.Mutex
	frame    int
	frames   []SpinnerFrame
	interval time.Duration
}

// NewSpinnerRenderer creates a new SpinnerRenderer with default settings.
func NewSpinnerRenderer() *SpinnerRenderer {
	return &SpinnerRenderer{
		frame:    0,
		frames:   SpinnerFrames,
		interval: 80 * time.Millisecond,
	}
}

// NewSpinnerRendererWithFrames creates a new SpinnerRenderer with custom frames.
func NewSpinnerRendererWithFrames(frames []SpinnerFrame, interval time.Duration) *SpinnerRenderer {
	return &SpinnerRenderer{
		frame:    0,
		frames:   frames,
		interval: interval,
	}
}

// NextFrame advances to the next frame and returns the current frame.
func (s *SpinnerRenderer) NextFrame() SpinnerFrame {
	s.mu.Lock()
	defer s.mu.Unlock()

	frame := s.frames[s.frame]
	s.frame = (s.frame + 1) % len(s.frames)

	return frame
}

type SpinnerTickMsg struct{}

// Tick is a command that waits for the spinner interval and sends a message.
// This enables chaining the animation in the Update loop.
func (s *SpinnerRenderer) Tick() tea.Cmd {
	return tea.Tick(s.interval, func(t time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

// Reset resets the spinner to the first frame.
func (s *SpinnerRenderer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.frame = 0
}

// Render returns the current spinner frame as a styled string.
func (s *SpinnerRenderer) Render() string {
	t := theme.CurrentTheme()
	frame := s.NextFrame()

	style := lipgloss.NewStyle().
		Foreground(t.Accent()).
		Bold(true)

	return style.Render(string(frame))
}

// RenderWithMessage returns the spinner with a message.
func (s *SpinnerRenderer) RenderWithMessage(message string) string {
	t := theme.CurrentTheme()
	frame := s.NextFrame()

	spinnerStyle := lipgloss.NewStyle().
		Foreground(t.Accent()).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted())

	return spinnerStyle.Render(string(frame)) + " " + messageStyle.Render(message)
}

// Interval returns the interval between frames.
func (s *SpinnerRenderer) Interval() time.Duration {
	return s.interval
}

// SpinnerRendererInstance is a global instance for convenient spinner rendering.
var SpinnerRendererInstance = NewSpinnerRenderer()

// SimpleSpinner renders a simple spinner without state management.
// Useful for one-off rendering in UI components.
func SimpleSpinner() string {
	t := theme.CurrentTheme()
	spinner := NewSpinnerRenderer()

	style := lipgloss.NewStyle().
		Foreground(t.Accent())

	return style.Render(string(spinner.NextFrame()))
}

// SimpleSpinnerWithMessage renders a simple spinner with a message.
func SimpleSpinnerWithMessage(message string) string {
	t := theme.CurrentTheme()
	spinner := NewSpinnerRenderer()

	spinnerStyle := lipgloss.NewStyle().
		Foreground(t.Accent())

	messageStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted())

	return spinnerStyle.Render(string(spinner.NextFrame())) + " " + messageStyle.Render(message)
}

// SpinnerModel implements the tea.Model interface for animated spinners.
type SpinnerModel struct {
	message  string
	frames   []SpinnerFrame
	frame    int
	interval time.Duration
	running  bool
}

// NewSpinnerModel creates a new SpinnerModel with the given message.
func NewSpinnerModel(message string) SpinnerModel {
	return SpinnerModel{
		message:  message,
		frames:   SpinnerFrames,
		frame:    0,
		interval: 80 * time.Millisecond,
		running:  true,
	}
}

// NewSpinnerModelWithFrames creates a new SpinnerModel with custom frames.
func NewSpinnerModelWithFrames(message string, frames []SpinnerFrame, interval time.Duration) SpinnerModel {
	return SpinnerModel{
		message:  message,
		frames:   frames,
		frame:    0,
		interval: interval,
		running:  true,
	}
}

// Init implements tea.Model Init method.
func (m SpinnerModel) Init() tea.Cmd {
	return m.Tick()
}

// Update implements tea.Model Update method.
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.running {
		return m, nil
	}

	switch msg.(type) {
	case SpinnerTickMsg:
		m.frame = (m.frame + 1) % len(m.frames)
		return m, m.Tick()
	case tea.WindowSizeMsg:
		// Spinner doesn't need to handle resize
		return m, nil
	}
	return m, nil
}

// View implements tea.Model View method.
func (m SpinnerModel) View() string {
	t := theme.CurrentTheme()

	spinnerStyle := lipgloss.NewStyle().
		Foreground(t.Accent())

	frame := m.frames[m.frame]

	if m.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(t.TextMuted())
		return spinnerStyle.Render(string(frame)) + " " + messageStyle.Render(m.message)
	}

	return spinnerStyle.Render(string(frame))
}

// Tick returns a command that waits for the interval and sends a SpinnerTickMsg.
func (m SpinnerModel) Tick() tea.Cmd {
	return tea.Tick(m.interval, func(t time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

// Stop stops the spinner animation.
func (m *SpinnerModel) Stop() {
	m.running = false
}

// Start starts the spinner animation.
func (m *SpinnerModel) Start() {
	m.running = true
}

// SetMessage updates the spinner's message.
func (m *SpinnerModel) SetMessage(msg string) {
	m.message = msg
}
