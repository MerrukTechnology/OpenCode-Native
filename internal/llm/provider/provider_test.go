package provider

import (
	"net/http"
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/models"
	"github.com/stretchr/testify/assert"
)

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		value EventType
		want  string
	}{
		{"EventContentStart", EventContentStart, "content_start"},
		{"EventToolUseStart", EventToolUseStart, "tool_use_start"},
		{"EventToolUseDelta", EventToolUseDelta, "tool_use_delta"},
		{"EventToolUseStop", EventToolUseStop, "tool_use_stop"},
		{"EventContentDelta", EventContentDelta, "content_delta"},
		{"EventThinkingDelta", EventThinkingDelta, "thinking_delta"},
		{"EventContentStop", EventContentStop, "content_stop"},
		{"EventComplete", EventComplete, "complete"},
		{"EventError", EventError, "error"},
		{"EventWarning", EventWarning, "warning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.value))
		})
	}
}

func TestTokenUsage(t *testing.T) {
	tests := []struct {
		name                string
		inputTokens         int64
		outputTokens        int64
		cacheCreationTokens int64
		cacheReadTokens     int64
		checkTokens         func(*testing.T, TokenUsage)
	}{
		{
			name:                "all tokens zero",
			inputTokens:         0,
			outputTokens:        0,
			cacheCreationTokens: 0,
			cacheReadTokens:     0,
			checkTokens: func(t *testing.T, tu TokenUsage) {
				assert.Equal(t, int64(0), tu.InputTokens)
				assert.Equal(t, int64(0), tu.OutputTokens)
			},
		},
		{
			name:                "with token values",
			inputTokens:         1000,
			outputTokens:        500,
			cacheCreationTokens: 100,
			cacheReadTokens:     200,
			checkTokens: func(t *testing.T, tu TokenUsage) {
				assert.Equal(t, int64(1000), tu.InputTokens)
				assert.Equal(t, int64(500), tu.OutputTokens)
				assert.Equal(t, int64(100), tu.CacheCreationTokens)
				assert.Equal(t, int64(200), tu.CacheReadTokens)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tu := TokenUsage{
				InputTokens:         tt.inputTokens,
				OutputTokens:        tt.outputTokens,
				CacheCreationTokens: tt.cacheCreationTokens,
				CacheReadTokens:     tt.cacheReadTokens,
			}
			tt.checkTokens(t, tu)
		})
	}
}

func TestProviderResponse(t *testing.T) {
	resp := ProviderResponse{
		Content:      "test content",
		ToolCalls:    nil,
		Usage:        TokenUsage{InputTokens: 100, OutputTokens: 50},
		FinishReason: "stop",
	}

	assert.Equal(t, "test content", resp.Content)
	assert.Equal(t, int64(100), resp.Usage.InputTokens)
	assert.Equal(t, int64(50), resp.Usage.OutputTokens)
}

func TestProviderEvent(t *testing.T) {
	tests := []struct {
		name       string
		eventType  EventType
		content    string
		thinking   string
		checkEvent func(*testing.T, ProviderEvent)
	}{
		{
			name:      "content start event",
			eventType: EventContentStart,
			content:   "Hello",
			checkEvent: func(t *testing.T, e ProviderEvent) {
				assert.Equal(t, EventContentStart, e.Type)
				assert.Equal(t, "Hello", e.Content)
			},
		},
		{
			name:      "thinking delta event",
			eventType: EventThinkingDelta,
			thinking:  " reasoning...",
			checkEvent: func(t *testing.T, e ProviderEvent) {
				assert.Equal(t, EventThinkingDelta, e.Type)
				assert.Equal(t, " reasoning...", e.Thinking)
			},
		},
		{
			name:      "error event",
			eventType: EventError,
			checkEvent: func(t *testing.T, e ProviderEvent) {
				assert.Equal(t, EventError, e.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := ProviderEvent{
				Type:     tt.eventType,
				Content:  tt.content,
				Thinking: tt.thinking,
			}
			tt.checkEvent(t, event)
		})
	}
}

func TestProviderClientOptions(t *testing.T) {
	tests := []struct {
		name  string
		opts  providerClientOptions
		check func(*testing.T, providerClientOptions)
	}{
		{
			name: "empty options",
			opts: providerClientOptions{},
			check: func(t *testing.T, o providerClientOptions) {
				assert.Nil(t, o.headers)
				assert.Empty(t, o.baseURL)
			},
		},
		{
			name: "with headers",
			opts: providerClientOptions{
				headers: map[string]string{
					"Authorization": "Bearer token",
				},
			},
			check: func(t *testing.T, o providerClientOptions) {
				assert.NotNil(t, o.headers)
				assert.Equal(t, "Bearer token", o.headers["Authorization"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.opts)
		})
	}
}

func TestProviderClientOptionsAsHeader(t *testing.T) {
	tests := []struct {
		name     string
		opts     providerClientOptions
		checkKey func(*testing.T, *http.Header)
	}{
		{
			name: "nil headers",
			opts: providerClientOptions{
				headers: nil,
			},
			checkKey: func(t *testing.T, header *http.Header) {
				// nil headers should return empty header
				assert.NotNil(t, header)
				assert.Empty(t, *header)
			},
		},
		{
			name: "with headers",
			opts: providerClientOptions{
				headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			checkKey: func(t *testing.T, header *http.Header) {
				assert.NotNil(t, header)
				assert.Equal(t, "application/json", header.Get("Content-Type"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := tt.opts.asHeader()
			tt.checkKey(t, header)
		})
	}
}

func TestProviderClientOptionFunctions(t *testing.T) {
	// Test the option functions
	t.Run("WithBaseURL", func(t *testing.T) {
		opts := &providerClientOptions{}
		WithBaseURL("https://api.example.com")(opts)
		assert.Equal(t, "https://api.example.com", opts.baseURL)
	})

	t.Run("WithAPIKey", func(t *testing.T) {
		opts := &providerClientOptions{}
		WithAPIKey("test-key")(opts)
		assert.Equal(t, "test-key", opts.apiKey)
	})

	t.Run("WithMaxTokens", func(t *testing.T) {
		opts := &providerClientOptions{}
		WithMaxTokens(4096)(opts)
		assert.Equal(t, int64(4096), opts.maxTokens)
	})

	t.Run("WithSystemMessage", func(t *testing.T) {
		opts := &providerClientOptions{}
		WithSystemMessage("You are a helpful assistant")(opts)
		assert.Equal(t, "You are a helpful assistant", opts.systemMessage)
	})

	t.Run("WithModel", func(t *testing.T) {
		opts := &providerClientOptions{}
		testModel := models.Model{Name: "test-model", ContextWindow: 4096}
		WithModel(testModel)(opts)
		assert.Equal(t, "test-model", opts.model.Name)
	})

	t.Run("WithHeaders", func(t *testing.T) {
		opts := &providerClientOptions{}
		WithHeaders(map[string]string{"X-Custom": "value"})(opts)
		assert.Equal(t, "value", opts.headers["X-Custom"])
	})
}

func TestNewProvider_Unsupported(t *testing.T) {
	// Test that NewProvider returns error for unsupported provider
	_, err := NewProvider("unsupported-provider")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not supported")
}

func TestMaxRetriesConstant(t *testing.T) {
	// Verify maxRetries constant
	assert.Equal(t, 8, maxRetries)
}
