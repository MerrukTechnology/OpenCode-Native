// Package layout provides responsive layout calculations for the OpenCode TUI.
// It offers utilities for dimension calculation, split panes, and container wrappers
// that automatically adapt to terminal window size changes.
//
// The package provides:
//   - LayoutEngine: Responsive dimension calculator for all UI sections
//   - SplitPaneLayout: Horizontal/vertical split container layout
//   - Container: Wrapper with padding and borders
//   - Sizeable, Bindings, Focusable: Core interfaces for tea.Model components
package layout

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Focusable is an interface for components that can receive focus.
// Implement this interface to enable keyboard navigation support.
type Focusable interface {
	// Focus gives focus to the component.
	Focus() tea.Cmd
	// Blur removes focus from the component.
	Blur() tea.Cmd
	// IsFocused returns true if the component currently has focus.
	IsFocused() bool
}

// Sizeable is an interface for components that can be sized.
// All visual components should implement this to support responsive layouts.
type Sizeable interface {
	// SetSize sets the dimensions of the component.
	SetSize(width, height int) tea.Cmd
	// GetSize returns the current dimensions of the component.
	GetSize() (int, int)
}

// Bindings is an interface for components that have keyboard bindings.
// Implement this to expose keyboard shortcuts to the application.
type Bindings interface {
	// BindingKeys returns the keyboard bindings for this component.
	BindingKeys() []key.Binding
}

// KeyMapToSlice extracts key bindings from a struct into a slice.
// The struct fields must be of type key.Binding.
func KeyMapToSlice(t any) (bindings []key.Binding) {
	typ := reflect.TypeOf(t)
	if typ.Kind() != reflect.Struct {
		return nil
	}
	for i := range typ.NumField() {
		v := reflect.ValueOf(t).Field(i)
		bindings = append(bindings, v.Interface().(key.Binding))
	}
	return
}
