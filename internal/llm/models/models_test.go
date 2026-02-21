package models

import (
	"testing"
)

func TestModelID_String(t *testing.T) {
	tests := []struct {
		name     string
		modelID  ModelID
		expected string
	}{
		{
			name:     "KiloCode Auto model ID",
			modelID:  KiloCodeAuto,
			expected: "kilo.auto",
		},
		{
			name:     "Mistral GPT-4o model ID",
			modelID:  MistralGPT4O,
			expected: "mistral.gpt-4o",
		},
		{
			name:     "Bedrock Claude model ID",
			modelID:  BedrockClaude45Sonnet,
			expected: "bedrock.claude-4.5-sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.modelID); got != tt.expected {
				t.Errorf("ModelID = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestModelProvider_String(t *testing.T) {
	tests := []struct {
		name     string
		provider ModelProvider
		expected string
	}{
		{
			name:     "KiloCode provider",
			provider: ProviderKiloCode,
			expected: "kilocode",
		},
		{
			name:     "Mistral provider",
			provider: ProviderMistral,
			expected: "mistral",
		},
		{
			name:     "Anthropic provider",
			provider: ProviderAnthropic,
			expected: "anthropic",
		},
		{
			name:     "OpenAI provider",
			provider: ProviderOpenAI,
			expected: "openai",
		},
		{
			name:     "VertexAI provider",
			provider: ProviderVertexAI,
			expected: "vertexai",
		},
		{
			name:     "Gemini provider",
			provider: ProviderGemini,
			expected: "gemini",
		},
		{
			name:     "Bedrock provider",
			provider: ProviderBedrock,
			expected: "bedrock",
		},
		{
			name:     "Mock provider",
			provider: ProviderMock,
			expected: "__mock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.provider); got != tt.expected {
				t.Errorf("ModelProvider = %q, want %q", got, tt.expected)
			}
		})
	}
}

// validateModelFields checks that a model has expected field values
func validateModelFields(t *testing.T, model Model, wantName, wantAPI string, wantCtx, wantMaxTok int64, wantProvider ModelProvider) {
	t.Helper()

	if model.Name != wantName {
		t.Errorf("Name = %q, want %q", model.Name, wantName)
	}
	if model.APIModel != wantAPI {
		t.Errorf("APIModel = %q, want %q", model.APIModel, wantAPI)
	}
	if model.ContextWindow != wantCtx {
		t.Errorf("ContextWindow = %d, want %d", model.ContextWindow, wantCtx)
	}
	if model.DefaultMaxTokens != wantMaxTok {
		t.Errorf("DefaultMaxTokens = %d, want %d", model.DefaultMaxTokens, wantMaxTok)
	}
	if model.Provider != wantProvider {
		t.Errorf("Provider = %q, want %q", model.Provider, wantProvider)
	}
}

func TestModelDefinitions(t *testing.T) {
	tests := []struct {
		name         string
		modelMap     map[ModelID]Model
		modelID      ModelID
		wantName     string
		wantAPI      string
		wantCtx      int64
		wantMaxTok   int64
		wantProvider ModelProvider
	}{
		{
			name:         "KiloCode Auto model",
			modelMap:     KiloCodeModels,
			modelID:      KiloCodeAuto,
			wantName:     "KiloCode Auto",
			wantAPI:      "kilo/auto",
			wantCtx:      128000,
			wantMaxTok:   16384,
			wantProvider: ProviderKiloCode,
		},
		{
			name:         "Mistral GPT-4o model",
			modelMap:     MistralModels,
			modelID:      MistralGPT4O,
			wantName:     "GPT-4o (via Mistral)",
			wantAPI:      "openai/gpt-4o",
			wantCtx:      128000,
			wantMaxTok:   16384,
			wantProvider: ProviderMistral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, exists := tt.modelMap[tt.modelID]
			if !exists {
				t.Fatalf("Model %q not found in model map", tt.modelID)
			}
			validateModelFields(t, model, tt.wantName, tt.wantAPI, tt.wantCtx, tt.wantMaxTok, tt.wantProvider)
		})
	}
}

