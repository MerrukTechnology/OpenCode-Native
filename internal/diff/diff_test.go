package diff

import (
	"testing"
)

// TestDataStructureFields tests that data structures have expected field values
func TestDataStructureFields(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "Segment",
			fn: func(t *testing.T) {
				segment := Segment{
					Start: 0,
					End:   10,
					Type:  LineAdded,
					Text:  "test text",
				}
				if segment.Start != 0 {
					t.Errorf("Start = %d, want 0", segment.Start)
				}
				if segment.End != 10 {
					t.Errorf("End = %d, want 10", segment.End)
				}
				if segment.Type != LineAdded {
					t.Errorf("Type = %d, want %d", segment.Type, LineAdded)
				}
				if segment.Text != "test text" {
					t.Errorf("Text = %q, want %q", segment.Text, "test text")
				}
			},
		},
		{
			name: "DiffLine",
			fn: func(t *testing.T) {
				line := DiffLine{
					OldLineNo: 1,
					NewLineNo: 2,
					Kind:      LineAdded,
					Content:   "added line",
					Segments:  []Segment{{Start: 0, End: 5, Type: LineAdded, Text: "added"}},
				}
				if line.OldLineNo != 1 {
					t.Errorf("OldLineNo = %d, want 1", line.OldLineNo)
				}
				if line.NewLineNo != 2 {
					t.Errorf("NewLineNo = %d, want 2", line.NewLineNo)
				}
				if line.Kind != LineAdded {
					t.Errorf("Kind = %d, want %d", line.Kind, LineAdded)
				}
				if line.Content != "added line" {
					t.Errorf("Content = %q, want %q", line.Content, "added line")
				}
				if len(line.Segments) != 1 {
					t.Errorf("Segments length = %d, want 1", len(line.Segments))
				}
			},
		},
		{
			name: "Hunk",
			fn: func(t *testing.T) {
				hunk := Hunk{
					Header: "@@ -1,3 +1,4 @@",
					Lines: []DiffLine{
						{OldLineNo: 1, NewLineNo: 1, Kind: LineContext, Content: "context"},
						{OldLineNo: 2, NewLineNo: 0, Kind: LineRemoved, Content: "removed"},
						{OldLineNo: 0, NewLineNo: 2, Kind: LineAdded, Content: "added"},
					},
				}
				if hunk.Header != "@@ -1,3 +1,4 @@" {
					t.Errorf("Header = %q, want %q", hunk.Header, "@@ -1,3 +1,4 @@")
				}
				if len(hunk.Lines) != 3 {
					t.Errorf("Lines length = %d, want 3", len(hunk.Lines))
				}
			},
		},
		{
			name: "DiffResult",
			fn: func(t *testing.T) {
				result := DiffResult{
					OldFile: "old.txt",
					NewFile: "new.txt",
					Hunks: []Hunk{
						{Header: "@@ -1,3 +1,4 @@"},
					},
				}
				if result.OldFile != "old.txt" {
					t.Errorf("OldFile = %q, want %q", result.OldFile, "old.txt")
				}
				if result.NewFile != "new.txt" {
					t.Errorf("NewFile = %q, want %q", result.NewFile, "new.txt")
				}
				if len(result.Hunks) != 1 {
					t.Errorf("Hunks length = %d, want 1", len(result.Hunks))
				}
			},
		},
		{
			name: "FileChange",
			fn: func(t *testing.T) {
				oldContent := "old"
				newContent := "new"
				movePath := "moved.txt"

				fc := FileChange{
					Type:       ActionUpdate,
					OldContent: &oldContent,
					NewContent: &newContent,
					MovePath:   &movePath,
				}

				if fc.Type != ActionUpdate {
					t.Errorf("Type = %q, want %q", fc.Type, ActionUpdate)
				}
				if fc.OldContent == nil || *fc.OldContent != "old" {
					t.Errorf("OldContent not set correctly")
				}
				if fc.NewContent == nil || *fc.NewContent != "new" {
					t.Errorf("NewContent not set correctly")
				}
				if fc.MovePath == nil || *fc.MovePath != "moved.txt" {
					t.Errorf("MovePath not set correctly")
				}
			},
		},
		{
			name: "Commit",
			fn: func(t *testing.T) {
				commit := Commit{
					Changes: make(map[string]FileChange),
				}
				commit.Changes["test.txt"] = FileChange{Type: ActionAdd}

				if len(commit.Changes) != 1 {
					t.Errorf("Changes length = %d, want 1", len(commit.Changes))
				}
				if commit.Changes["test.txt"].Type != ActionAdd {
					t.Errorf("Change type = %q, want %q", commit.Changes["test.txt"].Type, ActionAdd)
				}
			},
		},
		{
			name: "Chunk",
			fn: func(t *testing.T) {
				chunk := Chunk{
					OrigIndex: 5,
					DelLines:  []string{"line1", "line2"},
					InsLines:  []string{"newline"},
				}

				if chunk.OrigIndex != 5 {
					t.Errorf("OrigIndex = %d, want 5", chunk.OrigIndex)
				}
				if len(chunk.DelLines) != 2 {
					t.Errorf("DelLines length = %d, want 2", len(chunk.DelLines))
				}
				if len(chunk.InsLines) != 1 {
					t.Errorf("InsLines length = %d, want 1", len(chunk.InsLines))
				}
			},
		},
		{
			name: "PatchAction",
			fn: func(t *testing.T) {
				newFile := "new content"
				movePath := "new/path.txt"

				action := PatchAction{
					Type:     ActionUpdate,
					NewFile:  &newFile,
					Chunks:   []Chunk{{OrigIndex: 0, DelLines: []string{"old"}, InsLines: []string{"new"}}},
					MovePath: &movePath,
				}

				if action.Type != ActionUpdate {
					t.Errorf("Type = %q, want %q", action.Type, ActionUpdate)
				}
				if action.NewFile == nil || *action.NewFile != "new content" {
					t.Errorf("NewFile not set correctly")
				}
				if len(action.Chunks) != 1 {
					t.Errorf("Chunks length = %d, want 1", len(action.Chunks))
				}
				if action.MovePath == nil || *action.MovePath != "new/path.txt" {
					t.Errorf("MovePath not set correctly")
				}
			},
		},
		{
			name: "Patch",
			fn: func(t *testing.T) {
				patch := Patch{
					Actions: make(map[string]PatchAction),
				}
				patch.Actions["file.txt"] = PatchAction{Type: ActionAdd}

				if len(patch.Actions) != 1 {
					t.Errorf("Actions length = %d, want 1", len(patch.Actions))
				}
				if patch.Actions["file.txt"].Type != ActionAdd {
					t.Errorf("Action type = %q, want %q", patch.Actions["file.txt"].Type, ActionAdd)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

// TestConstants tests that constant values are defined correctly
func TestConstants(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "LineType",
			fn: func(t *testing.T) {
				tests := []struct {
					name string
					got  LineType
					want LineType
				}{
					{"context", LineContext, LineType(0)},
					{"added", LineAdded, LineType(1)},
					{"removed", LineRemoved, LineType(2)},
				}
				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						if tt.got < LineContext || tt.got > LineRemoved {
							t.Errorf("LineType value %d out of valid range", tt.got)
						}
					})
				}
			},
		},
		{
			name: "ActionType",
			fn: func(t *testing.T) {
				tests := []struct {
					name string
					got  string
					want string
				}{
					{"add", string(ActionAdd), "add"},
					{"delete", string(ActionDelete), "delete"},
					{"update", string(ActionUpdate), "update"},
				}
				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						if tt.got != tt.want {
							t.Errorf("got %q, want %q", tt.got, tt.want)
						}
					})
				}
			},
		},
		{
			name: "DiffError",
			fn: func(t *testing.T) {
				err := NewDiffError("test error")
				if err.Error() != "test error" {
					t.Errorf("Error() = %q, want %q", err.Error(), "test error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestParseUnifiedDiff(t *testing.T) {
	tests := []struct {
		name                string
		diff                string
		wantErr             bool
		wantOldFile         string
		wantNewFile         string
		wantHunks           int
		wantFirstHunkHeader string
	}{
		{
			name: "simple diff",
			diff: `--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,4 @@
 line1
-removed line
+added line
 line3`,
			wantErr:             false,
			wantOldFile:         "file.txt",
			wantNewFile:         "file.txt",
			wantHunks:           1,
			wantFirstHunkHeader: "@@ -1,3 +1,4 @@",
		},
		{
			name:      "empty diff",
			diff:      "",
			wantErr:   false,
			wantHunks: 0,
		},
		{
			name: "diff with no newline marker",
			diff: `--- a/file.txt
+++ b/file.txt
@@ -1,2 +1,2 @@
 line1
-removed
\ No newline at end of file
+added`,
			wantErr:     false,
			wantOldFile: "file.txt",
			wantNewFile: "file.txt",
			wantHunks:   1,
		},
		{
			name: "multiple hunks",
			diff: `--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 context
-removed
+added
 context
@@ -10,3 +10,3 @@
 more context
-old
+new
 end`,
			wantErr:     false,
			wantOldFile: "file.txt",
			wantNewFile: "file.txt",
			wantHunks:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseUnifiedDiff(tt.diff)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUnifiedDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.OldFile != tt.wantOldFile {
				t.Errorf("OldFile = %q, want %q", result.OldFile, tt.wantOldFile)
			}
			if result.NewFile != tt.wantNewFile {
				t.Errorf("NewFile = %q, want %q", result.NewFile, tt.wantNewFile)
			}
			if len(result.Hunks) != tt.wantHunks {
				t.Errorf("Hunks length = %d, want %d", len(result.Hunks), tt.wantHunks)
			}
			if tt.wantFirstHunkHeader != "" && len(result.Hunks) > 0 {
				if result.Hunks[0].Header != tt.wantFirstHunkHeader {
					t.Errorf("First hunk header = %q, want %q", result.Hunks[0].Header, tt.wantFirstHunkHeader)
				}
			}
		})
	}
}

func TestParseUnifiedDiff_LineNumbers(t *testing.T) {
	diff := `--- a/file.txt
+++ b/file.txt
@@ -1,5 +1,5 @@
 context1
 context2
-removed1
-removed2
+added1
+added2
 context3`

	result, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(result.Hunks) != 1 {
		t.Fatalf("Hunks length = %d, want 1", len(result.Hunks))
	}

	hunk := result.Hunks[0]
	if len(hunk.Lines) != 7 {
		t.Fatalf("Lines length = %d, want 7", len(hunk.Lines))
	}

	// Check context line numbers
	if hunk.Lines[0].Kind != LineContext {
		t.Errorf("Line 0 kind = %d, want %d", hunk.Lines[0].Kind, LineContext)
	}
	if hunk.Lines[0].OldLineNo != 1 || hunk.Lines[0].NewLineNo != 1 {
		t.Errorf("Line 0 numbers = (%d, %d), want (1, 1)", hunk.Lines[0].OldLineNo, hunk.Lines[0].NewLineNo)
	}

	// Check removed lines
	if hunk.Lines[2].Kind != LineRemoved {
		t.Errorf("Line 2 kind = %d, want %d", hunk.Lines[2].Kind, LineRemoved)
	}
	if hunk.Lines[2].OldLineNo != 3 || hunk.Lines[2].NewLineNo != 0 {
		t.Errorf("Line 2 numbers = (%d, %d), want (3, 0)", hunk.Lines[2].OldLineNo, hunk.Lines[2].NewLineNo)
	}

	// Check added lines
	if hunk.Lines[4].Kind != LineAdded {
		t.Errorf("Line 4 kind = %d, want %d", hunk.Lines[4].Kind, LineAdded)
	}
	if hunk.Lines[4].OldLineNo != 0 || hunk.Lines[4].NewLineNo != 3 {
		t.Errorf("Line 4 numbers = (%d, %d), want (0, 3)", hunk.Lines[4].OldLineNo, hunk.Lines[4].NewLineNo)
	}
}

func TestHighlightIntralineChanges(t *testing.T) {
	hunk := &Hunk{
		Header: "@@ -1,2 +1,2 @@",
		Lines: []DiffLine{
			{OldLineNo: 1, NewLineNo: 0, Kind: LineRemoved, Content: "hello world"},
			{OldLineNo: 0, NewLineNo: 1, Kind: LineAdded, Content: "hello universe"},
		},
	}

	HighlightIntralineChanges(hunk)

	// After highlighting, lines should have segments
	if len(hunk.Lines) != 2 {
		t.Errorf("Lines length = %d, want 2", len(hunk.Lines))
	}

	// Check that segments were added
	for i, line := range hunk.Lines {
		if len(line.Segments) == 0 {
			t.Errorf("Line %d has no segments after highlighting", i)
		}
	}
}

func TestHighlightIntralineChanges_NoChanges(t *testing.T) {
	hunk := &Hunk{
		Header: "@@ -1,3 +1,3 @@",
		Lines: []DiffLine{
			{OldLineNo: 1, NewLineNo: 1, Kind: LineContext, Content: "context"},
			{OldLineNo: 2, NewLineNo: 0, Kind: LineRemoved, Content: "removed"},
			{OldLineNo: 0, NewLineNo: 2, Kind: LineAdded, Content: "added"},
		},
	}

	HighlightIntralineChanges(hunk)

	// Context line should remain unchanged
	if hunk.Lines[0].Kind != LineContext {
		t.Errorf("Context line kind changed")
	}
}

// TestParseConfigOptions tests ParseConfig option functions
func TestParseConfigOptions(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "WithContextSize",
			fn: func(t *testing.T) {
				cfg := &ParseConfig{}
				WithContextSize(5)(cfg)
				if cfg.ContextSize != 5 {
					t.Errorf("ContextSize = %d, want 5", cfg.ContextSize)
				}
			},
		},
		{
			name: "WithContextSize_negativeIgnored",
			fn: func(t *testing.T) {
				cfg := &ParseConfig{}
				WithContextSize(-1)(cfg)
				if cfg.ContextSize != 0 {
					t.Errorf("ContextSize = %d, want 0 (negative should be ignored)", cfg.ContextSize)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

// TestSideBySideConfig tests SideBySideConfig functions
func TestSideBySideConfig(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "NewSideBySideConfig",
			fn: func(t *testing.T) {
				cfg := NewSideBySideConfig()
				if cfg.TotalWidth <= 0 {
					t.Errorf("TotalWidth = %d, want > 0", cfg.TotalWidth)
				}
			},
		},
		{
			name: "NewSideBySideConfig_WithOptions",
			fn: func(t *testing.T) {
				cfg := NewSideBySideConfig(WithTotalWidth(100))
				if cfg.TotalWidth != 100 {
					t.Errorf("TotalWidth = %d, want 100", cfg.TotalWidth)
				}
			},
		},
		{
			name: "WithTotalWidth",
			fn: func(t *testing.T) {
				cfg := &SideBySideConfig{}
				WithTotalWidth(200)(cfg)
				if cfg.TotalWidth != 200 {
					t.Errorf("TotalWidth = %d, want 200", cfg.TotalWidth)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

// TestIdentifyFiles tests file identification functions
func TestIdentifyFiles(t *testing.T) {
	tests := []struct {
		name      string
		fn        func(string) []string
		text      string
		wantCount int
	}{
		{
			name:      "IdentifyFilesNeeded_singleUpdate",
			fn:        IdentifyFilesNeeded,
			text:      "*** Begin Patch\n*** Update File: test.txt\ncontent\n*** End Patch",
			wantCount: 1,
		},
		{
			name:      "IdentifyFilesNeeded_singleDelete",
			fn:        IdentifyFilesNeeded,
			text:      "*** Begin Patch\n*** Delete File: test.txt\n*** End Patch",
			wantCount: 1,
		},
		{
			name:      "IdentifyFilesNeeded_multiple",
			fn:        IdentifyFilesNeeded,
			text:      "*** Begin Patch\n*** Update File: test1.txt\ncontent\n*** Update File: test2.txt\ncontent\n*** End Patch",
			wantCount: 2,
		},
		{
			name:      "IdentifyFilesNeeded_noFiles",
			fn:        IdentifyFilesNeeded,
			text:      "*** Begin Patch\n*** End Patch",
			wantCount: 0,
		},
		{
			name:      "IdentifyFilesNeeded_empty",
			fn:        IdentifyFilesNeeded,
			text:      "",
			wantCount: 0,
		},
		{
			name:      "IdentifyFilesAdded_single",
			fn:        IdentifyFilesAdded,
			text:      "*** Begin Patch\n*** Add File: test.txt\ncontent\n*** End Patch",
			wantCount: 1,
		},
		{
			name:      "IdentifyFilesAdded_multiple",
			fn:        IdentifyFilesAdded,
			text:      "*** Begin Patch\n*** Add File: test1.txt\ncontent\n*** Add File: test2.txt\ncontent\n*** End Patch",
			wantCount: 2,
		},
		{
			name:      "IdentifyFilesAdded_updateFile",
			fn:        IdentifyFilesAdded,
			text:      "*** Begin Patch\n*** Update File: test.txt\ncontent\n*** End Patch",
			wantCount: 0,
		},
		{
			name:      "IdentifyFilesAdded_empty",
			fn:        IdentifyFilesAdded,
			text:      "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := tt.fn(tt.text)
			if len(files) != tt.wantCount {
				t.Errorf("%s returned %d files, want %d", tt.name, len(files), tt.wantCount)
			}
		})
	}
}

// TestAssembleChanges tests the AssembleChanges function
func TestAssembleChanges(t *testing.T) {
	tests := []struct {
		name          string
		orig          map[string]string
		updated       map[string]string
		wantChangeCnt int
		wantFileType  string
		wantFileType2 string
		wantFileName  string
		wantFileName2 string
	}{
		{
			name: "updateAndAdd",
			orig: map[string]string{
				"file1.txt": "original content",
			},
			updated: map[string]string{
				"file1.txt": "updated content",
				"file2.txt": "new file",
			},
			wantChangeCnt: 2,
			wantFileType:  "update",
			wantFileType2: "add",
			wantFileName:  "file1.txt",
			wantFileName2: "file2.txt",
		},
		{
			name: "delete",
			orig: map[string]string{
				"file1.txt": "content",
				"file2.txt": "to be deleted",
			},
			updated: map[string]string{
				"file1.txt": "content",
				"file2.txt": "", // Empty content signals deletion
			},
			wantChangeCnt: 1,
			wantFileType:  "delete",
			wantFileName:  "file2.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit := AssembleChanges(tt.orig, tt.updated)

			if len(commit.Changes) != tt.wantChangeCnt {
				t.Errorf("Changes length = %d, want %d", len(commit.Changes), tt.wantChangeCnt)
			}

			if tt.wantFileName != "" {
				if change, ok := commit.Changes[tt.wantFileName]; ok {
					if string(change.Type) != tt.wantFileType {
						t.Errorf("%s type = %q, want %q", tt.wantFileName, change.Type, tt.wantFileType)
					}
				} else {
					t.Errorf("%s not found in changes", tt.wantFileName)
				}
			}

			if tt.wantFileName2 != "" {
				if change, ok := commit.Changes[tt.wantFileName2]; ok {
					if string(change.Type) != tt.wantFileType2 {
						t.Errorf("%s type = %q, want %q", tt.wantFileName2, change.Type, tt.wantFileType2)
					}
				} else {
					t.Errorf("%s not found in changes", tt.wantFileName2)
				}
			}
		})
	}
}

func TestNewParser(t *testing.T) {
	files := map[string]string{"test.txt": "content"}
	lines := []string{"line1", "line2"}

	parser := NewParser(files, lines)

	if parser.currentFiles == nil {
		t.Error("currentFiles should not be nil")
	}
	if len(parser.currentFiles) != 1 {
		t.Errorf("currentFiles length = %d, want 1", len(parser.currentFiles))
	}
	if parser.index != 0 {
		t.Errorf("index = %d, want 0", parser.index)
	}
	if parser.patch.Actions == nil {
		t.Error("patch.Actions should not be nil")
	}
}
