// Package diff provides shared diff parsing and rendering utilities for the OpenCode TUI.
// This package contains common types and functions for rendering unified diffs.
package diff

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffLineType represents the type of line in a diff.
type DiffLineType string

const (
	DiffLineAdded   DiffLineType = "+"
	DiffLineRemoved DiffLineType = "-"
	DiffLineContext DiffLineType = " "
	DiffLineHunk    DiffLineType = "@@"
)

// DiffLine represents a single line in a diff with its metadata.
type DiffLine struct {
	Content    string
	LineType   DiffLineType
	OldLineNum int
	NewLineNum int
}

// DiffHunk represents a hunk (group of changes) in a diff.
type DiffHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string
	Lines    []DiffLine
}

// ParseUnifiedDiff parses a unified diff string into structured DiffHunks.
func ParseUnifiedDiff(diffContent string) []DiffHunk {
	var hunks []DiffHunk
	var currentHunk *DiffHunk

	if diffContent == "" {
		return nil
	}
	lines := strings.Split(diffContent, "\n")

	for _, line := range lines {
		// Skip empty lines at the end
		if line == "" && len(lines) > 1 {
			continue
		}

		// Check for hunk header
		if strings.HasPrefix(line, "@@") {
			// Save previous hunk if exists
			if currentHunk != nil && len(currentHunk.Lines) > 0 {
				hunks = append(hunks, *currentHunk)
			}

			// Parse hunk header: @@ -oldStart,oldCount +newStart,newCount @@
			var oldStart, oldCount, newStart, newCount int
			fmt.Sscanf(line, "@@ -%d,%d +%d,%d @@", &oldStart, &oldCount, &newStart, &newCount)

			currentHunk = &DiffHunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
				Header:   line,
				Lines:    []DiffLine{},
			}
			continue
		}

		// Skip file headers
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			continue
		}

		// Process diff lines
		if currentHunk != nil {
			switch {
			case strings.HasPrefix(line, "+"):
				diffLine := DiffLine{
					LineType:   DiffLineAdded,
					Content:    strings.TrimPrefix(line, "+"),
					NewLineNum: currentHunk.NewStart,
				}
				currentHunk.NewStart++
				currentHunk.Lines = append(currentHunk.Lines, diffLine)
			case strings.HasPrefix(line, "-"):
				diffLine := DiffLine{
					LineType:   DiffLineRemoved,
					Content:    strings.TrimPrefix(line, "-"),
					OldLineNum: currentHunk.OldStart,
				}
				currentHunk.OldStart++
				currentHunk.Lines = append(currentHunk.Lines, diffLine)
			case strings.HasPrefix(line, " "):
				diffLine := DiffLine{
					LineType:   DiffLineContext,
					Content:    strings.TrimPrefix(line, " "),
					OldLineNum: currentHunk.OldStart,
					NewLineNum: currentHunk.NewStart,
				}
				currentHunk.OldStart++
				currentHunk.NewStart++
				currentHunk.Lines = append(currentHunk.Lines, diffLine)
			default:
				diffLine := DiffLine{
					LineType: DiffLineContext,
					Content:  line,
				}
				currentHunk.Lines = append(currentHunk.Lines, diffLine)
			}
		}
	}

	// Save last hunk
	if currentHunk != nil && len(currentHunk.Lines) > 0 {
		hunks = append(hunks, *currentHunk)
	}

	return hunks
}

// Viewer is a component for viewing file diffs with line-by-line coloring.
type Viewer struct {
	width        int
	height       int
	maxWidth     int
	focused      bool
	lineNumWidth int
	theme        theme.Theme
}

// NewViewer creates a new diff viewer component.
func NewViewer() *Viewer {
	return &Viewer{
		width:        80,
		height:       24,
		maxWidth:     0,
		focused:      false,
		lineNumWidth: 5,
		theme:        theme.CurrentTheme(),
	}
}

// NewViewerWithMaxWidth creates a new diff viewer with the given maximum width.
func NewViewerWithMaxWidth(maxWidth int) *Viewer {
	return &Viewer{
		width:        80,
		height:       24,
		maxWidth:     maxWidth,
		focused:      false,
		lineNumWidth: 5,
		theme:        theme.CurrentTheme(),
	}
}

// SetSize sets the dimensions of the diff viewer.
func (dv *Viewer) SetSize(width, height int) {
	dv.width = width
	dv.height = height
}

// SetMaxWidth sets the maximum width for the diff viewer.
func (dv *Viewer) SetMaxWidth(maxWidth int) {
	dv.maxWidth = maxWidth
}

// SetFocused sets the focused state of the diff viewer.
func (dv *Viewer) SetFocused(focused bool) {
	dv.focused = focused
}

// Focused returns whether the diff viewer is focused.
func (dv *Viewer) Focused() bool {
	return dv.focused
}

// SetTheme sets the theme for the diff viewer.
func (dv *Viewer) SetTheme(t theme.Theme) {
	dv.theme = t
}

// Init implements tea.Model interface.
func (dv *Viewer) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model interface.
func (dv *Viewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dv.SetSize(msg.Width, msg.Height)
	}
	return dv, nil
}

