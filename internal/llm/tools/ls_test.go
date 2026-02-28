package tools

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	mock_config "github.com/MerrukTechnology/OpenCode-Native/internal/config/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Create a test directory structure
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
		err := os.MkdirAll(dirPath, 0o755)
		require.NoError(t, err)
	}

	// Create files
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0o644)
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
		"dir1/file3.go",
		"dir1/subdir1/file4.txt",
		".hidden_file.txt",
	}

	// Create directories
	for _, dir := range listTestDirs {
		dirPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(dirPath, 0o755)
		require.NoError(t, err)
	}

	// Create files
	for _, file := range listTestFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0o644)
		require.NoError(t, err)
	}

	t.Run("lists files with no limit", func(t *testing.T) {
		files, truncated, err := listDirectory(context.Background(), tempDir, []string{}, 1000)
		require.NoError(t, err)
		assert.False(t, truncated)

		// Check that visible files and directories are included
		containsPath := func(paths []string, target string) bool {
			targetPath := filepath.Join(tempDir, target)
			for _, path := range paths {
				if strings.HasPrefix(path, targetPath) {
					return true
				}
			}
			return false
		}

		assert.True(t, containsPath(files, "dir1"))
		assert.True(t, containsPath(files, "file1.txt"))
		assert.True(t, containsPath(files, "file2.txt"))
		assert.True(t, containsPath(files, "dir1/file3.txt"))

		// Check that hidden files and directories are not included
		assert.False(t, containsPath(files, ".hidden_dir"))
		assert.False(t, containsPath(files, ".hidden_file.txt"))
	})

	t.Run("respects limit and returns truncated flag", func(t *testing.T) {
		files, truncated, err := listDirectory(context.Background(), tempDir, []string{}, 2)
		require.NoError(t, err)
		assert.True(t, truncated)
		assert.Len(t, files, 2)
	})

	t.Run("respects ignore patterns", func(t *testing.T) {
		files, truncated, err := listDirectory(context.Background(), tempDir, []string{"*.txt"}, 1000)
		require.NoError(t, err)
		assert.False(t, truncated)

		// Check that no .txt files are included
		for _, file := range files {
			assert.False(t, strings.HasSuffix(file, ".txt"), "Found .txt file: %s", file)
		}

		// But directories should still be included
		containsDir := false
		for _, file := range files {
			if strings.Contains(file, "dir1") {
				containsDir = true
				break
			}
		}
		assert.True(t, containsDir)
	})
}

func TestListDirectoryWithRipgrep(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep not installed, skipping ripgrep-specific tests")
	}

	tempDir, err := os.MkdirTemp("", "ls_rg_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize a git repo so .gitignore is respected
	gitInit := exec.Command("git", "init", tempDir)
	require.NoError(t, gitInit.Run())

	testDirs := []string{
		"src",
		"src/sub",
		"build",
	}
	testFiles := []string{
		"src/main.go",
		"src/sub/helper.go",
		"build/output.bin",
		"README.md",
	}

	for _, dir := range testDirs {
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, dir), 0o755))
	}
	for _, file := range testFiles {
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0o644))
	}

	// Write .gitignore that ignores build/
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte("build/\n"), 0o644))

	t.Run("respects gitignore", func(t *testing.T) {
		files, truncated, err := listDirectoryWithRipgrep(context.Background(), tempDir, nil, 1000)
		require.NoError(t, err)
		assert.False(t, truncated)

		containsPath := func(paths []string, target string) bool {
			targetPath := filepath.Join(tempDir, target)
			for _, p := range paths {
				if p == targetPath {
					return true
				}
			}
			return false
		}

		assert.True(t, containsPath(files, "src/main.go"))
		assert.True(t, containsPath(files, "src/sub/helper.go"))
		assert.True(t, containsPath(files, "README.md"))
		assert.False(t, containsPath(files, "build/output.bin"), "build/ should be excluded by .gitignore")
	})

	t.Run("user ignore patterns become glob flags", func(t *testing.T) {
		files, _, err := listDirectoryWithRipgrep(context.Background(), tempDir, []string{"*.md"}, 1000)
		require.NoError(t, err)

		for _, f := range files {
			assert.False(t, strings.HasSuffix(f, ".md"), "should have excluded .md files: %s", f)
		}
	})

	t.Run("truncation at limit returns lexicographically earliest entries", func(t *testing.T) {
		files, truncated, err := listDirectoryWithRipgrep(context.Background(), tempDir, nil, 2)
		require.NoError(t, err)
		assert.True(t, truncated)
		assert.Len(t, files, 2)
		assert.True(t, sort.StringsAreSorted(files), "truncated results should be sorted")

		// The 2 returned files must be the lexicographically smallest from the full set
		allFiles, _, err := listDirectoryWithRipgrep(context.Background(), tempDir, nil, 1000)
		require.NoError(t, err)
		sort.Strings(allFiles)
		assert.Equal(t, allFiles[:2], files, "truncated results should be the first entries in sorted order")
	})

	t.Run("empty directory returns empty results", func(t *testing.T) {
		emptyDir, err := os.MkdirTemp("", "ls_rg_empty")
		require.NoError(t, err)
		defer os.RemoveAll(emptyDir)

		files, truncated, err := listDirectoryWithRipgrep(context.Background(), emptyDir, nil, 1000)
		require.NoError(t, err)
		assert.False(t, truncated)
		assert.Empty(t, files)
	})

	t.Run("output is sorted", func(t *testing.T) {
		files, _, err := listDirectoryWithRipgrep(context.Background(), tempDir, nil, 1000)
		require.NoError(t, err)
		assert.True(t, sort.StringsAreSorted(files), "ripgrep output should be sorted")
	})
}

func TestListDirectoryWithWalk(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ls_walk_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testDirs := []string{
		"src",
		"node_modules",
		"__pycache__",
		".hidden",
	}
	testFiles := []string{
		"src/main.go",
		"node_modules/pkg.js",
		"__pycache__/mod.pyc",
		".hidden/secret.txt",
		"readme.txt",
	}

	for _, dir := range testDirs {
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, dir), 0o755))
	}
	for _, file := range testFiles {
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0o644))
	}

	t.Run("commonIgnored list is applied in walk fallback", func(t *testing.T) {
		files, _, err := listDirectoryWithWalk(tempDir, nil, 1000)
		require.NoError(t, err)

		containsPath := func(paths []string, substr string) bool {
			for _, p := range paths {
				// Handle both absolute and relative paths
				normalizedPath := filepath.ToSlash(p)
				if strings.Contains(normalizedPath, substr) {
					return true
				}
			}
			return false
		}

		// Check that expected files are present (using relative paths for matching)
		// The function returns absolute paths, but we check for the relative path substring
		assert.True(t, containsPath(files, "src/main.go"), "Expected src/main.go in results: %v", files)
		assert.True(t, containsPath(files, "readme.txt"), "Expected readme.txt in results: %v", files)
		assert.False(t, containsPath(files, "node_modules"), "node_modules should be skipped by commonIgnored")
		assert.False(t, containsPath(files, "__pycache__"), "__pycache__ should be skipped by commonIgnored")
		assert.False(t, containsPath(files, ".hidden"), "hidden dirs should be skipped")
	})
}
