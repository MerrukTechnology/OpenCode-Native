// Package styles provides styling utilities for the OpenCode TUI.
//
// This package includes:
//   - Icon constants for common UI elements (check, error, warning, etc.)
//   - Lipgloss style helpers for consistent theming
//   - Background color manipulation functions
//   - Markdown rendering styles
//
// The package integrates with the theme system to provide consistent
// styling across the entire TUI application.
package styles

// Icon constants used throughout the TUI for visual indicators.
const (
	// OpenCodeIcon is the logo icon displayed in the header.
	OpenCodeIcon string = "⌬"

	// CheckIcon is used to indicate success or completion.
	CheckIcon string = "✓"
	// ErrorIcon is used to indicate errors.
	ErrorIcon string = "✖"
	// WarningIcon is used to indicate warnings.
	WarningIcon string = "⚠"
	// InfoIcon is used for informational messages (currently empty).
	InfoIcon string = ""
	// HintIcon is used to display hints or tips.
	HintIcon string = "i"
	// SpinnerIcon is used during loading operations.
	SpinnerIcon string = "..."
	// LoadingIcon is used for loading states.
	LoadingIcon string = "⟳"
	// DocumentIcon is used for document or file icons.
	DocumentIcon string = "🖼"
)