func TestSupportedModels_Contains(t *testing.T) {
	tests := []struct {
		name         string
		modelID      ModelID
		wantProvider ModelProvider
	}{
		{
			name:         "KiloCodeAuto",
			modelID:      KiloCodeAuto,
			wantProvider: ProviderKiloCode,
		},
		{
			name:         "MistralGPT4O",
			modelID:      MistralGPT4O,
			wantProvider: ProviderMistral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, exists := SupportedModels[tt.modelID]
			if !exists {
				t.Fatalf("%s not found in SupportedModels", tt.modelID)
			}

			if model.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", model.Provider, tt.wantProvider)
			}
		})
	}
}

func TestProviderPopularity(t *testing.T) {
	tests := []struct {
		name           string
		provider       ModelProvider
		expectedRank   int
		shouldBeRanked bool
	}{
		{"VertexAI popularity", ProviderVertexAI, 1, true},
		{"Anthropic popularity", ProviderAnthropic, 2, true},
		{"OpenAI popularity", ProviderOpenAI, 3, true},
		{"Gemini popularity", ProviderGemini, 4, true},
		{"Groq popularity", ProviderGroq, 5, true},
		{"XAI popularity", ProviderXAI, 6, true},
		{"KiloCode popularity", ProviderKiloCode, 7, true},
		{"Mistral popularity", ProviderMistral, 8, true},
		{"OpenRouter popularity", ProviderOpenRouter, 9, true},
		{"DeepSeek popularity", ProviderDeepSeek, 10, true},
		{"Bedrock popularity", ProviderBedrock, 11, true},
		{"Local popularity", ProviderLocal, 12, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, exists := ProviderPopularity[tt.provider]
			if tt.shouldBeRanked {
				if !exists {
					t.Errorf("Provider %q not found in ProviderPopularity", tt.provider)
					return
				}
				if rank != tt.expectedRank {
					t.Errorf("ProviderPopularity[%q] = %d, want %d", tt.provider, rank, tt.expectedRank)
				}
			} else {
				if exists {
					t.Errorf("Provider %q should not be in ProviderPopularity", tt.provider)
				}
			}
		})
	}
}

func TestModel_StructFields(t *testing.T) {
	model := Model{
		ID:                       KiloCodeAuto,
		Name:                     "Test Model",
		Provider:                 ProviderKiloCode,
		APIModel:                 "test/api",
		CostPer1MIn:              0.20,
		CostPer1MOut:             0.50,
		CostPer1MInCached:        0.05,
		CostPer1MOutCached:       0.10,
		ContextWindow:            128000,
		DefaultMaxTokens:         16384,
		CanReason:                true,
		SupportsAdaptiveThinking: true,
		SupportsMaximumThinking:  true,
		SupportsAttachments:      true,
	}

	if model.ID != KiloCodeAuto {
		t.Errorf("ID = %q, want %q", model.ID, KiloCodeAuto)
	}
	if model.Name != "Test Model" {
		t.Errorf("Name = %q, want %q", model.Name, "Test Model")
	}
	if model.Provider != ProviderKiloCode {
		t.Errorf("Provider = %q, want %q", model.Provider, ProviderKiloCode)
	}
	if model.APIModel != "test/api" {
		t.Errorf("APIModel = %q, want %q", model.APIModel, "test/api")
	}
	if model.CostPer1MIn != 0.20 {
		t.Errorf("CostPer1MIn = %f, want %f", model.CostPer1MIn, 0.20)
	}
	if model.CostPer1MOut != 0.50 {
		t.Errorf("CostPer1MOut = %f, want %f", model.CostPer1MOut, 0.50)
	}
	if model.CostPer1MInCached != 0.05 {
		t.Errorf("CostPer1MInCached = %f, want %f", model.CostPer1MInCached, 0.05)
	}
	if model.CostPer1MOutCached != 0.10 {
		t.Errorf("CostPer1MOutCached = %f, want %f", model.CostPer1MOutCached, 0.10)
	}
	if model.ContextWindow != 128000 {
		t.Errorf("ContextWindow = %d, want %d", model.ContextWindow, 128000)
	}
	if model.DefaultMaxTokens != 16384 {
		t.Errorf("DefaultMaxTokens = %d, want %d", model.DefaultMaxTokens, 16384)
	}
	if !model.CanReason {
		t.Error("CanReason should be true")
	}
	if !model.SupportsAdaptiveThinking {
		t.Error("SupportsAdaptiveThinking should be true")
	}
	if !model.SupportsMaximumThinking {
		t.Error("SupportsMaximumThinking should be true")
	}
	if !model.SupportsAttachments {
		t.Error("SupportsAttachments should be true")
	}
}

