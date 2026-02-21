# OpenCode Providers and Models Guide

> **Note:** This document describes the provider and model architecture in OpenCode. This content is not yet integrated into the main documentation.

---

## Table of Contents

1. [Providers Overview](#providers-overview)
2. [OpenAI-Compatible API Providers](#openai-compatible-api-providers)
3. [Models by Provider](#models-by-provider)
4. [Tools](#tools)
5. [Package Architecture](#package-architecture)

---

## Providers Overview

OpenCode supports multiple LLM providers through a unified [`Provider`](internal/llm/provider/provider.go) interface. Each provider is implemented in the [`internal/llm/provider/`](internal/llm/provider/) directory.

### Native Providers (Custom Implementation)

These providers have dedicated client implementations:

| Provider | File | Description |
|----------|------|-------------|
| **OpenAI** | [`openai.go`](internal/llm/provider/openai.go) | OpenAI API (GPT-4, GPT-4o, O1, O3, GPT-5) |
| **Anthropic** | [`anthropic.go`](internal/llm/provider/anthropic.go) | Anthropic API (Claude 3, 4, 5 series) |
| **Google Gemini** | [`gemini.go`](internal/llm/provider/gemini.go) | Google Gemini API (2.0, 2.5, 3.0 series) |
| **Vertex AI** | [`vertexai.go`](internal/llm/provider/vertexai.go) | Google Vertex AI (Claude & Gemini via GCP) |
| **DeepSeek** | [`deepseek.go`](internal/llm/provider/deepseek.go) | DeepSeek API (Chat, Reasoner) |
| **Bedrock** | [`bedrock.go`](internal/llm/provider/bedrock.go) | AWS Bedrock (Claude via Bedrock) |

### OpenAI-Compatible API Providers

The following providers use the **OpenAI provider client** with a custom base URL. They implement the OpenAI Chat Completions API specification:

| Provider | Base URL | Description |
|----------|----------|-------------|
| **Groq** | `https://api.groq.com/openai/v1` | Fast inference for Llama, Qwen, DeepSeek R1 |
| **XAI** | `https://api.x.ai/v1` | xAI API (Grok models) |
| **OpenRouter** | `https://openrouter.ai/api/v1` | Unified API to 200+ models |
| **Mistral** | `https://api.mistral.ai/v1` | Mistral AI models |
| **KiloCode** | `https://api.kilo.ai/api/gateway` | KiloCode gateway |
| **Local/Ollama** | `LOCAL_ENDPOINT` env var | Local models via Ollama, LM Studio |

> **Important:** When adding support for a new OpenAI-compatible API provider, you only need to:

> 1. Add model definitions to [`internal/llm/models/`](internal/llm/models/)
> 2. Register the provider in [`provider.go`](internal/llm/provider/provider.go) with the appropriate base URL

---

## Models by Provider

### Anthropic Models

Official Anthropic models via [`anthropic.go`](internal/llm/models/anthropic.go):

| Model ID | Display Name | Context Window | Max Output | Reasoning | Attachments |
|----------|--------------|----------------|------------|-----------|-------------|
| `claude-3.5-sonnet` | Claude 3.5 Sonnet | 200K | 5K | ❌ | ✅ |
| `claude-3-haiku` | Claude 3 Haiku | 200K | 4K | ❌ | ✅ |
| `claude-3.7-sonnet` | Claude 3.7 Sonnet | 200K | 50K | ✅ | ✅ |
| `claude-3.5-haiku` | Claude 3.5 Haiku | 200K | 4K | ❌ | ✅ |
| `claude-3-opus` | Claude 3 Opus | 200K | 4K | ❌ | ✅ |
| `claude-4-sonnet` | Claude 4 Sonnet | 200K | 50K | ✅ | ✅ |
| `claude-4-opus` | Claude 4 Opus | 200K | 4K | ❌ | ✅ |
| `claude-4-5-sonnet[1m]` | Claude 4.5 Sonnet | 1M | 64K | ✅ | ✅ |
| `claude-4.5-opus` | Claude 4.5 Opus | 200K | 32K | ✅ | ✅ |
| `claude-4.6-opus` | Claude 4.6 Opus | 1M | 128K | ✅ | ✅ |
| `claude-4.6-sonnet` | Claude 4.6 Sonnet | 1M | 128K | ✅ | ✅ |

### OpenAI Models

Official OpenAI models via [`openai.go`](internal/llm/models/openai.go):

| Model ID | Display Name | Context Window | Max Output | Reasoning | Attachments |
|----------|--------------|----------------|------------|-----------|-------------|
| `gpt-4.1` | GPT 4.1 | 1M | 20K | ❌ | ✅ |
| `gpt-4.1-mini` | GPT 4.1 mini | 200K | 20K | ❌ | ✅ |
| `gpt-4.1-nano` | GPT 4.1 nano | 1M | 20K | ❌ | ✅ |
| `gpt-4.5-preview` | GPT 4.5 preview | 128K | 15K | ❌ | ✅ |
| `gpt-4o` | GPT 4o | 128K | 4K | ❌ | ✅ |
| `gpt-4o-mini` | GPT 4o mini | 128K | - | ❌ | ✅ |
| `o1` | O1 | 200K | 50K | ✅ | ✅ |
| `o1-pro` | O1 Pro | 200K | 50K | ✅ | ✅ |
| `o1-mini` | O1 mini | 128K | 50K | ✅ | ✅ |
| `o3` | O3 | 200K | - | ✅ | ✅ |
| `o3-mini` | O3 mini | 200K | 50K | ✅ | ❌ |
| `o4-mini` | O4 mini | 128K | 50K | ✅ | ✅ |
| `gpt-5` | GPT 5 | 400K | 128K | ✅ | ✅ |

### Google Gemini Models

Official Gemini models via [`gemini.go`](internal/llm/models/gemini.go):

| Model ID | Display Name | Context Window | Max Output | Reasoning | Attachments |
|----------|--------------|----------------|------------|-----------|-------------|
| `gemini-flash-2.0` | Gemini 2.0 Flash | 1M | 6K | ❌ | ✅ |
| `gemini-2.0-flash-lite` | Gemini 2.0 Flash Lite | 1M | 6K | ❌ | ✅ |
| `gemini-2.5-flash` | Gemini 2.5 Flash | 1M | 50K | ❌ | ✅ |
| `gemini-2.5` | Gemini 2.5 Pro | 1M | 50K | ❌ | ✅ |
| `gemini-3.0-pro` | Gemini 3.0 Pro | 1M | 64K | ✅ | ✅ |
| `gemini-3.0-flash` | Gemini 3.0 Flash | 1M | 64K | ✅ | ✅ |
| `gemini-3.1-pro-preview` | Gemini 3.1 Pro Preview | 1M | 64K | ✅ | ✅ |
| `gemini-3.1-flash-preview` | Gemini 3.1 Flash Preview | 1M | 64K | ✅ | ✅ |

### DeepSeek Models

DeepSeek models via [`deepseek.go`](internal/llm/models/deepseek.go):

| Model ID | Display Name | Context Window | Max Output | Reasoning | Attachments |
|----------|--------------|----------------|------------|-----------|-------------|
| `deepseek-chat` | DeepSeek Chat | 131K | 8K | ❌ | ✅ |
| `deepseek-reasoner` | DeepSeek Reasoner | 131K | 64K | ✅ | ✅ |

### Groq Models

Groq models via [`groq.go`](internal/llm/models/groq.go) (OpenAI-compatible):

| Model ID | Display Name | Context Window | Max Output | Reasoning | Attachments |
|----------|--------------|----------------|------------|-----------|-------------|
| `qwen-qwq` | Qwen Qwq | 128K | 50K | ❌ | ❌ |
| `meta-llama/llama-4-scout-17b-16e-instruct` | Llama 4 Scout | 128K | - | ❌ | ✅ |
| `meta-llama/llama-4-maverick-17b-128e-instruct` | Llama 4 Maverick | 128K | - | ❌ | ✅ |
| `llama-3.3-70b-versatile` | Llama 3.3 70B Versatile | 128K | - | ❌ | ❌ |
| `deepseek-r1-distill-llama-70b` | DeepSeek R1 Distill Llama 70B | 128K | - | ✅ | ❌ |
| `moonshotai/kimi-k2-instruct-0905` | Kimi K2 | 131K | 16K | ❌ | ✅ |

### XAI (Grok) Models

Grok models via [`xai.go`](internal/llm/models/xai.go) (OpenAI-compatible):

| Model ID | Display Name | Context Window | Max Output | Reasoning |
|----------|--------------|----------------|------------|-----------|
| `grok-4-1-fast-reasoning` | Grok 4.1 Fast Reasoning | 2M | 64K | ✅ |
| `grok-4-1-fast-non-reasoning` | Grok 4.1 Fast Non-Reasoning | 2M | 16K | ❌ |
| `grok-code-fast-1` | Grok Code Fast 1 | 256K | 32K | ❌ |
| `grok-4-fast-reasoning` | Grok 4 Fast Reasoning | 2M | 64K | ✅ |
| `grok-4-fast-non-reasoning` | Grok 4 Fast Non-Reasoning | 2M | 16K | ❌ |
| `grok-4-0709` | Grok 4 0709 | 256K | 20K | ❌ |

### OpenRouter Models

OpenRouter models via [`openrouter.go`](internal/llm/models/openrouter.go) (OpenAI-compatible). OpenRouter provides access to 200+ models:

| Model ID | Display Name | Context Window | Notes |
|----------|--------------|----------------|-------|
| `openrouter.free` | OpenRouter Free Router | 200K | Auto-selects free models |
| `openrouter.gpt-4.1` | GPT 4.1 | 1M | Via OpenAI |
| `openrouter.gpt-4o` | GPT 4o | 128K | Via OpenAI |
| `openrouter.gpt-4o-mini` | GPT 4o mini | 128K | Via OpenAI |
| `openrouter.o1` | O1 | 200K | Via OpenAI |
| `openrouter.o1-mini` | O1 mini | 128K | Via OpenAI |
| `openrouter.o3` | O3 | 200K | Via OpenAI |
| `openrouter.gemini-2.5-flash` | Gemini 2.5 Flash | 1M | Via Google |
| `openrouter.gemini-2.5` | Gemini 2.5 Pro | 1M | Via Google |
| `openrouter.claude-3.5-sonnet` | Claude 3.5 Sonnet | 200K | Via Anthropic |
| `openrouter.claude-3-haiku` | Claude 3 Haiku | 200K | Via Anthropic |
| `openrouter.deepseek-r1-free` | DeepSeek R1 (Free) | 164K | Free tier |
| `openrouter.deepseek-v3.2` | DeepSeek V3.2 | 128K | Via DeepSeek |
| `openrouter.devstral-2` | Devstral 2 | 256K | Via Mistral |
| `openrouter.mimo-v2` | MiMo-V2 | 256K | Via Xiaomi |
| `openrouter.grok-4-fast` | Grok 4 Fast | 2M | Via xAI |
| `openrouter.grok-4-fast:free` | Grok 4 Fast (Free) | 2M | Free tier |
| `openrouter.minimax-01` | MiniMax 01 | 1M | Via MiniMax |
| `openrouter.minimax-m2.5` | MiniMax M2.5 | 1M | Latest MoE model |
| `openrouter.nemotron-3-nano` | Nemotron 3 Nano | 262K | Via NVIDIA |
| `openrouter.glm-4.7-flash` | GLM 4.7 Flash | 200K | Via Z.AI |
| `openrouter.kimi-k2` | Kimi K2 | 200K | Via Moonshot |

### Vertex AI Models

Vertex AI models via [`vertexai.go`](internal/llm/models/vertexai.go):

| Model ID | Display Name | Context Window | Provider |
|----------|--------------|----------------|----------|
| `vertexai.gemini-3.0-pro` | Gemini 3.0 Pro | 1M | Google |
| `vertexai.gemini-3.0-flash` | Gemini 3.0 Flash | 1M | Google |
| `vertexai.claude-sonnet-4-5-m` | Claude Sonnet 4.5 [1m] | 1M | Anthropic |
| `vertexai.claude-opus-4-5` | Claude Opus 4.5 | 200K | Anthropic |
| `vertexai.claude-opus-4-6` | Claude Opus 4.6 | 1M | Anthropic |
| `vertexai.claude-sonnet-4-6` | Claude Sonnet 4.6 | 1M | Anthropic |

### Bedrock Models

AWS Bedrock models via [`bedrock.go`](internal/llm/provider/bedrock.go):

Uses Anthropic models via AWS Bedrock. Requires AWS credentials and region configuration.

### Local/Ollama Models

Local models via [`local.go`](internal/llm/models/local.go). These are dynamically loaded from:

- **Ollama**: `http://localhost:11434/v1/models`
- **LM Studio**: `http://localhost:1234/api/v0/models`

Configuration:

```bash
export LOCAL_ENDPOINT=http://localhost:11434
export LOCAL_ENDPOINT_API_KEY=ollama  # Optional
```

### Mistral Models

Mistral models via [`mistral.go`](internal/llm/models/mistral.go) (OpenAI-compatible):

| Model ID | Display Name | Context Window | Max Output |
|----------|--------------|----------------|------------|
| `mistral.gpt-4o` | GPT-4o (via Mistral) | 128K | 16K |

### KiloCode Models

KiloCode models via [`kilocode.go`](internal/llm/models/kilocode.go) (OpenAI-compatible):

| Model ID | Display Name | Context Window | Max Output |
|----------|--------------|----------------|------------|
| `kilo.auto` | KiloCode Auto | 128K | 16K |

---

## Provider Interface

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

---

## Model Struct

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

## Adding a New Model

To add a new model:

1. **Add model definition**: Edit the appropriate file in [`internal/llm/models/`](internal/llm/models/)
2. **Register model**: Add to the model's map (e.g., `AnthropicModels`, `OpenAIModels`)
3. **Update documentation**: Add the model to this guide

### Example: Adding an Anthropic Model

```go
// internal/llm/models/anthropic.go

const (
    // Add new model ID
    ClaudeNewModel ModelID = "claude-new-model"
)

var AnthropicModels = map[ModelID]Model{
    // Add model definition
    ClaudeNewModel: {
        ID:                  ClaudeNewModel,
        Name:                "Claude New Model",
        Provider:            ProviderAnthropic,
        APIModel:            "claude-new-model-latest",
        CostPer1MIn:         3.0,
        CostPer1MOut:        15.0,
        ContextWindow:       200000,
        DefaultMaxTokens:    5000,
        SupportsAttachments: true,
    },
}
```

---

## Adding an OpenAI-Compatible Provider

To add a new provider that uses the OpenAI API format:

1. **Add model definitions** to a new file in [`internal/llm/models/`](internal/llm/models/)
2. **Register provider** in [`internal/llm/provider/provider.go`](internal/llm/provider/provider.go):

```go
case models.ProviderNewProvider:
    clientOptions.openaiOptions = append(clientOptions.openaiOptions,
        WithOpenAIBaseURL("https://api.newprovider.com/v1"),
        WithOpenAIExtraHeaders(map[string]string{
            "HTTP-Referer": "opencode.ai",
            "X-Title":      "opencode",
        }),
    )
    return &baseProvider[OpenAIClient]{
        options: clientOptions,
        client:  newOpenAIClient(clientOptions),
    }, nil
```

---

## Configuration

Providers and models are configured in `.opencode.json`:

```json
{
  "model": {
    "provider": "anthropic",
    "model": "claude-4.5-sonnet[1m]",
    "maxTokens": 64000
  },
  "apiKeys": {
    "ANTHROPIC_API_KEY": "sk-..."
  }
}
```

### Environment Variables

| Variable | Provider | Description |
|----------|----------|-------------|
| `OPENAI_API_KEY` | OpenAI | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic | Anthropic API key |
| `GEMINI_API_KEY` | Gemini | Google Gemini API key |
| `DEEPSEEK_API_KEY` | DeepSeek | DeepSeek API key |
| `GROQ_API_KEY` | Groq | Groq API key |
| `XAI_API_KEY` | XAI | xAI API key |
| `OPENROUTER_API_KEY` | OpenRouter | OpenRouter API key |
| `MISTRAL_API_KEY` | Mistral | Mistral API key |
| `KILO_API_KEY` | KiloCode | KiloCode API key |
| `LOCAL_ENDPOINT` | Local | Ollama/LM Studio endpoint |
| `AWS_REGION` | Bedrock | AWS region |
| `VERTEXAI_PROJECT` | Vertex AI | GCP project ID |
| `VERTEXAI_LOCATION` | Vertex AI | GCP location |

---

## See Also

- [Skills Guide](skills.md) - Skill system documentation
- [Session Providers](session-providers.md) - Database providers
- [LSP Documentation](lsp.md) - Language Server Protocol integration
- [Structured Output](structured-output.md) - Structured output generation
