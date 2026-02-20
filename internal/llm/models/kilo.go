package models

const (
	ProviderKilo ModelProvider = "kilo"

	KiloGPT4O ModelID = "kilo.gpt-4o"
)

var KiloModels = map[ModelID]Model{
	KiloGPT4O: {
		ID:                 KiloGPT4O,
		Name:               "GPT-4o (via Kilo)",
		Provider:           ProviderKilo,
		APIModel:           "openai/gpt-4o",
		CostPer1MIn:        0.20,
		CostPer1MInCached:  0.05,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      128000,
		DefaultMaxTokens:   16384,
	},
}
