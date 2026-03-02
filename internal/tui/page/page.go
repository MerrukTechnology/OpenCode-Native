// Package page provides page components for the OpenCode TUI, including
// the main chat page, logs page, and page management utilities.
package page

// PageID is a unique identifier for a page
type PageID string

// PageChangeMsg is used to change the current page
type PageChangeMsg struct {
	ID PageID
}
