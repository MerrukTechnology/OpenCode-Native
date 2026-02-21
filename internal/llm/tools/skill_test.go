package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSampleSkillFiles(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T) (dir string, limit int)
		expectedMin   int
		expectedMax   int
		checkExcludes func(t *testing.T, files []string)
		checkErr      func(t *testing.T, err error)
	}{
		{
			name: "empty directory returns empty slice",
			setup: func(t *testing.T) (string, int) {
				dir := t.TempDir()
				return dir, 10
			},
			expectedMin: 0,
			expectedMax: 0,
			checkExcludes: func(t *testing.T, files []string) {
				if len(files) != 0 {
					t.Errorf("expected empty slice, got %v", files)
				}
			},
			checkErr: func(t *testing.T, err error) {
				// No error expected
			},
		},
		{
			name: "filters out SKILL.md",
			setup: func(t *testing.T) (string, int) {
				dir := t.TempDir()
				// Create SKILL.md (should be excluded)
				os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# Skill"), 0644)
				// Create other files (should be included)
				os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content"), 0644)
				return dir, 10
			},
			expectedMin: 2,
			expectedMax: 2,
			checkExcludes: func(t *testing.T, files []string) {
				for _, f := range files {
					if filepath.Base(f) == "SKILL.md" {
						t.Error("SKILL.md should be excluded from results")
					}
				}
			},
			checkErr: func(t *testing.T, err error) {},
		},
		{
			name: "respects limit",
			setup: func(t *testing.T) (string, int) {
				dir := t.TempDir()
				// Create 5 files
				for i := 0; i < 5; i++ {
					os.WriteFile(filepath.Join(dir, "file"+string(rune('0'+i))+".txt"), []byte("content"), 0644)
				}
				return dir, 3
			},
			expectedMin:   0,
			expectedMax:   3,
			checkExcludes: func(t *testing.T, files []string) {},
			checkErr:      func(t *testing.T, err error) {},
		},
		{
			name: "includes subdirectory files",
			setup: func(t *testing.T) (string, int) {
				dir := t.TempDir()
				subDir := filepath.Join(dir, "subdir")
				os.Mkdir(subDir, 0755)
				os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("content"), 0644)
				return dir, 10
			},
			expectedMin: 1,
			expectedMax: 10,
			checkExcludes: func(t *testing.T, files []string) {
				found := false
				for _, f := range files {
					if filepath.Base(f) == "nested.txt" {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected to find nested.txt from subdirectory")
				}
			},
			checkErr: func(t *testing.T, err error) {},
		},
		{
			name: "non-existent directory returns empty",
			setup: func(t *testing.T) (string, int) {
				return "/nonexistent/path", 10
			},
			expectedMin: 0,
			expectedMax: 0,
			checkExcludes: func(t *testing.T, files []string) {
				if len(files) != 0 {
					t.Errorf("expected empty slice for non-existent dir, got %v", files)
				}
			},
			checkErr: func(t *testing.T, err error) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, limit := tt.setup(t)
			files := sampleSkillFiles(dir, limit)

			if len(files) < tt.expectedMin || len(files) > tt.expectedMax {
				t.Errorf("expected between %d and %d files, got %d", tt.expectedMin, tt.expectedMax, len(files))
			}

			tt.checkExcludes(t, files)
			tt.checkErr(t, nil)
		})
	}
}

func TestCollectFiles(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (dir string, limit int)
		expectedMin int
		expectedMax int
		checkFiles  func(t *testing.T, files []string)
	}{
		{
			name: "empty directory returns empty",
			setup: func(t *testing.T) (string, int) {
				dir := t.TempDir()
				return dir, 10
			},
			expectedMin: 0,
			expectedMax: 0,
			checkFiles: func(t *testing.T, files []string) {
				if len(files) != 0 {
					t.Errorf("expected empty slice, got %v", files)
				}
			},
		},
		{
			name: "collects files from nested directories",
			setup: func(t *testing.T) (string, int) {
				dir := t.TempDir()
				// Create nested structure
				os.MkdirAll(filepath.Join(dir, "level1", "level2"), 0755)
				os.WriteFile(filepath.Join(dir, "root.txt"), []byte("root"), 0644)
				os.WriteFile(filepath.Join(dir, "level1", "l1.txt"), []byte("l1"), 0644)
				os.WriteFile(filepath.Join(dir, "level1", "level2", "l2.txt"), []byte("l2"), 0644)
				return dir, 10
			},
			expectedMin: 3,
			expectedMax: 3,
			checkFiles: func(t *testing.T, files []string) {
				// Verify all files are present
				if len(files) != 3 {
					t.Errorf("expected 3 files, got %d: %v", len(files), files)
				}
			},
		},
		{
			name: "respects limit across nested directories",
			setup: func(t *testing.T) (string, int) {
				dir := t.TempDir()
				// Create many files across directories
				for i := 0; i < 5; i++ {
					os.MkdirAll(filepath.Join(dir, "dir"+string(rune('0'+i))), 0755)
					os.WriteFile(filepath.Join(dir, "dir"+string(rune('0'+i)), "file.txt"), []byte("content"), 0644)
				}
				return dir, 3
			},
			expectedMin: 0,
			expectedMax: 3,
			checkFiles: func(t *testing.T, files []string) {
				if len(files) > 3 {
					t.Errorf("expected at most 3 files, got %d", len(files))
				}
			},
		},
		{
			name: "non-existent directory returns empty",
			setup: func(t *testing.T) (string, int) {
				return "/nonexistent/path", 10
			},
			expectedMin: 0,
			expectedMax: 0,
			checkFiles: func(t *testing.T, files []string) {
				if len(files) != 0 {
					t.Errorf("expected empty slice for non-existent dir, got %v", files)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, limit := tt.setup(t)
			files := collectFiles(dir, limit)

			if len(files) < tt.expectedMin || len(files) > tt.expectedMax {
				t.Errorf("expected between %d and %d files, got %d", tt.expectedMin, tt.expectedMax, len(files))
			}

			tt.checkFiles(t, files)
		})
	}
}
