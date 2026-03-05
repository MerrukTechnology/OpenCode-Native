package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/models"
	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/provider/sse"
	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/tools"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"github.com/MerrukTechnology/OpenCode-Native/internal/message"
)

const kiloDefaultBaseURL = "https://api.kilo.ai/api/gateway"

// ==========================================
// 1. Kilo SDK (Internal Logic)
// ==========================================

type kiloAPIClient struct {
	baseURL    string
	apiKey     string
	headers    map[string]string
	httpClient *http.Client
}

func newKiloSDK(baseURL, apiKey string, headers map[string]string) *kiloAPIClient {
	if baseURL == "" {
		baseURL = kiloDefaultBaseURL
	}
	return &kiloAPIClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		headers:    headers,
		httpClient: &http.Client{Timeout: 10 * time.Minute}, // 10m for reasoning models
	}
}

// createChatRequest builds the strictly typed payload Kilo expects.
func (k *kiloAPIClient) createChatRequest(model string, messages []message.Message, tools []tools.BaseTool, maxTokens int64, opts kiloOptions) ([]byte, error) {
	req := kiloRequest{
		Model:       model,
		Messages:    k.convertMessages(messages),
		Stream:      true, // Always stream to handle Kilo's protocol correctly
		MaxTokens:   maxTokens,
		Temperature: 0.7,
	}

	// Apply options
	if opts.reasoningEffort != "" {
		req.ReasoningEffort = &opts.reasoningEffort
	}
	if opts.serviceTier != "" {
		req.ServiceTier = &opts.serviceTier
	}
	if len(opts.modalities) > 0 {
		req.Modalities = opts.modalities
	}
	// Kilo supports "store"
	if opts.store {
		req.Store = true
	}
	if len(opts.promptCacheKey) > 0 {
		req.PromptCacheKey = opts.promptCacheKey
	}

	if len(tools) > 0 {
		req.Tools = k.convertTools(tools)
	}

	return json.Marshal(req)
}

// convertMessages maps internal messages to Kilo's strict schema.
func (k *kiloAPIClient) convertMessages(msgs []message.Message) []kiloReqMessage {
	var result []kiloReqMessage

	for _, msg := range msgs {
		km := kiloReqMessage{Role: string(msg.Role)}

		switch msg.Role {
		case message.User:
			if len(msg.BinaryContent()) > 0 {
				var parts []kiloContentPart
				parts = append(parts, kiloContentPart{Type: "text", Text: msg.Content().String()})
				for _, bin := range msg.BinaryContent() {
					parts = append(parts, kiloContentPart{
						Type:     "image_url",
						ImageURL: &kiloImage{URL: bin.String(models.ProviderOpenAI)},
					})
				}
				km.Content = parts
			} else {
				km.Content = msg.Content().String()
			}

		case message.Assistant:
			// Content must be explicit null if empty
			if msg.Content().String() == "" {
				km.Content = nil // Serializes to "content": null
			} else {
				km.Content = msg.Content().String()
			}

			if len(msg.ToolCalls()) > 0 {
				km.ToolCalls = make([]kiloReqToolCall, len(msg.ToolCalls()))
				for i, tc := range msg.ToolCalls() {
					km.ToolCalls[i] = kiloReqToolCall{
						ID:   tc.ID,
						Type: "function",
						Function: kiloReqFunction{
							Name:      tc.Name,
							Arguments: tc.Input, // This is already a complete JSON string
						},
					}
				}
			}

		case message.Tool:
			for _, res := range msg.ToolResults() {
				result = append(result, kiloReqMessage{
					Role:       "tool",
					Content:    res.Content,
					ToolCallID: res.ToolCallID,
				})
			}
			continue // Skip the default append
		}

		result = append(result, km)
	}
	return result
}

func (k *kiloAPIClient) convertTools(baseTools []tools.BaseTool) []kiloReqTool {
	var result []kiloReqTool
	for _, t := range baseTools {
		info := t.Info()
		result = append(result, kiloReqTool{
			Type: "function",
			Function: kiloReqDefFunction{
				Name:        info.Name,
				Description: info.Description,
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": info.Parameters,
					"required":   info.Required,
				},
			},
		})
	}
	return result
}

