package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	mock_permission "github.com/MerrukTechnology/OpenCode-Native/internal/permission/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupDeleteTest(t *testing.T) (context.Context, BaseTool, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockPerms := mock_permission.NewMockService(ctrl)
	mockPerms.EXPECT().Request(gomock.Any()).Return(true).AnyTimes()

	files := &stubHistoryService{}
	tool := NewDeleteTool(mockPerms, files, &stubRegistry{})

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session")
	ctx = context.WithValue(ctx, MessageIDContextKey, "test-message")

	return ctx, tool, ctrl
}

func runDelete(t *testing.T, tool BaseTool, ctx context.Context, params DeleteParams) ToolResponse {
	t.Helper()
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)
	resp, err := tool.Run(ctx, ToolCall{Name: DeleteToolName, Input: string(paramsJSON)})
	require.NoError(t, err)
	return resp
}

func runDeleteRaw(t *testing.T, tool BaseTool, ctx context.Context, input string) ToolResponse {
	t.Helper()
	resp, err := tool.Run(ctx, ToolCall{Name: DeleteToolName, Input: input})
	require.NoError(t, err)
	return resp
}

func createTempFileInWorkingDir(t *testing.T, pattern string) string {
	t.Helper()
	workingDir := config.WorkingDirectory()
	tmpFile, err := os.CreateTemp(workingDir, pattern)
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	t.Cleanup(func() {
		os.Remove(tmpPath)
	})
	return tmpPath
}

func createTempDirInWorkingDir(t *testing.T, pattern string) string {
	t.Helper()
	workingDir := config.WorkingDirectory()
	tmpDir, err := os.MkdirTemp(workingDir, pattern)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

// Helper to assert successful deletion response
func assertDeleteSuccess(t *testing.T, resp ToolResponse, path string, expectedFilesDeleted int) {
	t.Helper()
	assert.False(t, resp.IsError, "Expected no error, got: %s", resp.Content)
	assert.Contains(t, resp.Content, "successfully deleted")
	assert.Contains(t, resp.Content, path)

	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), "Path should be deleted")

	if expectedFilesDeleted >= 0 {
		var metadata DeleteResponseMetadata
		err = json.Unmarshal([]byte(resp.Metadata), &metadata)
		require.NoError(t, err)
		assert.Equal(t, expectedFilesDeleted, metadata.FilesDeleted)
	}
}

// Helper to assert error response
func assertDeleteError(t *testing.T, resp ToolResponse, expectedContent string) {
	t.Helper()
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, expectedContent)
}

func TestDeleteTool_Info(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPerms := mock_permission.NewMockService(ctrl)
	tool := NewDeleteTool(mockPerms, &stubHistoryService{}, &stubRegistry{})
	info := tool.Info()

	assert.Equal(t, DeleteToolName, info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Parameters, "path")
	assert.Contains(t, info.Required, "path")
}

func TestDeleteTool_DeleteFile(t *testing.T) {
	ctx, tool, ctrl := setupDeleteTest(t)
	defer ctrl.Finish()

	tmpPath := createTempFileInWorkingDir(t, "delete_test_*.txt")
	content := "test file content"
	require.NoError(t, os.WriteFile(tmpPath, []byte(content), 0644))

	resp := runDelete(t, tool, ctx, DeleteParams{Path: tmpPath})

	assertDeleteSuccess(t, resp, tmpPath, 1)

	var metadata DeleteResponseMetadata
	err := json.Unmarshal([]byte(resp.Metadata), &metadata)
	require.NoError(t, err)
	assert.Greater(t, metadata.Removals, 0)
	assert.NotEmpty(t, metadata.Diff)
}

func TestDeleteTool_DeleteDirectory(t *testing.T) {
	ctx, tool, ctrl := setupDeleteTest(t)
	defer ctrl.Finish()

	tmpDir := createTempDirInWorkingDir(t, "delete_test_dir_*")

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	subDir := filepath.Join(tmpDir, "subdir")
	file3 := filepath.Join(subDir, "file3.txt")

	require.NoError(t, os.WriteFile(file1, []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content2"), 0644))
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(file3, []byte("content3"), 0644))

	resp := runDelete(t, tool, ctx, DeleteParams{Path: tmpDir})

	assert.False(t, resp.IsError, "Expected no error, got: %s", resp.Content)
	assert.Contains(t, resp.Content, "successfully deleted")
	assert.Contains(t, resp.Content, tmpDir)
	assert.Contains(t, resp.Content, "3 files")

	_, err := os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err), "Directory should be deleted")

	var metadata DeleteResponseMetadata
	err = json.Unmarshal([]byte(resp.Metadata), &metadata)
	require.NoError(t, err)
	assert.Equal(t, 3, metadata.FilesDeleted)
	assert.Greater(t, metadata.Removals, 0)
}

