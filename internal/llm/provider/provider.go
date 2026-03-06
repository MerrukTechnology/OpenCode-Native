package provider

// Package provider provides LLM provider implementations supporting multiple backends.
//
// Supported Providers:
//   - OpenAI: OpenAI API with support for o1, o3, GPT-4, GPT-4o models
//   - Anthropic: Anthropic API with Claude models via anthropic-sdk-go
//   - Google Gemini: Gemini models via Google genai SDK
//   - AWS Bedrock: Claude models on AWS Bedrock
//   - Google Vertex AI: Claude models on Google Cloud Vertex AI
//   - DeepSeek: DeepSeek API via OpenAI-compatible SDK
//   - OpenRouter: Multi-provider aggregation via OpenRouter.ai
//   - Groq: Fast inference via GroqCloud
//   - xAI: Grok models via xAI API
//   - Mistral: Mistral models via Mistral API
//   - Kilo: Kilo's API gateway
//   - Local: Self-hosted OpenAI-compatible endpoints
//
// The package uses a generic baseProvider[C ProviderClient] pattern to share common
// logic across different provider implementations while allowing provider-specific clients.
//
// Key Features:
//   - Sentinel errors for clear error handling
//   - Automatic token counting with fallback estimation
//   - Dynamic max_tokens adjustment based on context window
//   - Message sanitization and tool pair validation
//   - Streaming support with proper event handling
//
// Usage:
//
//	p, err := provider.NewProvider(models.ProviderOpenAI,
//		provider.WithAPIKey("sk-..."),
//		provider.WithModel(models.Model{...}),
//	)
//	if err != nil {
//		// handle error
//	}
//	resp, err := p.SendMessages(ctx, messages, tools)
//
// For streaming:
//
//	events := p.StreamResponse(ctx, messages, tools)
//	for event := range events {
//		// handle event
//	}
import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/models"
	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/tools"
	toolsPkg "github.com/MerrukTechnology/OpenCode-Native/internal/llm/tools"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"github.com/MerrukTechnology/OpenCode-Native/internal/message"
)

// EventType represents the type of event during streaming.
type EventType string

const maxRetries = 8

// Sentinel errors for provider operations.
var (
	ErrProviderNotSupported = errors.New("provider not supported")
	ErrModelNotFound        = errors.New("model not found")
	ErrAPIKeyMissing        = errors.New("API key missing")
)

const (
	// EventContentStart indicates the start of content generation.
	EventContentStart EventType = "content_start"
	// EventToolUseStart indicates the start of a tool use.
	EventToolUseStart EventType = "tool_use_start"
	// EventToolUseDelta indicates a delta update for a tool use.
	EventToolUseDelta EventType = "tool_use_delta"
	// EventToolUseStop indicates the end of a tool use.
	EventToolUseStop EventType = "tool_use_stop"
	// EventContentDelta indicates a delta update for content.
	EventContentDelta EventType = "content_delta"
	// EventThinkingDelta indicates a delta update for thinking content.
	EventThinkingDelta EventType = "thinking_delta"
	// EventContentStop indicates the end of content generation.
	EventContentStop EventType = "content_stop"
	// EventComplete indicates the completion of the response.
	EventComplete EventType = "complete"
	// EventError indicates an error occurred.
	EventError EventType = "error"
	// EventWarning indicates a warning.
	EventWarning EventType = "warning"
)

// TokenUsage represents token consumption for a provider request.
type TokenUsage struct {
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
	TotalTokens         int64
}

// ProviderResponse represents a complete response from a provider.
type ProviderResponse struct {
	Content      string
	ToolCalls    []message.ToolCall
	Usage        TokenUsage
	FinishReason message.FinishReason
}

// ProviderEvent represents an event during streaming response.
type ProviderEvent struct {
	Type EventType

	Content  string
	Thinking string
	Response *ProviderResponse
	ToolCall *message.ToolCall
	Error    error
}

// Provider defines the interface for LLM providers.
type Provider interface {
	// SendMessages sends a list of messages to the provider and returns a response.
	SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error)

	// StreamResponse streams a response from the provider.
	StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent

	// Model returns the current model being used.
	Model() models.Model

	// CountTokens counts tokens for provided messages using underlying client OR fallback to default estimation strategy,
	// returns tokens count and whether a threshold has been hit based on the model context size,
	// threshold can be used to track an approaching limit to trigger compaction or other activities
	CountTokens(ctx context.Context, threshold float64, messages []message.Message, tools []toolsPkg.BaseTool) (tokens int64, hit bool)

	// AdjustMaxTokens calculates and sets new max_tokens if needed to be used by underlying client
	AdjustMaxTokens(estimatedTokens int64) int64
}

