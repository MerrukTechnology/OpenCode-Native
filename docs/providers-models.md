# OpenCode Providers and Models Guide

> **Note:** This document describes the provider and model architecture in OpenCode. This content is not yet integrated into the main documentation.

---

## Table of Contents

1. [Providers](#providers)
2. [Models](#models)
3. [Tools](#tools)
4. [Package Architecture](#package-architecture)

---

## Providers

OpenCode supports multiple LLM providers through a unified [`Provider`](internal/llm/provider/provider.go) interface.

### Provider Interface

```go
type Provider interface {
    // SendMessages sends a request and waits for complete response
    SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error)

    // StreamResponse streams the response as events
    StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent

    // Model returns the current model
    Model() models.Model

    // CountTokens counts tokens for messages and tools
    CountTokens(ctx context.Context, threshold float64, messages []message.Message, tools []toolsPkg.BaseTool) (tokens int64, hit bool)

    // AdjustMaxTokens adjusts max tokens based on estimated tokens
    AdjustMaxTokens(estimatedTokens int64) int64
}
```

### Event Types

Providers stream events during response generation:

```go
const (
    EventContentStart  EventType = "content_start"
    EventToolUseStart  EventType = "tool_use_start"
    EventToolUseDelta  EventType = "tool_use_delta"
    EventToolUseStop   EventType = "tool_use_stop"
    EventContentDelta  EventType = "content_delta"
    EventThinkingDelta EventType = "thinking_delta"
    EventContentStop   EventType = "content_stop"
    EventComplete      EventType = "complete"
    EventError         EventType = "error"
    EventWarning       EventType = "warning"
)
```

### Supported Providers

| Provider | Package | Description |
|----------|---------|-------------|
| **OpenAI** | [`internal/llm/provider/openai.go`](internal/llm/provider/openai.go) | OpenAI API (GPT-4, GPT-4o, etc.) |
| **Anthropic** | [`internal/llm/provider/anthropic.go`](internal/llm/provider/anthropic.go) | Anthropic API (Claude 3, Claude 4) |
| **Google Gemini** | [`internal/llm/provider/gemini.go`](internal/llm/provider/gemini.go) | Google Gemini models |
| **Vertex AI** | [`internal/llm/provider/vertexai.go`](internal/llm/provider/vertexai.go) | Google Vertex AI (Claude via Anthropic) |
| **Groq AI** | [`internal/llm/provider/groq.go`](internal/llm/provider/groq.go) | Groq AI models (Claude via Antropic) |
| **XAI** | [`internal/llm/provider/xai.go`](internal/llm/provider/xai.go) | XAI API (Groq) |
| **KiloCode** | [`internal/llm/models/kilocode.go`](internal/llm/models/kilocode.go) | KiloCode API |
| **Mistral** | [`internal/llm/models/mistral.go`](internal/llm/models/mistral.go) | Mistral API |
| **OpenRouter** | [`internal/llm/provider/openrouter.go`](internal/llm/provider/openrouter.go) | OpenRouter API |
| **DeepSeek** | [`internal/llm/provider/deepseek.go`](internal/llm/provider/deepseek.go) | DeepSeek API |
| **Bedrock** | [`internal/llm/provider/bedrock.go`](internal/llm/provider/bedrock.go) | AWS Bedrock (Claude via Bedrock) |
| **Local/Ollama** | [`internal/llm/models/local.go`](internal/llm/models/local.go) | Local models via Ollama |

### Provider Popularity Order

```go
var ProviderPopularity = map[ModelProvider]int{
    ProviderVertexAI:   1,  // Most popular
    ProviderAnthropic:  2,
    ProviderOpenAI:     3,
    ProviderGemini:     4,
    ProviderGroq:       5,
    ProviderXAI:        6,
    ProviderKiloCode:   7,
    ProviderMistral:    8,
    ProviderOpenRouter: 9,
    ProviderDeepSeek:   10,
    ProviderBedrock:    11,
    ProviderLocal:      12, // Least popular
}
```

---

## Models

Models are defined in [`internal/llm/models/models.go`](internal/llm/models/models.go).

### Model Struct

```go
type Model struct {
    ID                       ModelID       // Unique identifier (e.g., "anthropic.claude-sonnet-4-5-m")
    Name                     string        // Display name (e.g., "Claude Sonnet 4.5")
    Provider                 ModelProvider // Provider name (e.g., "anthropic")
    APIModel                 string        // API-specific model ID
    CostPer1MIn              float64       // Cost per 1M input tokens
    CostPer1MOut             float64       // Cost per 1M output tokens
    CostPer1MInCached        float64       // Cost per 1M cached input tokens
    CostPer1MOutCached       float64       // Cost per 1M cached output tokens
    ContextWindow            int64         // Max context window size
    DefaultMaxTokens         int64         // Default max output tokens
    CanReason                bool          // Supports reasoning/thinking
    SupportsAdaptiveThinking bool          // Supports adaptive thinking effort
    SupportsMaximumThinking  bool          // Supports maximum thinking effort
    SupportsAttachments      bool          // Supports file attachments
}
```

### Provider Types

```go
type ModelProvider string

const (
    ProviderVertexAI   ModelProvider = "vertexai"
    ProviderAnthropic  ModelProvider = "anthropic"
    ProviderOpenAI     ModelProvider = "openai"
    ProviderGemini     ModelProvider = "gemini"
    ProviderGroq       ModelProvider = "groq"
    ProviderXAI        ModelProvider = "xai"
    ProviderKiloCode   ModelProvider = "KiloCode"
    ProviderMistral    ModelProvider = "mistral"
    ProviderOpenRouter ModelProvider = "openrouter"
    ProviderDeepSeek   ModelProvider = "deepseek"
    ProviderBedrock    ModelProvider = "bedrock"
    ProviderLocal      ModelProvider = "local"
    ProviderMock       ModelProvider = "__mock"  // For testing
)
```

### Model ID Examples

```go
// Anthropic
anthropic.claude-sonnet-4-5-m
anthropic.claude-opus-4-5-20251120

// OpenAI
openai.gpt-4o
openai.gpt-4o-mini
openai.gpt-4-turbo

// Google
gemini.gemini-2.0-flash-exp
gemini.gemini-1.5-pro

// DeepSeek
deepseek.chat

// Vertex AI (Claude via Google)
vertexai.claude-sonnet-4-5-m
```

---

## Tools

Tools are defined in [`internal/llm/tools/`](internal/llm/tools/).

### Tool Interface

```go
type BaseTool interface {
    // Info returns the tool's schema information
    Info() ToolInfo

    // Run executes the tool with given parameters
    Run(ctx context.Context, params ToolCall) (ToolResponse, error)
}
```

### Tool Types

| Tool | File | Description |
|------|------|-------------|
| **Edit** | [`edit.go`](internal/llm/tools/edit.go) | Edit files with targeted changes |
| **Write** | [`write.go`](internal/llm/tools/write.go) | Write/create new files |
| **Read/View** | [`view.go`](internal/llm/tools/view.go) | Read file contents |
| **Delete** | [`delete.go`](internal/llm/tools/delete.go) | Delete files and directories |
| **Glob** | [`glob.go`](internal/llm/tools/glob.go) | Find files by pattern |
| **Grep** | [`grep.go`](internal/llm/tools/grep.go) | Search file contents |
| **Bash** | [`bash.go`](internal/llm/tools/bash.go) | Execute shell commands |
| **LS** | [`ls.go`](internal/llm/tools/ls.go) | List directory contents |
| **MultiEdit** | [`multiedit.go`](internal/llm/tools/multiedit.go) | Apply multiple edits atomically |
| **Patch** | [`patch.go`](internal/llm/tools/patch.go) | Apply patch files |
| **Skill** | [`skill.go`](internal/llm/tools/skill.go) | Invoke skill agents |
| **LSP** | [`lsp.go`](internal/llm/tools/lsp.go) | Language Server Protocol tools |
| **Diagnostics** | [`diagnostics.go`](internal/llm/tools/diagnostics.go) | Get linting/diagnostics |
| **Fetch** | [`fetch.go`](internal/llm/tools/fetch.go) | Fetch remote content |
| **ViewImage** | [`view_image.go`](internal/llm/tools/view_image.go) | View image files |
| **StructuredOutput** | [`struct_output.go`](internal/llm/tools/struct_output.go) | Generate structured output |

### Tool Response Types

```go
type ToolResponse struct {
    Type     toolResponseType  // "text" or "image"
    Content  string           // Response content
    Metadata string           // Optional JSON metadata
    IsError  bool            // Whether response is an error
}

// Response constructors
NewTextResponse(content string) toolResponse
NewImageResponse(content string) toolResponse
NewTextErrorResponse(content string) toolResponse
NewEmptyResponse() toolResponse
```

---

## Package Architecture

### Core Packages

```
internal/
├── agent/           # Agent registry and management
├── app/            # Application setup and LSP integration
├── completions/    # Shell completions
├── config/         # Configuration management
├── db/             # Database providers (SQLite, MySQL)
├── diff/           # Diff/patch functionality
├── fileutil/       # File utilities (unified API)
├── format/         # Code formatting
├── history/        # File history tracking
├── llm/
│   ├── agent/      # Agent implementation
│   ├── models/     # Model definitions
│   ├── prompt/     # System prompts
│   ├── provider/   # LLM providers
│   └── tools/      # Tool implementations
├── logging/        # Logging infrastructure
├── lsp/            # LSP client and server
├── message/        # Message types
├── permission/     # Permission system
├── pubsub/         # Pub/sub messaging
├── session/        # Session management
├── skill/          # Skill system
├── tui/            # Terminal UI
└── version/        # Version info
```

### Key Interfaces

| Interface | Package | Purpose |
|-----------|---------|---------|
| `Provider` | `llm/provider` | LLM API abstraction |
| `BaseTool` | `llm/tools` | Tool implementation |
| `SessionProvider` | `db` | Data persistence |
| `Querier` | `db` | Database queries |

---

## Adding a New Provider

To add a new LLM provider:

1. **Create provider file**: `internal/llm/provider/newprovider.go`
2. **Implement Provider interface**:
   ```go
   type newProviderClient struct {
       model models.Model
       // ...
   }

   func NewProvider(opts ...ProviderClientOption) (Provider, error) { ... }

   func (p *newProviderClient) SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error) { ... }

   func (p *newProviderClient) StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent { ... }

   func (p *newProviderClient) Model() models.Model { ... }

   func (p *newProviderClient) CountTokens(ctx context.Context, threshold float64, messages []message.Message, tools []toolsPkg.BaseTool) (int64, bool) { ... }

   func (p *newProviderClient) AdjustMaxTokens(estimatedTokens int64) int64 { ... }
   ```
3. **Add model definitions**: Update `internal/llm/models/models.go`
4. **Register provider**: Add to config loading in `internal/config/`

---

## Configuration

Providers and models are configured in `.opencode.json`:

```json
{
  "model": {
    "provider": "anthropic",
    "model": "anthropic.claude-sonnet-4-5-m",
    "maxTokens": 64000
  },
  "apiKeys": {
    "ANTHROPIC_API_KEY": "sk-..."
  }
}
```

---

## See Also

- [Skills Guide](skills.md) - Skill system documentation
- [Session Providers](session-providers.md) - Database providers
- [LSP Documentation](lsp.md) - Language Server Protocol integration
- [Structured Output](structured-output.md) - Structured output generation
