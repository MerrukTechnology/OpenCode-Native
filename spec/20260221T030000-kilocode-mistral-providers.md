# KiloCode and Mistral Provider Support

**Date:** 2026-02-21
**Status:** Implemented
**Author:** OpenCode Team

## Overview

This specification documents the addition of KiloCode and Mistral as supported LLM providers in OpenCode. Both providers are configured via environment variables and are automatically detected when their respective API keys are present.

## Motivation

Users requested support for additional LLM providers to expand the range of available models. KiloCode offers an auto model that intelligently routes to the best available model, while Mistral provides access to GPT-4o through their API.

## Implementation

### Provider Constants

```go
// internal/llm/models/kilocode.go
const (
    ProviderKiloCode ModelProvider = "kilocode"
    KiloCodeAuto     ModelID       = "kilo.auto"
)

// internal/llm/models/mistral.go
const (
    ProviderMistral ModelProvider = "mistral"
    MistralGPT4O    ModelID       = "Mistral.gpt-4o"
)
```

### Model Definitions

#### KiloCode Auto

| Property | Value |
|----------|-------|
| ID | `kilo.auto` |
| Name | KiloCode Auto |
| Provider | `kilocode` |
| API Model | `kilo/auto` |
| Context Window | 128,000 tokens |
| Default Max Tokens | 16,384 |
| Cost Per 1M Input | $0.20 |
| Cost Per 1M Output | $0.50 |

#### Mistral GPT-4o

| Property | Value |
|----------|-------|
| ID | `Mistral.gpt-4o` |
| Name | GPT-4o (via Mistral) |
| Provider | `mistral` |
| API Model | `openai/gpt-4o` |
| Context Window | 128,000 tokens |
| Default Max Tokens | 16,384 |
| Cost Per 1M Input | $0.20 |
| Cost Per 1M Output | $0.50 |

### Environment Variables

| Variable | Provider | Description |
|----------|----------|-------------|
| `KILO_API_KEY` | KiloCode | API key for KiloCode models |
| `MISTRAL_API_KEY` | Mistral | API key for Mistral models |

### Provider Detection

Providers are automatically enabled when their API keys are detected. The TUI model selection dialog checks both:

1. **Configuration-based providers**: Providers explicitly configured in `.opencode.json`
2. **Environment-based providers**: Providers with API keys available via environment variables

This is implemented in [`internal/tui/components/dialog/models.go`](../internal/tui/components/dialog/models.go) via the `getEnabledProviders()` function.

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

## Usage

### Configuration

Add to `.opencode.json`:

```json
{
  "providers": {
    "kilocode": {
      "apiKey": "your-kilo-api-key"
    },
    "mistral": {
      "apiKey": "your-mistral-api-key"
    }
  },
  "agents": {
    "coder": {
      "model": "kilo.auto"
    }
  }
}
```

### Environment Variables

```bash
# KiloCode
export KILO_API_KEY=your-kilo-api-key

# Mistral
export MISTRAL_API_KEY=your-mistral-api-key
```

### Model Selection

In the TUI, press `Ctrl+O` to open the model selection dialog. KiloCode and Mistral will appear in the provider list if their API keys are configured.

## Files Modified

| File | Changes |
|------|---------|
| `internal/llm/models/kilocode.go` | New file - KiloCode model definitions |
| `internal/llm/models/mistral.go` | New file - Mistral model definitions |
| `internal/config/config.go` | Exported `GetProviderAPIKey()` function |
| `internal/tui/components/dialog/models.go` | Added environment-based provider detection |
| `README.md` | Updated provider list and environment variables |
| `docs/providers-models.md` | Updated provider documentation |

## Testing

Unit tests verify:
- Model ID constants are correctly defined
- Model definitions have required fields
- Provider detection works with environment variables

## Future Considerations

1. **Additional Models**: Both providers may add more models in the future
2. **Streaming Support**: Ensure streaming responses work correctly
3. **Token Counting**: Verify accurate token counting for cost estimation

## See Also

- [Providers and Models Guide](../docs/providers-models.md)
- [Configuration Documentation](../README.md#configuration)
