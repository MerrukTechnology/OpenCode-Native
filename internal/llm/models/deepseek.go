package models

const (
	ProviderDeepSeek ModelProvider = "deepseek"

	DeepSeekChat     ModelID = "deepseek-chat"
	DeepSeekReasoner ModelID = "deepseek-reasoner"
)

var DeepSeekModels = map[ModelID]Model{
	DeepSeekChat: {
		ID:                  DeepSeekChat,
		Name:                "DeepSeek Chat",
		Provider:            ProviderDeepSeek,
		APIModel:            "deepseek-chat",
		CostPer1MIn:         0.28,
		CostPer1MInCached:   0.028,
		CostPer1MOut:        0.42,
		ContextWindow:       131_072,
		DefaultMaxTokens:    8192,
		SupportsAttachments: true,
	},
	DeepSeekReasoner: {
		ID:                  DeepSeekReasoner,
		Name:                "DeepSeek Reasoner",
		Provider:            ProviderDeepSeek,
		APIModel:            "deepseek-reasoner",
		CostPer1MIn:         0.28,
		CostPer1MInCached:   0.028,
		CostPer1MOut:        0.42,
		ContextWindow:       131072,
		DefaultMaxTokens:    65536,
		CanReason:           true,
		SupportsAttachments: true, // Reasoner supports tools!
	},
}
