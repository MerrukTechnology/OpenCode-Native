package agent

import (
	"context"
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
)

// TestAgentID tests the AgentID method - tests the basic getter functionality
func TestAgentID(t *testing.T) {
	// Test that AgentID returns the correct agent ID
	// This tests the simplest method that doesn't require external dependencies
	a := &agent{
		agentID: config.AgentCoder,
	}
	result := a.AgentID()
	if result != config.AgentCoder {
		t.Errorf("AgentID() = %v, want %v", result, config.AgentCoder)
	}
}

// TestAgentID_VariousAgents tests AgentID with various agent types
func TestAgentID_VariousAgents(t *testing.T) {
	tests := []struct {
		name     string
		agentID  config.AgentName
		expected config.AgentName
	}{
		{"coder agent", config.AgentCoder, config.AgentCoder},
		{"explorer agent", config.AgentExplorer, config.AgentExplorer},
		{"hivemind agent", config.AgentHivemind, config.AgentHivemind},
		{"workhorse agent", config.AgentWorkhorse, config.AgentWorkhorse},
		{"summarizer agent", config.AgentSummarizer, config.AgentSummarizer},
		{"descriptor agent", config.AgentDescriptor, config.AgentDescriptor},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &agent{
				agentID: tc.agentID,
			}
			result := a.AgentID()
			if result != tc.expected {
				t.Errorf("AgentID() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// TestIsBusy tests the IsBusy method
func TestIsBusy(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*agent)
		expected bool
	}{
		{
			name:     "empty requests",
			setup:    func(a *agent) {},
			expected: false,
		},
		{
			name: "with active request",
			setup: func(a *agent) {
				_, cancel := context.WithCancel(context.Background())
				a.activeRequests.Store("session-1", cancel)
			},
			expected: true,
		},
		{
			name: "with multiple active requests",
			setup: func(a *agent) {
				_, cancel1 := context.WithCancel(context.Background())
				_, cancel2 := context.WithCancel(context.Background())
				a.activeRequests.Store("session-1", cancel1)
				a.activeRequests.Store("session-2", cancel2)
			},
			expected: true,
		},
		{
			name: "with nil cancel func",
			setup: func(a *agent) {
				a.activeRequests.Store("session-1", nil)
			},
			expected: false, // nil cancel func should not be considered busy
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &agent{}
			tc.setup(a)
			result := a.IsBusy()
			if result != tc.expected {
				t.Errorf("IsBusy() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// TestIsSessionBusy tests the IsSessionBusy method
func TestIsSessionBusy(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		setup     func(*agent)
		expected  bool
	}{
		{
			name:      "session not busy",
			sessionID: "session-1",
			setup:     func(a *agent) {},
			expected:  false,
		},
		{
			name:      "session is busy",
			sessionID: "session-1",
			setup: func(a *agent) {
				a.activeRequests.Store("session-1", func() {})
			},
			expected: true,
		},
		{
			name:      "different session not busy",
			sessionID: "session-2",
			setup: func(a *agent) {
				a.activeRequests.Store("session-1", func() {})
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &agent{}
			tc.setup(a)
			result := a.IsSessionBusy(tc.sessionID)
			if result != tc.expected {
				t.Errorf("IsSessionBusy(%q) = %v, want %v", tc.sessionID, result, tc.expected)
			}
		})
	}
}

// TestCancel tests the Cancel method
func TestCancel(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		setup     func(*agent) map[string]any
		check     func(*agent, string) bool
	}{
		{
			name:      "cancel regular request",
			sessionID: "session-1",
			setup: func(a *agent) map[string]any {
				a.activeRequests.Store("session-1", func() {})
				return map[string]any{"session-1": true}
			},
			check: func(a *agent, sessionID string) bool {
				_, exists := a.activeRequests.Load(sessionID)
				return !exists
			},
		},
		{
			name:      "cancel summarize request",
			sessionID: "session-1",
			setup: func(a *agent) map[string]any {
				a.activeRequests.Store("session-1-summarize", func() {})
				return map[string]any{"session-1-summarize": true}
			},
			check: func(a *agent, sessionID string) bool {
				_, exists := a.activeRequests.Load(sessionID + "-summarize")
				return !exists
			},
		},
		{
			name:      "cancel non-existent request",
			sessionID: "session-nonexistent",
			setup: func(a *agent) map[string]any {
				return nil
			},
			check: func(a *agent, sessionID string) bool {
				return true // should not panic
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &agent{}
			tc.setup(a)
			a.Cancel(tc.sessionID)
			if !tc.check(a, tc.sessionID) {
				t.Errorf("Cancel(%q) failed check", tc.sessionID)
			}
		})
	}
}

// TestConstants tests the error constants
func TestConstants(t *testing.T) {
	if ErrRequestCancelled == nil {
		t.Error("ErrRequestCancelled should not be nil")
	}
	if ErrSessionBusy == nil {
		t.Error("ErrSessionBusy should not be nil")
	}
	if ErrRequestCancelled.Error() != "request cancelled by user" {
		t.Errorf("ErrRequestCancelled = %q", ErrRequestCancelled.Error())
	}
	if ErrSessionBusy.Error() != "session is currently processing another request" {
		t.Errorf("ErrSessionBusy = %q", ErrSessionBusy.Error())
	}
}

// TestAgentEventType tests the event type constants
func TestAgentEventType(t *testing.T) {
	tests := []struct {
		name     string
		got      AgentEventType
		expected string
	}{
		{"error event", AgentEventTypeError, "error"},
		{"response event", AgentEventTypeResponse, "response"},
		{"summarize event", AgentEventTypeSummarize, "summarize"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if string(tc.got) != tc.expected {
				t.Errorf("AgentEventType = %q, want %q", tc.got, tc.expected)
			}
		})
	}
}

// TestAutoCompactionThreshold tests the compaction threshold constant
func TestAutoCompactionThreshold(t *testing.T) {
	if AutoCompactionThreshold != 0.95 {
		t.Errorf("AutoCompactionThreshold = %v, want 0.95", AutoCompactionThreshold)
	}
}
