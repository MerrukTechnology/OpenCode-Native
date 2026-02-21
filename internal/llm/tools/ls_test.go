package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mock_config "github.com/MerrukTechnology/OpenCode-Native/internal/config/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Test directory structure constants
const (
	testDir1   = "dir1"
	testDir2   = "dir2"
	testDir3   = "dir3"
	testSubDir = "dir2/subdir1"
)

var testDirs = []string{
	"dir1",
	"dir2",
	"dir2/subdir1",
	"dir2/subdir2",
	"dir3",
	"dir3/.hidden_dir",
	"__pycache__",
}

var testFiles = []string{
	"file1.txt",
	"file2.txt",
	"dir1/file3.txt",
	"dir2/file4.txt",
	"dir2/subdir1/file5.txt",
	"dir2/subdir2/file6.txt",
	"dir3/file7.txt",
	"dir3/.hidden_file.txt",
	"__pycache__/cache.pyc",
	".hidden_root_file.txt",
}

// setupTestDir creates a temporary directory with test files and directories.
func setupTestDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "ls_tool_test")
	require.NoError(t, err)

	// Create directories
	for _, dir := range testDirs {
		dirPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(dirPath, 0755)
		require.NoError(t, err)
	}

	// Create files
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	return tempDir
}

// newTestToolCall creates a ToolCall for the LS tool with given params.
func newTestToolCall(params LSParams) ToolCall {
	paramsJSON, _ := json.Marshal(params)
	return ToolCall{
		Name:  LSToolName,
		Input: string(paramsJSON),
	}
}

func TestLsTool_Info(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := mock_config.NewMockConfigurator(ctrl)
	config.EXPECT().WorkingDirectory().Times(0)

	tool := NewLsTool(config)
	info := tool.Info()

	assert.Equal(t, LSToolName, info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Parameters, "path")
	assert.Contains(t, info.Parameters, "ignore")
	assert.Contains(t, info.Required, "path")
}

func TestLsTool_Run(t *testing.T) {
	tempDir := setupTestDir(t)
	defer os.RemoveAll(tempDir)

	// Define test cases for TestLsTool_Run
	testCases := []struct {
		name         string
		path         string
		ignore       []string
		setupConfig  func(*mock_config.MockConfigurator)
		assertResult func(*testing.T, ToolResponse)
	}{
		{
			name: "lists directory successfully",
			path: tempDir,
			assertResult: func(t *testing.T, resp ToolResponse) {
				assert.Contains(t, resp.Content, "dir1")
				assert.Contains(t, resp.Content, "dir2")
				assert.Contains(t, resp.Content, "dir3")
				assert.Contains(t, resp.Content, "file1.txt")
				assert.Contains(t, resp.Content, "file2.txt")
				assert.NotContains(t, resp.Content, ".hidden_dir")
				assert.NotContains(t, resp.Content, ".hidden_file.txt")
				assert.NotContains(t, resp.Content, ".hidden_root_file.txt")
				assert.NotContains(t, resp.Content, "__pycache__")
			},
		},
		{
			name: "handles non-existent path",
			path: filepath.Join(tempDir, "non_existent_dir"),
			assertResult: func(t *testing.T, resp ToolResponse) {
				assert.Contains(t, resp.Content, "path does not exist")
			},
		},
		{
			name: "handles empty path parameter",
			path: "",
			setupConfig: func(m *mock_config.MockConfigurator) {
				m.EXPECT().WorkingDirectory().Return("").Times(2)
			},
			assertResult: func(t *testing.T, resp ToolResponse) {
				assert.NotEmpty(t, resp.Content)
			},
		},
		{
			name: "handles invalid parameters",
			path: "",
			setupConfig: func(m *mock_config.MockConfigurator) {
				m.EXPECT().WorkingDirectory().Times(0)
			},
			assertResult: func(t *testing.T, resp ToolResponse) {
				assert.Contains(t, resp.Content, "error parsing parameters")
			},
		},
		{
			name:   "respects ignore patterns",
			path:   tempDir,
			ignore: []string{"file1.txt", "dir1"},
			assertResult: func(t *testing.T, resp ToolResponse) {
				assert.NotContains(t, resp.Content, "- file1.txt")
				assert.NotContains(t, resp.Content, "- dir1/")
			},
		},
		{
			name: "handles relative path",
			path: "",
			setupConfig: func(m *mock_config.MockConfigurator) {
				m.EXPECT().WorkingDirectory().Times(1)
			},
			assertResult: func(t *testing.T, resp ToolResponse) {
				assert.Contains(t, resp.Content, "dir1")
				assert.Contains(t, resp.Content, "file1.txt")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCtrl := gomock.NewController(t)
			defer testCtrl.Finish()

			config := mock_config.NewMockConfigurator(testCtrl)

			// Handle relative path test specially
			if tc.name == "handles relative path" {
				origWd, err := os.Getwd()
				require.NoError(t, err)
				defer os.Chdir(origWd)

				parentDir := filepath.Dir(tempDir)
				err = os.Chdir(parentDir)
				require.NoError(t, err)

				tc.path = filepath.Base(tempDir)
			}

			if tc.setupConfig != nil {
				tc.setupConfig(config)
			} else {
				config.EXPECT().WorkingDirectory().Times(0)
			}

			tool := NewLsTool(config)

			var call ToolCall
			if tc.name == "handles invalid parameters" {
				call = ToolCall{
					Name:  LSToolName,
					Input: "invalid json",
				}
			} else {
				params := LSParams{
					Path:   tc.path,
					Ignore: tc.ignore,
				}
				call = newTestToolCall(params)
			}

			response, err := tool.Run(context.Background(), call)
			require.NoError(t, err)

			tc.assertResult(t, response)
		})
	}
}

