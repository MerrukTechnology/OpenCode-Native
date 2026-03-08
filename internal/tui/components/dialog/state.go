// Package dialog provides UI components for the OpenCode TUI.
package dialog

// Dialog constants for map-based dialog state management.
const (
	// DialogPermissions is the permission dialog identifier.
	DialogPermissions = "permissions"
	// DialogHelp is the help dialog identifier.
	DialogHelp = "help"
	// DialogQuit is the quit confirmation dialog identifier.
	DialogQuit = "quit"
	// DialogSession is the session management dialog identifier.
	DialogSession = "session"
	// DialogDeleteSession is the delete session dialog identifier.
	DialogDeleteSession = "deleteSession"
	// DialogCommand is the command selection dialog identifier.
	DialogCommand = "command"
	// DialogModel is the model selection dialog identifier.
	DialogModel = "model"
	// DialogInit is the initialization dialog identifier.
	DialogInit = "init"
	// DialogFilepicker is the file picker dialog identifier.
	DialogFilepicker = "filepicker"
	// DialogTheme is the theme selection dialog identifier.
	DialogTheme = "theme"
	// DialogMultiArguments is the multi-arguments dialog identifier.
	DialogMultiArguments = "multiArguments"
)

// DialogState is a map-based state for managing dialog visibility.
type DialogState map[string]bool

// NewDialogState creates a new DialogState with all dialogs hidden.
func NewDialogState() DialogState {
	return DialogState{
		DialogPermissions:     false,
		DialogHelp:            false,
		DialogQuit:            false,
		DialogSession:         false,
		DialogDeleteSession:   false,
		DialogCommand:         false,
		DialogModel:           false,
		DialogInit:            false,
		DialogFilepicker:      false,
		DialogTheme:           false,
		DialogMultiArguments: false,
	}
}

// ShowDialog sets a dialog to visible.
func (d DialogState) ShowDialog(name string) {
	d[name] = true
}

// HideDialog sets a dialog to hidden.
func (d DialogState) HideDialog(name string) {
	d[name] = false
}

// ToggleDialog toggles a dialog's visibility.
func (d DialogState) ToggleDialog(name string) {
	d[name] = !d[name]
}

// IsDialogOpen checks if a specific dialog is open.
func (d DialogState) IsDialogOpen(name string) bool {
	return d[name]
}

// IsAnyDialogOpen returns true if any dialog is currently shown.
func (d DialogState) IsAnyDialogOpen() bool {
	for _, v := range d {
		if v {
			return true
		}
	}
	return false
}

// CloseAllDialogs closes all open dialogs. Returns true if any were closed.
func (d DialogState) CloseAllDialogs() bool {
	anyClosed := false
	for name := range d {
		if d[name] {
			d[name] = false
			anyClosed = true
		}
	}
	return anyClosed
}
