// Package sse provides SSE (Server-Sent Events) decoding for Kilo Gateway's
// non-standard streaming protocol with concatenated data:, reasoning, and
// fragmented tool_calls.
package sse

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"github.com/MerrukTechnology/OpenCode-Native/internal/message"
)

// EventType represents the type of SSE event
type EventType string

const (
	EventContentDelta  EventType = "content_delta"
	EventThinkingDelta EventType = "thinking_delta"
	EventToolUseStart  EventType = "tool_use_start"
	EventToolUseDelta  EventType = "tool_use_delta"
	EventComplete      EventType = "complete"
	EventError         EventType = "error"
)

// SSEEvent represents a decoded SSE event
type SSEEvent struct {
	Type     EventType
	Content  string
	Thinking string
	ToolCall *message.ToolCall
	Error    error
	Response *SSECompleteResponse
}

// SSECompleteResponse contains the final response data
type SSECompleteResponse struct {
	FinishReason message.FinishReason
	Usage        TokenUsage
}

// TokenUsage represents token usage information
type TokenUsage struct {
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
}

// KiloDecoder handles Kilo Gateway's non-standard SSE (concatenated data:, reasoning, fragmented tool_calls)
type KiloDecoder struct {
	cfg config.Config
}

// NewKiloDecoder creates a new KiloDecoder instance.
func NewKiloDecoder() *KiloDecoder {
	return &KiloDecoder{cfg: *config.Get()}
}

// Parse reads SSE events from the body and returns a channel of SSE events.
func (d *KiloDecoder) Parse(body io.Reader) <-chan SSEEvent {
	eventChan := make(chan SSEEvent)
	go d.parse(body, eventChan)
	return eventChan
}

func (d *KiloDecoder) parse(body io.Reader, eventChan chan<- SSEEvent) {
	defer close(eventChan)
	reader := bufio.NewReader(body)

	// We keep track of the last usage we saw, because sometimes
	// usage comes in a chunk before the final [DONE]
	var accumulatedUsage TokenUsage

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				eventChan <- SSEEvent{Type: EventError, Error: err}
			}
			return
		}

		line = strings.TrimSpace(line)

		// 1. Skip Kilo Processing logs and empty lines
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// 2. Extract standard data lines
		if !strings.HasPrefix(line, "data: ") {
			// Not SSE format - log this for debugging (could be normal JSON response)
			logging.Warn("Kilo SSE: Received non-SSE data", "line", line)
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		// 3. Unmarshal
		var streamResp kiloStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			// Always log malformed JSON to help debug tool call issues
			logging.Warn("Kilo SSE: Malformed JSON in stream", "data", data, "error", err.Error())
			continue
		}

		// 4. Handle Usage (Critical: This often comes in a chunk with 0 choices)
		if streamResp.Usage != nil {
			accumulatedUsage = TokenUsage{
				InputTokens:  streamResp.Usage.PromptTokens,
				OutputTokens: streamResp.Usage.CompletionTokens,
				TotalTokens:  streamResp.Usage.TotalTokens,
			}
			// If this chunk has no choices, it's likely a final usage update.
			// We can emit a partial complete or store it for the finish event.
		}

		// 5. If no choices, skip delta processing (but we captured usage above)
		if len(streamResp.Choices) == 0 {
			continue
		}

		choice := streamResp.Choices[0]
		delta := choice.Delta

		// --- Handle Content ---
		// Check for null content explicitly to avoid sending "null" strings
		if delta.Content != nil && *delta.Content != "" {
			eventChan <- SSEEvent{Type: EventContentDelta, Content: *delta.Content}
		}

		// --- Handle Reasoning/Thinking ---
		if delta.Reasoning != "" {
			eventChan <- SSEEvent{Type: EventThinkingDelta, Thinking: delta.Reasoning}
		}

		// --- Handle Tool Calls ---
		for _, tc := range delta.ToolCalls {
			if tc == nil {
				continue
			}
			idx := tc.Index

			// Case A: ID is present -> Start of tool call
			if tc.ID != "" {
				toolCall := &message.ToolCall{
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Type:  "function",
					Index: idx,
				}
				eventChan <- SSEEvent{Type: EventToolUseStart, ToolCall: toolCall}
				// Always log tool use start for debugging
				logging.Debug("Kilo SSE: ToolUseStart", "id", tc.ID, "name", tc.Function.Name, "index", idx)
			}

			// Case B: Arguments are present (even empty string) -> Delta
			if tc.Function != nil {
				// We explicitly allow empty string arguments to pass through
				// so the accumulator knows the field exists
				eventChan <- SSEEvent{
					Type: EventToolUseDelta,
					ToolCall: &message.ToolCall{
						Index: idx,
						Input: tc.Function.Arguments,
					},
				}
				// Always log tool use delta for debugging
				logging.Debug("Kilo SSE: ToolUseDelta", "index", idx, "input", tc.Function.Arguments)
			}
		}

		// --- Handle Finish Reason ---
		if choice.FinishReason != "" {
			fr := message.FinishReasonEndTurn
			switch choice.FinishReason {
			case "tool_calls":
				fr = message.FinishReasonToolUse
			case "length":
				fr = message.FinishReasonMaxTokens
			case "stop":
				fr = message.FinishReasonEndTurn
			}

			eventChan <- SSEEvent{
				Type: EventComplete,
				Response: &SSECompleteResponse{
					FinishReason: fr,
					Usage:        accumulatedUsage,
				},
			}
		}
	}
}

// extractSSEDataEvents handles Kilo's known corruption (concatenated data: events).
// Uses strings.Cut for modern Go 1.24+ idioms.
func extractSSEDataEvents(rawLine string) []string {
	rawLine = strings.TrimSpace(rawLine)
	if !strings.HasPrefix(rawLine, "data: ") {
		return nil
	}

	data := strings.TrimPrefix(rawLine, "data: ")
	var events []string

	for {
		data = strings.TrimSpace(data)
		if data == "" || data == "[DONE]" {
			break
		}

		// Use strings.Cut instead of strings.Index for modern Go idioms
		prefix, rest, found := strings.Cut(data, "data: ")
		if !found {
			events = append(events, data)
			break
		}

		// prefix contains everything before "data: " - that's one event
		chunk := strings.TrimSpace(prefix)
		if chunk != "" && chunk != "[DONE]" {
			events = append(events, chunk)
		}
		// rest contains everything after "data: " - continue processing
		data = rest
	}
	return events
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// SSE Response Structures (shared across Kilo implementations)

type kiloStreamResponse struct {
	ID      string             `json:"id"`
	Choices []kiloStreamChoice `json:"choices"`
	Usage   *kiloUsage         `json:"usage,omitempty"`
}

type kiloStreamChoice struct {
	Index        int       `json:"index"`
	Delta        kiloDelta `json:"delta"`
	FinishReason string    `json:"finish_reason"`
}

type kiloDelta struct {
	// Pointer to string handles "content": null in JSON
	Content   *string          `json:"content"`
	Role      string           `json:"role,omitempty"`
	Reasoning string           `json:"reasoning,omitempty"`
	ToolCalls []*kiloDeltaTool `json:"tool_calls,omitempty"`
}

type kiloDeltaTool struct {
	Index    int                `json:"index"`
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	Function *kiloDeltaFunction `json:"function,omitempty"`
}

type kiloDeltaFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"` // Note: Received as 'arguments', not 'parameters'
}

type kiloUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}
