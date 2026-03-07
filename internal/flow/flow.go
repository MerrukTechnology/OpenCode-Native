package flow

import (
	"errors"
	"strings"
)

var (
	ErrFlowNotFound     = errors.New("flow not found")
	ErrFlowDisabled     = errors.New("flow is disabled")
	ErrInvalidFlowName  = errors.New("invalid flow name")
	ErrDuplicateFlowID  = errors.New("duplicate flow ID")
	ErrInvalidStepID    = errors.New("invalid step ID")
	ErrDuplicateStepID  = errors.New("duplicate step ID")
	ErrInvalidRule      = errors.New("rule references non-existent step")
	ErrInvalidFallback  = errors.New("fallback references non-existent step")
	ErrNoSteps          = errors.New("flow has no steps")
	ErrInvalidYAML      = errors.New("invalid flow YAML")
	ErrInvalidPredicate = errors.New("invalid predicate")
	// TODO(#): Implement cycle detection in step graph validation
	// Use DFS or Tarjan's algorithm to detect circular dependencies
	ErrCycleDetected = errors.New("cycle detected in step graph")
)

// Flow represents a discovered flow definition.
type Flow struct {
	ID          string
	Name        string
	Disabled    bool
	Description string
	Spec        FlowSpec
	Location    string
}

// FlowSession controls session behavior at the flow level.
type FlowSession struct {
	Prefix string `yaml:"prefix,omitempty"`
}

// FlowSpec contains the flow's args schema and step definitions.
type FlowSpec struct {
	Args    map[string]any `yaml:"args,omitempty"`
	Session FlowSession    `yaml:"session,omitempty"`
	Steps   []Step         `yaml:"steps"`
}

// Step defines a single step in the flow graph.
type Step struct {
	ID       string      `yaml:"id"`
	Agent    string      `yaml:"agent,omitempty"`
	Session  StepSession `yaml:"session,omitempty"`
	Prompt   string      `yaml:"prompt"`
	Output   *StepOutput `yaml:"output,omitempty"`
	Rules    []Rule      `yaml:"rules,omitempty"`
	Fallback *Fallback   `yaml:"fallback,omitempty"`
}

// StepSession controls session behavior for a step.
type StepSession struct {
	Fork bool `yaml:"fork,omitempty"`
}

// StepOutput defines optional structured output for a step.
type StepOutput struct {
	Schema map[string]any `yaml:"schema"`
}

// Rule defines a conditional routing rule evaluated after step completion.
type Rule struct {
	If       string `yaml:"if"`
	Then     string `yaml:"then"`
	Postpone bool   `yaml:"postpone,omitempty"`
}

// Fallback defines retry and error-routing behavior for a step.
type Fallback struct {
	Retry int    `yaml:"retry"`
	Delay int    `yaml:"delay,omitempty"`
	To    string `yaml:"to,omitempty"`
}

// FlowConflict holds duplicate flow IDs and their locations.
type FlowConflict struct {
	ID        string
	Locations []string
}

// Conflicts aggregates all duplicate flow ID conflicts.
type Conflicts struct {
	Conflicts []FlowConflict
}

// Error returns a formatted error string for the conflicts.
func (c *Conflicts) Error() string {
	if len(c.Conflicts) == 0 {
		return ""
	}
	result := "duplicate flow IDs found:"
	var resultSB1 strings.Builder
	for _, conflict := range c.Conflicts {
		resultSB1.WriteString("\n  - " + conflict.ID + ":")
		var resultSB2 strings.Builder
		for _, loc := range conflict.Locations {
			resultSB2.WriteString("\n    - " + loc)
		}
		result += resultSB2.String()
	}
	result += resultSB1.String()
	return result
}

// HasConflicts returns true if there are any conflicts.
func (c *Conflicts) HasConflicts() bool {
	return len(c.Conflicts) > 0
}
