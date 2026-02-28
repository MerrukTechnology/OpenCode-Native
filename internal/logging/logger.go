// Package logging provides structured logging functionality for the application.
// It wraps the standard slog package with additional features like session-based
// log persistence, panic recovery, and message directory management.
package logging

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

func getCaller() string {
	var caller string
	if _, file, line, ok := runtime.Caller(2); ok {
		// caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
		caller = fmt.Sprintf("%s:%d", file, line)
	} else {
		caller = "unknown"
	}
	return caller
}

// Info logs a message at INFO level with the caller's source location.
func Info(msg string, args ...any) {
	source := getCaller()
	slog.Info(msg, append([]any{"source", source}, args...)...)
}

// Debug logs a message at DEBUG level with the caller's source location.
func Debug(msg string, args ...any) {
	source := getCaller()
	slog.Debug(msg, append([]any{"source", source}, args...)...)
}

// Warn logs a message at WARN level with the caller's source location.
func Warn(msg string, args ...any) {
	source := getCaller()
	slog.Warn(msg, append([]any{"source", source}, args...)...)
}

// Error logs a message at ERROR level with the caller's source location.
func Error(msg string, args ...any) {
	source := getCaller()
	slog.Error(msg, append([]any{"source", source}, args...)...)
}

func logPersist(level slog.Level, msg string, args ...any) {
	args = append(args, persistKeyArg, true)
	switch level {
	case slog.LevelInfo:
		slog.Info(msg, args...)
	case slog.LevelDebug:
		slog.Debug(msg, args...)
	case slog.LevelWarn:
		slog.Warn(msg, args...)
	case slog.LevelError:
		slog.Error(msg, args...)
	default:
		slog.Info(msg, args...)
	}
}

// InfoPersist logs a message at INFO level that should be persisted to the status bar.
func InfoPersist(msg string, args ...any) {
	logPersist(slog.LevelInfo, msg, args...)
}

// DebugPersist logs a message at DEBUG level that should be persisted to the status bar.
func DebugPersist(msg string, args ...any) {
	logPersist(slog.LevelDebug, msg, args...)
}

// WarnPersist logs a message at WARN level that should be persisted to the status bar.
func WarnPersist(msg string, args ...any) {
	logPersist(slog.LevelWarn, msg, args...)
}

// ErrorPersist logs a message at ERROR level that should be persisted to the status bar.
func ErrorPersist(msg string, args ...any) {
	logPersist(slog.LevelError, msg, args...)
}

// RecoverPanic is a common function to handle panics gracefully.
// It logs the error, creates a panic log file with stack trace,
// and executes an optional cleanup function before returning.
func RecoverPanic(name string, cleanup func()) {
	if r := recover(); r != nil {
		// Log the panic
		ErrorPersist(fmt.Sprintf("Panic in %s: %v", name, r))

		// Create a timestamped panic log file
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("opencode-panic-%s-%s.log", name, timestamp)

		file, err := os.Create(filename)
		if err != nil {
			ErrorPersist(fmt.Sprintf("Failed to create panic log: %v", err))
		} else {
			defer file.Close()

			// Write panic information and stack trace
			fmt.Fprintf(file, "Panic in %s: %v\n\n", name, r)
			fmt.Fprintf(file, "Time: %s\n\n", time.Now().Format(time.RFC3339))
			fmt.Fprintf(file, "Stack Trace:\n%s\n", debug.Stack())

			InfoPersist("Panic details written to " + filename)
		}

		// Execute cleanup function if provided
		if cleanup != nil {
			cleanup()
		}
	}
}

// MessageDir is the directory where session messages are stored.
var (
	MessageDir   string
	messageDirMu sync.RWMutex
)

// SetMessageDir sets the directory where session messages are stored.
func SetMessageDir(dir string) {
	messageDirMu.Lock()
	defer messageDirMu.Unlock()
	MessageDir = dir
}

// GetMessageDir returns the directory where session messages are stored.
func GetMessageDir() string {
	messageDirMu.RLock()
	defer messageDirMu.RUnlock()
	return MessageDir
}

// GetSessionPrefix returns a shortened version of the session ID for use in filenames.
// If the session ID is shorter than 8 characters, it returns the full ID.
// If the session ID is empty, it returns an empty string.
func GetSessionPrefix(sessionId string) string {
	// Ensure we don't slice beyond the string length.
	if len(sessionId) == 0 {
		return ""
	}
	if len(sessionId) < 8 {
		return sessionId
	}
	return sessionId[:8]
}

var sessionLogMutex sync.Mutex

