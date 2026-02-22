package db

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuerier(t *testing.T) {
	// Test New function
	q := New(nil)
	assert.NotNil(t, q)
}

func TestQueriesStruct(t *testing.T) {
	// Test Queries struct initialization
	q := &Queries{}
	assert.NotNil(t, q)
}

func TestFileModel(t *testing.T) {
	file := File{
		ID:        "test-id",
		SessionID: "session-id",
		Path:      "/path/to/file.go",
		Content:   "file content",
		Version:   "1.0",
		CreatedAt: 1234567890,
		UpdatedAt: 1234567890,
	}

	assert.Equal(t, "test-id", file.ID)
	assert.Equal(t, "session-id", file.SessionID)
	assert.Equal(t, "/path/to/file.go", file.Path)
	assert.Equal(t, "file content", file.Content)
	assert.Equal(t, "1.0", file.Version)
	assert.Equal(t, int64(1234567890), file.CreatedAt)
	assert.Equal(t, int64(1234567890), file.UpdatedAt)
}

func TestMessageModel(t *testing.T) {
	msg := Message{
		ID:         "msg-id",
		SessionID:  "session-id",
		Role:       "user",
		Parts:      `,{"type":"text","text":"Hello"}`,
		Model:      sql.NullString{String: "gpt-4", Valid: true},
		CreatedAt:  1234567890,
		UpdatedAt:  1234567890,
		FinishedAt: sql.NullInt64{Int64: 1234567891, Valid: true},
	}

	assert.Equal(t, "msg-id", msg.ID)
	assert.Equal(t, "session-id", msg.SessionID)
	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, `,{"type":"text","text":"Hello"}`, msg.Parts)
	assert.True(t, msg.Model.Valid)
	assert.Equal(t, "gpt-4", msg.Model.String)
}

func TestSessionModel(t *testing.T) {
	session := Session{
		ID:               "session-id",
		ParentSessionID:  sql.NullString{String: "parent-id", Valid: true},
		Title:            "Test Session",
		MessageCount:     10,
		PromptTokens:     1000,
		CompletionTokens: 500,
		Cost:             0.05,
		UpdatedAt:        1234567890,
		CreatedAt:        1234567890,
		SummaryMessageID: sql.NullString{String: "summary-id", Valid: true},
		ProjectID:        sql.NullString{String: "project-id", Valid: true},
		RootSessionID:    sql.NullString{String: "root-id", Valid: true},
	}

	assert.Equal(t, "session-id", session.ID)
	assert.Equal(t, "Test Session", session.Title)
	assert.Equal(t, int64(10), session.MessageCount)
	assert.Equal(t, int64(1000), session.PromptTokens)
	assert.Equal(t, int64(500), session.CompletionTokens)
	assert.Equal(t, 0.05, session.Cost)
	assert.True(t, session.ParentSessionID.Valid)
	assert.True(t, session.SummaryMessageID.Valid)
	assert.True(t, session.ProjectID.Valid)
	assert.True(t, session.RootSessionID.Valid)
}

func TestQueriesWithTx(t *testing.T) {
	// Test WithTx method
	q := &Queries{}
	_ = q.WithTx(nil) // Just verify it doesn't panic
}
