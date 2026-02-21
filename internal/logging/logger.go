package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"encoding/json"
	"path/filepath"
	"runtime"
	"runtime/debug"
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

func Info(msg string, args ...any) {
	source := getCaller()
	slog.Info(msg, append([]any{"source", source}, args...)...)
}

func Debug(msg string, args ...any) {
	source := getCaller()
	slog.Debug(msg, append([]any{"source", source}, args...)...)
}

func Warn(msg string, args ...any) {
	source := getCaller()
	slog.Warn(msg, append([]any{"source", source}, args...)...)
}

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

func InfoPersist(msg string, args ...any) {
	logPersist(slog.LevelInfo, msg, args...)
}

func DebugPersist(msg string, args ...any) {
	logPersist(slog.LevelDebug, msg, args...)
}

func WarnPersist(msg string, args ...any) {
	logPersist(slog.LevelWarn, msg, args...)
}

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

			InfoPersist(fmt.Sprintf("Panic details written to %s", filename))
		}

		// Execute cleanup function if provided
		if cleanup != nil {
			cleanup()
		}
	}
}

// Message Logging for Debug
var (
	MessageDir   string
	messageDirMu sync.RWMutex
)

func SetMessageDir(dir string) {
	messageDirMu.Lock()
	defer messageDirMu.Unlock()
	MessageDir = dir
}
func GetMessageDir() string {
	messageDirMu.RLock()
	defer messageDirMu.RUnlock()
	return MessageDir
}

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
		// Best effort close; log if it fails but prioritize the write error.
		if cerr := f.Close(); cerr != nil {
			Error("Failed to close session log file after write error", "filepath", filePath, "error", cerr)
		}
		if cerr := f.Close(); cerr != nil {
			Error("Failed to close session log file after write error", "filepath", filePath, "error", cerr)

	// Ensure data is flushed and close the file, handling any error explicitly.
	if err := f.Close(); err != nil {
		Error("Failed to close session log file", "filepath", filePath, "error", err)
		return ""
	}

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

func WriteRequestMessageJson(sessionId string, requestSeqId int, message any) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	msgJson, err := json.Marshal(message)
	if err != nil {
		Error("Failed to marshal message", "session_id", sessionId, "request_seq_id", requestSeqId, "error", err)
		return ""
	}
	return WriteRequestMessage(sessionId, requestSeqId, string(msgJson))
}

func WriteRequestMessage(sessionId string, requestSeqId int, message string) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	filename := fmt.Sprintf("%d_request.json", requestSeqId)

	return AppendToSessionLogFile(sessionId, filename, message)
}

func AppendToStreamSessionLogJson(sessionId string, requestSeqId int, jsonableChunk any) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	chunkJson, err := json.Marshal(jsonableChunk)
	if err != nil {
		Error("Failed to marshal message", "session_id", sessionId, "request_seq_id", requestSeqId, "error", err)
		return ""
	}
	return AppendToStreamSessionLog(sessionId, requestSeqId, string(chunkJson))
}

func AppendToStreamSessionLog(sessionId string, requestSeqId int, chunk string) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	filename := fmt.Sprintf("%d_response_stream.log", requestSeqId)
	return AppendToSessionLogFile(sessionId, filename, chunk)
}

func WriteChatResponseJson(sessionId string, requestSeqId int, response any) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		Error("Failed to marshal response", "session_id", sessionId, "request_seq_id", requestSeqId, "error", err)
		return ""
	}
	filename := fmt.Sprintf("%d_response.json", requestSeqId)

	return AppendToSessionLogFile(sessionId, filename, string(responseJson))
}

func WriteToolResultsJson(sessionId string, requestSeqId int, toolResults any) string {
	if MessageDir == "" || sessionId == "" || requestSeqId <= 0 {
		return ""
	}
	toolResultsJson, err := json.Marshal(toolResults)
	if err != nil {
		Error("Failed to marshal tool results", "session_id", sessionId, "request_seq_id", requestSeqId, "error", err)
		return ""
	}
	filename := fmt.Sprintf("%d_tool_results.json", requestSeqId)
	return AppendToSessionLogFile(sessionId, filename, string(toolResultsJson))
}
