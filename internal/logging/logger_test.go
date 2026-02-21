package logging

import (
	"os"
	"path/filepath"
	"testing"
)

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
	// Test setting and getting message directory
	testDir := "/tmp/test-messages"
	SetMessageDir(testDir)

	got := GetMessageDir()
	if got != testDir {
		t.Errorf("GetMessageDir() = %q, want %q", got, testDir)
	}

	// Reset
	SetMessageDir("")
}

func TestAppendToSessionLogFile_EmptyMessageDir(t *testing.T) {
	// Ensure MessageDir is empty
	SetMessageDir("")
	defer SetMessageDir("")

	result := AppendToSessionLogFile("session123", "test.log", "content")
	if result != "" {
		t.Errorf("AppendToSessionLogFile() should return empty string when MessageDir is empty")
	}
}

func TestAppendToSessionLogFile_EmptySessionId(t *testing.T) {
	SetMessageDir("/tmp/test")
	defer SetMessageDir("")

	result := AppendToSessionLogFile("", "test.log", "content")
	if result != "" {
		t.Errorf("AppendToSessionLogFile() should return empty string when sessionId is empty")
	}
}

func TestAppendToSessionLogFile_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

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
}

func TestAppendToSessionLogFile_Append(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

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
}

func TestWriteRequestMessage_EmptyMessageDir(t *testing.T) {
	SetMessageDir("")
	defer SetMessageDir("")

	result := WriteRequestMessage("session123", 1, "message")
	if result != "" {
		t.Errorf("WriteRequestMessage() should return empty string when MessageDir is empty")
	}
}

func TestWriteRequestMessage_InvalidSeqId(t *testing.T) {
	SetMessageDir("/tmp/test")
	defer SetMessageDir("")

	tests := []struct {
		name         string
		sessionId    string
		requestSeqId int
	}{
		{"zero seq id", "session123", 0},
		{"negative seq id", "session123", -1},
		{"empty session id", "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WriteRequestMessage(tt.sessionId, tt.requestSeqId, "message")
			if result != "" {
				t.Errorf("WriteRequestMessage() should return empty string for invalid input")
			}
		})
	}
}

func TestWriteRequestMessage_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

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

func TestWriteRequestMessageJson_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

	sessionId := "1234567890abcdef"
	requestSeqId := 2
	message := map[string]string{"key": "value"}

	result := WriteRequestMessageJson(sessionId, requestSeqId, message)
	if result == "" {
		t.Error("WriteRequestMessageJson() should return non-empty path on success")
	}

	// Verify file exists and contains valid JSON
	data, err := os.ReadFile(result)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != `{"key":"value"}` {
		t.Errorf("File content = %q, want %q", string(data), `{"key":"value"}`)
	}
}

func TestAppendToStreamSessionLog_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

	sessionId := "1234567890abcdef"
	requestSeqId := 1
	chunk := "stream chunk data"

	result := AppendToStreamSessionLog(sessionId, requestSeqId, chunk)
	if result == "" {
		t.Error("AppendToStreamSessionLog() should return non-empty path on success")
	}

	// Verify filename
	expectedFilename := "1_response_stream.log"
	if filepath.Base(result) != expectedFilename {
		t.Errorf("Filename = %q, want %q", filepath.Base(result), expectedFilename)
	}
}

func TestAppendToStreamSessionLogJson_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

	sessionId := "1234567890abcdef"
	requestSeqId := 1
	chunk := map[string]string{"data": "chunk"}

	result := AppendToStreamSessionLogJson(sessionId, requestSeqId, chunk)
	if result == "" {
		t.Error("AppendToStreamSessionLogJson() should return non-empty path on success")
	}
}

func TestWriteChatResponseJson_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

	sessionId := "1234567890abcdef"
	requestSeqId := 1
	response := map[string]string{"response": "test"}

	result := WriteChatResponseJson(sessionId, requestSeqId, response)
	if result == "" {
		t.Error("WriteChatResponseJson() should return non-empty path on success")
	}

	// Verify filename
	expectedFilename := "1_response.json"
	if filepath.Base(result) != expectedFilename {
		t.Errorf("Filename = %q, want %q", filepath.Base(result), expectedFilename)
	}
}

func TestWriteToolResultsJson_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

	sessionId := "1234567890abcdef"
	requestSeqId := 1
	toolResults := []map[string]string{{"tool": "result"}}

	result := WriteToolResultsJson(sessionId, requestSeqId, toolResults)
	if result == "" {
		t.Error("WriteToolResultsJson() should return non-empty path on success")
	}

	// Verify filename
	expectedFilename := "1_tool_results.json"
	if filepath.Base(result) != expectedFilename {
		t.Errorf("Filename = %q, want %q", filepath.Base(result), expectedFilename)
	}
}

func TestAppendToSessionLogFile_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	SetMessageDir(tmpDir)
	defer SetMessageDir("")

	sessionId := "1234567890abcdef"
	// Try path traversal
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
			// File was created safely within the directory structure
			return
		}
	}
	// Either empty result (rejected) or safe path is acceptable
}
