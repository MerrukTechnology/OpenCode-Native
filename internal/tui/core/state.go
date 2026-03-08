package core

import (
	"sync"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
)

// Define the Page IDs here so everything can share them without cycles
type PageID string

const (
	PageChat PageID = "chat"
	PageLogs PageID = "logs"
)

// UIState represents the shared application state across all TUI components.
type UIState struct {
	// Window dimensions
	Width  int
	Height int

	// Page navigation state
	CurrentPage  PageID
	PreviousPage PageID

	// Layout engine for responsive dimensions
	LayoutEngine *layout.LayoutEngine

	// Focus state
	InputFocused bool

	// AI execution state
	IsExecuting    bool
	ExecutingTool  string
	ExecutingAgent string

	// Page tracking - maps PageID to loaded state (thread-safe)
	pageLoaded     map[PageID]bool
	pageLoadedLock sync.RWMutex
}

// NewUIState creates a new UIState with default values.
func NewUIState() *UIState {
	return &UIState{
		Width:        80,
		Height:       24,
		CurrentPage:  PageChat,
		LayoutEngine: layout.NewLayoutEngine(),
	}
}

// SetPage sets the current page and tracks navigation history.
func (s *UIState) SetPage(pageID PageID) {
	s.PreviousPage = s.CurrentPage
	s.CurrentPage = pageID
}

// GoBack returns to the previous page.
func (s *UIState) GoBack() {
	s.CurrentPage, s.PreviousPage = s.PreviousPage, s.CurrentPage
}

// IsPageLoaded checks if a page has been initialized.
func (s *UIState) IsPageLoaded(pageID PageID) bool {
	s.pageLoadedLock.RLock()
	defer s.pageLoadedLock.RUnlock()
	if s.pageLoaded == nil {
		return false
	}
	return s.pageLoaded[pageID]
}

// MarkPageLoaded marks a page as initialized.
func (s *UIState) MarkPageLoaded(pageID PageID) {
	s.pageLoadedLock.Lock()
	defer s.pageLoadedLock.Unlock()
	if s.pageLoaded == nil {
		s.pageLoaded = make(map[PageID]bool)
	}
	s.pageLoaded[pageID] = true
}

// SetDimensions updates window dimensions and recalculates layout.
func (s *UIState) SetDimensions(width, height int) {
	s.Width = width
	s.Height = height

	if s.LayoutEngine != nil {
		s.LayoutEngine.Calculate(width, height)
	}
}

// StartExecution marks the start of AI tool execution.
func (s *UIState) StartExecution(agent, tool string) {
	s.IsExecuting = true
	s.ExecutingAgent = agent
	s.ExecutingTool = tool
}

// EndExecution marks the end of AI tool execution.
func (s *UIState) EndExecution() {
	s.IsExecuting = false
	s.ExecutingTool = ""
	s.ExecutingAgent = ""
}