func TestDeleteTool_ErrorCases(t *testing.T) {
	ctx, tool, ctrl := setupDeleteTest(t)
	defer ctrl.Finish()

	workingDir := config.WorkingDirectory()

	tests := []struct {
		name             string
		params           DeleteParams
		expectedError    string
		prepare          func(t *testing.T) string // returns path for verification
		verifyNotDeleted func(t *testing.T, path string)
	}{
		{
			name:          "empty path",
			params:        DeleteParams{Path: ""},
			expectedError: "path is required",
		},
		{
			name:          "non-existent path",
			params:        DeleteParams{Path: filepath.Join(workingDir, "nonexistent_file_12345.txt")},
			expectedError: "does not exist",
		},
		{
			name:   "path outside working directory",
			params: DeleteParams{Path: "/tmp/outside_test_12345.txt"},
			prepare: func(t *testing.T) string {
				tmpFile, err := os.CreateTemp("/tmp", "outside_test_*.txt")
				require.NoError(t, err)
				tmpPath := tmpFile.Name()
				tmpFile.Close()
				t.Cleanup(func() { os.Remove(tmpPath) })
				return tmpPath
			},
			expectedError: "outside the working directory",
			verifyNotDeleted: func(t *testing.T, path string) {
				_, err := os.Stat(path)
				assert.False(t, os.IsNotExist(err), "File should NOT be deleted")
			},
		},
		{
			name:          "invalid JSON",
			params:        DeleteParams{Path: "test.txt"},
			expectedError: "error parsing parameters",
			// Override to use raw input
			prepare: func(t *testing.T) string { return "" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp ToolResponse

			if tt.name == "invalid JSON" {
				resp = runDeleteRaw(t, tool, ctx, "invalid json")
			} else {
				var verifyPath string
				if tt.prepare != nil {
					verifyPath = tt.prepare(t)
				}
				resp = runDelete(t, tool, ctx, tt.params)

				if tt.verifyNotDeleted != nil && verifyPath != "" {
					tt.verifyNotDeleted(t, verifyPath)
				}
			}

			assertDeleteError(t, resp, tt.expectedError)
		})
	}
}

func TestDeleteTool_DeleteSymlink(t *testing.T) {
	ctx, tool, ctrl := setupDeleteTest(t)
	defer ctrl.Finish()

	tmpDir := createTempDirInWorkingDir(t, "delete_symlink_test_*")

	targetFile := filepath.Join(tmpDir, "target.txt")
	require.NoError(t, os.WriteFile(targetFile, []byte("target content"), 0644))

	symlinkPath := filepath.Join(tmpDir, "symlink.txt")
	require.NoError(t, os.Symlink(targetFile, symlinkPath))

	resp := runDelete(t, tool, ctx, DeleteParams{Path: symlinkPath})

	assert.False(t, resp.IsError, "Expected no error, got: %s", resp.Content)
	assert.Contains(t, resp.Content, "successfully deleted")

	_, err := os.Lstat(symlinkPath)
	assert.True(t, os.IsNotExist(err), "Symlink should be deleted")

	_, err = os.Stat(targetFile)
	assert.False(t, os.IsNotExist(err), "Target file should still exist")

	content, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	assert.Equal(t, "target content", string(content))
}

func TestDeleteTool_RelativePath(t *testing.T) {
	ctx, tool, ctrl := setupDeleteTest(t)
	defer ctrl.Finish()

	workingDir := config.WorkingDirectory()
	tmpFile := filepath.Join(workingDir, "delete_relative_test.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("content"), 0644))
	t.Cleanup(func() {
		os.Remove(tmpFile)
	})

	resp := runDelete(t, tool, ctx, DeleteParams{Path: "delete_relative_test.txt"})

	assert.False(t, resp.IsError, "Expected no error, got: %s", resp.Content)

	_, err := os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err), "File should be deleted")
}

func TestDeleteTool_DirectoryWithManyFiles(t *testing.T) {
	ctx, tool, ctrl := setupDeleteTest(t)
	defer ctrl.Finish()

	tmpDir := createTempDirInWorkingDir(t, "delete_many_files_test_*")

	for i := 0; i < 10; i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))
	}

	resp := runDelete(t, tool, ctx, DeleteParams{Path: tmpDir})

	assert.False(t, resp.IsError, "Expected no error, got: %s", resp.Content)
	assert.Contains(t, resp.Content, "10 files")

	_, err := os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err), "Directory should be deleted")
}

func TestDeleteTool_DirectoryWithSymlinks(t *testing.T) {
	ctx, tool, ctrl := setupDeleteTest(t)
	defer ctrl.Finish()

	tmpDir := createTempDirInWorkingDir(t, "delete_symlinks_test_*")

	targetFile := filepath.Join(tmpDir, "target.txt")
	require.NoError(t, os.WriteFile(targetFile, []byte("target"), 0644))

	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	symlinkPath := filepath.Join(subDir, "symlink.txt")
	require.NoError(t, os.Symlink(targetFile, symlinkPath))

	regularFile := filepath.Join(subDir, "regular.txt")
	require.NoError(t, os.WriteFile(regularFile, []byte("regular"), 0644))

	resp := runDelete(t, tool, ctx, DeleteParams{Path: tmpDir})

	assert.False(t, resp.IsError, "Expected no error, got: %s", resp.Content)

	_, err := os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err), "Directory should be deleted")

	var metadata DeleteResponseMetadata
	err = json.Unmarshal([]byte(resp.Metadata), &metadata)
	require.NoError(t, err)
	assert.Equal(t, 2, metadata.FilesDeleted, "Should count regular files only, not symlinks")
}

func TestDeleteTool_EmptyDirectory(t *testing.T) {
	ctx, tool, ctrl := setupDeleteTest(t)
	defer ctrl.Finish()

	tmpDir := createTempDirInWorkingDir(t, "delete_empty_dir_test_*")

	resp := runDelete(t, tool, ctx, DeleteParams{Path: tmpDir})

	assert.False(t, resp.IsError, "Expected no error, got: %s", resp.Content)
	assert.Contains(t, resp.Content, "0 files")

	_, err := os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err), "Directory should be deleted")

	var metadata DeleteResponseMetadata
	err = json.Unmarshal([]byte(resp.Metadata), &metadata)
	require.NoError(t, err)
	assert.Equal(t, 0, metadata.FilesDeleted)
}