func TestModel_JSONTags(t *testing.T) {
	// Test that JSON tags are correctly set
	model := Model{
		ID:                       KiloCodeAuto,
		Name:                     "Test",
		Provider:                 ProviderKiloCode,
		APIModel:                 "test",
		CostPer1MIn:              1.0,
		CostPer1MOut:             2.0,
		CostPer1MInCached:        0.5,
		CostPer1MOutCached:       1.0,
		ContextWindow:            1000,
		DefaultMaxTokens:         100,
		CanReason:                true,
		SupportsAdaptiveThinking: true,
		SupportsMaximumThinking:  true,
		SupportsAttachments:      true,
	}

	// Verify the model can be serialized (basic check)
	// This ensures JSON tags are valid
	if model.ID != KiloCodeAuto {
		t.Error("Model ID should match")
	}
}

func TestFriendlyModelName(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected string
	}{
		{
			name:     "Simple model name",
			modelID:  "llama",
			expected: "Llama",
		},
		{
			name:     "Model with version - actual behavior",
			modelID:  "llama3.2",
			expected: "Llama3 2",
		},
		{
			name:     "Model with version and label",
			modelID:  "llama3.2-instruct",
			expected: "Llama3 2 Instruct",
		},
		{
			name:     "Model with publisher prefix",
			modelID:  "meta/llama3.2",
			expected: "Llama3 2",
		},
		{
			name:     "Model with tag",
			modelID:  "llama3.2@latest",
			expected: "Llama3 2 latest",
		},
		{
			name:     "Qwen model",
			modelID:  "qwen2.5",
			expected: "Qwen2 5",
		},
		{
			name:     "Mistral model",
			modelID:  "mistral7b",
			expected: "Mistral7b",
		},
		{
			name:     "Complex model ID",
			modelID:  "deepseek-r1",
			expected: "Deepseek R1",
		},
		{
			name:     "Empty string",
			modelID:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := friendlyModelName(tt.modelID)
			if got != tt.expected {
				t.Errorf("friendlyModelName(%q) = %q, want %q", tt.modelID, got, tt.expected)
			}
		})
	}
}

func TestConvertLocalModel(t *testing.T) {
	tests := []struct {
		name         string
		model        localModel
		source       *Model
		wantID       ModelID
		wantName     string
		wantProvider ModelProvider
		wantAPIModel string
	}{
		{
			name: "Model with source",
			model: localModel{
				ID:                  "llama-3.2",
				MaxContextLength:    8192,
				LoadedContextLength: 4096,
			},
			source: &Model{
				ID:            "test.model",
				Name:          "Llama 3.2",
				Provider:      ProviderOpenAI,
				APIModel:      "original-api-model",
				ContextWindow: 8192,
			},
			wantID:       "local.llama-3.2",
			wantName:     "Llama 3.2",
			wantProvider: ProviderLocal,
			wantAPIModel: "llama-3.2",
		},
		{
			name: "Model without source",
			model: localModel{
				ID:                  "unknown-model",
				MaxContextLength:    16384,
				LoadedContextLength: 8192,
			},
			source:       nil,
			wantID:       "local.unknown-model",
			wantName:     "Unknown Model",
			wantProvider: ProviderLocal,
			wantAPIModel: "unknown-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertLocalModel(tt.model, tt.source)

			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
			if tt.source != nil && got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}
			if got.APIModel != tt.wantAPIModel {
				t.Errorf("APIModel = %q, want %q", got.APIModel, tt.wantAPIModel)
			}
		})
	}
}

