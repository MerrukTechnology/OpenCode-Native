package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// ProTheme implements the Theme interface with a professional dark theme
// optimized for coding with orange accents.
type ProTheme struct {
	BaseTheme
}

// NewProTheme creates a new instance of the Pro theme.
func NewProTheme() *ProTheme {
	// Pro theme color palette - based on your design
	bgBase := "#0d0d0d"
	bgPanel := "#181818"
	borderDim := "#333333"
	borderFocus := "#ff9800"

	textMain := "#d4d4d4"
	textMuted := "#777777"

	accentOrange := "#ff9800"
	accentBlue := "#569cd6"
	accentGreen := "#4caf50"
	accentRed := "#f44336"

	// Diff colors
	diffBg := "#050505"
	diffAddBg := "#1a3a1a"
	diffAddTxt := "#89d185"
	diffDelBg := "#4a1f1f"
	diffDelTxt := "#f48771"

	theme := &ProTheme{}

	// Base colors
	theme.PrimaryColor = lipgloss.AdaptiveColor{
		Dark:  accentOrange,
		Light: accentOrange,
	}
	theme.SecondaryColor = lipgloss.AdaptiveColor{
		Dark:  accentBlue,
		Light: accentBlue,
	}
	theme.AccentColor = lipgloss.AdaptiveColor{
		Dark:  accentOrange,
		Light: accentOrange,
	}

	// Status colors
	theme.ErrorColor = lipgloss.AdaptiveColor{
		Dark:  accentRed,
		Light: accentRed,
	}
	theme.WarningColor = lipgloss.AdaptiveColor{
		Dark:  accentOrange,
		Light: accentOrange,
	}
	theme.SuccessColor = lipgloss.AdaptiveColor{
		Dark:  accentGreen,
		Light: accentGreen,
	}
	theme.InfoColor = lipgloss.AdaptiveColor{
		Dark:  accentBlue,
		Light: accentBlue,
	}

	// Text colors
	theme.TextColor = lipgloss.AdaptiveColor{
		Dark:  textMain,
		Light: textMain,
	}
	theme.TextMutedColor = lipgloss.AdaptiveColor{
		Dark:  textMuted,
		Light: textMuted,
	}
	theme.TextEmphasizedColor = lipgloss.AdaptiveColor{
		Dark:  textMain,
		Light: textMain,
	}

	// Background colors
	theme.BackgroundColor = lipgloss.AdaptiveColor{
		Dark:  bgBase,
		Light: bgBase,
	}
	theme.BackgroundSecondaryColor = lipgloss.AdaptiveColor{
		Dark:  bgPanel,
		Light: bgPanel,
	}
	theme.BackgroundDarkerColor = lipgloss.AdaptiveColor{
		Dark:  diffBg,
		Light: diffBg,
	}

	// Border colors
	theme.BorderNormalColor = lipgloss.AdaptiveColor{
		Dark:  borderDim,
		Light: borderDim,
	}
	theme.BorderFocusedColor = lipgloss.AdaptiveColor{
		Dark:  borderFocus,
		Light: borderFocus,
	}
	theme.BorderDimColor = lipgloss.AdaptiveColor{
		Dark:  borderDim,
		Light: borderDim,
	}

	// Diff view colors
	theme.DiffAddedColor = lipgloss.AdaptiveColor{
		Dark:  diffAddTxt,
		Light: diffAddTxt,
	}
	theme.DiffRemovedColor = lipgloss.AdaptiveColor{
		Dark:  diffDelTxt,
		Light: diffDelTxt,
	}
	theme.DiffContextColor = lipgloss.AdaptiveColor{
		Dark:  textMuted,
		Light: textMuted,
	}
	theme.DiffHunkHeaderColor = lipgloss.AdaptiveColor{
		Dark:  accentOrange,
		Light: accentOrange,
	}
	theme.DiffHighlightAddedColor = lipgloss.AdaptiveColor{
		Dark:  diffAddTxt,
		Light: diffAddTxt,
	}
	theme.DiffHighlightRemovedColor = lipgloss.AdaptiveColor{
		Dark:  diffDelTxt,
		Light: diffDelTxt,
	}
	theme.DiffAddedBgColor = lipgloss.AdaptiveColor{
		Dark:  diffAddBg,
		Light: diffAddBg,
	}
	theme.DiffRemovedBgColor = lipgloss.AdaptiveColor{
		Dark:  diffDelBg,
		Light: diffDelBg,
	}
	theme.DiffContextBgColor = lipgloss.AdaptiveColor{
		Dark:  diffBg,
		Light: diffBg,
	}
	theme.DiffLineNumberColor = lipgloss.AdaptiveColor{
		Dark:  "#555555",
		Light: "#555555",
	}
	theme.DiffAddedLineNumberBgColor = lipgloss.AdaptiveColor{
		Dark:  "#1a3a1a",
		Light: "#1a3a1a",
	}
	theme.DiffRemovedLineNumberBgColor = lipgloss.AdaptiveColor{
		Dark:  "#4a1f1f",
		Light: "#4a1f1f",
	}

	// Markdown colors
	theme.MarkdownTextColor = lipgloss.AdaptiveColor{
		Dark:  textMain,
		Light: textMain,
	}
	theme.MarkdownHeadingColor = lipgloss.AdaptiveColor{
		Dark:  accentOrange,
		Light: accentOrange,
	}
	theme.MarkdownLinkColor = lipgloss.AdaptiveColor{
		Dark:  accentBlue,
		Light: accentBlue,
	}
	theme.MarkdownLinkTextColor = lipgloss.AdaptiveColor{
		Dark:  accentBlue,
		Light: accentBlue,
	}
	theme.MarkdownCodeColor = lipgloss.AdaptiveColor{
		Dark:  accentGreen,
		Light: accentGreen,
	}
	theme.MarkdownBlockQuoteColor = lipgloss.AdaptiveColor{
		Dark:  textMuted,
		Light: textMuted,
	}
	theme.MarkdownEmphColor = lipgloss.AdaptiveColor{
		Dark:  textMuted,
		Light: textMuted,
	}
	theme.MarkdownStrongColor = lipgloss.AdaptiveColor{
		Dark:  textMain,
		Light: textMain,
	}
	theme.MarkdownHorizontalRuleColor = lipgloss.AdaptiveColor{
		Dark:  borderDim,
		Light: borderDim,
	}
	theme.MarkdownListItemColor = lipgloss.AdaptiveColor{
		Dark:  textMain,
		Light: textMain,
	}
	theme.MarkdownListEnumerationColor = lipgloss.AdaptiveColor{
		Dark:  textMuted,
		Light: textMuted,
	}
	theme.MarkdownImageColor = lipgloss.AdaptiveColor{
		Dark:  accentBlue,
		Light: accentBlue,
	}
	theme.MarkdownImageTextColor = lipgloss.AdaptiveColor{
		Dark:  textMuted,
		Light: textMuted,
	}
	theme.MarkdownCodeBlockColor = lipgloss.AdaptiveColor{
		Dark:  accentGreen,
		Light: accentGreen,
	}

	// Border style
	theme.BorderStyleValue = BorderStyle{
		Normal:  lipgloss.NormalBorder(),
		Focused: lipgloss.RoundedBorder(),
		Dim:     lipgloss.NormalBorder(),
	}

	// Diff styles for diff rendering
	theme.DiffStylesValue = DiffStyles{
		HunkHeader:      lipgloss.NewStyle().Foreground(lipgloss.Color(accentOrange)).Bold(true),
		AddLineNumber:   lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")),
		AddIndicator:    lipgloss.NewStyle().Foreground(lipgloss.Color(diffAddTxt)),
		Addition:        lipgloss.NewStyle().Foreground(lipgloss.Color(diffAddTxt)),
		Deletion:        lipgloss.NewStyle().Foreground(lipgloss.Color(diffDelTxt)),
		DelLineNumber:   lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")),
		DelIndicator:    lipgloss.NewStyle().Foreground(lipgloss.Color(diffDelTxt)),
		LineNumberStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")),
		ContextStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color(textMuted)),
	}

	// Common UI colors
	theme.ColorsValue = Colors{
		BorderFocus: lipgloss.AdaptiveColor{Dark: borderFocus, Light: borderFocus},
		BorderDim:   lipgloss.AdaptiveColor{Dark: borderDim, Light: borderDim},
		AccentGreen: lipgloss.AdaptiveColor{Dark: accentGreen, Light: accentGreen},
		AccentRed:   lipgloss.AdaptiveColor{Dark: accentRed, Light: accentRed},
		BgSurface:   lipgloss.AdaptiveColor{Dark: bgPanel, Light: bgPanel},
		TextMain:    lipgloss.AdaptiveColor{Dark: textMain, Light: textMain},
	}

	// Chat bubble styles
	theme.ChatStylesValue = ChatStyles{
		User:      lipgloss.NewStyle().Background(lipgloss.Color(accentOrange)).Foreground(lipgloss.Color(textMain)),
		Assistant: lipgloss.NewStyle().Background(lipgloss.Color(bgPanel)).Foreground(lipgloss.Color(textMain)),
		Role:      lipgloss.NewStyle().Foreground(lipgloss.Color(textMuted)).Bold(true),
	}

	// Syntax theme name
	theme.SyntaxThemeName = "gruvbox-dark"

	return theme
}

func init() {
	// Register the Pro theme with the theme manager
	RegisterTheme("pro", NewProTheme())
}
