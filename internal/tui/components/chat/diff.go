// Package chat provides UI components for rendering chat messages and diffs.
// This file contains the diff viewer with the half-block indicator (▐).
package chat

import (
	"fmt"
	"strconv"
	"strings"

	diff "github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/diff"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
)

// ============================================================================
// Diff Viewer - Renders unified diffs with ▐ indicator
// ============================================================================

// DiffViewer renders unified diff format with the ▐ indicator.
// Uses the shared diff types and parsing from the diff package.
type DiffViewer struct {
	theme        theme.Theme
	maxWidth     int
	lineNumWidth int
}

// NewDiffViewer creates a new diff viewer with the given maximum width.
func NewDiffViewer(maxWidth int) *DiffViewer {
	return &DiffViewer{
		theme:        theme.CurrentTheme(),
		maxWidth:     maxWidth,
		lineNumWidth: 5, // Width for line number column
	}
}

// RenderFileHeader renders a file header with the file path.
func (dv *DiffViewer) RenderFileHeader(filePath string) string {
	if dv.theme == nil {
		dv.theme = theme.CurrentTheme()
	}

	var builder strings.Builder

	// Render file path header
	headerStyle := dv.theme.Diff().HunkHeader
	builder.WriteString(headerStyle.Render("--- a/" + filePath))
	builder.WriteString("\n")
	builder.WriteString(headerStyle.Render("+++ b/" + filePath))
	builder.WriteString("\n")

	return builder.String()
}

// SetTheme sets the theme for the diff viewer.
func (dv *DiffViewer) SetTheme(t theme.Theme) {
	dv.theme = t
}

// ParseUnifiedDiff parses a unified diff string into structured DiffHunks.
// This is a wrapper that delegates to the shared diff package.
func ParseUnifiedDiff(diffContent string) []diff.DiffHunk {
	return diff.ParseUnifiedDiff(diffContent)
}

// Render renders the diff viewer and returns the formatted string.
func (dv *DiffViewer) Render(diffContent string) string {
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
func (dv *DiffViewer) renderHunk(hunk diff.DiffHunk) string {
	var builder strings.Builder

	// Render hunk header
	hunkStyle := dv.theme.Diff().HunkHeader
	builder.WriteString(hunkStyle.Render(hunk.Header))
	builder.WriteString("\n")

	// Render each line
	oldLineNum := hunk.OldStart
	newLineNum := hunk.NewStart

	for _, line := range hunk.Lines {
		switch line.LineType {
		case diff.DiffLineAdded:
			addLineNumStyle := dv.theme.Diff().AddLineNumber
			indicatorStyle := dv.theme.Diff().AddIndicator
			lineNumStr := strconv.Itoa(newLineNum)
			builder.WriteString(addLineNumStyle.Render(fmt.Sprintf("%*s", dv.lineNumWidth, lineNumStr)))
			builder.WriteString(indicatorStyle.Render("▐"))
			builder.WriteString(dv.theme.Diff().Addition.Render(line.Content))
			builder.WriteString("\n")
			newLineNum++
		case diff.DiffLineRemoved:
			delLineNumStyle := dv.theme.Diff().DelLineNumber
			indicatorStyle := dv.theme.Diff().DelIndicator
			lineNumStr := strconv.Itoa(oldLineNum)
			builder.WriteString(delLineNumStyle.Render(fmt.Sprintf("%*s", dv.lineNumWidth, lineNumStr)))
			builder.WriteString(indicatorStyle.Render("▐"))
			builder.WriteString(dv.theme.Diff().Deletion.Render(line.Content))
			builder.WriteString("\n")
			oldLineNum++
		default:
			lineNumStyle := dv.theme.Diff().LineNumberStyle
			oldLineNumStr := strconv.Itoa(oldLineNum)
			newLineNumStr := strconv.Itoa(newLineNum)
			builder.WriteString(lineNumStyle.Render(fmt.Sprintf("%s/%s",
				fmt.Sprintf("%*s", dv.lineNumWidth/2, oldLineNumStr),
				fmt.Sprintf("%*s", dv.lineNumWidth/2, newLineNumStr))))
			builder.WriteString(" ")
			builder.WriteString(dv.theme.Diff().ContextStyle.Render(line.Content))
			builder.WriteString("\n")
			oldLineNum++
			newLineNum++
		}
	}

	return builder.String()
}
