// Package editor provides UI components for code viewing and editing in the TUI.
// This file contains the diff viewer with line-by-line coloring.
// Uses the shared diff package for parsing and common functionality.
package editor

import (
	diff "github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/diff"
	tea "github.com/charmbracelet/bubbletea"
)

// DiffViewer is a wrapper around the shared diff.Viewer for the editor component.
// It provides tea.Model interface compatibility.
type DiffViewer struct {
	*diff.Viewer
}

// NewDiffViewer creates a new diff viewer component.
func NewDiffViewer() *DiffViewer {
	return &DiffViewer{
		Viewer: diff.NewViewer(),
	}
}

// SetSize sets the dimensions of the diff viewer.
func (dv *DiffViewer) SetSize(width, height int) {
	dv.Viewer.SetSize(width, height)
}

// SetFocused sets the focused state of the diff viewer.
func (dv *DiffViewer) SetFocused(focused bool) {
	dv.Viewer.SetFocused(focused)
}

// Focused returns whether the diff viewer is focused.
func (dv *DiffViewer) Focused() bool {
	return dv.Viewer.Focused()
}

// Init implements tea.Model interface.
func (dv *DiffViewer) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model interface.
func (dv *DiffViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return dv.Viewer.Update(msg)
}

// View implements tea.Model interface.
func (dv *DiffViewer) View() string {
	return dv.Viewer.View()
}

// SetDiffContent sets the diff content to display.
func (dv *DiffViewer) SetDiffContent(content string) string {
	return dv.Viewer.Render(content)
}

// ParseUnifiedDiff parses a unified diff string into structured DiffHunks.
// This is a wrapper that delegates to the shared diff package.
func ParseUnifiedDiff(diffContent string) []diff.DiffHunk {
	return diff.ParseUnifiedDiff(diffContent)
}

// Render renders the diff viewer and returns the formatted string.
func (dv *DiffViewer) Render(diffContent string) string {
	return dv.Viewer.Render(diffContent)
}

// RenderSimple renders a diff without borders, useful for inline display.
func (dv *DiffViewer) RenderSimple(diffContent string) string {
	return dv.Viewer.RenderSimple(diffContent)
}
