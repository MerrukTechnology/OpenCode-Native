// Package util provides utility types and functions for the OpenCode TUI.
//
// This package includes:
//   - InfoMsg: Messages for displaying status information (info, warning, error)
//   - ClearStatusMsg: Message for clearing status display
//   - Helper functions for reporting errors, warnings, and info messages
//
// These utilities are used across the TUI to display transient status messages
// to the user, such as error notifications or operation progress updates.
package util

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// CmdHandler wraps a message into a tea.Cmd for asynchronous handling.
// This is a convenience function for creating commands that return a specific message.
func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

// ReportError creates a command that reports an error message to the user.
// The error will be displayed as an InfoMsg with InfoTypeError.
func ReportError(err error) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeError,
		Msg:  err.Error(),
	})
}

// InfoType represents the type of information message.
type InfoType int

const (
	// InfoTypeInfo indicates an informational message.
	InfoTypeInfo InfoType = iota
	// InfoTypeWarn indicates a warning message.
	InfoTypeWarn
	// InfoTypeError indicates an error message.
	InfoTypeError
)

// ReportInfo creates a command that reports an informational message to the user.
// The message will be displayed as an InfoMsg with InfoTypeInfo.
func ReportInfo(info string) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeInfo,
		Msg:  info,
	})
}

// ReportWarn creates a command that reports a warning message to the user.
// The message will be displayed as an InfoMsg with InfoTypeWarn.
func ReportWarn(warn string) tea.Cmd {
	return CmdHandler(InfoMsg{
		Type: InfoTypeWarn,
		Msg:  warn,
	})
}

// InfoMsg is a message type for displaying status information to the user.
// It includes the message type (info, warning, error), the message text,
// and an optional time-to-live (TTL) for auto-dismissal.
type (
	InfoMsg struct {
		// Type is the kind of message (Info, Warn, or Error).
		Type InfoType
		// Msg is the text content of the message.
		Msg string
		// TTL is the time duration after which the message should be cleared.
		// A value of 0 means the message persists until manually cleared.
		TTL time.Duration
	}
	// ClearStatusMsg is sent to clear any displayed status message.
	ClearStatusMsg struct{}
)

func Clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
