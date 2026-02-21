package skill

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "git-release", false},
		{"valid single word", "docker", false},
		{"valid multiple hyphens", "my-cool-skill", false},
		{"valid with numbers", "skill-123", false},
		{"empty name", "", true},
		{"uppercase", "Git-Release", true},
		{"starts with hyphen", "-skill", true},
		{"ends with hyphen", "skill-", true},
		{"consecutive hyphens", "skill--name", true},
		{"underscore", "skill_name", true},
		{"space", "skill name", true},
		{"too long", string(make([]byte, 65)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid description", "A helpful skill", false},
		{"max length", string(make([]byte, 1024)), false},
		{"empty description", "", true},
		{"too long", string(make([]byte, 1025)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDescription(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSplitFrontmatter(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantFrontmatter string
		wantContent     string
		wantErr         bool
	}{
		{
			name: "valid frontmatter",
			input: `---
name: test
description: A test skill
---

# Content here`,
			wantFrontmatter: "name: test\ndescription: A test skill",
			wantContent:     "\n# Content here",
			wantErr:         false,
		},
		{
			name: "no content",
			input: `---
name: test
description: A test skill
---`,
			wantFrontmatter: "name: test\ndescription: A test skill",
			wantContent:     "",
			wantErr:         false,
		},
		{
			name:    "missing start delimiter",
			input:   "name: test\n---",
			wantErr: true,
		},
		{
			name: "missing end delimiter",
			input: `---
name: test
description: A test skill`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter, content, err := splitFrontmatter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if frontmatter != tt.wantFrontmatter {
					t.Errorf("splitFrontmatter() frontmatter = %q, want %q", frontmatter, tt.wantFrontmatter)
				}
				if content != tt.wantContent {
					t.Errorf("splitFrontmatter() content = %q, want %q", content, tt.wantContent)
				}
			}
		})
	}
}

func TestParseSkillFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create a valid skill file
	validSkillDir := filepath.Join(tmpDir, "git-release")
	if err := os.MkdirAll(validSkillDir, 0755); err != nil {
		t.Fatal(err)
	}

	validSkillContent := `---
name: git-release
description: Create consistent releases and changelogs
license: MIT
compatibility: opencode
metadata:
  audience: maintainers
---

## What I do

- Draft release notes
- Propose version bump
`

	validSkillPath := filepath.Join(validSkillDir, "SKILL.md")
	if err := os.WriteFile(validSkillPath, []byte(validSkillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create skill with name mismatch
	mismatchDir := filepath.Join(tmpDir, "wrong-name")
	if err := os.MkdirAll(mismatchDir, 0755); err != nil {
		t.Fatal(err)
	}

	mismatchContent := `---
name: different-name
description: This name doesn't match directory
---

Content here
`

	mismatchPath := filepath.Join(mismatchDir, "SKILL.md")
	if err := os.WriteFile(mismatchPath, []byte(mismatchContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		check   func(*testing.T, *Info)
	}{
		{
			name:    "valid skill",
			path:    validSkillPath,
			wantErr: false,
			check: func(t *testing.T, info *Info) {
				if info.Name != "git-release" {
					t.Errorf("Name = %q, want %q", info.Name, "git-release")
				}
				if info.Description != "Create consistent releases and changelogs" {
					t.Errorf("Description = %q, want %q", info.Description, "Create consistent releases and changelogs")
				}
				if info.License != "MIT" {
					t.Errorf("License = %q, want %q", info.License, "MIT")
				}
				if info.Location != validSkillPath {
					t.Errorf("Location = %q, want %q", info.Location, validSkillPath)
				}
				if !contains(info.Content, "Draft release notes") {
					t.Errorf("Content missing expected text")
				}
			},
		},
		{
			name:    "name mismatch",
			path:    mismatchPath,
			wantErr: true,
		},
		{
			name:    "nonexistent file",
			path:    filepath.Join(tmpDir, "nonexistent", "SKILL.md"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parseSkillFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSkillFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, info)
			}
		})
	}
}

func TestGetWorktreeRoot(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create nested directories
	gitDir := filepath.Join(tmpDir, "project")
	subDir := filepath.Join(gitDir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .git directory
	dotGit := filepath.Join(gitDir, ".git")
	if err := os.MkdirAll(dotGit, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		workingDir string
		want       string
	}{
		{
			name:       "from git root",
			workingDir: gitDir,
			want:       gitDir,
		},
		{
			name:       "from subdirectory",
			workingDir: subDir,
			want:       gitDir,
		},
		{
			name:       "not in git repo",
			workingDir: tmpDir,
			want:       tmpDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getWorktreeRoot(tt.workingDir)
			if got != tt.want {
				t.Errorf("getWorktreeRoot() = %q, want %q", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDiscoverProjectSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .git directory to mark as worktree root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create skill in root .opencode/skills/
	rootSkillDir := filepath.Join(tmpDir, ".opencode", "skills", "root-skill")
	if err := os.MkdirAll(rootSkillDir, 0755); err != nil {
		t.Fatal(err)
	}
	rootSkillContent := `---
name: root-skill
description: Skill at root level
---
Root content`
	if err := os.WriteFile(filepath.Join(rootSkillDir, "SKILL.md"), []byte(rootSkillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create skill in subdirectory .opencode/skills/
	subSkillDir := filepath.Join(subDir, ".opencode", "skills", "sub-skill")
	if err := os.MkdirAll(subSkillDir, 0755); err != nil {
		t.Fatal(err)
	}
	subSkillContent := `---
name: sub-skill
description: Skill in subdirectory
---
Sub content`
	if err := os.WriteFile(filepath.Join(subSkillDir, "SKILL.md"), []byte(subSkillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Discover from subdirectory - should find both
	worktreeRoot := getWorktreeRoot(subDir)
	skills := discoverProjectSkills(subDir, worktreeRoot)

	if len(skills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(skills))
	}

	// Verify both skills were found
	foundRoot := false
	foundSub := false
	for _, s := range skills {
		if s.Name == "root-skill" {
			foundRoot = true
		}
		if s.Name == "sub-skill" {
			foundSub = true
		}
	}

	if !foundRoot {
		t.Errorf("root-skill not found")
	}
	if !foundSub {
		t.Errorf("sub-skill not found")
	}
}

func TestDiscoverCustomPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create custom skill directory
	customDir := filepath.Join(tmpDir, "my-skills", "custom-skill")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}

	customContent := `---
name: custom-skill
description: Custom skill from custom path
---
Custom content`
	if err := os.WriteFile(filepath.Join(customDir, "SKILL.md"), []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test absolute path
	paths := []string{filepath.Join(tmpDir, "my-skills")}
	skills := discoverCustomPaths(paths, tmpDir)

	if len(skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(skills))
	}

	if len(skills) > 0 && skills[0].Name != "custom-skill" {
		t.Errorf("Expected custom-skill, got %s", skills[0].Name)
	}

	// Test relative path
	relPaths := []string{"my-skills"}
	relSkills := discoverCustomPaths(relPaths, tmpDir)

	if len(relSkills) != 1 {
		t.Errorf("Expected 1 skill from relative path, got %d", len(relSkills))
	}
}

func TestInvalidSkillFiles(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		dirName     string
		content     string
		shouldError bool
	}{
		{
			name:    "missing name",
			dirName: "missing-name",
			content: `---
description: Missing name field
---
Content`,
			shouldError: true,
		},
		{
			name:    "missing description",
			dirName: "missing-desc",
			content: `---
name: missing-desc
---
Content`,
			shouldError: true,
		},
		{
			name:    "invalid name format",
			dirName: "invalid-name",
			content: `---
name: Invalid_Name
description: Has underscore
---
Content`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skillDir := filepath.Join(tmpDir, tt.dirName)
			if err := os.MkdirAll(skillDir, 0755); err != nil {
				t.Fatal(err)
			}

			skillPath := filepath.Join(skillDir, "SKILL.md")
			if err := os.WriteFile(skillPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := parseSkillFile(skillPath)
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for %s, got %v", tt.name, err)
			}
		})
	}
}

func TestSkillError(t *testing.T) {
	tests := []struct {
		name     string
		skillErr *SkillError
		want     string
	}{
		{
			name: "with wrapped error",
			skillErr: &SkillError{
				Path:    "/path/to/skill.md",
				Message: "failed to parse",
				Err:     ErrInvalidFrontmatter,
			},
			want: "/path/to/skill.md: failed to parse: invalid skill frontmatter",
		},
		{
			name: "without wrapped error",
			skillErr: &SkillError{
				Path:    "/path/to/skill.md",
				Message: "some error",
			},
			want: "/path/to/skill.md: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.skillErr.Error(); got != tt.want {
				t.Errorf("SkillError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSkillErrorUnwrap(t *testing.T) {
	wrappedErr := ErrInvalidFrontmatter
	skillErr := &SkillError{
		Path:    "/path/to/skill.md",
		Message: "failed",
		Err:     wrappedErr,
	}

	unwrapped := skillErr.Unwrap()
	if unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}

	// Test nil error
	skillErrNoWrap := &SkillError{
		Path:    "/path/to/skill.md",
		Message: "no wrap",
	}
	if skillErrNoWrap.Unwrap() != nil {
		t.Errorf("Unwrap() should return nil for no wrapped error")
	}
}

func TestSplitFrontmatter_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantFrontmatter string
		wantContent     string
		wantErr         bool
	}{
		{
			name:            "CRLF line endings",
			input:           "---\r\nname: test\r\ndescription: desc\r\n---\r\n\r\nContent",
			wantFrontmatter: "name: test\r\ndescription: desc\r",
			wantContent:     "\r\nContent",
			wantErr:         false,
		},
		{
			name:    "only start delimiter",
			input:   "---\nname: test",
			wantErr: true,
		},
		{
			name:    "empty file",
			input:   "",
			wantErr: true,
		},
		{
			name: "content with dashes",
			input: `---
name: test
description: desc
---

---
This has dashes
---`,
			wantFrontmatter: "name: test\ndescription: desc",
			wantContent:     "\n---\nThis has dashes\n---",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter, content, err := splitFrontmatter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if frontmatter != tt.wantFrontmatter {
					t.Errorf("frontmatter = %q, want %q", frontmatter, tt.wantFrontmatter)
				}
				if content != tt.wantContent {
					t.Errorf("content = %q, want %q", content, tt.wantContent)
				}
			}
		})
	}
}

func TestParseSkillFile_ContentTooLarge(t *testing.T) {
	tmpDir := t.TempDir()

	skillDir := filepath.Join(tmpDir, "large-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create content larger than maxContentSize (100KB)
	largeContent := make([]byte, 101*1024)
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	content := fmt.Sprintf(`---
name: large-skill
description: A large skill
---
%s`, string(largeContent))

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := parseSkillFile(skillPath)
	if err == nil {
		t.Error("Expected error for content too large, got nil")
	}

	var skillErr *SkillError
	if errors.As(err, &skillErr) {
		if skillErr.Err != ErrContentTooLarge {
			t.Errorf("Expected ErrContentTooLarge, got %v", skillErr.Err)
		}
	} else {
		t.Errorf("Expected SkillError, got %T", err)
	}
}

func TestScanDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create skill directory structure
	skillDir := filepath.Join(tmpDir, "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	skillContent := `---
name: test-skill
description: A test skill
---
Test content`

	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test scanning
	skills := scanDirectory(filepath.Join(tmpDir, "skills"), "**/SKILL.md")
	if len(skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(skills))
		return
	}

	if skills[0].Name != "test-skill" {
		t.Errorf("Expected name 'test-skill', got %q", skills[0].Name)
	}
}

func TestScanDirectory_NonexistentDir(t *testing.T) {
	skills := scanDirectory("/nonexistent/path", "**/SKILL.md")
	if len(skills) != 0 {
		t.Errorf("Expected 0 skills for nonexistent directory, got %d", len(skills))
	}
}

func TestIsClaudeSkillsDisabled(t *testing.T) {
	// Save original value
	orig := os.Getenv("OPENCODE_DISABLE_CLAUDE_SKILLS")
	defer os.Setenv("OPENCODE_DISABLE_CLAUDE_SKILLS", orig)

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"disabled", "true", true},
		{"enabled empty", "", false},
		{"enabled other", "false", false},
		{"enabled random", "yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OPENCODE_DISABLE_CLAUDE_SKILLS", tt.value)
			got := isClaudeSkillsDisabled()
			if got != tt.want {
				t.Errorf("isClaudeSkillsDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfoStruct(t *testing.T) {
	info := Info{
		Name:          "test-skill",
		Description:   "A test skill",
		License:       "MIT",
		Compatibility: "opencode",
		Metadata: map[string]any{
			"audience": "developers",
			"version":  1.0,
		},
		Location: "/path/to/skill",
		Content:  "Skill content here",
	}

	if info.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", info.Name, "test-skill")
	}
	if info.License != "MIT" {
		t.Errorf("License = %q, want %q", info.License, "MIT")
	}
	if info.Metadata["audience"] != "developers" {
		t.Errorf("Metadata[audience] = %v, want %v", info.Metadata["audience"], "developers")
	}
}

func TestValidateFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		skill   *Info
		wantErr bool
	}{
		{
			name: "valid skill",
			skill: &Info{
				Name:        "valid-skill",
				Description: "A valid description",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			skill: &Info{
				Name:        "",
				Description: "Has description",
			},
			wantErr: true,
		},
		{
			name: "empty description",
			skill: &Info{
				Name:        "some-name",
				Description: "",
			},
			wantErr: true,
		},
		{
			name: "invalid name format",
			skill: &Info{
				Name:        "Invalid_Name",
				Description: "Has description",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFrontmatter(tt.skill)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscoverGlobalSkills(t *testing.T) {
	// This test verifies the function doesn't panic
	// Actual discovery depends on home directory structure
	skills := discoverGlobalSkills()
	// Should return a slice (possibly empty)
	if skills == nil {
		t.Error("discoverGlobalSkills() should not return nil")
	}
}

func TestAll(t *testing.T) {
	// Invalidate cache to ensure clean state
	Invalidate()

	// All() should return a slice (possibly empty)
	skills := All()
	if skills == nil {
		t.Error("All() should not return nil")
	}
}

func TestGet_NotFound(t *testing.T) {
	// Invalidate cache
	Invalidate()

	_, err := Get("nonexistent-skill-xyz")
	if err == nil {
		t.Error("Get() should return error for nonexistent skill")
	}
	if !errors.Is(err, ErrSkillNotFound) {
		t.Errorf("Get() error should wrap ErrSkillNotFound, got %v", err)
	}
}

func TestValidateName_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"single char", "a", false},
		{"numbers only", "123", false},
		{"max length", string(make([]byte, 64)), false},
		{"unicode", "skill-日本語", true},
		{"dot", "skill.name", true},
		{"slash", "skill/name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill max length test with valid chars
			input := tt.input
			if tt.name == "max length" {
				for i := range input {
					input = input[:i] + "a" + input[i+1:]
				}
			}

			err := validateName(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateName(%q) error = %v, wantErr %v", input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"single char", "a", false},
		{"whitespace only", "   ", false},
		{"unicode", "日本語の説明", false},
		{"max length", string(make([]byte, 1024)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDescription(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
