package fileutil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"github.com/bmatcuk/doublestar/v4"
)

var (
	rgPath  string
	fzfPath string
)

// Standard permissions for a Native application
const (
	DefaultDirPerms  = 0755
	DefaultFilePerms = 0644
	// MaxReadSize limits AI from reading massive files (1MB is usually plenty for context)
	MaxReadSize = 1 * 1024 * 1024
)

// CommonIgnoredDirs is a map of directory names that should be ignored in file operations
var CommonIgnoredDirs = map[string]bool{
	".opencode":        true,
	"node_modules":     true,
	"vendor":           true,
	"dist":             true,
	"build":            true,
	"target":           true,
	".git":             true,
	".svn":             true,
	".hg":              true,
	".idea":            true,
	".vscode":          true,
	".cursor":          true,
	".claude":          true,
	".vscode-insiders": true,
	".gradle":          true,
	".cache":           true,
	".gitignore":       true,
	".gitattributes":   true,
	".gitmodules":      true,
	".gitkeep":         true,
	".DS_Store":        true,
	".env":             true,
	".github":          true,
	".githooks":        true,
	".agents":          true,
	"__pycache__":      true,
	"bin":              true,
	"obj":              true,
	"out":              true,
	"coverage":         true,
	"tmp":              true,
	"temp":             true,
	"logs":             true,
	"generated":        true,
	"bower_components": true,
	"jspm_packages":    true,
}

// CommonIgnoredExtensions is a map of file extensions that should be ignored
var CommonIgnoredExtensions = map[string]bool{
	".swp":   true,
	".swo":   true,
	".tmp":   true,
	".temp":  true,
	".bak":   true, // maybe we need to read this
	".log":   true, // maybe we need to read this
	".obj":   true,
	".out":   true,
	".pyc":   true, // Python bytecode
	".pyo":   true, // Python optimized bytecode
	".pyd":   true, // Python extension modules
	".o":     true, // Object files
	".so":    true, // Shared libraries
	".dylib": true, // macOS shared libraries
	".dll":   true, // Windows shared libraries
	".a":     true, // Static libraries
	".exe":   true, // Windows executables
	".lock":  true, // Lock files
}

// ImageTypes maps image extensions to their type names
var ImageTypes = map[string]string{
	".jpg":  "JPEG",
	".jpeg": "JPEG",
	".png":  "PNG",
	".gif":  "GIF",
	".bmp":  "BMP",
	".svg":  "SVG",
	".webp": "WebP",
}

// FileOperation represents the type of file operation being performed
type FileOperation int

const (
	OpRead FileOperation = iota
	OpWrite
	OpEdit
	OpDelete
	OpCreate
)

// FileInfo contains basic file information
type FileInfo struct {
	Path    string
	ModTime time.Time
	Size    int64
	IsDir   bool
}

// FileValidationResult contains the result of file validation
type FileValidationResult struct {
	AbsPath      string
	Exists       bool
	IsDirectory  bool
	IsModified   bool
	LastReadTime time.Time
	ModTime      time.Time
	Size         int64
	Error        error
}

// FileReadResult contains the result of reading a file
type FileReadResult struct {
	Content    string
	TotalLines int
	Error      error
}

// PathConfig holds configuration for path operations
type PathConfig struct {
	WorkingDir     string
	IgnorePatterns []string
}

// DefaultPathConfig returns a default path configuration
func DefaultPathConfig(workingDir string) PathConfig {
	return PathConfig{
		WorkingDir:     workingDir,
		IgnorePatterns: []string{},
	}
}

// MaxLineLength is the maximum allowed line length when reading files
const MaxLineLength = 2000

func init() {
	ReloadTools()
}

// ReloadTools refreshes the path for external dependencies
func ReloadTools() {
	var err error
	rgPath, err = exec.LookPath("rg")
	if err != nil {
		logging.Warn("Ripgrep (rg) not found in $PATH. Some features might be limited or slower.")
		rgPath = ""
	}
	fzfPath, err = exec.LookPath("fzf")
	if err != nil {
		logging.Warn("FZF not found in $PATH. Some features might be limited or slower.")
		fzfPath = ""
	}
}