// StreamChat sends the request and returns a channel of parsed events.
func (k *kiloAPIClient) StreamChat(ctx context.Context, payload []byte) <-chan ProviderEvent {
	eventChan := make(chan ProviderEvent)
	cfg := config.Get()

	go func() {
		defer close(eventChan)

		attempts := 0
		for {
			attempts++
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, k.baseURL+"/chat/completions", bytes.NewBuffer(payload))
			if err != nil {
				eventChan <- ProviderEvent{Type: EventError, Error: err}
				return
			}

			req.Header.Set("Content-Type", "application/json")
			if k.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+k.apiKey)
			}
			for key, val := range k.headers {
				req.Header.Set(key, val)
			}

			if cfg.Debug {
				logging.Debug("KiloSDK: Sending Request", "attempt", attempts)
			}

			resp, err := k.httpClient.Do(req)
			if err != nil {
				if k.shouldRetry(attempts, err) {
					logging.Warn("KiloSDK: Network error, retrying...", "error", err)
					select {
					case <-ctx.Done():
						return
					case <-time.After(1 * time.Second):
						continue
					}
				}
				eventChan <- ProviderEvent{Type: EventError, Error: err}
				return
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				apiErr := fmt.Errorf("kilo api error: %d %s", resp.StatusCode, string(body))

				// 422 is a schema error (like missing content), do not retry, fail immediately
				if resp.StatusCode == http.StatusUnprocessableEntity {
					eventChan <- ProviderEvent{Type: EventError, Error: apiErr}
					return
				}

				if k.shouldRetry(attempts, apiErr) {
					logging.Warn("KiloSDK: API error, retrying...", "status", resp.StatusCode)
					select {
					case <-ctx.Done():
						return
					case <-time.After(1 * time.Second):
						continue
					}
				}
				eventChan <- ProviderEvent{Type: EventError, Error: apiErr}
				return
			}

			// Use the KiloDecoder from sse package for SSE parsing
			decoder := sse.NewKiloDecoder()
			for sseEvent := range decoder.Parse(resp.Body) {
				eventChan <- convertSSEEvent(sseEvent)
			}
			resp.Body.Close()
			return
		}
	}()

	return eventChan
}

func (k *kiloAPIClient) shouldRetry(attempts int, err error) bool {
	if attempts >= 8 {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "500") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "unexpected EOF") ||
		strings.Contains(msg, "connection reset")
}

// convertSSEEvent maps your internal sse.SSEEvent to the provider's generic ProviderEvent
func convertSSEEvent(e sse.SSEEvent) ProviderEvent {
	switch e.Type {
	case sse.EventContentDelta:
		return ProviderEvent{Type: EventContentDelta, Content: e.Content}
	case sse.EventThinkingDelta:
		return ProviderEvent{Type: EventThinkingDelta, Thinking: e.Thinking}
	case sse.EventToolUseStart:
		return ProviderEvent{
			Type:     EventToolUseStart,
			ToolCall: e.ToolCall,
		}
	case sse.EventToolUseDelta:
		return ProviderEvent{
			Type:     EventToolUseDelta,
			ToolCall: e.ToolCall,
		}
	case sse.EventComplete:
		return ProviderEvent{
			Type: EventComplete,
			Response: &ProviderResponse{
				FinishReason: e.Response.FinishReason,
				Usage:        convertTokenUsage(e.Response.Usage),
			},
		}
	case sse.EventError:
		return ProviderEvent{Type: EventError, Error: e.Error}
	default:
		return ProviderEvent{}
	}
}

func convertTokenUsage(u sse.TokenUsage) TokenUsage {
	return TokenUsage{
		InputTokens:  u.InputTokens,
		OutputTokens: u.OutputTokens,
		TotalTokens:  u.TotalTokens,
	}
}

// ==========================================
// 2. Provider Implementation
// ==========================================

// --- Options (Exported) ---

type kiloOptions struct {
	baseURL         string
	extraHeaders    map[string]string
	reasoningEffort string   // low, medium, high for reasoning models
	serviceTier     string   // auto, flex for pricing
	modalities      []string // text, audio for output
	store           bool     // store completions
	promptCacheKey  []string // prompt caching keys
}

// KiloOption is a function that configures Kilo provider options.
type KiloOption func(*kiloOptions)

// WithKiloBaseURL sets a custom base URL for the Kilo API.
func WithKiloBaseURL(url string) KiloOption {
	return func(o *kiloOptions) {
		o.baseURL = url
	}
}