// AppendToSessionLogFile appends content to a session log file.
// It returns the absolute path to the file on success, or an empty string on failure.
// The function validates that the resulting path stays within the message directory
// to prevent path traversal attacks.
func AppendToSessionLogFile(sessionId string, filename string, content string) string {
	if GetMessageDir() == "" || sessionId == "" {
		return ""
	}
	sessionPrefix := GetSessionPrefix(sessionId)

	safeSessionPrefix := filepath.Base(sessionPrefix)
	safeFilename := filepath.Base(filename)

	sessionLogMutex.Lock()
	defer sessionLogMutex.Unlock()

	sessionPath := filepath.Join(MessageDir, safeSessionPrefix)

	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		// 0o755 is more standard/secure than 0o766 (removes world-write)
		if err := os.MkdirAll(sessionPath, 0o755); err != nil {
			Error("Failed to create session directory", "dirpath", sessionPath, "error", err)
			return ""
		}
	}

	filePath := filepath.Join(sessionPath, safeFilename)

	absMessageDir, err := filepath.Abs(MessageDir)
	if err != nil {
		Error("Failed to resolve absolute path for MessageDir", "messagedir", MessageDir, "error", err)
		return ""
	}

	absFinalPath, err := filepath.Abs(filePath)
	if err != nil {
		Error("Failed to resolve absolute path for session log file", "filepath", filePath, "error", err)
		return ""
	}

	if !strings.HasPrefix(absFinalPath, absMessageDir) {
		Error("Security violation: Path traversal detected", "path", filePath)
		return ""
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		Error("Failed to open session log file", "filepath", filePath, "error", err)
		return ""
	}

	// Append chunk to file
	_, err = f.WriteString(content)
	if err != nil {
		Error("Failed to write chunk to session log file", "filepath", filePath, "error", err)
		// Best effort close; log if it fails but prioritize the write error.
		if cerr := f.Close(); cerr != nil {
			Error("Failed to close session log file after write error", "filepath", filePath, "error", cerr)
		}
		return ""
	}

	// Ensure data is flushed and close the file, handling any error explicitly.
	if err := f.Close(); err != nil {
		Error("Failed to close session log file", "filepath", filePath, "error", err)
		return ""
	}

	return filePath
}

// marshalAndWrite is a helper that marshals data to JSON and returns the string.
// It returns an empty string if MessageDir is not set, sessionId is empty, or requestSeqId is <= 0.
func marshalAndWrite(sessionId string, requestSeqId int, data any, errorMsg string) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	dataJson, err := json.Marshal(data)
	if err != nil {
		Error(errorMsg, "session_id", sessionId, "request_seq_id", requestSeqId, "error", err)
		return ""
	}
	return string(dataJson)
}

// WriteRequestMessageJson marshals the message to JSON and writes it to a session log file.
func WriteRequestMessageJson(sessionId string, requestSeqId int, message any) string {
	jsonStr := marshalAndWrite(sessionId, requestSeqId, message, "Failed to marshal message")
	if jsonStr == "" {
		return ""
	}
	return WriteRequestMessage(sessionId, requestSeqId, jsonStr)
}

// WriteRequestMessage writes a request message to a session log file.
func WriteRequestMessage(sessionId string, requestSeqId int, message string) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	filename := fmt.Sprintf("%d_request.json", requestSeqId)

	return AppendToSessionLogFile(sessionId, filename, message)
}

// AppendToStreamSessionLogJson marshals the chunk to JSON and appends it to a session stream log.
func AppendToStreamSessionLogJson(sessionId string, requestSeqId int, jsonableChunk any) string {
	jsonStr := marshalAndWrite(sessionId, requestSeqId, jsonableChunk, "Failed to marshal message")
	if jsonStr == "" {
		return ""
	}
	return AppendToStreamSessionLog(sessionId, requestSeqId, jsonStr)
}

// AppendToStreamSessionLog appends a chunk to a session stream log file.
func AppendToStreamSessionLog(sessionId string, requestSeqId int, chunk string) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	filename := fmt.Sprintf("%d_response_stream.log", requestSeqId)
	return AppendToSessionLogFile(sessionId, filename, chunk)
}

// WriteChatResponseJson marshals the response to JSON and writes it to a session log file.
func WriteChatResponseJson(sessionId string, requestSeqId int, response any) string {
	jsonStr := marshalAndWrite(sessionId, requestSeqId, response, "Failed to marshal response")
	if jsonStr == "" {
		return ""
	}
	filename := fmt.Sprintf("%d_response.json", requestSeqId)
	return AppendToSessionLogFile(sessionId, filename, jsonStr)
}

// WriteToolResultsJson marshals the tool results to JSON and writes them to a session log file.
func WriteToolResultsJson(sessionId string, requestSeqId int, toolResults any) string {
	jsonStr := marshalAndWrite(sessionId, requestSeqId, toolResults, "Failed to marshal tool results")
	if jsonStr == "" {
		return ""
	}
	filename := fmt.Sprintf("%d_tool_results.json", requestSeqId)
	return AppendToSessionLogFile(sessionId, filename, jsonStr)
}