// GetRgCmd returns a command for ripgrep with the given glob pattern
func GetRgCmd(globPattern string) *exec.Cmd {
	if rgPath == "" {
		return nil
	}
	rgArgs := []string{
		"--files",
		"-L",
		"--null",
	}
	if globPattern != "" {
		if !filepath.IsAbs(globPattern) && !strings.HasPrefix(globPattern, string(filepath.Separator)) {
			globPattern = string(filepath.Separator) + globPattern
		}
		rgArgs = append(rgArgs, "--glob", globPattern)
	}
	cmd := exec.Command(rgPath, rgArgs...)
	cmd.Dir = "."
	return cmd
}

// GetFzfCmd returns a command for fzf with the given query
func GetFzfCmd(query string) *exec.Cmd {
	if fzfPath == "" {
		return nil
	}
	fzfArgs := []string{
		"--filter",
		query,
		"--read0",
		"--print0",
	}
	cmd := exec.Command(fzfPath, fzfArgs...)
	cmd.Dir = "."
	return cmd
}

// ============================================
// PATH RESOLUTION AND SANITIZATION
// ============================================

// ResolvePath converts a path to absolute using the working directory.
// It cleans the path to prevent basic traversal but does NOT enforce boundaries.
func ResolvePath(path, workingDir string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(workingDir, path))
}

// ResolvePaths resolves multiple paths to absolute paths
func ResolvePaths(paths []string, workingDir string) []string {
	result := make([]string, len(paths))
	for i, path := range paths {
		result[i] = ResolvePath(path, workingDir)
	}
	return result
}

// SecureResolvePath ensures the resolved path is trapped within the workingDir.
// Use this for any AI-generated path to prevent Path Traversal attacks.
func SecureResolvePath(path, workingDir string) (string, error) {
	absBase, err := filepath.Abs(workingDir)
	if err != nil {
		return "", err
	}

	target := ResolvePath(path, absBase)
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	// Ensure the final path actually starts with the working directory path
	if !strings.HasPrefix(absTarget, absBase) {
		return "", fmt.Errorf("security: path traversal attempt blocked: %s", path)
	}

	return absTarget, nil
}

// IsInWorkingDir checks if a path is within the working directory
func IsInWorkingDir(path, workingDir string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return false
	}
	return strings.HasPrefix(absPath, absWorkingDir)
}

// GetParentDir returns the parent directory of a path
func GetParentDir(path string) string {
	return filepath.Dir(path)
}

// NormalizePath normalizes a path by resolving . and .. components
func NormalizePath(path string) string {
	return filepath.Clean(path)
}

// ============================================
// FILE EXISTENCE AND TYPE CHECKING
// ============================================

// FileExists checks if a path exists and is a file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Exists checks if a path exists (file or directory)
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// NormalizePathForAI ensures paths always use forward slashes, which AI models handle better.
func NormalizePathForAI(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if a path is a file (not a directory)
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsImageFile checks if a file is an image based on extension
func IsImageFile(path string) (bool, string) {
	ext := strings.ToLower(filepath.Ext(path))
	if imgType, ok := ImageTypes[ext]; ok {
		return true, imgType
	}
	return false, ""
}

// GetFileSize returns the size of a file
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ============================================
// FILE MODIFICATION CHECKING
// ============================================

// GetFileModTime returns the modification time of a file
func GetFileModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// IsFileModifiedSince checks if a file has been modified since a given time
func IsFileModifiedSince(path string, since time.Time) (bool, error) {
	modTime, err := GetFileModTime(path)
	if err != nil {
		return false, err
	}
	return modTime.After(since), nil
}

// ============================================
// HIDDEN FILE AND IGNORED PATH CHECKING
// ============================================

// IsHiddenFile checks if a file is hidden (starts with dot)
func IsHiddenFile(path string) bool {
	base := filepath.Base(path)
	return base != "." && strings.HasPrefix(base, ".")
}

// IsIgnoredDir checks if a directory should be ignored
func IsIgnoredDir(path string) bool {
	base := filepath.Base(path)
	return CommonIgnoredDirs[base]
}

// IsIgnoredExtension checks if a file extension should be ignored
func IsIgnoredExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return CommonIgnoredExtensions[ext]
}

