package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobTool_Info(t *testing.T) {
	tool := NewGlobTool()
	info := tool.Info()

	assert.Equal(t, GlobToolName, info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Parameters, "pattern")
	assert.Contains(t, info.Parameters, "path")
	assert.Contains(t, info.Required, "pattern")
}

func TestGlobTool_Run(t *testing.T) {
	tool := NewGlobTool()

	tests := []struct {
		name         string
		params       GlobParams
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
			name:      "handles valid pattern with files found",
			callInput: `{"pattern":"*.go"}`,
			assertResult: func(t *testing.T, resp ToolResponse) {
				// Should return some content (possibly empty if no files found)
				assert.NotEmpty(t, resp.Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := ToolCall{
				Name:  GlobToolName,
				Input: tt.callInput,
			}

			response, err := tool.Run(context.Background(), call)
			require.NoError(t, err)
			tt.assertResult(t, response)
		})
	}
}

func TestGlobResponseMetadata(t *testing.T) {
	tests := []struct {
		name          string
		numberOfFiles int
		truncated     bool
		checkMetadata func(*testing.T, string)
	}{
		{
			name:          "files found",
			numberOfFiles: 5,
			truncated:     false,
			checkMetadata: func(t *testing.T, metadata string) {
				assert.Contains(t, metadata, "number_of_files")
				assert.Contains(t, metadata, "5")
			},
		},
		{
			name:          "truncated results",
			numberOfFiles: 100,
			truncated:     true,
			checkMetadata: func(t *testing.T, metadata string) {
				assert.Contains(t, metadata, "truncated")
				assert.Contains(t, metadata, "true")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := WithResponseMetadata(
				NewTextResponse("test"),
				GlobResponseMetadata{
					NumberOfFiles: tt.numberOfFiles,
					Truncated:     tt.truncated,
				},
			)
			tt.checkMetadata(t, response.Metadata)
		})
	}
}

func TestEscapeRegexPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "dot escaped",
			input:    "test.value",
			expected: `test\.value`,
		},
		{
			name:     "plus escaped",
			input:    "test+value",
			expected: `test\+value`,
		},
		{
			name:     "star escaped",
			input:    "test*value",
			expected: `test\*value`,
		},
		{
			name:     "question escaped",
			input:    "test?value",
			expected: `test\?value`,
		},
		{
			name:     "parentheses escaped",
			input:    "test(value)",
			expected: `test\(value\)`,
		},
		{
			name:     "brackets escaped",
			input:    "test[value]",
			expected: `test\[value\]`,
		},
		{
			name:     "braces escaped",
			input:    "test{value}",
			expected: `test\{value\}`,
		},
		{
			name:     "caret escaped",
			input:    "test^value",
			expected: `test\^value`,
		},
		{
			name:     "dollar escaped",
			input:    "test$value",
			expected: `test\$value`,
		},
		{
			name:     "pipe escaped",
			input:    "test|value",
			expected: `test\|value`,
		},
		{
			name:     "backslash escaped",
			input:    `test\value`,
			expected: `test\\value`,
		},
		{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeRegexPattern(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple glob",
			input:    "*.txt",
			expected: ".*\\.txt",
		},
		{
			name:     "question mark",
			input:    "file?.txt",
			expected: "file.\\.txt",
		},
		{
			name:     "brace expansion single",
			input:    "*.{txt,md}",
			expected: ".*\\.(txt|md)",
		},
		{
			name:     "brace expansion multiple",
			input:    "*.{js,ts,jsx,tsx}",
			expected: ".*\\.(js|ts|jsx|tsx)",
		},
		{
			name:     "no glob characters",
			input:    "file.txt",
			expected: "file\\.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := globToRegex(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