// View implements tea.Model interface.
func (dv *Viewer) View() string {
	return dv.Render("")
}

// SetDiffContent sets the diff content to display.
func (dv *Viewer) SetDiffContent(content string) string {
	return dv.Render(content)
}

// Render renders the diff viewer and returns the formatted string.
func (dv *Viewer) Render(diffContent string) string {
	if dv.theme == nil {
		dv.theme = theme.CurrentTheme()
	}

	hunks := ParseUnifiedDiff(diffContent)
	var builder strings.Builder

	borderStyle := dv.getBorderStyle()

	for _, hunk := range hunks {
		builder.WriteString(dv.renderHunk(hunk))
	}

	// Apply border if we have content
	if builder.Len() > 0 {
		content := builder.String()
		styledContent := borderStyle.Height(dv.height).Width(dv.width).Render(content)
		return styledContent
	}

	return builder.String()
}

// RenderSimple renders a diff without borders, useful for inline display.
func (dv *Viewer) RenderSimple(diffContent string) string {
	if dv.theme == nil {
		dv.theme = theme.CurrentTheme()
	}

	hunks := ParseUnifiedDiff(diffContent)
	var builder strings.Builder

	for _, hunk := range hunks {
		builder.WriteString(dv.renderHunk(hunk))
	}

	return builder.String()
}

// renderHunk renders a single hunk with the @@ header.
func (dv *Viewer) renderHunk(hunk DiffHunk) string {
	var builder strings.Builder

	// Render hunk header
	hunkStyle := lipgloss.NewStyle().Foreground(dv.theme.DiffHunkHeader())
	builder.WriteString(hunkStyle.Render(hunk.Header))
	builder.WriteString("\n")

	// Render each line
	oldLineNum := hunk.OldStart
	newLineNum := hunk.NewStart

	addedStyle := dv.getAddedLineStyle()
	removedStyle := dv.getRemovedLineStyle()
	contextStyle := dv.getContextLineStyle()
	lineNumStyle := dv.getLineNumberStyle()
	addedLineNumStyle := dv.getAddedLineNumberStyle()
	removedLineNumStyle := dv.getRemovedLineNumberStyle()

	for _, line := range hunk.Lines {
		switch line.LineType {
		case DiffLineAdded:
			lineNumStr := fmt.Sprintf("%*s", dv.lineNumWidth, strconv.Itoa(newLineNum))
			builder.WriteString(addedLineNumStyle.Render(lineNumStr))
			builder.WriteString(addedStyle.Render("+" + line.Content))
			builder.WriteString("\n")
			newLineNum++
		case DiffLineRemoved:
			lineNumStr := fmt.Sprintf("%*s", dv.lineNumWidth, strconv.Itoa(oldLineNum))
			builder.WriteString(removedLineNumStyle.Render(lineNumStr))
			builder.WriteString(removedStyle.Render("-" + line.Content))
			builder.WriteString("\n")
			oldLineNum++
		default:
			oldLineNumStr := fmt.Sprintf("%*s", dv.lineNumWidth/2, strconv.Itoa(oldLineNum))
			newLineNumStr := fmt.Sprintf("%*s", dv.lineNumWidth/2, strconv.Itoa(newLineNum))
			builder.WriteString(lineNumStyle.Render(oldLineNumStr + "/" + newLineNumStr))
			builder.WriteString(contextStyle.Render(" " + line.Content))
			builder.WriteString("\n")
			oldLineNum++
			newLineNum++
		}
	}

	return builder.String()
}

// getBorderStyle returns the border style based on focus state.
func (dv *Viewer) getBorderStyle() lipgloss.Style {
	borderStyle := dv.theme.GetBorderStyle()

	if dv.focused {
		return lipgloss.NewStyle().
			Border(borderStyle.Focused).
			BorderForeground(dv.theme.BorderFocused())
	}

	return lipgloss.NewStyle().
		Border(borderStyle.Normal).
		BorderForeground(dv.theme.BorderNormal())
}

// getAddedLineStyle returns the style for added lines (green).
func (dv *Viewer) getAddedLineStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(dv.theme.Success()).
		Background(dv.theme.DiffAddedBg())
}

// getRemovedLineStyle returns the style for removed lines (red).
func (dv *Viewer) getRemovedLineStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(dv.theme.Error()).
		Background(dv.theme.DiffRemovedBg())
}

// getContextLineStyle returns the style for unchanged lines.
func (dv *Viewer) getContextLineStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(dv.theme.DiffContext())
}

// getLineNumberStyle returns the style for line numbers.
func (dv *Viewer) getLineNumberStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(dv.theme.DiffLineNumber())
}

// getAddedLineNumberStyle returns the style for added line numbers.
func (dv *Viewer) getAddedLineNumberStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(dv.theme.DiffLineNumber()).
		Background(dv.theme.DiffAddedLineNumberBg())
}

// getRemovedLineNumberStyle returns the style for removed line numbers.
func (dv *Viewer) getRemovedLineNumberStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(dv.theme.DiffLineNumber()).
		Background(dv.theme.DiffRemovedLineNumberBg())
}