// SkipHidden checks if a path should be skipped (hidden or ignored).
// Optimized to avoid heavy allocations like strings.Split.
func SkipHidden(path string) bool {
	// Standardize to forward slashes for cross-platform matching
	cleanPath := filepath.ToSlash(filepath.Clean(path))

	// 1. Check for any hidden segment in the path (e.g., "src/.secret/file.go")
	if strings.Contains(cleanPath, "/.") || strings.HasPrefix(cleanPath, ".") {
		// Safety check: don't ignore the current directory itself
		if cleanPath != "." && cleanPath != "./" {
			return true
		}
	}

	// 2. Iterate through path segments to catch ignored directories
	// Example: "internal/node_modules/react/index.js"
	start := 0
	for {
		end := strings.IndexByte(cleanPath[start:], '/')
		var segment string

		if end == -1 {
			segment = cleanPath[start:]
		} else {
			segment = cleanPath[start : start+end]
		}

		if segment != "" && IsIgnoredDir(segment) {
			return true
		}

		if end == -1 {
			break
		}
		start += end + 1
	}

	// 3. Final check for forbidden extensions (e.g., .exe, .dll)
	return IsIgnoredExtension(cleanPath)
}

// ShouldSkipPath checks if a path should be skipped based on ignore patterns
// This is a more comprehensive version that combines hidden check with custom patterns
func ShouldSkipPath(path string, ignorePatterns []string) bool {
	// First, check the built-in "smart" ignores
	if SkipHidden(path) {
		return true
	}

	// Then, check custom ignore patterns using doublestar for recursive support
	for _, pattern := range ignorePatterns {
		// Use doublestar.Match for '**/temp/*' style patterns
		matched, err := doublestar.Match(pattern, path)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// ============================================
// FILE READING AND WRITING
// ============================================

// IsTextFile checks if a file is likely text vs binary to prevent AI from reading garbage.
func IsTextFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read the first 512 bytes to determine the content type
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	contentType := http.DetectContentType(buffer[:n])
	return strings.HasPrefix(contentType, "text/") || contentType == "application/octet-stream", nil
}

// ReadFile reads a file and returns its content
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SafeReadFile reads a file only if it meets security and size requirements.
func SafeReadFile(path, workingDir string) (string, error) {
	safePath, err := SecureResolvePath(path, workingDir)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(safePath)
	if err != nil {
		return "", err
	}

	if info.Size() > MaxReadSize {
		return "", fmt.Errorf("file too large (%d bytes). Max allowed: %d", info.Size(), MaxReadSize)
	}

	isText, err := IsTextFile(safePath)
	if err != nil || !isText {
		return "", errors.New("file appears to be binary or unsupported format")
	}

	content, err := os.ReadFile(safePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// ReadFileWithLimit reads a file with line offset and limit
func ReadFileWithLimit(path string, offset, limit int) (FileReadResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return FileReadResult{}, err
	}
	defer file.Close()

	// Count total lines first
	totalLines := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		totalLines++
	}
	if err := scanner.Err(); err != nil {
		return FileReadResult{}, err
	}

	// Reset to beginning
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return FileReadResult{}, err
	}

	// Skip to offset
	lineNum := 0
	scanner = bufio.NewScanner(file)
	for lineNum < offset && scanner.Scan() {
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		return FileReadResult{}, err
	}

	// Read up to limit lines
	var lines []string
	for scanner.Scan() && len(lines) < limit {
		lineText := scanner.Text()
		// Truncate long lines
		if len(lineText) > MaxLineLength {
			lineText = lineText[:MaxLineLength] + "..."
		}
		lines = append(lines, lineText)
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return FileReadResult{}, err
	}

	return FileReadResult{
		Content:    strings.Join(lines, "\n"),
		TotalLines: totalLines,
	}, nil
}

// WriteFile writes content to a file
func WriteFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// CreateFile creates a new file with the given content
func CreateFile(path, content string) error {
	if DirExists(path) {
		return errors.New("path is a directory, not a file")
	}
	if FileExists(path) {
		return errors.New("file already exists")
	}
	return WriteFile(path, content)
}

// ============================================
// FILE DELETION
// ============================================

