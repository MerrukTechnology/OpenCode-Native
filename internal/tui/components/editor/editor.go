// Package editor provides UI components for code viewing and editing in the TUI.
// This file contains the main Editor component for displaying code with syntax highlighting.
package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/layout"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Editor is a component for viewing code with syntax highlighting.
type Editor struct {
	rect         layout.Rect // Uses Rect from layout package
	focused      bool
	content      string
	filename     string
	language     string
	showLineNums bool
	lineNumWidth int
}

// NewEditor creates a new editor component.
func NewEditor() *Editor {
	return &Editor{
		rect:         layout.Rect{Width: 80, Height: 24},
		focused:      false,
		content:      "",
		filename:     "",
		language:     "",
		showLineNums: true,
		lineNumWidth: 4,
	}
}

// NewEditorWithContent creates a new editor component with the given content.
func NewEditorWithContent(content string) *Editor {
	editor := NewEditor()
	editor.SetContent(content)
	return editor
}

// SetContent sets the content of the editor.
func (e *Editor) SetContent(content string) {
	e.content = content
}

// GetContent returns the content of the editor.
func (e *Editor) GetContent() string {
	return e.content
}

// SetFilename sets the filename for language detection.
func (e *Editor) SetFilename(filename string) {
	e.filename = filename
}

// SetLanguage sets the language for syntax highlighting.
func (e *Editor) SetLanguage(language string) {
	e.language = language
}

// SetShowLineNumbers sets whether to show line numbers.
func (e *Editor) SetShowLineNumbers(show bool) {
	e.showLineNums = show
}

// SetRect sets the dimensions of the editor using a Rect.
func (e *Editor) SetRect(rect layout.Rect) {
	e.rect = rect

	// Adjust line number width based on content size
	lines := strings.Split(e.content, "\n")
	if len(lines) > 0 {
		maxLineNum := len(lines)
		e.lineNumWidth = len(strconv.Itoa(maxLineNum))
		if e.lineNumWidth < 4 {
			e.lineNumWidth = 4
		}
	}
}

// SetSize sets the dimensions of the editor (legacy compatibility).
func (e *Editor) SetSize(width, height int) {
	e.rect.Width = width
	e.rect.Height = height

	// Adjust line number width based on content size
	lines := strings.Split(e.content, "\n")
	if len(lines) > 0 {
		maxLineNum := len(lines)
		e.lineNumWidth = len(strconv.Itoa(maxLineNum))
		if e.lineNumWidth < 4 {
			e.lineNumWidth = 4
		}
	}
}

// SetFocused sets the focused state of the editor.
func (e *Editor) SetFocused(focused bool) {
	e.focused = focused
}

// Focused returns whether the editor is focused.
func (e *Editor) Focused() bool {
	return e.focused
}

// Init implements tea.Model interface.
func (e *Editor) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model interface.
func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.SetSize(msg.Width, msg.Height)
	}
	return e, nil
}

// View implements tea.Model interface.
func (e *Editor) View() string {
	return e.Render()
}

// Render renders the editor with syntax highlighting.
func (e *Editor) Render() string {
	t := theme.CurrentTheme()
	if t == nil {
		t = theme.CurrentTheme()
	}

	borderStyle := e.getBorderStyle(t)

	// Get highlighted content
	highlightedContent := e.getHighlightedContent(t)

	// Add line numbers if enabled
	if e.showLineNums {
		highlightedContent = e.addLineNumbers(highlightedContent, t)
	}

	// Apply border
	contentStyle := borderStyle.Height(e.rect.Height - 2).Width(e.rect.Width - 2)
	result := contentStyle.Render(highlightedContent)

	// Add outer border
	outerBorder := e.getOuterBorderStyle(t)
	outerBorder = outerBorder.Height(e.rect.Height).Width(e.rect.Width)

	return outerBorder.Render(result)
}

// RenderSimple renders the editor without borders.
func (e *Editor) RenderSimple() string {
	t := theme.CurrentTheme()
	if t == nil {
		t = theme.CurrentTheme()
	}

	// Get highlighted content
	highlightedContent := e.getHighlightedContent(t)

	// Add line numbers if enabled
	if e.showLineNums {
		highlightedContent = e.addLineNumbers(highlightedContent, t)
	}

	return highlightedContent
}

// getHighlightedContent returns the syntax-highlighted content.
func (e *Editor) getHighlightedContent(t theme.Theme) string {
	if e.content == "" {
		return ""
	}

	// Determine language
	language := e.language
	if language == "" && e.filename != "" {
		language = detectLanguageFromFilename(e.filename)
	}

	// Highlight the code
	return HighlightWithFilename(e.content, language, e.filename)
}

// addLineNumbers adds line numbers to the content.
func (e *Editor) addLineNumbers(content string, t theme.Theme) string {
	lines := strings.Split(content, "\n")
	var result []string

	lineNumStyle := e.getLineNumberStyle(t)
	textStyle := lipgloss.NewStyle().Foreground(t.Text())

	for i, line := range lines {
		lineNum := fmt.Sprintf("%*d", e.lineNumWidth, i+1)
		result = append(result, lineNumStyle.Render(lineNum)+" "+textStyle.Render(line))
	}

	return strings.Join(result, "\n")
}

// getBorderStyle returns the border style based on focus state.
func (e *Editor) getBorderStyle(t theme.Theme) lipgloss.Style {
	borderStyle := t.GetBorderStyle()

	if e.focused {
		return lipgloss.NewStyle().
			Border(borderStyle.Focused).
			BorderForeground(t.BorderFocused())
	}

	return lipgloss.NewStyle().
		Border(borderStyle.Normal).
		BorderForeground(t.BorderNormal())
}

// getOuterBorderStyle returns the outer border style.
func (e *Editor) getOuterBorderStyle(t theme.Theme) lipgloss.Style {
	borderStyle := t.GetBorderStyle()

	if e.focused {
		return lipgloss.NewStyle().
			Border(borderStyle.Focused).
			BorderForeground(t.BorderFocused())
	}

	return lipgloss.NewStyle().
		Border(borderStyle.Normal).
		BorderForeground(t.BorderNormal())
}

// getLineNumberStyle returns the style for line numbers.
func (e *Editor) getLineNumberStyle(t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.TextMuted())
}

// EditorFromFile creates an editor component from file content.
func EditorFromFile(filename string, content string) *Editor {
	editor := NewEditor()
	editor.SetFilename(filename)
	editor.SetContent(content)
	return editor
}

// EditorWithLanguage creates an editor component with explicit language.
func EditorWithLanguage(filename string, content string, language string) *Editor {
	editor := NewEditor()
	editor.SetFilename(filename)
	editor.SetLanguage(language)
	editor.SetContent(content)
	return editor
}
