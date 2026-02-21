// Package tools provides tool implementations for the LLM agent system.
// It includes file operations (read, write, edit, delete), shell commands,
// search utilities, and integration with LSP servers.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/fileutil"
)

type ToolInfo struct {
	Name        string
	Description string
	Parameters  map[string]any
	Required    []string
	// TODO: Consider to add Output parameters: https://modelcontextprotocol.io/specification/2025-06-18/server/tools#output-schema
}

type toolResponseType string

type (
	sessionIDContextKey   string
	messageIDContextKey   string
	isTaskAgentContextKey string
	agentIDContextKey     string
)

const (
	ToolResponseTypeText  toolResponseType = "text"
	ToolResponseTypeImage toolResponseType = "image"

	SessionIDContextKey   sessionIDContextKey   = "session_id"
	MessageIDContextKey   messageIDContextKey   = "message_id"
	IsTaskAgentContextKey isTaskAgentContextKey = "is_task_agent"
	AgentIDContextKey     agentIDContextKey     = "agent_id"

	// MaxToolResponseTokens is the maximum number of tokens allowed in a tool response
	// to prevent context overflow. ~1200KB of text content.
	MaxToolResponseTokens = 300_000
)

type toolResponse struct {
	Type     toolResponseType `json:"type"`
	Content  string           `json:"content"`
	Metadata string           `json:"metadata,omitempty"`
	IsError  bool             `json:"is_error"`
}

// ToolResponse is the public interface for tool responses
type ToolResponse = toolResponse

// validateAndTruncate validates the tool response size and truncates if necessary
func validateAndTruncate(response toolResponse) toolResponse {
	// Rough estimation: ~4 characters per token
	estimatedTokens := len(response.Content) / 4

	if estimatedTokens > MaxToolResponseTokens {
		maxChars := MaxToolResponseTokens * 4
		truncated := response.Content[:maxChars]
		response.Content = truncated + "\n\n[Output truncated due to size limit. Consider using more specific search parameters or viewing smaller sections.]"
	}

	return response
}

func NewTextResponse(content string) toolResponse {
	return validateAndTruncate(toolResponse{
		Type:    ToolResponseTypeText,
		Content: content,
	})
}

func NewImageResponse(content string) toolResponse {
	return toolResponse{
		Type:    ToolResponseTypeImage,
		Content: content,
	}
}

func NewEmptyResponse() toolResponse {
	return toolResponse{
		Type:    ToolResponseTypeText,
		Content: "",
	}
}

func WithResponseMetadata(response toolResponse, metadata any) toolResponse {
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return response
		}
		response.Metadata = string(metadataBytes)
	}
	return response
}

func NewTextErrorResponse(content string) toolResponse {
	return validateAndTruncate(toolResponse{
		Type:    ToolResponseTypeText,
		Content: content,
		IsError: true,
	})
}

// ValidatePathInWorkingDirectory checks if a path is within the working directory
// to prevent path traversal attacks. It returns the absolute path if valid, or an error if not.
// This is a wrapper around fileutil.SecureResolvePath that uses config.WorkingDirectory().
func ValidatePathInWorkingDirectory(filePath string) (string, error) {
	absPath, err := fileutil.SecureResolvePath(filePath, config.WorkingDirectory())
	if err != nil {
		return "", fmt.Errorf("invalid file path: %s attempts to escape working directory (outside the working directory)", filePath)
	}
	return absPath, nil
}

type ToolCall struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Input string `json:"input"`
}

type BaseTool interface {
	Info() ToolInfo
	Run(ctx context.Context, params ToolCall) (ToolResponse, error)
}

func GetContextValues(ctx context.Context) (string, string) {
	sessionID := ctx.Value(SessionIDContextKey)
	messageID := ctx.Value(MessageIDContextKey)
	if sessionID == nil {
		return "", ""
	}
	if messageID == nil {
		return sessionID.(string), ""
	}
	return sessionID.(string), messageID.(string)
}

// IsTaskAgent returns true if the context indicates this is a task agent
func IsTaskAgent(ctx context.Context) bool {
	isTaskAgent := ctx.Value(IsTaskAgentContextKey)
	if isTaskAgent == nil {
		return false
	}
	if val, ok := isTaskAgent.(bool); ok {
		return val
	}
	return false
}

// GetAgentID returns the agent name from context, or empty string if not set
func GetAgentID(ctx context.Context) config.AgentName {
	agentName := ctx.Value(AgentIDContextKey)
	if agentName == nil {
		return ""
	}
	if val, ok := agentName.(config.AgentName); ok {
		return val
	}
	return ""
}