// DeleteFile deletes a file
func DeleteFile(path string) error {
	return os.Remove(path)
}

// DeleteDir deletes a directory recursively
func DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// ============================================
// GLOB AND PATTERN MATCHING
// ============================================

// GlobWithDoublestar finds files matching a pattern
func GlobWithDoublestar(pattern, searchPath string, limit int) ([]string, bool, error) {
	fsys := os.DirFS(searchPath)
	relPattern := strings.TrimPrefix(pattern, "/")
	var matches []FileInfo

	err := doublestar.GlobWalk(fsys, relPattern, func(path string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		if SkipHidden(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		absPath := path
		if !strings.HasPrefix(absPath, searchPath) && searchPath != "." {
			absPath = filepath.Join(searchPath, absPath)
		} else if !strings.HasPrefix(absPath, "/") && searchPath == "." {
			absPath = filepath.Join(searchPath, absPath)
		}

		matches = append(matches, FileInfo{Path: absPath, ModTime: info.ModTime()})
		if limit > 0 && len(matches) >= limit*2 {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, false, fmt.Errorf("glob walk error: %w", err)
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ModTime.After(matches[j].ModTime)
	})

	truncated := false
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
		truncated = true
	}

	results := make([]string, len(matches))
	for i, m := range matches {
		results[i] = m.Path
	}
	return results, truncated, nil
}

// GlobToRegex converts a glob pattern to a regex pattern
func GlobToRegex(glob string) string {
	regexPattern := strings.ReplaceAll(glob, ".", "\\.")
	regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "?", ".")

	re := regexp.MustCompile(`\{([^}]+)\}`)
	regexPattern = re.ReplaceAllStringFunc(regexPattern, func(match string) string {
		inner := match[1 : len(match)-1]
		return "(" + strings.ReplaceAll(inner, ",", "|") + ")"
	})

	return regexPattern
}

// ============================================
// REGEX PATTERN ESCAPING
// ============================================

// EscapeRegexPattern escapes special regex characters
func EscapeRegexPattern(pattern string) string {
	specialChars := []string{"\\", ".", "+", "*", "?", "(", ")", "[", "]", "{", "}", "^", "$", "|"}
	escaped := pattern

	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, "\\"+char)
	}

	return escaped
}

// ============================================
// DIRECTORY LISTING
// ============================================

// ListDirectory lists files in a directory
func ListDirectory(initialPath string, ignorePatterns []string, limit int) ([]string, bool, error) {
	var results []string
	truncated := false

	err := filepath.Walk(initialPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we don't have permission to access
		}

		if ShouldSkipPath(path, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if path != initialPath {
			if info.IsDir() {
				path = path + string(filepath.Separator)
			}
			results = append(results, path)
		}

		if len(results) >= limit {
			truncated = true
			return filepath.SkipAll
		}

		return nil
	})
	if err != nil {
		return nil, truncated, err
	}

	return results, truncated, nil
}

// ============================================
// FILE CONTENT SEARCHING
// ============================================

// FileContainsPattern checks if a file contains a pattern
func FileContainsPattern(filePath string, pattern *regexp.Regexp) (bool, int, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, 0, "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if pattern.MatchString(line) {
			return true, lineNum, line, nil
		}
	}

	return false, 0, "", scanner.Err()
}

// ============================================
// UNIFIED FILE VALIDATION
// ============================================

// ValidateFileOperation performs comprehensive validation for file operations
func ValidateFileOperation(path, workingDir string, operation FileOperation) *FileValidationResult {
	result := &FileValidationResult{
		AbsPath: ResolvePath(path, workingDir),
	}

	// Check if path is within working directory
	if !IsInWorkingDir(result.AbsPath, workingDir) {
		result.Error = errors.New("cannot access files outside the working directory")
		return result
	}

	// Get file info
	info, err := os.Stat(result.AbsPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Exists = false
			// For create operations, non-existence is OK
			if operation == OpCreate {
				return result
			}
			result.Error = fmt.Errorf("file not found: %s", result.AbsPath)
			return result
		}
		result.Error = fmt.Errorf("failed to access file: %w", err)
		return result
	}

	result.Exists = true
	result.IsDirectory = info.IsDir()
	result.ModTime = info.ModTime()
	result.Size = info.Size()

	// For read operations, check if file was modified since last read
	// This is handled by the caller using LastReadTime

	return result
}

