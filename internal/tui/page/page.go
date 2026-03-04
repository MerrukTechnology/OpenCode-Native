// Package page provides page components for the OpenCode TUI, including
// the main chat page, logs page, and page management utilities.
package page

import (
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	tea "github.com/charmbracelet/bubbletea"
)

// PageID is a unique identifier for a page
type PageID string

// PageChangeMsg is used to change the current page
type PageChangeMsg struct {
	ID PageID
}

// Page defines the interface that all pages must implement.
// It combines tea.Model with layout-specific interfaces for
// responsive sizing and keyboard bindings.
type Page interface {
	tea.Model
	layout.Sizeable
	layout.Bindings
}
