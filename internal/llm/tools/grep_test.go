package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrepTool_Info(t *testing.T) {
	tool := NewGrepTool()
	info := tool.Info()

	assert.Equal(t, GrepToolName, info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Parameters, "pattern")
	assert.Contains(t, info.Parameters, "path")
	assert.Contains(t, info.Parameters, "include")
	assert.Contains(t, info.Parameters, "literal_text")
	assert.Contains(t, info.Required, "pattern")
}

func TestGrepTool_Run(t *testing.T) {
	tool := NewGrepTool()

	tests := []struct {
		name         string
		callInput    string
		assertResult func(*testing.T, ToolResponse)
	}{
		{
			name:      "returns error for empty pattern",
			callInput: `{}`,
			assertResult: func(t *testing.T, resp ToolResponse) {
				assert.Contains(t, resp.Content, "pattern is required")
			},
		},
		{
			name:      "handles valid search",
			callInput: `{"pattern":"test"}`,
			assertResult: func(t *testing.T, resp ToolResponse) {
				// Should return some content (possibly empty if no files found)
				assert.NotEmpty(t, resp.Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := ToolCall{
				Name:  GrepToolName,
				Input: tt.callInput,
			}

			response, err := tool.Run(context.Background(), call)
			require.NoError(t, err)
			tt.assertResult(t, response)
		})
	}
}

func TestGrepResponseMetadata(t *testing.T) {
	tests := []struct {
		name            string
		numberOfMatches int
		truncated       bool
		checkMetadata   func(*testing.T, string)
	}{
		{
			name:            "matches found",
			numberOfMatches: 10,
			truncated:       false,
			checkMetadata: func(t *testing.T, metadata string) {
				assert.Contains(t, metadata, "number_of_matches")
			},
		},
		{
			name:            "truncated results",
			numberOfMatches: 100,
			truncated:       true,
			checkMetadata: func(t *testing.T, metadata string) {
				assert.Contains(t, metadata, "truncated")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := WithResponseMetadata(
				NewTextResponse("test"),
				GrepResponseMetadata{
					NumberOfMatches: tt.numberOfMatches,
					Truncated:       tt.truncated,
				},
			)
			tt.checkMetadata(t, response.Metadata)
		})
	}
}

func TestGrepMatch(t *testing.T) {
	// Test grepMatch struct initialization
	match := grepMatch{
		path:     "/path/to/file.go",
		lineNum:  42,
		lineText: "func test()",
	}

	assert.Equal(t, "/path/to/file.go", match.path)
	assert.Equal(t, 42, match.lineNum)
	assert.Equal(t, "func test()", match.lineText)
}
