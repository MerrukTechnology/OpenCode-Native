package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchTool_Info(t *testing.T) {
	// Create a mock permission service
	permService := permission.NewPermissionService()
	tool := NewFetchTool(permService)
	info := tool.Info()

	assert.Equal(t, FetchToolName, info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Parameters, "url")
	assert.Contains(t, info.Parameters, "format")
	assert.Contains(t, info.Parameters, "timeout")
	assert.Contains(t, info.Required, "url")
}

func TestFetchTool_Run(t *testing.T) {
	// Create a mock permission service
	permService := permission.NewPermissionService()
	tool := NewFetchTool(permService)

	tests := []struct {
		name         string
		params       FetchParams
		call         ToolCall
		assertResult func(*testing.T, ToolResponse, error)
	}{
		{
			name: "returns error for empty URL",
			params: FetchParams{
				URL:    "",
				Format: "text",
			},
			assertResult: func(t *testing.T, resp ToolResponse, err error) {
				assert.Contains(t, resp.Content, "URL parameter is required")
			},
		},
		{
			name: "returns error for invalid format",
			params: FetchParams{
				URL:    "https://example.com",
				Format: "invalid",
			},
			assertResult: func(t *testing.T, resp ToolResponse, err error) {
				assert.Contains(t, resp.Content, "Format must be one of")
			},
		},
		{
			name: "returns error for invalid URL protocol",
			params: FetchParams{
				URL:    "ftp://example.com",
				Format: "text",
			},
			assertResult: func(t *testing.T, resp ToolResponse, err error) {
				assert.Contains(t, resp.Content, "must start with http:// or https://")
			},
		},
		{
			name: "handles invalid JSON parameters",
			call: ToolCall{
				Name:  FetchToolName,
				Input: "invalid json",
			},
			assertResult: func(t *testing.T, resp ToolResponse, err error) {
				assert.Contains(t, resp.Content, "Failed to parse fetch parameters")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var call ToolCall

			if tt.call.Input != "" {
				call = tt.call
			} else {
				paramsJSON, err := json.Marshal(tt.params)
				require.NoError(t, err)
				call = ToolCall{
					Name:  FetchToolName,
					Input: string(paramsJSON),
				}
			}

			// Add required context values
			ctx := context.Background()
			ctx = context.WithValue(ctx, SessionIDContextKey, "test-session")
			ctx = context.WithValue(ctx, MessageIDContextKey, "test-message")

			response, err := tool.Run(ctx, call)
			require.NoError(t, err)
			tt.assertResult(t, response, err)
		})
	}
}

func TestFetchParams(t *testing.T) {
	tests := []struct {
		name   string
		params FetchParams
		check  func(*testing.T, FetchParams)
	}{
		{
			name: "valid params",
			params: FetchParams{
				URL:     "https://example.com",
				Format:  "text",
				Timeout: 30,
			},
			check: func(t *testing.T, params FetchParams) {
				assert.Equal(t, "https://example.com", params.URL)
				assert.Equal(t, "text", params.Format)
				assert.Equal(t, 30, params.Timeout)
			},
		},
		{
			name: "default timeout",
			params: FetchParams{
				URL:    "https://example.com",
				Format: "markdown",
			},
			check: func(t *testing.T, params FetchParams) {
				assert.Equal(t, 0, params.Timeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.params)
		})
	}
}

func TestFetchPermissionsParams(t *testing.T) {
	params := FetchPermissionsParams{
		URL:     "https://example.com",
		Format:  "html",
		Timeout: 60,
	}

	assert.Equal(t, "https://example.com", params.URL)
	assert.Equal(t, "html", params.Format)
	assert.Equal(t, 60, params.Timeout)
}

func TestExtractTextFromHTML(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple HTML",
			html:     "<html><body><p>Hello World</p></body></html>",
			expected: "Hello World",
		},
		{
			name:     "HTML with multiple elements",
			html:     "<html><head><title>Test</title></head><body><h1>Title</h1><p>Paragraph</p></body></html>",
			expected: "TestTitleParagraph",
		},
		{
			name:     "HTML with whitespace",
			html:     "<html><body>   <p>Text   with   spaces</p>   </body></html>",
			expected: "Text with spaces",
		},
		{
			name:     "empty HTML",
			html:     "",
			expected: "",
		},
		{
			name:     "plain text",
			html:     "Just plain text",
			expected: "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractTextFromHTML(tt.html)
			if tt.name == "empty HTML" {
				assert.NoError(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
