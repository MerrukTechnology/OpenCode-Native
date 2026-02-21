package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
)

func TestValidateAndTruncate(t *testing.T) {
	t.Run("small response not truncated", func(t *testing.T) {
		response := NewTextResponse("Hello, world!")
		if len(response.Content) != 13 {
			t.Errorf("expected length 13, got %d", len(response.Content))
		}
		if strings.Contains(response.Content, "[Output truncated") {
			t.Error("small response should not be truncated")
		}
	})

	t.Run("large response truncated", func(t *testing.T) {
		// Create content larger than MaxToolResponseTokens
		// 1.5M chars = ~375k tokens, which exceeds 300k limit
		largeContent := strings.Repeat("A", 1_500_000)
		response := NewTextResponse(largeContent)

		expectedMaxChars := MaxToolResponseTokens * 4
		if len(response.Content) <= expectedMaxChars {
			t.Errorf("expected content to be truncated to ~%d chars, got %d", expectedMaxChars, len(response.Content))
		}
		if !strings.Contains(response.Content, "[Output truncated") {
			t.Error("large response should contain truncation message")
		}
	})

	t.Run("error response also truncated", func(t *testing.T) {
		largeContent := strings.Repeat("Error: ", 200_000)
		response := NewTextErrorResponse(largeContent)

		if !response.IsError {
			t.Error("response should be marked as error")
		}
		if !strings.Contains(response.Content, "[Output truncated") {
			t.Error("large error response should be truncated")
		}
	})

	t.Run("empty response not truncated", func(t *testing.T) {
		response := NewEmptyResponse()
		if response.Content != "" {
			t.Errorf("expected empty content, got %q", response.Content)
		}
		if response.Type != ToolResponseTypeText {
			t.Errorf("expected text type, got %v", response.Type)
		}
	})

	t.Run("image response truncated not truncated (will be corrupted)", func(t *testing.T) {
		largeContent := strings.Repeat("base64data", 200_000)
		response := NewImageResponse(largeContent)

		if response.Type != ToolResponseTypeImage {
			t.Errorf("expected image type, got %v", response.Type)
		}
		if strings.Contains(response.Content, "[Output truncated") {
			t.Error("large image response should not be truncated")
		}
	})

	t.Run("response at exact limit not truncated", func(t *testing.T) {
		// Create content exactly at the limit
		exactContent := strings.Repeat("A", MaxToolResponseTokens*4)
		response := NewTextResponse(exactContent)

		if strings.Contains(response.Content, "[Output truncated") {
			t.Error("response at exact limit should not be truncated")
		}
	})

	t.Run("response just over limit truncated", func(t *testing.T) {
		// Create content just over the limit (need more than 4 chars to exceed 1 token)
		overContent := strings.Repeat("A", MaxToolResponseTokens*4+4)
		response := NewTextResponse(overContent)

		if !strings.Contains(response.Content, "[Output truncated") {
			t.Error("response just over limit should be truncated")
		}
	})
}

func TestIsTaskAgent(t *testing.T) {
	t.Run("returns false for empty context", func(t *testing.T) {
		ctx := context.Background()
		if IsTaskAgent(ctx) {
			t.Error("expected false for empty context")
		}
	})

	t.Run("returns true when context has task agent flag", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), IsTaskAgentContextKey, true)
		if !IsTaskAgent(ctx) {
			t.Error("expected true when context has task agent flag")
		}
	})

	t.Run("returns false when context has false flag", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), IsTaskAgentContextKey, false)
		if IsTaskAgent(ctx) {
			t.Error("expected false when context has false flag")
		}
	})

	t.Run("returns false when context has wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), IsTaskAgentContextKey, "not a bool")
		if IsTaskAgent(ctx) {
			t.Error("expected false when context has wrong type")
		}
	})
}

