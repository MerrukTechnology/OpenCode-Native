package completions

import (
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/components/dialog"
	"github.com/stretchr/testify/assert"
)

func TestFilesAndFoldersContextGroup_GetId(t *testing.T) {
	cg := &filesAndFoldersContextGroup{
		prefix: "test_prefix",
	}

	assert.Equal(t, "test_prefix", cg.GetId())
}

func TestFilesAndFoldersContextGroup_GetEntry(t *testing.T) {
	cg := &filesAndFoldersContextGroup{
		prefix: "file",
	}

	entry := cg.GetEntry()
	assert.NotNil(t, entry)

	// Test that entry implements CompletionItemI
	_, ok := entry.(dialog.CompletionItemI)
	assert.True(t, ok)
}

func TestProcessNullTerminatedOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []string
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: []string{},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "single item no trailing null",
			input:    []byte("file1.txt"),
			expected: []string{"file1.txt"},
		},
		{
			name:     "multiple items null terminated",
			input:    []byte("file1.txt\x00file2.txt\x00file3.txt\x00"),
			expected: []string{"file1.txt", "file2.txt", "file3.txt"},
		},
		{
			name:     "trailing null at end",
			input:    []byte("file1.txt\x00"),
			expected: []string{"file1.txt"},
		},
		{
			name:     "empty items skipped",
			input:    []byte("file1.txt\x00\x00file2.txt"),
			expected: []string{"file1.txt", "file2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processNullTerminatedOutput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewFileAndFolderContextGroup(t *testing.T) {
	provider := NewFileAndFolderContextGroup()
	assert.NotNil(t, provider)

	cg, ok := provider.(*filesAndFoldersContextGroup)
	assert.True(t, ok)
	assert.Equal(t, "file", cg.prefix)
}