// WithKiloExtraHeaders sets extra headers for the Kilo API.
func WithKiloExtraHeaders(headers map[string]string) KiloOption {
	return func(o *kiloOptions) {
		o.extraHeaders = headers
	}
}

// WithKiloReasoningEffort sets the reasoning effort for Kilo reasoning models.
// Values: "low", "medium", "high". Only applies to reasoning models.
func WithKiloReasoningEffort(effort string) KiloOption {
	return func(o *kiloOptions) {
		defaultEffort := "medium"
		switch effort {
		case "low", "medium", "high":
			defaultEffort = effort
		}
		o.reasoningEffort = defaultEffort
	}
}

// WithKiloServiceTier sets the service tier for pricing.
// Values: "auto" (default), "flex"
func WithKiloServiceTier(tier string) KiloOption {
	return func(o *kiloOptions) {
		o.serviceTier = tier
	}
}

// WithKiloModalities sets the output modalities.
// Values: []string{"text"} (default) or []string{"text", "audio"}
func WithKiloModalities(modalities []string) KiloOption {
	return func(o *kiloOptions) {
		o.modalities = modalities
	}
}

// WithKiloStore enables storing completions.
func WithKiloStore(store bool) KiloOption {
	return func(o *kiloOptions) {
		o.store = store
	}
}

// WithKiloPromptCacheKey enables prompt caching with specific cache keys.
func WithKiloPromptCacheKey(keys []string) KiloOption {
	return func(o *kiloOptions) {
		o.promptCacheKey = keys
	}
}

// --- Client ---

type kiloClient struct {
	providerOptions providerClientOptions
	options         kiloOptions
	sdk             *kiloAPIClient
}

type KiloClient ProviderClient

// newKiloClient initializes the Provider using the internal SDK.
func newKiloClient(opts providerClientOptions) KiloClient {
	kiloOpts := kiloOptions{}
	for _, o := range opts.kiloOptions {
		o(&kiloOpts)
	}

	// Set default base URL if not provided
	if kiloOpts.baseURL == "" {
		kiloOpts.baseURL = kiloDefaultBaseURL
	}

	logging.Info("Kilo: Initializing provider", "model", opts.model.APIModel, "baseURL", kiloOpts.baseURL)

	return &kiloClient{
		providerOptions: opts,
		options:         kiloOpts,
		sdk:             newKiloSDK(kiloOpts.baseURL, opts.apiKey, kiloOpts.extraHeaders),
	}
}

