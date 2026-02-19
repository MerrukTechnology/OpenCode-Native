package models
const (
	ProviderMistral ModelProvider = "mistral"

	MistralGPT4O    ModelID = "Mistral.gpt-4o"
)

var MistralModels = map[ModelID]Model{
	MistralGPT4O: {
		ID:                 MistralGPT4O,
		Name:               "GPT-4o (via Mistral)",
		Provider:           ProviderMistral,
		APIModel:           "openai/gpt-4o",
		CostPer1MIn:        0.20,
		CostPer1MInCached:  0.05,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      128000,
		DefaultMaxTokens:   16384,
	},
}
