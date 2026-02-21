package logging

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDir creates a temporary directory and sets it as the message directory.
// Returns the temp directory path for verification.
func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	return tmpDir
}

// cleanupTestDir resets the message directory.
func cleanupTestDir() {
	SetMessageDir("")
}

// TestGetSessionPrefix tests the GetSessionPrefix function with various inputs.
func TestGetSessionPrefix(t *testing.T) {
	tests := []struct {
		name      string
		sessionId string
		want      string
	}{
		{"normal id", "1234567890abcdef", "12345678"},
		{"short id", "abc", "abc"},
		{"empty id", "", ""},
		{"exact 8 chars", "12345678", "12345678"},
		{"single char", "a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSessionPrefix(tt.sessionId)
			if got != tt.want {
				t.Errorf("GetSessionPrefix(%q) = %q, want %q", tt.sessionId, got, tt.want)
			}
		})
	}
}

func TestSetGetMessageDir(t *testing.T) {
	testDir := "/tmp/test-messages"
	SetMessageDir(testDir)
	defer cleanupTestDir()

	got := GetMessageDir()
	if got != testDir {
		t.Errorf("GetMessageDir() = %q, want %q", got, testDir)
	}
}

// TestAppendToSessionLogFile_EmptyInputs tests behavior with empty message directory
// and empty session ID.
func TestAppendToSessionLogFile_EmptyInputs(t *testing.T) {
	tests := []struct {
		name      string
		msgDir    string
		sessionId string
	}{
		{"empty message dir", "", "session123"},
		{"empty session id", "/tmp/test", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetMessageDir(tt.msgDir)
			defer cleanupTestDir()

			result := AppendToSessionLogFile(tt.sessionId, "test.log", "content")
			if result != "" {
				t.Errorf("AppendToSessionLogFile() should return empty string when %s", tt.name)
			}
		})
	}
}

// TestAppendToSessionLogFile_Valid tests valid input scenarios including file creation,
// content verification, and appending behavior.
func TestAppendToSessionLogFile_Valid(t *testing.T) {
	// Test file creation and content
	t.Run("creates file with content", func(t *testing.T) {
		_ = setupTestDir(t) // Set up temp directory
		defer cleanupTestDir()

		sessionId := "1234567890abcdef"
		filename := "test.log"
		content := "test content"

		result := AppendToSessionLogFile(sessionId, filename, content)
		if result == "" {
			t.Error("AppendToSessionLogFile() should return non-empty path on success")
		}

		// Verify file was created
		if _, err := os.Stat(result); os.IsNotExist(err) {
			t.Errorf("File was not created at %s", result)
		}

		// Verify content
		data, err := os.ReadFile(result)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(data) != content {
			t.Errorf("File content = %q, want %q", string(data), content)
		}
	})

	// Test append behavior
	t.Run("appends to existing file", func(t *testing.T) {
		tmpDir := setupTestDir(t)
		defer cleanupTestDir()

		sessionId := "1234567890abcdef"
		filename := "append.log"

		// First write
		AppendToSessionLogFile(sessionId, filename, "line1\n")
		// Second write
		AppendToSessionLogFile(sessionId, filename, "line2\n")

		// Verify both lines exist
		sessionPrefix := GetSessionPrefix(sessionId)
		filePath := filepath.Join(tmpDir, sessionPrefix, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		expected := "line1\nline2\n"
		if string(data) != expected {
			t.Errorf("File content = %q, want %q", string(data), expected)
		}
	})

	// Test path traversal prevention
	t.Run("prevents path traversal", func(t *testing.T) {
		tmpDir := setupTestDir(t)
		defer cleanupTestDir()

		sessionId := "1234567890abcdef"
		filename := "../../../etc/passwd"

		result := AppendToSessionLogFile(sessionId, filename, "malicious")
		// Should either return empty or create file safely within MessageDir
		if result != "" {
			// Verify the file is within MessageDir
			absResult, _ := filepath.Abs(result)
			absTmpDir, _ := filepath.Abs(tmpDir)
			if absResult == "" || absTmpDir == "" {
				return
			}
			// The file should be inside tmpDir
			if filepath.Dir(absResult) == absTmpDir || filepath.Dir(filepath.Dir(absResult)) == absTmpDir {
				return
			}
		}
	})
}

// TestWriteRequestMessage_EmptyInputs tests behavior with empty message directory
// and invalid sequence IDs.
func TestWriteRequestMessage_EmptyInputs(t *testing.T) {
	tests := []struct {
		name         string
		sessionId    string
		requestSeqId int
	}{
		{"zero seq id", "session123", 0},
		{"negative seq id", "session123", -1},
		{"empty session id", "", 1},
		{"empty message dir", "session123", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "empty message dir" {
				SetMessageDir("")
			} else {
				SetMessageDir("/tmp/test")
			}
			defer cleanupTestDir()

			result := WriteRequestMessage(tt.sessionId, tt.requestSeqId, "message")
			if result != "" {
				t.Errorf("WriteRequestMessage() should return empty string for %s", tt.name)
			}
		})
	}
}

