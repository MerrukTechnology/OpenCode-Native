package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
)

// TestValidateAndTruncate tests response truncation functionality
func TestValidateAndTruncate(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		truncate     bool
		responseType string
	}{
		{"small response not truncated", "Hello, world!", false, "text"},
		{"large response truncated", strings.Repeat("A", 1_500_000), true, "text"},
		{"error response also truncated", strings.Repeat("Error: ", 200_000), true, "error"},
		{"empty response not truncated", "", false, "text"},
		{"image response not truncated (will be corrupted)", strings.Repeat("base64data", 200_000), false, "image"},
		{"response at exact limit not truncated", strings.Repeat("A", MaxToolResponseTokens*4), false, "text"},
		{"response just over limit truncated", strings.Repeat("A", MaxToolResponseTokens*4+4), true, "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response ToolResponse
			switch tt.responseType {
			case "text":
				response = NewTextResponse(tt.content)
			case "error":
				response = NewTextErrorResponse(tt.content)
			case "image":
				response = NewImageResponse(tt.content)
			}

			// Verify type
			expectedType := ToolResponseTypeText
			if tt.responseType == "image" {
				expectedType = ToolResponseTypeImage
			}
			if response.Type != expectedType {
				t.Errorf("expected type %v, got %v", expectedType, response.Type)
			}

			// Verify truncation
			if tt.truncate {
				if !strings.Contains(response.Content, "[Output truncated") {
					t.Error("expected truncation message")
				}
			} else {
				if strings.Contains(response.Content, "[Output truncated") {
					t.Error("should not be truncated")
				}
			}

			// Specific checks for empty and error responses
			if tt.name == "empty response not truncated" && response.Content != "" {
				t.Errorf("expected empty content, got %q", response.Content)
			}
			if tt.responseType == "error" && !response.IsError {
				t.Error("response should be marked as error")
			}
		})
	}
}

// TestContextFunctions tests context retrieval functions with various scenarios
func TestContextFunctions(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		assertions func(*testing.T, context.Context)
	}{
		{
			name:     "IsTaskAgent - empty context returns false",
			setupCtx: func() context.Context { return context.Background() },
			assertions: func(t *testing.T, ctx context.Context) {
				if IsTaskAgent(ctx) {
					t.Error("expected false for empty context")
				}
			},
		},
		{
			name: "IsTaskAgent - context with true flag returns true",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), IsTaskAgentContextKey, true)
			},
			assertions: func(t *testing.T, ctx context.Context) {
				if !IsTaskAgent(ctx) {
					t.Error("expected true when context has task agent flag")
				}
			},
		},
		{
			name: "IsTaskAgent - context with false flag returns false",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), IsTaskAgentContextKey, false)
			},
			assertions: func(t *testing.T, ctx context.Context) {
				if IsTaskAgent(ctx) {
					t.Error("expected false when context has false flag")
				}
			},
		},
		{
			name: "IsTaskAgent - context with wrong type returns false",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), IsTaskAgentContextKey, "not a bool")
			},
			assertions: func(t *testing.T, ctx context.Context) {
				if IsTaskAgent(ctx) {
					t.Error("expected false when context has wrong type")
				}
			},
		},
		{
			name:     "GetContextValues - empty context returns empty strings",
			setupCtx: func() context.Context { return context.Background() },
			assertions: func(t *testing.T, ctx context.Context) {
				sessionID, messageID := GetContextValues(ctx)
				if sessionID != "" || messageID != "" {
					t.Errorf("expected empty strings, got sessionID=%q, messageID=%q", sessionID, messageID)
				}
			},
		},
		{
			name: "GetContextValues - only sessionID is set",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")
			},
			assertions: func(t *testing.T, ctx context.Context) {
				sessionID, messageID := GetContextValues(ctx)
				if sessionID != "test-session-123" {
					t.Errorf("expected sessionID=test-session-123, got %q", sessionID)
				}
				if messageID != "" {
					t.Errorf("expected empty messageID, got %q", messageID)
				}
			},
		},
		{
			name: "GetContextValues - both values are set",
			setupCtx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, SessionIDContextKey, "session-456")
				return context.WithValue(ctx, MessageIDContextKey, "message-789")
			},
			assertions: func(t *testing.T, ctx context.Context) {
				sessionID, messageID := GetContextValues(ctx)
				if sessionID != "session-456" {
					t.Errorf("expected sessionID=session-456, got %q", sessionID)
				}
				if messageID != "message-789" {
					t.Errorf("expected messageID=message-789, got %q", messageID)
				}
			},
		},
		{
			name:     "GetAgentID - empty context returns empty string",
			setupCtx: func() context.Context { return context.Background() },
			assertions: func(t *testing.T, ctx context.Context) {
				agentID := GetAgentID(ctx)
				if agentID != "" {
					t.Errorf("expected empty string, got %q", agentID)
				}
			},
		},
		{
			name: "GetAgentID - context with agent ID set",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), AgentIDContextKey, config.AgentCoder)
			},
			assertions: func(t *testing.T, ctx context.Context) {
				agentID := GetAgentID(ctx)
				if agentID != config.AgentCoder {
					t.Errorf("expected agentID=%q, got %q", config.AgentCoder, agentID)
				}
			},
		},
		{
			name: "GetAgentID - wrong type returns empty string",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), AgentIDContextKey, 12345)
			},
			assertions: func(t *testing.T, ctx context.Context) {
				agentID := GetAgentID(ctx)
				if agentID != "" {
					t.Errorf("expected empty string for wrong type, got %q", agentID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			tt.assertions(t, ctx)
		})
	}
}

func TestWithResponseMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata any
	}{
		{"adds metadata to response", map[string]string{"key": "value"}},
		{"returns unchanged response for nil metadata", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewTextResponse("test content")
			originalMetadata := response.Metadata
			response = WithResponseMetadata(response, tt.metadata)

			if tt.metadata == nil {
				if response.Metadata != originalMetadata {
					t.Error("expected metadata to be unchanged for nil input")
				}
			} else {
				if response.Metadata == "" {
					t.Error("expected metadata to be set")
				}
				if !strings.Contains(response.Metadata, "key") {
					t.Errorf("expected metadata to contain 'key', got %q", response.Metadata)
				}
			}
		})
	}
}

// TestStructs tests tool-related struct initialization
func TestStructs(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "ToolInfo struct",
			fn: func(t *testing.T) {
				info := ToolInfo{
					Name:        "test_tool",
					Description: "A test tool",
					Parameters:  map[string]any{"type": "object"},
					Required:    []string{"path"},
				}
				if info.Name != "test_tool" {
					t.Errorf("Name = %q, want %q", info.Name, "test_tool")
				}
				if len(info.Required) != 1 {
					t.Errorf("Required length = %d, want %d", len(info.Required), 1)
				}
			},
		},
		{
			name: "ToolCall struct",
			fn: func(t *testing.T) {
				call := ToolCall{
					ID:    "call-123",
					Name:  "read_file",
					Input: `{"path": "/test/file.txt"}`,
				}
				if call.ID != "call-123" {
					t.Errorf("ID = %q, want %q", call.ID, "call-123")
				}
				if call.Name != "read_file" {
					t.Errorf("Name = %q, want %q", call.Name, "read_file")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

// TestConstants tests tool-related constants
func TestConstants(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"ToolResponseTypeText", string(ToolResponseTypeText), "text"},
		{"ToolResponseTypeImage", string(ToolResponseTypeImage), "image"},
		{"SessionIDContextKey", string(SessionIDContextKey), "session_id"},
		{"MessageIDContextKey", string(MessageIDContextKey), "message_id"},
		{"IsTaskAgentContextKey", string(IsTaskAgentContextKey), "is_task_agent"},
		{"AgentIDContextKey", string(AgentIDContextKey), "agent_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.value.(type) {
			case string:
				if v != tt.want {
					t.Errorf("%s = %q, want %q", tt.name, v, tt.want)
				}
			default:
				t.Errorf("unexpected type for %s: %T", tt.name, tt.value)
			}
		})
	}
}
