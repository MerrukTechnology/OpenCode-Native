package models

const (
	ProviderXAI ModelProvider = "xai"

	XAIGrok41FastReasoning    ModelID = "grok-4-1-fast-reasoning"
	XAIGrok41FastNonReasoning ModelID = "grok-4-1-fast-non-reasoning"
	XAIGrokCodeFast1          ModelID = "grok-code-fast-1"
	XAIGrok4FastReasoning     ModelID = "grok-4-fast-reasoning"
	XAIGrok4FastNonReasoning  ModelID = "grok-4-fast-non-reasoning"
	XAIGrok40709              ModelID = "grok-4-0709"
)

var XAIModels = map[ModelID]Model{
	XAIGrok41FastReasoning: {
		ID:                 XAIGrok41FastReasoning,
		Name:               "Grok 4.1 Fast Reasoning",
		Provider:           ProviderXAI,
		APIModel:           "grok-4-1-fast-reasoning",
		CostPer1MIn:        0.20,
		CostPer1MInCached:  0.05,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      2_000_000,
		DefaultMaxTokens:   64_000,
	},
	XAIGrok41FastNonReasoning: {
		ID:                 XAIGrok41FastNonReasoning,
		Name:               "Grok 4.1 Fast Non-Reasoning",
		Provider:           ProviderXAI,
		APIModel:           "grok-4-1-fast-non-reasoning",
		CostPer1MIn:        0.20,
		CostPer1MInCached:  0.05,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      2_000_000,
		DefaultMaxTokens:   16_000,
	},
	XAIGrokCodeFast1: {
		ID:                 XAIGrokCodeFast1,
		Name:               "Grok Code Fast 1",
		Provider:           ProviderXAI,
		APIModel:           "grok-code-fast-1",
		CostPer1MIn:        0.20,
		CostPer1MInCached:  0.02,
		CostPer1MOut:       1.50,
		CostPer1MOutCached: 0,
		ContextWindow:      256_000,
		DefaultMaxTokens:   32_000,
	},
	XAIGrok4FastReasoning: {
		ID:                 XAIGrok4FastReasoning,
		Name:               "Grok 4 Fast Reasoning",
		Provider:           ProviderXAI,
		APIModel:           "grok-4-fast-reasoning",
		CostPer1MIn:        0.20,
		CostPer1MInCached:  0.05,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      2_000_000,
		DefaultMaxTokens:   64_000,
	},
	XAIGrok4FastNonReasoning: {
		ID:                 XAIGrok4FastNonReasoning,
		Name:               "Grok 4 Fast Non-Reasoning",
		Provider:           ProviderXAI,
		APIModel:           "grok-4-fast-non-reasoning",
		CostPer1MIn:        0.20,
		CostPer1MInCached:  0.05,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      2_000_000,
		DefaultMaxTokens:   16_000,
	},
	XAIGrok40709: {
		ID:                 XAIGrok40709,
		Name:               "Grok 4 0709",
		Provider:           ProviderXAI,
		APIModel:           "grok-4-0709",
		CostPer1MIn:        3.0,
		CostPer1MInCached:  0.75,
		CostPer1MOut:       15.0,
		CostPer1MOutCached: 0.75,
		ContextWindow:      256_000,
		DefaultMaxTokens:   20_000,
	},
}
