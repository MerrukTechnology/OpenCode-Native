package dialog

import "sync"

// CommandRegistry manages commands for the command dialog.
type CommandRegistry struct {
	mu       sync.RWMutex
	commands []Command
}

// NewCommandRegistry creates a new CommandRegistry.
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make([]Command, 0),
	}
}

// Register adds a command to the registry.
func (r *CommandRegistry) Register(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = append(r.commands, cmd)
}

// Find searches for a command by its ID.
// Returns the command and true if found, or an empty command and false if not found.
func (r *CommandRegistry) Find(id string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cmd := range r.commands {
		if cmd.ID == id {
			return cmd, true
		}
	}
	return Command{}, false
}

// All returns a copy of all registered commands.
func (r *CommandRegistry) All() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Command, len(r.commands))
	copy(result, r.commands)
	return result
}

// SetCommands replaces all commands in the registry.
func (r *CommandRegistry) SetCommands(commands []Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = commands
}

// DefaultRegistry is the package-level command registry used by the application.
var DefaultRegistry = NewCommandRegistry()