func TestShouldSkip(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		ignorePatterns []string
		expected       bool
	}{
		{
			name:           "hidden file",
			path:           "/path/to/.hidden_file",
			ignorePatterns: []string{},
			expected:       true,
		},
		{
			name:           "hidden directory",
			path:           "/path/to/.hidden_dir",
			ignorePatterns: []string{},
			expected:       true,
		},
		{
			name:           "pycache directory",
			path:           "/path/to/__pycache__/file.pyc",
			ignorePatterns: []string{},
			expected:       true,
		},
		{
			name:           "node_modules directory",
			path:           "/path/to/node_modules/package",
			ignorePatterns: []string{},
			expected:       true,
		},
		{
			name:           "normal file",
			path:           "/path/to/normal_file.txt",
			ignorePatterns: []string{},
			expected:       false,
		},
		{
			name:           "normal directory",
			path:           "/path/to/normal_dir",
			ignorePatterns: []string{},
			expected:       false,
		},
		{
			name:           "ignored by pattern",
			path:           "/path/to/ignore_me.txt",
			ignorePatterns: []string{"ignore_*.txt"},
			expected:       true,
		},
		{
			name:           "not ignored by pattern",
			path:           "/path/to/keep_me.txt",
			ignorePatterns: []string{"ignore_*.txt"},
			expected:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := shouldSkip(tc.path, tc.ignorePatterns)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateFileTree(t *testing.T) {
	paths := []string{
		"/path/to/file1.txt",
		"/path/to/dir1/file2.txt",
		"/path/to/dir1/subdir/file3.txt",
		"/path/to/dir2/file4.txt",
	}

	tree := createFileTree(paths)

	// Check the structure of the tree
	assert.Len(t, tree, 1) // Should have one root node

	// Check the root node
	rootNode := tree[0]
	assert.Equal(t, "path", rootNode.Name)
	assert.Equal(t, "directory", rootNode.Type)
	assert.Len(t, rootNode.Children, 1)

	// Check the "to" node
	toNode := rootNode.Children[0]
	assert.Equal(t, "to", toNode.Name)
	assert.Equal(t, "directory", toNode.Type)
	assert.Len(t, toNode.Children, 3) // file1.txt, dir1, dir2

	// Find the dir1 node
	var dir1Node *TreeNode
	for _, child := range toNode.Children {
		if child.Name == "dir1" {
			dir1Node = child
			break
		}
	}

	require.NotNil(t, dir1Node)
	assert.Equal(t, "directory", dir1Node.Type)
	assert.Len(t, dir1Node.Children, 2) // file2.txt and subdir
}

func TestPrintTree(t *testing.T) {
	tree := []*TreeNode{
		{
			Name: "dir1",
			Path: "dir1",
			Type: "directory",
			Children: []*TreeNode{
				{
					Name: "file1.txt",
					Path: "dir1/file1.txt",
					Type: "file",
				},
				{
					Name: "subdir",
					Path: "dir1/subdir",
					Type: "directory",
					Children: []*TreeNode{
						{
							Name: "file2.txt",
							Path: "dir1/subdir/file2.txt",
							Type: "file",
						},
					},
				},
			},
		},
		{
			Name: "file3.txt",
			Path: "file3.txt",
			Type: "file",
		},
	}

	result := printTree(tree, "/root")

	// Check the output format
	assert.Contains(t, result, "- /root/")
	assert.Contains(t, result, "  - dir1/")
	assert.Contains(t, result, "    - file1.txt")
	assert.Contains(t, result, "    - subdir/")
	assert.Contains(t, result, "      - file2.txt")
	assert.Contains(t, result, "  - file3.txt")
}

// containsPath checks if any path in the slice starts with the target path.
func containsPath(paths []string, tempDir, target string) bool {
	targetPath := filepath.Join(tempDir, target)
	for _, path := range paths {
		if strings.HasPrefix(path, targetPath) {
			return true
		}
	}
	return false
}

func TestListDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "list_directory_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test directory structure (minimal set for this test)
	listTestDirs := []string{
		"dir1",
		"dir1/subdir1",
		".hidden_dir",
	}

	listTestFiles := []string{
		"file1.txt",
		"file2.txt",
		"dir1/file3.txt",
		"dir1/subdir1/file4.txt",
		".hidden_file.txt",
	}

	// Create directories
	for _, dir := range listTestDirs {
		dirPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(dirPath, 0755)
		require.NoError(t, err)
	}

	// Create files
	for _, file := range listTestFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	testCases := []struct {
		name        string
		ignorePats  []string
		limit       int
		checkTrunc  bool
		assertFiles func(*testing.T, []string)
	}{
		{
			name:       "lists files with no limit",
			ignorePats: []string{},
			limit:      1000,
			checkTrunc: false,
			assertFiles: func(t *testing.T, files []string) {
				assert.True(t, containsPath(files, tempDir, "dir1"))
				assert.True(t, containsPath(files, tempDir, "file1.txt"))
				assert.True(t, containsPath(files, tempDir, "file2.txt"))
				assert.True(t, containsPath(files, tempDir, "dir1/file3.txt"))
				assert.False(t, containsPath(files, tempDir, ".hidden_dir"))
				assert.False(t, containsPath(files, tempDir, ".hidden_file.txt"))
			},
		},
		{
			name:       "respects limit and returns truncated flag",
			ignorePats: []string{},
			limit:      2,
			checkTrunc: true,
			assertFiles: func(t *testing.T, files []string) {
				assert.Len(t, files, 2)
			},
		},
		{
			name:       "respects ignore patterns",
			ignorePats: []string{"*.txt"},
			limit:      1000,
			checkTrunc: false,
			assertFiles: func(t *testing.T, files []string) {
				for _, file := range files {
					assert.False(t, strings.HasSuffix(file, ".txt"), "Found .txt file: %s", file)
				}
				containsDir := false
				for _, file := range files {
					if strings.Contains(file, "dir1") {
						containsDir = true
						break
					}
				}
				assert.True(t, containsDir)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, truncated, err := listDirectory(tempDir, tc.ignorePats, tc.limit)
			require.NoError(t, err)

			if tc.checkTrunc {
				assert.True(t, truncated)
			} else {
				assert.False(t, truncated)
			}

			tc.assertFiles(t, files)
		})
	}
}