func TestGetContextValues(t *testing.T) {
	t.Run("returns empty strings for empty context", func(t *testing.T) {
		ctx := context.Background()
		sessionID, messageID := GetContextValues(ctx)
		if sessionID != "" || messageID != "" {
			t.Errorf("expected empty strings, got sessionID=%q, messageID=%q", sessionID, messageID)
		}
	})

	t.Run("returns sessionID when only sessionID is set", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")
		sessionID, messageID := GetContextValues(ctx)
		if sessionID != "test-session-123" {
			t.Errorf("expected sessionID=test-session-123, got %q", sessionID)
		}
		if messageID != "" {
			t.Errorf("expected empty messageID, got %q", messageID)
		}
	})

	t.Run("returns both values when both are set", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, SessionIDContextKey, "session-456")
		ctx = context.WithValue(ctx, MessageIDContextKey, "message-789")
		sessionID, messageID := GetContextValues(ctx)
		if sessionID != "session-456" {
			t.Errorf("expected sessionID=session-456, got %q", sessionID)
		}
		if messageID != "message-789" {
			t.Errorf("expected messageID=message-789, got %q", messageID)
		}
	})
}

func TestGetAgentID(t *testing.T) {
	t.Run("returns empty string for empty context", func(t *testing.T) {
		ctx := context.Background()
		agentID := GetAgentID(ctx)
		if agentID != "" {
			t.Errorf("expected empty string, got %q", agentID)
		}
	})

	t.Run("returns agentID when set", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), AgentIDContextKey, config.AgentCoder)
		agentID := GetAgentID(ctx)
		if agentID != config.AgentCoder {
			t.Errorf("expected agentID=%q, got %q", config.AgentCoder, agentID)
		}
	})

	t.Run("returns empty string when wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), AgentIDContextKey, 12345)
		agentID := GetAgentID(ctx)
		if agentID != "" {
			t.Errorf("expected empty string for wrong type, got %q", agentID)
		}
	})
}

func TestWithResponseMetadata(t *testing.T) {
	t.Run("adds metadata to response", func(t *testing.T) {
		response := NewTextResponse("test content")
		metadata := map[string]string{"key": "value"}
		response = WithResponseMetadata(response, metadata)

		if response.Metadata == "" {
			t.Error("expected metadata to be set")
		}
		if !strings.Contains(response.Metadata, "key") {
			t.Errorf("expected metadata to contain 'key', got %q", response.Metadata)
		}
	})

	t.Run("returns unchanged response for nil metadata", func(t *testing.T) {
		response := NewTextResponse("test content")
		originalMetadata := response.Metadata
		response = WithResponseMetadata(response, nil)

		if response.Metadata != originalMetadata {
			t.Error("expected metadata to be unchanged for nil input")
		}
	})
}

func TestToolInfoStruct(t *testing.T) {
	info := ToolInfo{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]any{
			"type": "object",
		},
		Required: []string{"path"},
	}

	if info.Name != "test_tool" {
		t.Errorf("Name = %q, want %q", info.Name, "test_tool")
	}
	if len(info.Required) != 1 {
		t.Errorf("Required length = %d, want %d", len(info.Required), 1)
	}
}

func TestToolCallStruct(t *testing.T) {
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
}

func TestToolResponseTypes(t *testing.T) {
	t.Run("text response type", func(t *testing.T) {
		if ToolResponseTypeText != "text" {
			t.Errorf("ToolResponseTypeText = %q, want %q", ToolResponseTypeText, "text")
		}
	})

	t.Run("image response type", func(t *testing.T) {
		if ToolResponseTypeImage != "image" {
			t.Errorf("ToolResponseTypeImage = %q, want %q", ToolResponseTypeImage, "image")
		}
	})
}

func TestContextKeys(t *testing.T) {
	t.Run("session ID context key", func(t *testing.T) {
		if SessionIDContextKey != "session_id" {
			t.Errorf("SessionIDContextKey = %q, want %q", SessionIDContextKey, "session_id")
		}
	})

	t.Run("message ID context key", func(t *testing.T) {
		if MessageIDContextKey != "message_id" {
			t.Errorf("MessageIDContextKey = %q, want %q", MessageIDContextKey, "message_id")
		}
	})

	t.Run("is task agent context key", func(t *testing.T) {
		if IsTaskAgentContextKey != "is_task_agent" {
			t.Errorf("IsTaskAgentContextKey = %q, want %q", IsTaskAgentContextKey, "is_task_agent")
		}
	})

	t.Run("agent ID context key", func(t *testing.T) {
		if AgentIDContextKey != "agent_id" {
			t.Errorf("AgentIDContextKey = %q, want %q", AgentIDContextKey, "agent_id")
		}
	})
}
