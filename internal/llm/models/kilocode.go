package models

const (
	ProviderKiloCode ModelProvider = "kilocode"

	KiloCodeAuto ModelID = "kilo.auto"
)

var KiloCodeModels = map[ModelID]Model{
	KiloCodeAuto: {
		ID:                 KiloCodeAuto,
		Name:               "KiloCode Auto",
		Provider:           ProviderKiloCode,
		APIModel:           "kilo/auto",
		CostPer1MIn:        0.20,
		CostPer1MInCached:  0.05,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      128000,
		DefaultMaxTokens:   16384,
	},
}
