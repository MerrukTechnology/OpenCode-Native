package message

import (
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/models"
	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/tools"
)

// Helper function to create a Message with parts
func newMessage(parts ...ContentPart) Message {
	return Message{Parts: parts}
}

// TestContentString is a unified test for all content types with String() methods
func TestContentString(t *testing.T) {
	tests := []struct {
		name     string
		content  interface{ String() string }
		expected string
	}{
		// TextContent tests
		{name: "TextContent/simple text", content: TextContent{Text: "Hello, World!"}, expected: "Hello, World!"},
		{name: "TextContent/empty text", content: TextContent{Text: ""}, expected: ""},
		{name: "TextContent/multiline text", content: TextContent{Text: "Line 1\nLine 2\nLine 3"}, expected: "Line 1\nLine 2\nLine 3"},

		// ReasoningContent tests
		{name: "ReasoningContent/simple thinking", content: ReasoningContent{Thinking: "Let me think about this..."}, expected: "Let me think about this..."},
		{name: "ReasoningContent/empty thinking", content: ReasoningContent{Thinking: ""}, expected: ""},

		// ImageURLContent tests
		{name: "ImageURLContent/simple URL", content: ImageURLContent{URL: "https://example.com/image.png"}, expected: "https://example.com/image.png"},
		{name: "ImageURLContent/URL with detail", content: ImageURLContent{URL: "https://example.com/image.png", Detail: "high"}, expected: "https://example.com/image.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.content.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBinaryContentString(t *testing.T) {
	// PNG file signature: 0x89 0x50 0x4E 0x47 0x0D 0x0A 0x1A 0x0A
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	tests := []struct {
		name     string
		content  BinaryContent
		provider models.ModelProvider
		expected string
	}{
		{name: "OpenAI/text plain", content: BinaryContent{MIMEType: "text/plain", Data: []byte("Hello")}, provider: models.ProviderOpenAI, expected: "data:text/plain;base64,SGVsbG8="},
		{name: "Anthropic/text plain", content: BinaryContent{MIMEType: "text/plain", Data: []byte("Hello")}, provider: models.ProviderAnthropic, expected: "SGVsbG8="},
		{name: "OpenAI/PNG signature", content: BinaryContent{MIMEType: "image/png", Data: pngSignature}, provider: models.ProviderOpenAI, expected: "data:image/png;base64,iVBORw0KGgo="},
		{name: "Anthropic/PNG signature", content: BinaryContent{MIMEType: "image/png", Data: pngSignature}, provider: models.ProviderAnthropic, expected: "iVBORw0KGgo="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.content.String(tt.provider)
			if result != tt.expected {
				t.Errorf("BinaryContent.String(%s) = %q, want %q", tt.provider, result, tt.expected)
			}
		})
	}
}

// TestConstants is a unified test for all constant types with string conversions
func TestConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		// MessageRole constants
		{name: "MessageRole/assistant", got: string(Assistant), want: "assistant"},
		{name: "MessageRole/user", got: string(User), want: "user"},
		{name: "MessageRole/system", got: string(System), want: "system"},
		{name: "MessageRole/tool", got: string(Tool), want: "tool"},

		// FinishReason constants
		{name: "FinishReason/end_turn", got: string(FinishReasonEndTurn), want: "end_turn"},
		{name: "FinishReason/max_tokens", got: string(FinishReasonMaxTokens), want: "max_tokens"},
		{name: "FinishReason/tool_use", got: string(FinishReasonToolUse), want: "tool_use"},
		{name: "FinishReason/canceled", got: string(FinishReasonCanceled), want: "canceled"},
		{name: "FinishReason/error", got: string(FinishReasonError), want: "error"},
		{name: "FinishReason/permission_denied", got: string(FinishReasonPermissionDenied), want: "permission_denied"},
		{name: "FinishReason/unknown", got: string(FinishReasonUnknown), want: "unknown"},

		// ToolResultType constants
		{name: "ToolResultType/text", got: string(ToolResultTypeText), want: "text"},
		{name: "ToolResultType/image", got: string(ToolResultTypeImage), want: "image"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("constant = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestMessageContent(t *testing.T) {
	tests := []struct {
		name     string
		msg      Message
		expected string
	}{
		{name: "text content", msg: newMessage(TextContent{Text: "Hello"}), expected: "Hello"},
		{name: "no content", msg: newMessage(), expected: ""},
		{name: "reasoning only", msg: newMessage(ReasoningContent{Thinking: "Thinking..."}), expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.Content()
			if result.Text != tt.expected {
				t.Errorf("Message.Content() = %q, want %q", result.Text, tt.expected)
			}
		})
	}
}

func TestMessageReasoningContent(t *testing.T) {
	msg := newMessage(ReasoningContent{Thinking: "Let me think..."})
	result := msg.ReasoningContent()
	if result.Thinking != "Let me think..." {
		t.Errorf("Message.ReasoningContent() = %q, want %q", result.Thinking, "Let me think...")
	}
}

func TestMessageImageURLContent(t *testing.T) {
	msg := newMessage(
		ImageURLContent{URL: "https://example.com/img1.png"},
		ImageURLContent{URL: "https://example.com/img2.png"},
	)
	results := msg.ImageURLContent()
	if len(results) != 2 {
		t.Errorf("Message.ImageURLContent() returned %d, want 2", len(results))
	}
}

func TestMessageBinaryContent(t *testing.T) {
	msg := newMessage(BinaryContent{MIMEType: "image/png", Data: []byte{0x89, 0x50}})
	results := msg.BinaryContent()
	if len(results) != 1 {
		t.Errorf("Message.BinaryContent() returned %d, want 1", len(results))
	}
}

func TestMessageToolCalls(t *testing.T) {
	msg := newMessage(ToolCall{ID: "call_123", Name: "test_tool"})
	results := msg.ToolCalls()
	if len(results) != 1 || results[0].ID != "call_123" {
		t.Errorf("Message.ToolCalls() = %v, want [{call_123 test_tool}]", results)
	}
}

func TestMessageToolResults(t *testing.T) {
	msg := newMessage(
		ToolResult{ToolCallID: "tool_result_1", Type: ToolResultTypeText, Content: "Result 1"},
		ToolResult{ToolCallID: "tool_result_2", Type: ToolResultTypeText, Content: "Result 2"},
	)
	results := msg.ToolResults()
	if len(results) != 2 {
		t.Errorf("Message.ToolResults() returned %d, want 2", len(results))
	}
}

func TestMessageToolResultsByToolName(t *testing.T) {
	msg := newMessage(
		ToolResult{Name: "tool_a", Type: ToolResultTypeText, Content: "A"},
		ToolResult{Name: "tool_b", Type: ToolResultTypeText, Content: "B"},
	)
	results := msg.ToolResultsByToolName("tool_a")
	if len(results) != 1 || results[0].Content != "A" {
		t.Errorf("Message.ToolResultsByToolName(tool_a) = %v, want [{A}]", results)
	}
}

func TestMessageIsFinished(t *testing.T) {
	tests := []struct {
		name     string
		msg      Message
		expected bool
	}{
		{name: "finished", msg: newMessage(Finish{Reason: FinishReasonEndTurn}), expected: true},
		{name: "unfinished", msg: newMessage(TextContent{Text: "Hello"}), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.IsFinished()
			if result != tt.expected {
				t.Errorf("Message.IsFinished() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMessageFinishReason(t *testing.T) {
	msg := newMessage(Finish{Reason: FinishReasonMaxTokens})
	reason := msg.FinishReason()
	if reason != FinishReasonMaxTokens {
		t.Errorf("Message.FinishReason() = %q, want %q", reason, FinishReasonMaxTokens)
	}
}

func TestMessageIsThinking(t *testing.T) {
	tests := []struct {
		name     string
		msg      Message
		expected bool
	}{
		{name: "thinking", msg: newMessage(ReasoningContent{Thinking: "Thinking..."}), expected: true},
		{name: "not thinking - has text", msg: newMessage(ReasoningContent{Thinking: "Thinking..."}, TextContent{Text: "Hello"}), expected: false},
		{name: "not thinking - finished", msg: newMessage(ReasoningContent{Thinking: "Thinking..."}, Finish{Reason: FinishReasonEndTurn}), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.IsThinking()
			if result != tt.expected {
				t.Errorf("Message.IsThinking() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMessageAppendContent(t *testing.T) {
	tests := []struct {
		name        string
		initialMsg  *Message
		appendStr   string
		expectedStr string
	}{
		{
			name:        "append to existing",
			initialMsg:  &Message{Parts: []ContentPart{TextContent{Text: "Hello"}}},
			appendStr:   " World",
			expectedStr: "Hello World",
		},
		{
			name:        "append to empty",
			initialMsg:  &Message{Parts: []ContentPart{}},
			appendStr:   "Hello",
			expectedStr: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initialMsg.AppendContent(tt.appendStr)
			content := tt.initialMsg.Content()
			if content.Text != tt.expectedStr {
				t.Errorf("Message.AppendContent() = %q, want %q", content.Text, tt.expectedStr)
			}
		})
	}
}

func TestMessageAppendReasoningContent(t *testing.T) {
	msg := &Message{Parts: []ContentPart{ReasoningContent{Thinking: "Step 1"}}}
	msg.AppendReasoningContent(" Step 2")

	reasoning := msg.ReasoningContent()
	if reasoning.Thinking != "Step 1 Step 2" {
		t.Errorf("Message.AppendReasoningContent() = %q, want %q", reasoning.Thinking, "Step 1 Step 2")
	}
}

func TestMessageAddToolCall(t *testing.T) {
	msg := &Message{Parts: []ContentPart{}}
	msg.AddToolCall(ToolCall{ID: "call_1", Name: "test_tool"})

	calls := msg.ToolCalls()
	if len(calls) != 1 {
		t.Errorf("Message.ToolCalls() returned %d, want 1", len(calls))
	}
}

func TestMessageSetToolCalls(t *testing.T) {
	msg := &Message{Parts: []ContentPart{TextContent{Text: "original"}}}
	msg.SetToolCalls([]ToolCall{
		{ID: "call_1", Name: "tool_a"},
		{ID: "call_2", Name: "tool_b"},
	})

	calls := msg.ToolCalls()
	if len(calls) != 2 {
		t.Errorf("Message.ToolCalls() returned %d, want 2", len(calls))
	}
}

func TestMessageUpdateToolCall(t *testing.T) {
	msg := &Message{Parts: []ContentPart{
		ToolCall{ID: "call_1", Name: "old_tool"},
	}}
	msg.UpdateToolCall(ToolCall{ID: "call_1", Name: "new_tool"})

	calls := msg.ToolCalls()
	if calls[0].Name != "new_tool" {
		t.Errorf("Message.ToolCalls()[0].Name = %q, want %q", calls[0].Name, "new_tool")
	}
}

func TestMessageAddToolResult(t *testing.T) {
	msg := &Message{Parts: []ContentPart{}}
	msg.AddToolResult(ToolResult{ToolCallID: "result_1", Type: ToolResultTypeText, Content: "Done"})

	results := msg.ToolResults()
	if len(results) != 1 {
		t.Errorf("Message.ToolResults() returned %d, want 1", len(results))
	}
}

func TestMessageSetToolResults(t *testing.T) {
	msg := &Message{Parts: []ContentPart{}}
	msg.SetToolResults([]ToolResult{
		{ToolCallID: "result_1", Type: ToolResultTypeText, Content: "A"},
		{ToolCallID: "result_2", Type: ToolResultTypeText, Content: "B"},
	})

	results := msg.ToolResults()
	if len(results) != 2 {
		t.Errorf("Message.ToolResults() returned %d, want 2", len(results))
	}
}

func TestMessageAddFinish(t *testing.T) {
	msg := &Message{Parts: []ContentPart{TextContent{Text: "Hello"}}}
	msg.AddFinish(FinishReasonEndTurn)

	if !msg.IsFinished() {
		t.Error("Message.IsFinished() = false, want true")
	}
}

func TestMessageAddImageURL(t *testing.T) {
	msg := &Message{Parts: []ContentPart{}}
	msg.AddImageURL("https://example.com/image.png", "high")

	images := msg.ImageURLContent()
	if len(images) != 1 || images[0].URL != "https://example.com/image.png" {
		t.Errorf("Message.ImageURLContent() = %v, want [{https://example.com/image.png}]", images)
	}
}

func TestMessageAddBinary(t *testing.T) {
	msg := &Message{Parts: []ContentPart{}}
	msg.AddBinary("image/png", []byte{0x89, 0x50})

	binaries := msg.BinaryContent()
	if len(binaries) != 1 || binaries[0].MIMEType != "image/png" {
		t.Errorf("Message.BinaryContent() = %v, want [{image/png}]", binaries)
	}
}

func TestToolResultIsImageToolResponse(t *testing.T) {
	tests := []struct {
		name     string
		result   ToolResult
		expected bool
	}{
		{name: "image result", result: ToolResult{Type: ToolResultTypeImage}, expected: true},
		{name: "text result", result: ToolResult{Type: ToolResultTypeText}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.IsImageToolResponse()
			if result != tt.expected {
				t.Errorf("ToolResult.IsImageToolResponse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMessageFinishPart(t *testing.T) {
	msg := newMessage(Finish{Reason: FinishReasonEndTurn})
	finish := msg.FinishPart()

	if finish == nil || finish.Reason != FinishReasonEndTurn {
		t.Errorf("Message.FinishPart() = %v, want {Reason: %s}", finish, FinishReasonEndTurn)
	}
}

func TestMessageStructOutput(t *testing.T) {
	tests := []struct {
		name       string
		msg        Message
		wantResult bool
	}{
		{
			name:       "found struct output",
			msg:        newMessage(ToolResult{Name: tools.StructOutputToolName, Type: ToolResultTypeText, Content: `{"key": "value"}`}),
			wantResult: true,
		},
		{
			name:       "not found",
			msg:        newMessage(ToolResult{Name: "other_tool", Type: ToolResultTypeText, Content: "result"}),
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := tt.msg.StructOutput()
			if ok != tt.wantResult {
				t.Errorf("Message.StructOutput() ok = %v, want %v", ok, tt.wantResult)
			}
			if tt.wantResult && result == nil {
				t.Error("Message.StructOutput() = nil, want non-nil")
			}
			if !tt.wantResult && result != nil {
				t.Error("Message.StructOutput() = non-nil, want nil")
			}
		})
	}
}