// THE SEND METHOD - FIXED ACCUMULATION LOGIC
func (k *kiloClient) send(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error) {
	// Re-use stream logic for unified handling
	eventChan := k.stream(ctx, messages, tools)

	var fullContent strings.Builder

	// Map ID -> ToolCall
	activeToolCalls := make(map[string]*message.ToolCall)
	// Map ID -> Builder (Accumulator)
	activeToolBuilders := make(map[string]*strings.Builder)
	// Order tracking
	toolCallOrder := make([]string, 0)

	var usage TokenUsage
	finishReason := message.FinishReasonEndTurn
	var lastErr error

	cfg := config.Get()
	if cfg.Debug {
		logging.Debug("Kilo: Starting accumulation loop")
	}

	for event := range eventChan {
		switch event.Type {
		case EventContentDelta:
			fullContent.WriteString(event.Content)

		case EventToolUseStart:
			if event.ToolCall != nil && event.ToolCall.ID != "" {
				id := event.ToolCall.ID

				if _, exists := activeToolCalls[id]; !exists {
					activeToolCalls[id] = &message.ToolCall{
						ID:   id,
						Name: event.ToolCall.Name,
						Type: "function",
					}
					activeToolBuilders[id] = &strings.Builder{}
					toolCallOrder = append(toolCallOrder, id)
					if cfg.Debug {
						logging.Debug("🔧 Kilo: ToolUseStart", "id", id, "name", event.ToolCall.Name)
					}
				}
			}

		case EventToolUseDelta:
			if event.ToolCall != nil && event.ToolCall.ID != "" {
				id := event.ToolCall.ID
				// Since we rely on SSE decoder for ID resolution, we can trust ID exists
				if builder, exists := activeToolBuilders[id]; exists {
					builder.WriteString(event.ToolCall.Input)
				} else {
					// Fallback if Start wasn't caught (unlikely with this setup)
					activeToolBuilders[id] = &strings.Builder{}
					activeToolBuilders[id].WriteString(event.ToolCall.Input)
					activeToolCalls[id] = &message.ToolCall{ID: id, Type: "function"}
					toolCallOrder = append(toolCallOrder, id)
				}
			}

		case EventComplete:
			if event.Response != nil {
				finishReason = event.Response.FinishReason
				usage = event.Response.Usage
			}

		case EventError:
			lastErr = event.Error
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	// Finalize tool calls in the order they were started
	var finalToolCalls []message.ToolCall
	for _, id := range toolCallOrder {
		if tc, ok := activeToolCalls[id]; ok {
			// Use string builder for final input
			if builder, ok := activeToolBuilders[id]; ok {
				tc.Input = builder.String()
			}
			tc.Finished = true
			if cfg.Debug {
				logging.Debug("🔧 Kilo: FINAL ToolCall", "id", id, "name", tc.Name, "json", tc.Input)
			}
			finalToolCalls = append(finalToolCalls, *tc)
		}
	}

	if len(finalToolCalls) > 0 && finishReason == message.FinishReasonEndTurn {
		finishReason = message.FinishReasonToolUse
	}

	return &ProviderResponse{
		Content:      fullContent.String(),
		ToolCalls:    finalToolCalls,
		Usage:        usage,
		FinishReason: finishReason,
	}, nil
}

func (k *kiloClient) stream(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent {
	// Build request payload once
	payload, err := k.sdk.createChatRequest(
		k.providerOptions.model.APIModel,
		messages,
		tools,
		k.providerOptions.maxTokens,
		k.options,
	)
	if err != nil {
		ch := make(chan ProviderEvent, 1)
		ch <- ProviderEvent{Type: EventError, Error: err}
		close(ch)
		return ch
	}

	return k.sdk.StreamChat(ctx, payload)
}

// countTokens estimates token count for messages and tools.
// Uses the API's usage data from responses when available.
func (k *kiloClient) countTokens(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (int64, error) {
	// Estimate tokens using a simple word-based heuristic
	// This is a fallback since Kilo doesn't provide a /tokenize endpoint
	var total int64

	for _, msg := range messages {
		// Rough estimate: 1 token ≈ 4 characters
		content := msg.Content().String()
		total += int64(len(content) / 4)

		// Add tool calls if present
		for _, tc := range msg.ToolCalls() {
			total += int64(len(tc.Name)/4) + int64(len(tc.Input)/4)
		}
	}

	// Add tool definitions
	for _, tool := range tools {
		info := tool.Info()
		total += int64(len(info.Name)/4) + int64(len(info.Description)/4)
	}

	// Add overhead for message structure (~10 tokens per message)
	total += int64(len(messages) * 10)

	return total, nil
}

func (k *kiloClient) maxTokens() int64 {
	return k.providerOptions.maxTokens
}

func (k *kiloClient) setMaxTokens(maxTokens int64) {
	k.providerOptions.maxTokens = maxTokens
}

// ==========================================
// 3. Strict JSON Request Schemas (Internal)
// ==========================================

type kiloRequest struct {
	Model           string           `json:"model"`
	Messages        []kiloReqMessage `json:"messages"`
	Stream          bool             `json:"stream"`
	MaxTokens       int64            `json:"max_tokens,omitempty"`
	Temperature     float64          `json:"temperature,omitempty"`
	Tools           []kiloReqTool    `json:"tools,omitempty"`
	ReasoningEffort *string          `json:"reasoning_effort,omitempty"`
	ServiceTier     *string          `json:"service_tier,omitempty"`
	Modalities      []string         `json:"modalities,omitempty"`
	Store           bool             `json:"store,omitempty"`
	PromptCacheKey  []string         `json:"prompt_cache_key,omitempty"`
}

type kiloReqMessage struct {
	Role       string            `json:"role"`
	Content    interface{}       `json:"content"` // Must be nil or string
	ToolCalls  []kiloReqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
}

type kiloReqToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function kiloReqFunction `json:"function"`
}

type kiloReqFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type kiloReqTool struct {
	Type     string             `json:"type"`
	Function kiloReqDefFunction `json:"function"`
}

type kiloReqDefFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters"`
}

type kiloContentPart struct {
	Type     string     `json:"type"`
	Text     string     `json:"text,omitempty"`
	ImageURL *kiloImage `json:"image_url,omitempty"`
}

type kiloImage struct {
	URL string `json:"url"`
}