type providerClientOptions struct {
	apiKey        string
	model         models.Model
	maxTokens     int64
	systemMessage string
	baseURL       string
	headers       map[string]string

	anthropicOptions []AnthropicOption
	openaiOptions    []OpenAIOption
	geminiOptions    []GeminiOption
	bedrockOptions   []BedrockOption
	deepSeekOptions  []DeepSeekOption
	kiloOptions      []KiloOption
}

func (opts *providerClientOptions) asHeader() *http.Header {
	header := http.Header{}
	if opts.headers == nil {
		return &header
	}
	for k, v := range opts.headers {
		header.Add(k, v)
	}
	return &header
}

// ProviderClientOption is a function that configures provider client options.
type ProviderClientOption func(*providerClientOptions)

// ProviderClient is the interface for provider-specific clients.
type ProviderClient interface {
	send(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error)
	stream(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent
	countTokens(ctx context.Context, messages []message.Message, tools []toolsPkg.BaseTool) (int64, error)
	maxTokens() int64
	setMaxTokens(maxTokens int64)
}

type baseProvider[C ProviderClient] struct {
	options providerClientOptions
	client  C
	name    models.ModelProvider // cached provider name for ListModels caching
}

// NewProvider creates a new provider instance for the given provider name.
func NewProvider(providerName models.ModelProvider, opts ...ProviderClientOption) (Provider, error) {
	clientOptions := providerClientOptions{}
	for _, o := range opts {
		o(&clientOptions)
	}
	switch providerName {
	case models.ProviderVertexAI:
		return &baseProvider[VertexAIClient]{
			options: clientOptions,
			client:  newVertexAIClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderAnthropic:
		return &baseProvider[AnthropicClient]{
			options: clientOptions,
			client:  newAnthropicClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderOpenAI:
		return &baseProvider[OpenAIClient]{
			options: clientOptions,
			client:  newOpenAIClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderGemini:
		return &baseProvider[GeminiClient]{
			options: clientOptions,
			client:  newGeminiClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderBedrock:
		return &baseProvider[BedrockClient]{
			options: clientOptions,
			client:  newBedrockClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderGroq:
		clientOptions.openaiOptions = append(clientOptions.openaiOptions,
			WithOpenAIBaseURL("https://api.groq.com/openai/v1"),
		)
		return &baseProvider[OpenAIClient]{
			options: clientOptions,
			client:  newOpenAIClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderOpenRouter:
		clientOptions.openaiOptions = append(clientOptions.openaiOptions,
			WithOpenAIBaseURL("https://openrouter.ai/api/v1"),
			WithOpenAIExtraHeaders(map[string]string{
				"HTTP-Referer": "opencode.ai",
				"X-Title":      "opencode",
			}),
		)
		return &baseProvider[OpenAIClient]{
			options: clientOptions,
			client:  newOpenAIClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderXAI:
		clientOptions.openaiOptions = append(clientOptions.openaiOptions,
			WithOpenAIBaseURL("https://api.x.ai/v1"),
		)
		return &baseProvider[OpenAIClient]{
			options: clientOptions,
			client:  newOpenAIClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderMistral:
		clientOptions.openaiOptions = append(clientOptions.openaiOptions,
			WithOpenAIBaseURL("https://api.mistral.ai/v1"),
		)
		return &baseProvider[OpenAIClient]{
			options: clientOptions,
			client:  newOpenAIClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderKilo:
		return &baseProvider[KiloClient]{
			options: clientOptions,
			client:  newKiloClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderDeepSeek:
		return &baseProvider[DeepSeekClient]{
			options: clientOptions,
			client:  newDeepSeekClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderLocal:
		if clientOptions.baseURL == "" {
			clientOptions.baseURL = os.Getenv("LOCAL_ENDPOINT")
		}
		return &baseProvider[OpenAIClient]{
			options: clientOptions,
			client:  newOpenAIClient(clientOptions),
			name:    providerName,
		}, nil
	case models.ProviderMock:
		// Mock provider is not supported for direct instantiation
		// TODO: Impliment a mock provider that can be used for testing and local development without special setup
		return nil, fmt.Errorf("%w: mock provider requires special setup", ErrProviderNotSupported)
	}
	return nil, fmt.Errorf("provider not supported: %s", providerName)
}

func (p *baseProvider[C]) cleanMessages(messages []message.Message) (cleaned []message.Message) {
	for _, msg := range messages {
		// The message has no content parts at all
		if len(msg.Parts) == 0 {
			continue
		}
		// Skip assistant messages that have no text content and no tool calls
		// (e.g., canceled messages that only contain a Finish part)
		if msg.Role == message.Assistant && msg.Content().String() == "" && len(msg.ToolCalls()) == 0 {
			logging.Warn("Skipping assistant message with no content or tool calls (likely canceled)",
				"message_id", msg.ID,
			)
			continue
		}
		cleaned = append(cleaned, msg)
	}
	return
}

// sanitizeToolPairs ensures that tool_use/tool_result message pairs are consistent.
// With seq-based ordering, messages are guaranteed to be in correct order.
// This function handles crash recovery and proxy ID rewrite:
// 1. An Assistant message with tool calls not followed by a Tool message → synthesize error tool results
// 2. Incomplete tool results (some tool_use IDs missing) → synthesize missing ones
// 3. Mismatched tool_result IDs (proxy rewrite) → fix by positional match
// 4. Orphaned tool result messages → skip
func (p *baseProvider[C]) sanitizeToolPairs(messages []message.Message) []message.Message {
	var result []message.Message
	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		if msg.Role == message.Assistant && len(msg.ToolCalls()) > 0 {
			result = append(result, msg)
			toolCalls := msg.ToolCalls()

			if i+1 < len(messages) && messages[i+1].Role == message.Tool {
				i++
				toolMsg := messages[i]
				toolResults := toolMsg.ToolResults()

				validIDs := make(map[string]bool, len(toolCalls))
				for _, tc := range toolCalls {
					validIDs[tc.ID] = true
				}

				resultIDs := make(map[string]bool, len(toolResults))
				allValid := true
				for _, tr := range toolResults {
					if !validIDs[tr.ToolCallID] {
						allValid = false
						break
					}
					resultIDs[tr.ToolCallID] = true
				}

				allComplete := allValid
				if allValid {
					for _, tc := range toolCalls {
						if !resultIDs[tc.ID] {
							allComplete = false
							break
						}
					}
				}

				if allComplete {
					result = append(result, toolMsg)
				} else if allValid {
					logging.Warn("Synthesizing missing tool results for incomplete tool_result set",
						"message_id", toolMsg.ID,
						"tool_call_count", len(toolCalls),
						"tool_result_count", len(toolResults),
					)
					fixedParts := make([]message.ContentPart, 0, len(toolMsg.Parts)+len(toolCalls))
					fixedParts = append(fixedParts, toolMsg.Parts...)
					for _, tc := range toolCalls {
						if !resultIDs[tc.ID] {
							fixedParts = append(fixedParts, message.ToolResult{
								ToolCallID: tc.ID,
								Name:       tc.Name,
								Content:    "Tool execution was interrupted",
								IsError:    true,
							})
						}
					}
					toolMsg.Parts = fixedParts
					result = append(result, toolMsg)
				} else {
					logging.Warn("Fixing mismatched tool_result IDs",
						"message_id", toolMsg.ID,
						"tool_call_count", len(toolCalls),
						"tool_result_count", len(toolResults),
					)
					fixedParts := make([]message.ContentPart, 0, len(toolMsg.Parts))
					for _, part := range toolMsg.Parts {
						if tr, ok := part.(message.ToolResult); ok {
							if !validIDs[tr.ToolCallID] {
								resultIdx := -1
								for j, origTR := range toolResults {
									if origTR.ToolCallID == tr.ToolCallID {
										resultIdx = j
										break
									}
								}
								if resultIdx >= 0 && resultIdx < len(toolCalls) {
									tr.ToolCallID = toolCalls[resultIdx].ID
								} else {
									logging.Warn("Dropping unmatched tool result",
										"tool_call_id", tr.ToolCallID,
										"message_id", toolMsg.ID,
									)
									continue
								}
							}
							fixedParts = append(fixedParts, tr)
						} else {
							fixedParts = append(fixedParts, part)
						}
					}
					toolMsg.Parts = fixedParts
					result = append(result, toolMsg)
				}
			} else {
				logging.Warn("Synthesizing missing tool results for orphaned tool_use blocks",
					"message_id", msg.ID,
					"tool_call_count", len(toolCalls),
				)
				parts := make([]message.ContentPart, len(toolCalls))
				for j, tc := range toolCalls {
					parts[j] = message.ToolResult{
						ToolCallID: tc.ID,
						Name:       tc.Name,
						Content:    "Tool execution was interrupted",
						IsError:    true,
					}
				}
				result = append(result, message.Message{
					Role:      message.Tool,
					SessionID: msg.SessionID,
					Parts:     parts,
				})
			}
			continue
		}

		if msg.Role == message.Tool && len(msg.ToolResults()) > 0 {
			hasMatchingAssistant := false
			if len(result) > 0 {
				prev := result[len(result)-1]
				if prev.Role == message.Assistant && len(prev.ToolCalls()) > 0 {
					hasMatchingAssistant = true
				}
			}
			if !hasMatchingAssistant {
				logging.Warn("Skipping orphaned tool result message without preceding assistant tool_use",
					"message_id", msg.ID,
				)
				continue
			}
		}

		result = append(result, msg)
	}
	return result
}

func (p *baseProvider[C]) SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error) {
	messages = p.cleanMessages(messages)
	messages = p.sanitizeToolPairs(messages)
	return p.client.send(ctx, messages, tools)
}

func (p *baseProvider[C]) Model() models.Model {
	return p.options.model
}

func (p *baseProvider[C]) StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent {
	messages = p.cleanMessages(messages)
	messages = p.sanitizeToolPairs(messages)
	return p.client.stream(ctx, messages, tools)
}

func (p *baseProvider[C]) CountTokens(ctx context.Context, threshold float64, messages []message.Message, tools []toolsPkg.BaseTool) (int64, bool) {
	estimatedTokens, err := p.client.countTokens(ctx, messages, tools)
	// Fallback to local estimation
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			logging.Warn("Provider doesn't support countTokens endpoint, using local strategy for max_tokens", "model", p.options.model.Name, "cause", err.Error())
		}
		estimatedTokens = message.EstimateTokens(messages, tools)
	}
	contextWindow := p.Model().ContextWindow
	if contextWindow <= 0 {
		return estimatedTokens, false
	}
	thresholdAbs := int64(float64(contextWindow) * threshold)
	hitThreshold := estimatedTokens >= thresholdAbs
	logging.Debug("Token estimation for auto-compaction",
		"estimated_tokens", estimatedTokens,
		"threshold", thresholdAbs,
		"context_window", contextWindow,
		"auto-compaction required", hitThreshold,
	)
	return estimatedTokens, hitThreshold
}

func (p *baseProvider[C]) AdjustMaxTokens(estimatedTokens int64) int64 {
	maxTokens := p.client.maxTokens()
	model := p.options.model
	// Safeguard
	if estimatedTokens >= model.ContextWindow {
		logging.Warn(
			"Estimated token count higher than context window, use existing max_tokens",
			"model",
			model.Name,
			"context",
			model.ContextWindow,
			"max_tokens",
			maxTokens,
			"estimated",
			estimatedTokens,
		)
		return 0
	}

	newMaxTokens := maxTokens
	for estimatedTokens+newMaxTokens >= model.ContextWindow {
		newMaxTokens = newMaxTokens / 2
		if float64(newMaxTokens) < float64(model.ContextWindow)*0.05 {
			logging.Warn(
				"New max_tokens is below 5% of total context, can't shrink further, proceeding",
				"model",
				model.Name,
				"context",
				model.ContextWindow,
				"new_max_tokens",
				newMaxTokens,
				"estimated",
				estimatedTokens,
			)
			break
		}
	}
	if maxTokens != newMaxTokens {
		p.client.setMaxTokens(newMaxTokens)
		logging.Info("max_tokens value has changed", "model", model.Name, "old", maxTokens, "new", newMaxTokens)
	}

	return newMaxTokens
}

// WithBaseURL sets the base URL for the provider.
func WithBaseURL(baseURL string) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.baseURL = baseURL
	}
}

// WithHeaders sets custom headers for the provider.
func WithHeaders(headers map[string]string) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.headers = headers
	}
}

// WithAPIKey sets the API key for the provider.
func WithAPIKey(apiKey string) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.apiKey = apiKey
	}
}

// WithModel sets the model for the provider.
func WithModel(model models.Model) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.model = model
	}
}

