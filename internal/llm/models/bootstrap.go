package models

import "maps"

func init() {
	maps.Copy(SupportedModels, AnthropicModels)
	maps.Copy(SupportedModels, OpenAIModels)
	maps.Copy(SupportedModels, GeminiModels)
	maps.Copy(SupportedModels, GroqModels)
	maps.Copy(SupportedModels, XAIModels)
	maps.Copy(SupportedModels, OpenRouterModels)
	maps.Copy(SupportedModels, DeepSeekModels)
	maps.Copy(SupportedModels, VertexAIGeminiModels)
	maps.Copy(SupportedModels, VertexAIAnthropicModels)

	initLocalModels()
}