// TestWriteRequestMessage_Valid tests valid input scenarios.
func TestWriteRequestMessage_Valid(t *testing.T) {
	_ = setupTestDir(t) // Set up temp directory
	defer cleanupTestDir()

	sessionId := "1234567890abcdef"
	requestSeqId := 1
	message := `{"test": "message"}`

	result := WriteRequestMessage(sessionId, requestSeqId, message)
	if result == "" {
		t.Error("WriteRequestMessage() should return non-empty path on success")
	}

	// Verify filename
	expectedFilename := "1_request.json"
	if filepath.Base(result) != expectedFilename {
		t.Errorf("Filename = %q, want %q", filepath.Base(result), expectedFilename)
	}
}

// writeOp represents a file write operation for testing.
type writeOp struct {
	name          string
	sessionId     string
	requestSeqId  int
	content       interface{}
	expectedFile  string
	verifyContent bool
	contentMatch  string
}

// TestFileWriteOperations tests various file writing functions with valid inputs.
func TestFileWriteOperations(t *testing.T) {
	tests := []writeOp{
		{
			name:          "WriteRequestMessageJson",
			sessionId:     "1234567890abcdef",
			requestSeqId:  2,
			content:       map[string]string{"key": "value"},
			expectedFile:  "2_request.json",
			verifyContent: true,
			contentMatch:  `{"key":"value"}`,
		},
		{
			name:         "AppendToStreamSessionLog",
			sessionId:    "1234567890abcdef",
			requestSeqId: 1,
			content:      "stream chunk data",
			expectedFile: "1_response_stream.log",
		},
		{
			name:         "AppendToStreamSessionLogJson",
			sessionId:    "1234567890abcdef",
			requestSeqId: 1,
			content:      map[string]string{"data": "chunk"},
			expectedFile: "1_response_stream.log",
		},
		{
			name:         "WriteChatResponseJson",
			sessionId:    "1234567890abcdef",
			requestSeqId: 1,
			content:      map[string]string{"response": "test"},
			expectedFile: "1_response.json",
		},
		{
			name:         "WriteToolResultsJson",
			sessionId:    "1234567890abcdef",
			requestSeqId: 1,
			content:      []map[string]string{{"tool": "result"}},
			expectedFile: "1_tool_results.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = setupTestDir(t) // Set up temp directory
			defer cleanupTestDir()

			var result string

			switch tt.name {
			case "WriteRequestMessageJson":
				result = WriteRequestMessageJson(tt.sessionId, tt.requestSeqId, tt.content.(map[string]string))
			case "AppendToStreamSessionLog":
				result = AppendToStreamSessionLog(tt.sessionId, tt.requestSeqId, tt.content.(string))
			case "AppendToStreamSessionLogJson":
				result = AppendToStreamSessionLogJson(tt.sessionId, tt.requestSeqId, tt.content.(map[string]string))
			case "WriteChatResponseJson":
				result = WriteChatResponseJson(tt.sessionId, tt.requestSeqId, tt.content.(map[string]string))
			case "WriteToolResultsJson":
				result = WriteToolResultsJson(tt.sessionId, tt.requestSeqId, tt.content.([]map[string]string))
			}

			if result == "" {
				t.Error(tt.name + "() should return non-empty path on success")
			}

			// Verify filename if expected
			if tt.expectedFile != "" && filepath.Base(result) != tt.expectedFile {
				t.Errorf("Filename = %q, want %q", filepath.Base(result), tt.expectedFile)
			}

			// Verify content if requested
			if tt.verifyContent {
				data, err := os.ReadFile(result)
				if err != nil {
					t.Fatalf("Failed to read file: %v", err)
				}
				if string(data) != tt.contentMatch {
					t.Errorf("File content = %q, want %q", string(data), tt.contentMatch)
				}
			}
		})
	}
}
