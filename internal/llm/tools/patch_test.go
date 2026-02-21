package tools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchTool_Info(t *testing.T) {
	tool := NewPatchTool(nil, nil, nil, nil)
	info := tool.Info()

	assert.Equal(t, PatchToolName, info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Parameters, "patch_text")
	assert.Contains(t, info.Required, "patch_text")
}

func TestPatchTool_Run(t *testing.T) {
	tool := NewPatchTool(nil, nil, nil, nil)

	tests := []struct {
		name         string
		params       PatchParams
		call         ToolCall
		assertResult func(*testing.T, ToolResponse, error)
	}{
		{
			name: "returns error for empty patch_text",
			params: PatchParams{
				PatchText: "",
			},
			assertResult: func(t *testing.T, resp ToolResponse, err error) {
				assert.Contains(t, resp.Content, "patch_text is required")
			},
		},
		{
			name:   "handles invalid JSON parameters",
			params: PatchParams{},
			assertResult: func(t *testing.T, resp ToolResponse, err error) {
				assert.Contains(t, resp.Content, "invalid parameters")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var call ToolCall

			if tt.name == "returns error for empty patch_text" {
				call = ToolCall{
					Name:  PatchToolName,
					Input: `{"patch_text": ""}`,
				}
			} else if tt.name == "handles invalid JSON parameters" {
				call = ToolCall{
					Name:  PatchToolName,
					Input: "invalid json",
				}
			}

			response, err := tool.Run(nil, call)
			require.NoError(t, err)
			tt.assertResult(t, response, err)
		})
	}
}

func TestPatchResponseMetadata(t *testing.T) {
	tests := []struct {
		name          string
		metadata      PatchResponseMetadata
		checkMetadata func(*testing.T, string)
	}{
		{
			name: "single file changed",
			metadata: PatchResponseMetadata{
				FilesChanged: []string{"/path/to/file.go"},
				Additions:    10,
				Removals:     5,
			},
			checkMetadata: func(t *testing.T, metadata string) {
				assert.Contains(t, metadata, "files_changed")
				assert.Contains(t, metadata, "10")
				assert.Contains(t, metadata, "5")
			},
		},
		{
			name: "multiple files changed",
			metadata: PatchResponseMetadata{
				FilesChanged: []string{"/path/to/file1.go", "/path/to/file2.go"},
				Additions:    20,
				Removals:     15,
			},
			checkMetadata: func(t *testing.T, metadata string) {
				assert.Contains(t, metadata, "file1.go")
				assert.Contains(t, metadata, "file2.go")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := WithResponseMetadata(
				NewTextResponse("test"),
				tt.metadata,
			)
			tt.checkMetadata(t, response.Metadata)
		})
	}
}

func TestPatchParams(t *testing.T) {
	tests := []struct {
		name   string
		params PatchParams
		check  func(*testing.T, PatchParams)
	}{
		{
			name: "valid params with patch text",
			params: PatchParams{
				PatchText: "*** Begin Patch\n*** Update File: /test.go\n@@\n-old\n+new\n*** End Patch",
			},
			check: func(t *testing.T, params PatchParams) {
				assert.NotEmpty(t, params.PatchText)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paramsJSON, err := json.Marshal(tt.params)
			require.NoError(t, err)

			var params PatchParams
			err = json.Unmarshal(paramsJSON, &params)
			require.NoError(t, err)

			tt.check(t, params)
		})
	}
}