// WithMaxTokens sets the maximum tokens for the provider.
func WithMaxTokens(maxTokens int64) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.maxTokens = maxTokens
	}
}

// WithSystemMessage sets the system message for the provider.
func WithSystemMessage(systemMessage string) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.systemMessage = systemMessage
	}
}

// WithOpenAIOptions sets OpenAI-specific options.
func WithOpenAIOptions(openaiOptions ...OpenAIOption) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.openaiOptions = openaiOptions
	}
}

// WithAnthropicOptions sets Anthropic-specific options.
func WithAnthropicOptions(anthropicOptions ...AnthropicOption) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.anthropicOptions = anthropicOptions
	}
}

// WithGeminiOptions sets Gemini-specific options.
func WithGeminiOptions(geminiOptions ...GeminiOption) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.geminiOptions = geminiOptions
	}
}

// WithBedrockOptions sets Bedrock-specific options.
func WithBedrockOptions(bedrockOptions ...BedrockOption) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.bedrockOptions = bedrockOptions
	}
}

// WithDeepSeekOptions sets DeepSeek-specific options.
func WithDeepSeekOptions(deepSeekOptions ...DeepSeekOption) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.deepSeekOptions = deepSeekOptions
	}
}

// WithKiloOptions sets Kilo-specific options.
func WithKiloOptions(kiloOptions ...KiloOption) ProviderClientOption {
	return func(options *providerClientOptions) {
		options.kiloOptions = kiloOptions
	}
}
