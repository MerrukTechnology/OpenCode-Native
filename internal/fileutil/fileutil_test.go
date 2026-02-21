package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolvePath(t *testing.T) {
	workingDir := "/home/user/project"

	tests := []struct {
		name       string
		path       string
		workingDir string
		expected   string
	}{
		{
			name:       "absolute path stays absolute",
			path:       "/absolute/path/file.go",
			workingDir: workingDir,
			expected:   "/absolute/path/file.go",
		},
		{
			name:       "relative path is resolved",
			path:       "src/file.go",
			workingDir: workingDir,
			expected:   "/home/user/project/src/file.go",
		},
		{
			name:       "empty working dir",
			path:       "file.go",
			workingDir: "",
			expected:   "file.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolvePath(tt.path, tt.workingDir)
			if result != tt.expected {
				t.Errorf("ResolvePath(%q, %q) = %q, want %q", tt.path, tt.workingDir, result, tt.expected)
			}
		})
	}
}

func TestIsInWorkingDir(t *testing.T) {
	workingDir := "/home/user/project"

	tests := []struct {
		name       string
		path       string
		workingDir string
		expected   bool
	}{
		{
			name:       "file in working dir",
			path:       "/home/user/project/src/main.go",
			workingDir: workingDir,
			expected:   true,
		},
		{
			name:       "file outside working dir",
			path:       "/etc/passwd",
			workingDir: workingDir,
			expected:   false,
		},
		{
			name:       "same as working dir",
			path:       "/home/user/project",
			workingDir: workingDir,
			expected:   true,
		},
		{
			name:       "sibling directory",
			path:       "/home/user/other",
			workingDir: workingDir,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInWorkingDir(tt.path, tt.workingDir)
			if result != tt.expected {
				t.Errorf("IsInWorkingDir(%q, %q) = %v, want %v", tt.path, tt.workingDir, result, tt.expected)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "existing file",
			path:     tmpFile,
			expected: true,
		},
		{
			name:     "non-existing file",
			path:     filepath.Join(tmpDir, "nonexistent.txt"),
			expected: false,
		},
		{
			name:     "directory path",
			path:     tmpDir,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FileExists(tt.path)
			if result != tt.expected {
				t.Errorf("FileExists(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "existing directory",
			path:     tmpDir,
			expected: true,
		},
		{
			name:     "non-existing directory",
			path:     filepath.Join(tmpDir, "nonexistent"),
			expected: false,
		},
		{
			name:     "file path",
			path:     tmpFile,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DirExists(tt.path)
			if result != tt.expected {
				t.Errorf("DirExists(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsHiddenFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "hidden file",
			path:     "/path/to/.gitignore",
			expected: true,
		},
		{
			name:     "visible file",
			path:     "/path/to/main.go",
			expected: false,
		},
		{
			name:     "dot as filename",
			path:     "/path/to/.",
			expected: false,
		},
		{
			name:     "double dot file",
			path:     "/path/to/..gitkeep",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHiddenFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsHiddenFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsIgnoredDir(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "node_modules",
			path:     "node_modules",
			expected: true,
		},
		{
			name:     "git directory",
			path:     ".git",
			expected: true,
		},
		{
			name:     "pycache",
			path:     "__pycache__",
			expected: true,
		},
		{
			name:     "regular dir",
			path:     "src",
			expected: false,
		},
		{
			name:     "full path",
			path:     "/path/to/node_modules",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIgnoredDir(tt.path)
			if result != tt.expected {
				t.Errorf("IsIgnoredDir(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsIgnoredExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "pyc file",
			path:     "test.pyc",
			expected: true,
		},
		{
			name:     "so file",
			path:     "test.so",
			expected: true,
		},
		{
			name:     "go file",
			path:     "main.go",
			expected: false,
		},
		{
			name:     "js file",
			path:     "app.js",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIgnoredExtension(tt.path)
			if result != tt.expected {
				t.Errorf("IsIgnoredExtension(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestSkipHidden(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "hidden file",
			path:     "/path/.hidden",
			expected: true,
		},
		{
			name:     "visible file",
			path:     "/path/to/file.txt",
			expected: false,
		},
		{
			name:     "node_modules in path",
			path:     "/path/node_modules/package",
			expected: true,
		},
		{
			name:     "git directory in path",
			path:     "/project/.git/config",
			expected: true,
		},
		{
			name:     "pycache in path",
			path:     "/project/src/__pycache__/module.pyc",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SkipHidden(tt.path)
			if result != tt.expected {
				t.Errorf("SkipHidden(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		ignorePatterns []string
		expected       bool
	}{
		{
			name:           "hidden file",
			path:           "/path/.hidden",
			ignorePatterns: []string{},
			expected:       true,
		},
		{
			name:           "ignored directory",
			path:           "/path/node_modules",
			ignorePatterns: []string{},
			expected:       true,
		},
		{
			name:           "custom pattern match",
			path:           "/path/test_temp.txt",
			ignorePatterns: []string{"**/*_temp.txt"},
			expected:       true,
		},
		{
			name:           "visible file no patterns",
			path:           "/path/to/file.txt",
			ignorePatterns: []string{},
			expected:       false,
		},
		{
			name:           "custom pattern no match",
			path:           "/path/to/real.txt",
			ignorePatterns: []string{"*_temp.txt"},
			expected:       false,
		},
		{
			name:           "ignored extension",
			path:           "/path/to/test.pyc",
			ignorePatterns: []string{},
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldSkipPath(tt.path, tt.ignorePatterns)
			if result != tt.expected {
				t.Errorf("ShouldSkipPath(%q, %v) = %v, want %v", tt.path, tt.ignorePatterns, result, tt.expected)
			}
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedOK   bool
		expectedType string
	}{
		{
			name:         "jpg file",
			path:         "photo.jpg",
			expectedOK:   true,
			expectedType: "JPEG",
		},
		{
			name:         "png file",
			path:         "image.PNG",
			expectedOK:   true,
			expectedType: "PNG",
		},
		{
			name:         "gif file",
			path:         "anim.gif",
			expectedOK:   true,
			expectedType: "GIF",
		},
		{
			name:         "svg file",
			path:         "vector.svg",
			expectedOK:   true,
			expectedType: "SVG",
		},
		{
			name:         "go file",
			path:         "main.go",
			expectedOK:   false,
			expectedType: "",
		},
		{
			name:         "txt file",
			path:         "readme.txt",
			expectedOK:   false,
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, imgType := IsImageFile(tt.path)
			if ok != tt.expectedOK {
				t.Errorf("IsImageFile(%q) ok = %v, want %v", tt.path, ok, tt.expectedOK)
			}
			if imgType != tt.expectedType {
				t.Errorf("IsImageFile(%q) type = %q, want %q", tt.path, imgType, tt.expectedType)
			}
		})
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		name     string
		glob     string
		expected string
	}{
		{
			name:     "simple glob",
			glob:     "*.go",
			expected: ".*\\.go",
		},
		{
			name:     "question mark",
			glob:     "file?.txt",
			expected: "file.\\.txt",
		},
		{
			name:     "character class",
			glob:     "file[123].txt",
			expected: "file[123]\\.txt",
		},
		{
			name:     "multiple extensions",
			glob:     "*.{js,ts,tsx}",
			expected: ".*\\.(js|ts|tsx)",
		},
		{
			name:     "literal dot",
			glob:     "file.name.go",
			expected: "file\\.name\\.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GlobToRegex(tt.glob)
			if result != tt.expected {
				t.Errorf("GlobToRegex(%q) = %q, want %q", tt.glob, result, tt.expected)
			}
		})
	}
}

func TestEscapeRegexPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "simple text",
			pattern:  "hello",
			expected: "hello",
		},
		{
			name:     "special chars",
			pattern:  "test.file+",
			expected: "test\\.file\\+",
		},
		{
			name:     "brackets",
			pattern:  "test[1]",
			expected: "test\\[1\\]",
		},
		{
			name:     "all special",
			pattern:  "\\.+*?()[]{}|^$",
			expected: "\\\\\\.\\+\\*\\?\\(\\)\\[\\]\\{\\}\\|\\^\\$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeRegexPattern(tt.pattern)
			if result != tt.expected {
				t.Errorf("EscapeRegexPattern(%q) = %q, want %q", tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestValidateFileForRead(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("failed to stat temp file: %v", err)
	}
	initialModTime := info.ModTime()

	tests := []struct {
		name           string
		path           string
		workingDir     string
		lastReadTime   time.Time
		expectError    bool
		expectExists   bool
		expectModified bool
	}{
		{
			name:         "valid file with zero last read time",
			path:         tmpFile,
			workingDir:   tmpDir,
			lastReadTime: time.Time{},
			expectError:  false,
			expectExists: true,
		},
		{
			name:         "valid file with recent last read time",
			path:         tmpFile,
			workingDir:   tmpDir,
			lastReadTime: time.Now(),
			expectError:  false,
			expectExists: true,
		},
		{
			name:         "file outside working directory",
			path:         "/etc/passwd",
			workingDir:   tmpDir,
			lastReadTime: time.Time{},
			expectError:  true,
			expectExists: false,
		},
		{
			name:         "non-existent file",
			path:         filepath.Join(tmpDir, "nonexistent.txt"),
			workingDir:   tmpDir,
			lastReadTime: time.Time{},
			expectError:  true,
			expectExists: false,
		},
		{
			name:         "directory instead of file",
			path:         tmpDir,
			workingDir:   tmpDir,
			lastReadTime: time.Time{},
			expectError:  true,
			expectExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFileForRead(tt.path, tt.workingDir, tt.lastReadTime)

			if tt.expectError && result.Error == nil {
				t.Errorf("ValidateFileForRead(%q, %q) expected error but got none", tt.path, tt.workingDir)
			}
			if !tt.expectError && result.Error != nil {
				t.Errorf("ValidateFileForRead(%q, %q) unexpected error: %v", tt.path, tt.workingDir, result.Error)
			}
			if result.Exists != tt.expectExists {
				t.Errorf("ValidateFileForRead(%q, %q).Exists = %v, want %v", tt.path, tt.workingDir, result.Exists, tt.expectExists)
			}
		})
	}

	t.Run("file modified since last read", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond)

		if err := os.WriteFile(tmpFile, []byte("modified content"), 0644); err != nil {
			t.Fatalf("failed to modify temp file: %v", err)
		}

		result := ValidateFileForRead(tmpFile, tmpDir, initialModTime)
		if !result.IsModified {
			t.Errorf("Expected file to be marked as modified")
		}
		if result.Error == nil {
			t.Errorf("Expected error for modified file")
		}
	})
}

func TestValidateFileForWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name         string
		path         string
		workingDir   string
		lastReadTime time.Time
		expectError  bool
		expectExists bool
	}{
		{
			name:         "valid file with zero last read time",
			path:         tmpFile,
			workingDir:   tmpDir,
			lastReadTime: time.Time{},
			expectError:  false,
			expectExists: true,
		},
		{
			name:         "new file",
			path:         filepath.Join(tmpDir, "newfile.txt"),
			workingDir:   tmpDir,
			lastReadTime: time.Time{},
			expectError:  false,
			expectExists: false,
		},
		{
			name:         "directory instead of file",
			path:         tmpDir,
			workingDir:   tmpDir,
			lastReadTime: time.Time{},
			expectError:  true,
			expectExists: true,
		},
		{
			name:         "file outside working directory",
			path:         "/etc/passwd",
			workingDir:   tmpDir,
			lastReadTime: time.Time{},
			expectError:  true,
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFileForWrite(tt.path, tt.workingDir, tt.lastReadTime)

			if tt.expectError && result.Error == nil {
				t.Errorf("ValidateFileForWrite(%q, %q) expected error but got none", tt.path, tt.workingDir)
			}
			if !tt.expectError && result.Error != nil {
				t.Errorf("ValidateFileForWrite(%q, %q) unexpected error: %v", tt.path, tt.workingDir, result.Error)
			}
			if result.Exists != tt.expectExists {
				t.Errorf("ValidateFileForWrite(%q, %q).Exists = %v, want %v", tt.path, tt.workingDir, result.Exists, tt.expectExists)
			}
		})
	}
}

func TestValidateFileForDelete(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name         string
		path         string
		workingDir   string
		expectError  bool
		expectExists bool
	}{
		{
			name:         "existing file",
			path:         tmpFile,
			workingDir:   tmpDir,
			expectError:  false,
			expectExists: true,
		},
		{
			name:         "existing directory",
			path:         tmpDir,
			workingDir:   tmpDir,
			expectError:  false,
			expectExists: true,
		},
		{
			name:         "non-existent path",
			path:         filepath.Join(tmpDir, "nonexistent.txt"),
			workingDir:   tmpDir,
			expectError:  true,
			expectExists: false,
		},
		{
			name:         "file outside working directory",
			path:         "/etc/passwd",
			workingDir:   tmpDir,
			expectError:  true,
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFileForDelete(tt.path, tt.workingDir)

			if tt.expectError && result.Error == nil {
				t.Errorf("ValidateFileForDelete(%q, %q) expected error but got none", tt.path, tt.workingDir)
			}
			if !tt.expectError && result.Error != nil {
				t.Errorf("ValidateFileForDelete(%q, %q) unexpected error: %v", tt.path, tt.workingDir, result.Error)
			}
			if result.Exists != tt.expectExists {
				t.Errorf("ValidateFileForDelete(%q, %q).Exists = %v, want %v", tt.path, tt.workingDir, result.Exists, tt.expectExists)
			}
		})
	}
}

func TestGetParentDir(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple path",
			path:     "/path/to/file.txt",
			expected: "/path/to",
		},
		{
			name:     "root file",
			path:     "/file.txt",
			expected: "/",
		},
		{
			name:     "relative path",
			path:     "src/main.go",
			expected: "src",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetParentDir(tt.path)
			if result != tt.expected {
				t.Errorf("GetParentDir(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "remove double dots",
			path:     "/path/to/../file.txt",
			expected: "/path/file.txt",
		},
		{
			name:     "remove single dots",
			path:     "/path/./to/file.txt",
			expected: "/path/to/file.txt",
		},
		{
			name:     "collapse separators",
			path:     "/path//to///file.txt",
			expected: "/path/to/file.txt",
		},
		{
			name:     "trailing separator",
			path:     "/path/to/",
			expected: "/path/to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.path)
			if result != tt.expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestListDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "subdir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "subdir2"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, ".hidden"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hiddenfile"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir1", "nested.txt"), []byte("test"), 0644)

	files, truncated, err := ListDirectory(tmpDir, []string{}, 100)
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	if truncated {
		t.Errorf("Expected no truncation")
	}

	for _, f := range files {
		if strings.Contains(f, ".hidden") {
			t.Errorf("Hidden path should be skipped: %s", f)
		}
		if strings.Contains(f, "node_modules") {
			t.Errorf("node_modules should be skipped: %s", f)
		}
	}

	files, truncated, err = ListDirectory(tmpDir, []string{}, 1)
	if err != nil {
		t.Fatalf("ListDirectory with limit failed: %v", err)
	}
	if !truncated {
		t.Errorf("Expected truncation with limit 1")
	}

	files, _, err = ListDirectory(tmpDir, []string{"**/*.go"}, 100)
	if err != nil {
		t.Fatalf("ListDirectory with pattern failed: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, ".go") {
			t.Errorf("Go files should be skipped: %s", f)
		}
	}
}

func TestReadWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!"

	err := WriteFile(testFile, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	readContent, err := ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if readContent != content {
		t.Errorf("ReadFile returned %q, want %q", readContent, content)
	}

	_, err = ReadFile(filepath.Join(tmpDir, "nonexistent.txt"))
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}
}

func TestCreateFile(t *testing.T) {
	tmpDir := t.TempDir()

	newFile := filepath.Join(tmpDir, "new.txt")
	err := CreateFile(newFile, "content")
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	content, err := ReadFile(newFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if content != "content" {
		t.Errorf("Content = %q, want %q", content, "content")
	}

	err = CreateFile(newFile, "new content")
	if err == nil {
		t.Errorf("Expected error for existing file")
	}

	err = CreateFile(tmpDir, "content")
	if err == nil {
		t.Errorf("Expected error for directory path")
	}
}

func TestDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := WriteFile(testFile, "test"); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err := DeleteFile(testFile)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}

	if FileExists(testFile) {
		t.Errorf("File should be deleted")
	}
}

func TestGlobWithDoublestar(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "src", "util"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "util.go"), []byte("util"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "util", "helper.go"), []byte("helper"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden.go"), []byte("hidden"), 0644)

	files, _, err := GlobWithDoublestar("*.go", tmpDir, 10)
	if err != nil {
		t.Fatalf("GlobWithDoublestar failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	files, _, err = GlobWithDoublestar("**/*.go", tmpDir, 10)
	if err != nil {
		t.Fatalf("GlobWithDoublestar recursive failed: %v", err)
	}
	found := false
	for _, f := range files {
		if strings.HasSuffix(f, "main.go") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find main.go")
	}

	for _, f := range files {
		if strings.Contains(f, ".hidden") {
			t.Errorf("Hidden files should be skipped: %s", f)
		}
	}
}

func TestGetFileInfo(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	info, err := GetFileInfo(testFile)
	if err != nil {
		t.Fatalf("GetFileInfo failed: %v", err)
	}
	if info.Size() != int64(len(content)) {
		t.Errorf("Size = %d, want %d", info.Size(), len(content))
	}
	if info.IsDir() {
		t.Errorf("Expected file, not directory")
	}

	_, err = GetFileInfo(filepath.Join(tmpDir, "nonexistent.txt"))
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}
}

func TestSafeReadFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     []byte
		fileName    string
		expectError bool
		errorMatch  string
	}{
		{
			name:        "small text file",
			content:     []byte("hello world"),
			fileName:    "test.txt",
			expectError: false,
		},
		{
			name:        "empty file",
			content:     []byte{},
			fileName:    "empty.txt",
			expectError: false,
		},
		{
			name:        "text file exactly 512 bytes",
			content:     make([]byte, 512),
			fileName:    "exact512.txt",
			expectError: false,
		},
		{
			name:        "text file larger than 512 bytes",
			content:     append(make([]byte, 512), []byte("additional content")...),
			fileName:    "large.txt",
			expectError: false,
		},
		{
			name:        "binary file - PNG header",
			content:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			fileName:    "image.png",
			expectError: true,
			errorMatch:  "binary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill text content with printable characters for text files
			if !tt.expectError && len(tt.content) > 0 && !strings.Contains(tt.name, "binary") && !strings.Contains(tt.name, "PNG") {
				for i := range tt.content {
					tt.content[i] = 'a'
				}
			}

			testFile := filepath.Join(tmpDir, tt.fileName)
			if err := os.WriteFile(testFile, tt.content, 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			result, err := SafeReadFile(testFile, tmpDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("SafeReadFile(%q) expected error but got none", testFile)
				} else if tt.errorMatch != "" && !strings.Contains(err.Error(), tt.errorMatch) {
					t.Errorf("SafeReadFile(%q) error = %v, want error containing %q", testFile, err, tt.errorMatch)
				}
			} else {
				if err != nil {
					t.Errorf("SafeReadFile(%q) unexpected error: %v", testFile, err)
				}
				if string(tt.content) != result {
					t.Errorf("SafeReadFile(%q) = %q, want %q", testFile, result, string(tt.content))
				}
			}
		})
	}
}

func TestSafeReadFile_TooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create a file larger than MaxReadSize
	largeContent := make([]byte, MaxReadSize+1)
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	if err := os.WriteFile(testFile, largeContent, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := SafeReadFile(testFile, tmpDir)
	if err == nil {
		t.Errorf("SafeReadFile(%q) expected error for large file", testFile)
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("SafeReadFile(%q) error = %v, want error containing 'too large'", testFile, err)
	}
}

func TestSafeReadFile_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to read a file outside the working directory
	_, err := SafeReadFile("/etc/passwd", tmpDir)
	if err == nil {
		t.Errorf("SafeReadFile('/etc/passwd') expected error for path traversal")
	}
}

func TestSafeReadFile_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := SafeReadFile(filepath.Join(tmpDir, "nonexistent.txt"), tmpDir)
	if err == nil {
		t.Errorf("SafeReadFile(nonexistent) expected error")
	}
}

func TestIsTextFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     []byte
		fileName    string
		expectText  bool
		expectError bool
	}{
		{
			name:       "text file",
			content:    []byte("hello world"),
			fileName:   "test.txt",
			expectText: true,
		},
		{
			name:       "empty file",
			content:    []byte{},
			fileName:   "empty.txt",
			expectText: true, // Empty files are considered text (application/octet-stream)
		},
		{
			name:       "binary file - PNG header",
			content:    []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			fileName:   "image.png",
			expectText: false,
		},
		{
			name:       "go source file",
			content:    []byte("package main\n\nfunc main() {}"),
			fileName:   "main.go",
			expectText: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.fileName)
			if err := os.WriteFile(testFile, tt.content, 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			isText, err := IsTextFile(testFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("IsTextFile(%q) expected error but got none", testFile)
				}
			} else {
				if err != nil {
					t.Errorf("IsTextFile(%q) unexpected error: %v", testFile, err)
				}
				if isText != tt.expectText {
					t.Errorf("IsTextFile(%q) = %v, want %v", testFile, isText, tt.expectText)
				}
			}
		})
	}
}