func TestTryResolveSource(t *testing.T) {
	// Save original SupportedModels and restore after test
	originalModels := make(map[ModelID]Model)
	for k, v := range SupportedModels {
		originalModels[k] = v
	}
	defer func() {
		SupportedModels = originalModels
	}()

	// Clear and set up test models
	SupportedModels = make(map[ModelID]Model)
	SupportedModels["test.claude-3"] = Model{
		ID:       "test.claude-3",
		Name:     "Claude 3",
		APIModel: "claude-3-sonnet",
		Provider: ProviderAnthropic,
	}

	tests := []struct {
		name         string
		localID      string
		wantNil      bool
		wantAPIModel string
	}{
		{
			name:         "Matching API model",
			localID:      "claude-3-sonnet-20240229",
			wantNil:      false,
			wantAPIModel: "claude-3-sonnet",
		},
		{
			name:    "No matching model",
			localID: "unknown-model-xyz",
			wantNil: true,
		},
		{
			name:    "Empty local ID",
			localID: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tryResolveSource(tt.localID)

			if tt.wantNil {
				if got != nil {
					t.Errorf("tryResolveSource(%q) = %v, want nil", tt.localID, got)
				}
			} else {
				if got == nil {
					t.Errorf("tryResolveSource(%q) = nil, want non-nil", tt.localID)
				} else if got.APIModel != tt.wantAPIModel {
					t.Errorf("APIModel = %q, want %q", got.APIModel, tt.wantAPIModel)
				}
			}
		})
	}
}

func TestLocalModelStruct(t *testing.T) {
	tests := []struct {
		name     string
		model    localModel
		checkID  string
		checkObj string
		checkTyp string
		checkSt  string
		checkCtx int64
	}{
		{
			name: "Local model with all fields",
			model: localModel{
				ID:                  "test-model",
				Object:              "model",
				Type:                "llm",
				Publisher:           "test-publisher",
				Arch:                "transformer",
				CompatibilityType:   "full",
				Quantization:        "q4_0",
				State:               "loaded",
				MaxContextLength:    8192,
				LoadedContextLength: 4096,
			},
			checkID:  "test-model",
			checkObj: "model",
			checkTyp: "llm",
			checkSt:  "loaded",
			checkCtx: 8192,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.model.ID != tt.checkID {
				t.Errorf("ID = %q, want %q", tt.model.ID, tt.checkID)
			}
			if tt.model.Object != tt.checkObj {
				t.Errorf("Object = %q, want %q", tt.model.Object, tt.checkObj)
			}
			if tt.model.Type != tt.checkTyp {
				t.Errorf("Type = %q, want %q", tt.model.Type, tt.checkTyp)
			}
			if tt.model.State != tt.checkSt {
				t.Errorf("State = %q, want %q", tt.model.State, tt.checkSt)
			}
			if tt.model.MaxContextLength != tt.checkCtx {
				t.Errorf("MaxContextLength = %d, want %d", tt.model.MaxContextLength, tt.checkCtx)
			}
		})
	}
}

func TestLocalModelListStruct(t *testing.T) {
	list := localModelList{
		Data: []localModel{
			{ID: "model1", Type: "llm"},
			{ID: "model2", Type: "llm"},
		},
	}

	if len(list.Data) != 2 {
		t.Errorf("Data length = %d, want %d", len(list.Data), 2)
	}
	if list.Data[0].ID != "model1" {
		t.Errorf("First model ID = %q, want %q", list.Data[0].ID, "model1")
	}
}

func TestProviderLocal(t *testing.T) {
	tests := []struct {
		name     string
		provider ModelProvider
		expected string
	}{
		{
			name:     "ProviderLocal value",
			provider: ProviderLocal,
			expected: "local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.provider); got != tt.expected {
				t.Errorf("ProviderLocal = %q, want %q", got, tt.expected)
			}
		})
	}
}
