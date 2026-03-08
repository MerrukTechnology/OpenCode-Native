// Package layout provides responsive layout calculations for the OpenCode TUI.
// This file contains overlay placement functions for rendering content on top of backgrounds.
//
// The overlay functionality is borrowed and modified from the lipgloss library
// (https://github.com/charmbracelet/lipgloss/pull/102) to support:
//   - Placing foreground content over background content
//   - Optional shadow rendering
//   - Precise x, y coordinate placement
package layout

import (
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/styles"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/util"
	"github.com/charmbracelet/lipgloss"
	chAnsi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/ansi"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/termenv"
)

// getLines splits a string into lines and calculates the widest line width.
// This is used internally for overlay positioning calculations.
func getLines(s string) (lines []string, widest int) {
	lines = strings.Split(s, "\n")

	for _, l := range lines {
		w := ansi.PrintableRuneWidth(l)
		if widest < w {
			widest = w
		}
	}

	return lines, widest
}

// PlaceOverlay places fg (foreground) on top of bg (background) at position (x, y).
// The shadow parameter enables a drop shadow effect behind the foreground.
// Additional WhitespaceOption parameters can customize the placement behavior.
//
// This function is useful for rendering dialogs, tooltips, or other overlay content
// on top of the main UI background.
func PlaceOverlay(
	x, y int,
	fg, bg string,
	shadow bool, opts ...WhitespaceOption,
) string {
	fgLines, fgWidth := getLines(fg)
	bgLines, bgWidth := getLines(bg)
	bgHeight := len(bgLines)
	fgHeight := len(fgLines)

	if shadow {
		t := theme.CurrentTheme()
		baseStyle := styles.BaseStyle()

		shadowbg := ""

		shadowchar := lipgloss.NewStyle().
			Background(t.BackgroundDarker()).
			Foreground(t.Background()).
			Render("░")
		bgchar := baseStyle.Render(" ")
		var shadowbgSb57 strings.Builder
		for i := 0; i <= fgHeight; i++ {
			if i == 0 {
				shadowbgSb57.WriteString(bgchar + strings.Repeat(bgchar, fgWidth) + "\n")
			} else {
				shadowbgSb57.WriteString(bgchar + strings.Repeat(shadowchar, fgWidth) + "\n")
			}
		}
		shadowbg += shadowbgSb57.String()

		fg = PlaceOverlay(0, 0, fg, shadowbg, false, opts...)
		fgLines, fgWidth = getLines(fg)
		fgHeight = len(fgLines)
	}

	if fgWidth >= bgWidth && fgHeight >= bgHeight {
		// Foreground is larger than background in both dimensions.
		// Return foreground as-is since it contains all the content.
		return fg
	}
	// Clamp coordinates to ensure overlay stays within background bounds.
	// To allow placement outside, add a new WhitespaceOption or remove clamping.
	x = util.Clamp(x, 0, bgWidth-fgWidth)
	y = util.Clamp(y, 0, bgHeight-fgHeight)

	ws := &whitespace{}
	for _, opt := range opts {
		opt(ws)
	}

	var b strings.Builder
	for i, bgLine := range bgLines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if i < y || i >= y+fgHeight {
			b.WriteString(bgLine)
			continue
		}

		pos := 0
		if x > 0 {
			left := truncate.String(bgLine, uint(x))
			pos = ansi.PrintableRuneWidth(left)
			b.WriteString(left)
			if pos < x {
				b.WriteString(ws.render(x - pos))
				pos = x
			}
		}

		fgLine := fgLines[i-y]
		b.WriteString(fgLine)
		pos += ansi.PrintableRuneWidth(fgLine)

		right := cutLeft(bgLine, pos)
		bgWidth := ansi.PrintableRuneWidth(bgLine)
		rightWidth := ansi.PrintableRuneWidth(right)
		if rightWidth <= bgWidth-pos {
			b.WriteString(ws.render(bgWidth - rightWidth - pos))
		}

		b.WriteString(right)
	}

	return b.String()
}

// cutLeft cuts printable characters from the left.
// This function is heavily based on muesli's ansi and truncate packages.
func cutLeft(s string, cutWidth int) string {
	return chAnsi.Cut(s, cutWidth, lipgloss.Width(s))
}

// whitespace is a style for whitespace. It allows you to specify a string of characters to use for filling in gaps,
type whitespace struct {
	style termenv.Style
	chars string
}

// Render whitespaces.
func (w whitespace) render(width int) string {
	if w.chars == "" {
		w.chars = " "
	}

	r := []rune(w.chars)
	j := 0
	b := strings.Builder{}

	// Cycle through runes and print them into the whitespace.
	for i := 0; i < width; {
		b.WriteRune(r[j])
		j++
		if j >= len(r) {
			j = 0
		}
		i += ansi.PrintableRuneWidth(string(r[j]))
	}

	// Fill any extra gaps white spaces. This might be necessary if any runes
	// are more than one cell wide, which could leave a one-rune gap.
	short := width - ansi.PrintableRuneWidth(b.String())
	if short > 0 {
		b.WriteString(strings.Repeat(" ", short))
	}

	return w.style.Styled(b.String())
}

// WhitespaceOption sets a styling rule for rendering whitespace.
type WhitespaceOption func(*whitespace)