// ValidateFileForRead validates a file for reading
func ValidateFileForRead(path, workingDir string, lastReadTime time.Time) *FileValidationResult {
	result := ValidateFileForReadWithoutLastRead(path, workingDir)
	if result.Error != nil {
		return result
	}

	// Check modification time against last read
	if !lastReadTime.IsZero() && result.ModTime.After(lastReadTime) {
		result.IsModified = true
		result.LastReadTime = lastReadTime
		result.Error = fmt.Errorf("file has been modified since it was last read (mod time: %s, last read: %s)",
			result.ModTime.Format(time.RFC3339), lastReadTime.Format(time.RFC3339))
	}

	return result
}

// ValidateFileForReadWithoutLastRead validates a file for reading without last read check
func ValidateFileForReadWithoutLastRead(path, workingDir string) *FileValidationResult {
	result := &FileValidationResult{
		AbsPath: ResolvePath(path, workingDir),
	}

	// Check if path is within working directory
	if !IsInWorkingDir(result.AbsPath, workingDir) {
		result.Error = errors.New("cannot access files outside the working directory")
		return result
	}

	info, err := os.Stat(result.AbsPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Exists = false
			result.Error = fmt.Errorf("file not found: %s", result.AbsPath)
			return result
		}
		result.Error = fmt.Errorf("failed to access file: %w", err)
		return result
	}

	result.Exists = true
	result.IsDirectory = info.IsDir()
	result.ModTime = info.ModTime()
	result.Size = info.Size()

	if result.IsDirectory {
		result.Error = fmt.Errorf("path is a directory, not a file: %s", result.AbsPath)
	}

	return result
}

// ValidateFileForWrite validates a file for writing
func ValidateFileForWrite(path, workingDir string, lastReadTime time.Time) *FileValidationResult {
	result := &FileValidationResult{
		AbsPath: ResolvePath(path, workingDir),
	}

	// Check if path is within working directory
	if !IsInWorkingDir(result.AbsPath, workingDir) {
		result.Error = errors.New("cannot write to files outside the working directory")
		return result
	}

	info, err := os.Stat(result.AbsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - OK for create
			result.Exists = false
			return result
		}
		result.Error = fmt.Errorf("failed to access file: %w", err)
		return result
	}

	result.Exists = true
	result.IsDirectory = info.IsDir()
	result.ModTime = info.ModTime()
	result.Size = info.Size()

	if result.IsDirectory {
		result.Error = fmt.Errorf("path is a directory, not a file: %s", result.AbsPath)
		return result
	}

	// Check if file was modified since last read
	if !lastReadTime.IsZero() && result.ModTime.After(lastReadTime) {
		result.IsModified = true
		result.LastReadTime = lastReadTime
		result.Error = fmt.Errorf("file has been modified since it was last read (mod time: %s, last read: %s)",
			result.ModTime.Format(time.RFC3339), lastReadTime.Format(time.RFC3339))
	}

	return result
}

// ValidateFileForDelete validates a file for deletion
func ValidateFileForDelete(path, workingDir string) *FileValidationResult {
	result := &FileValidationResult{
		AbsPath: ResolvePath(path, workingDir),
	}

	// Check if path is within working directory
	if !IsInWorkingDir(result.AbsPath, workingDir) {
		result.Error = errors.New("cannot delete files outside the working directory")
		return result
	}

	info, err := os.Lstat(result.AbsPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Exists = false
			result.Error = fmt.Errorf("file or directory does not exist: %s", result.AbsPath)
			return result
		}
		result.Error = fmt.Errorf("failed to access path: %w", err)
		return result
	}

	result.Exists = true
	result.IsDirectory = info.IsDir()
	result.ModTime = info.ModTime()
	result.Size = info.Size()

	return result
}

// EnsureParentDir ensures the parent directory of a path exists
func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}
	return nil
}

// ============================================
// FILE STAT OPERATIONS
// ============================================

// GetFileInfo returns file information
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// GetFileInfoL returns file information without following symlinks
func GetFileInfoL(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}
