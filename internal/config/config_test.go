package config

import (
	"os"
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/models"
)

func TestGetProviderAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		provider  models.ModelProvider
		envKey    string
		envValue  string
		wantValue string
		wantEmpty bool
	}{
		{
			name:      "Anthropic API key from environment",
			provider:  models.ProviderAnthropic,
			envKey:    "ANTHROPIC_API_KEY",
			envValue:  "test-anthropic-key",
			wantValue: "test-anthropic-key",
		},
		{
			name:      "OpenAI API key from environment",
			provider:  models.ProviderOpenAI,
			envKey:    "OPENAI_API_KEY",
			envValue:  "test-openai-key",
			wantValue: "test-openai-key",
		},
		{
			name:      "Gemini API key from environment",
			provider:  models.ProviderGemini,
			envKey:    "GEMINI_API_KEY",
			envValue:  "test-gemini-key",
			wantValue: "test-gemini-key",
		},
		{
			name:      "Groq API key from environment",
			provider:  models.ProviderGroq,
			envKey:    "GROQ_API_KEY",
			envValue:  "test-groq-key",
			wantValue: "test-groq-key",
		},
		{
			name:      "KiloCode API key from environment",
			provider:  models.ProviderKiloCode,
			envKey:    "KILO_API_KEY",
			envValue:  "test-kilo-key",
			wantValue: "test-kilo-key",
		},
		{
			name:      "Mistral API key from environment",
			provider:  models.ProviderMistral,
			envKey:    "MISTRAL_API_KEY",
			envValue:  "test-mistral-key",
			wantValue: "test-mistral-key",
		},
		{
			name:      "OpenRouter API key from environment",
			provider:  models.ProviderOpenRouter,
			envKey:    "OPENROUTER_API_KEY",
			envValue:  "test-openrouter-key",
			wantValue: "test-openrouter-key",
		},
		{
			name:      "XAI API key from environment",
			provider:  models.ProviderXAI,
			envKey:    "XAI_API_KEY",
			envValue:  "test-xai-key",
			wantValue: "test-xai-key",
		},
		{
			name:      "DeepSeek API key from environment",
			provider:  models.ProviderDeepSeek,
			envKey:    "DEEPSEEK_API_KEY",
			envValue:  "test-deepseek-key",
			wantValue: "test-deepseek-key",
		},
		{
			name:      "Empty API key when not set",
			provider:  models.ProviderAnthropic,
			envKey:    "ANTHROPIC_API_KEY",
			envValue:  "",
			wantEmpty: true,
		},
		{
			name:      "Unknown provider returns empty",
			provider:  models.ModelProvider("unknown"),
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env value
			originalValue := os.Getenv(tt.envKey)
			defer os.Setenv(tt.envKey, originalValue)

			// Set test value
			if tt.envKey != "" {
				if tt.envValue != "" {
					os.Setenv(tt.envKey, tt.envValue)
				} else {
					os.Unsetenv(tt.envKey)
				}
			}

			got := GetProviderAPIKey(tt.provider)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("GetProviderAPIKey(%q) = %q, want empty", tt.provider, got)
				}
			} else {
				if got != tt.wantValue {
					t.Errorf("GetProviderAPIKey(%q) = %q, want %q", tt.provider, got, tt.wantValue)
				}
			}
		})
	}
}

func TestGetProviderAPIKey_KiloCode(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("KILO_API_KEY")
	defer os.Setenv("KILO_API_KEY", originalValue)

	// Test with key set
	os.Setenv("KILO_API_KEY", "kilo-test-key-123")
	if got := GetProviderAPIKey(models.ProviderKiloCode); got != "kilo-test-key-123" {
		t.Errorf("GetProviderAPIKey(ProviderKiloCode) = %q, want %q", got, "kilo-test-key-123")
	}

	// Test with key unset
	os.Unsetenv("KILO_API_KEY")
	if got := GetProviderAPIKey(models.ProviderKiloCode); got != "" {
		t.Errorf("GetProviderAPIKey(ProviderKiloCode) = %q, want empty", got)
	}
}

func TestGetProviderAPIKey_Mistral(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("MISTRAL_API_KEY")
	defer os.Setenv("MISTRAL_API_KEY", originalValue)

	// Test with key set
	os.Setenv("MISTRAL_API_KEY", "mistral-test-key-456")
	if got := GetProviderAPIKey(models.ProviderMistral); got != "mistral-test-key-456" {
		t.Errorf("GetProviderAPIKey(ProviderMistral) = %q, want %q", got, "mistral-test-key-456")
	}

	// Test with key unset
	os.Unsetenv("MISTRAL_API_KEY")
	if got := GetProviderAPIKey(models.ProviderMistral); got != "" {
		t.Errorf("GetProviderAPIKey(ProviderMistral) = %q, want empty", got)
	}
}

func TestValidateSessionProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      SessionProviderConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid SQLite configuration",
			config: SessionProviderConfig{
				Type: ProviderSQLite,
			},
			expectError: false,
		},
		{
			name: "Valid MySQL configuration with DSN",
			config: SessionProviderConfig{
				Type: ProviderMySQL,
				MySQL: MySQLConfig{
					DSN: "user:pass@tcp(localhost:3306)/dbname",
				},
			},
			expectError: false,
		},
		{
			name: "Valid MySQL configuration with individual fields",
			config: SessionProviderConfig{
				Type: ProviderMySQL,
				MySQL: MySQLConfig{
					Host:     "localhost",
					Port:     3306,
					Database: "opencode",
					Username: "user",
					Password: "pass",
				},
			},
			expectError: false,
		},
		{
			name: "MySQL without DSN or host",
			config: SessionProviderConfig{
				Type: ProviderMySQL,
				MySQL: MySQLConfig{
					Database: "opencode",
					Username: "user",
					Password: "pass",
				},
			},
			expectError: true,
			errorMsg:    "MySQL host is required",
		},
		{
			name: "MySQL without database",
			config: SessionProviderConfig{
				Type: ProviderMySQL,
				MySQL: MySQLConfig{
					Host:     "localhost",
					Username: "user",
					Password: "pass",
				},
			},
			expectError: true,
			errorMsg:    "MySQL database is required",
		},
		{
			name: "MySQL without username",
			config: SessionProviderConfig{
				Type: ProviderMySQL,
				MySQL: MySQLConfig{
					Host:     "localhost",
					Database: "opencode",
					Password: "pass",
				},
			},
			expectError: true,
			errorMsg:    "MySQL username is required",
		},
		{
			name: "MySQL without password",
			config: SessionProviderConfig{
				Type: ProviderMySQL,
				MySQL: MySQLConfig{
					Host:     "localhost",
					Database: "opencode",
					Username: "user",
				},
			},
			expectError: true,
			errorMsg:    "MySQL password is required",
		},
		{
			name: "Invalid provider type",
			config: SessionProviderConfig{
				Type: "postgres",
			},
			expectError: true,
			errorMsg:    "invalid session provider type",
		},
		{
			name:        "Empty type defaults to SQLite",
			config:      SessionProviderConfig{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up a temporary config
			cfg = &Config{
				SessionProvider: tt.config,
			}

			err := validateSessionProvider()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error message %q does not contain %q", err.Error(), tt.errorMsg)
				}
			}

			// Clean up
			cfg = nil
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMCPType(t *testing.T) {
	tests := []struct {
		name     string
		mcpType  MCPType
		expected string
	}{
		{"Stdio type", MCPStdio, "stdio"},
		{"SSE type", MCPSse, "sse"},
		{"HTTP type", MCPHttp, "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mcpType) != tt.expected {
				t.Errorf("MCPType = %q, want %q", tt.mcpType, tt.expected)
			}
		})
	}
}

func TestMCPServerStruct(t *testing.T) {
	server := MCPServer{
		Command: "test-command",
		Env:     []string{"KEY=value"},
		Args:    []string{"--flag"},
		Type:    MCPStdio,
		URL:     "http://localhost:8080",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}

	if server.Command != "test-command" {
		t.Errorf("Command = %q, want %q", server.Command, "test-command")
	}
	if len(server.Env) != 1 {
		t.Errorf("Env length = %d, want %d", len(server.Env), 1)
	}
	if len(server.Args) != 1 {
		t.Errorf("Args length = %d, want %d", len(server.Args), 1)
	}
	if server.Type != MCPStdio {
		t.Errorf("Type = %q, want %q", server.Type, MCPStdio)
	}
}

func TestAgentMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     AgentMode
		expected string
	}{
		{"Agent mode", AgentModeAgent, "agent"},
		{"Subagent mode", AgentModeSubagent, "subagent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("AgentMode = %q, want %q", tt.mode, tt.expected)
			}
		})
	}
}

func TestAgentConstants(t *testing.T) {
	agentNames := []struct {
		name     string
		actual   AgentName
		expected string
	}{
		{"Coder", AgentCoder, "coder"},
		{"Summarizer", AgentSummarizer, "summarizer"},
		{"Explorer", AgentExplorer, "explorer"},
		{"Descriptor", AgentDescriptor, "descriptor"},
		{"Workhorse", AgentWorkhorse, "workhorse"},
		{"Hivemind", AgentHivemind, "hivemind"},
	}

	for _, tt := range agentNames {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("AgentName = %q, want %q", tt.actual, tt.expected)
			}
		})
	}
}

func TestAgentStruct(t *testing.T) {
	agent := Agent{
		Model:           models.Claude35Sonnet,
		MaxTokens:       4096,
		ReasoningEffort: "medium",
		Permission:      map[string]any{"bash": "allow"},
		Tools:           map[string]bool{"skill": true},
		Mode:            AgentModeAgent,
		Name:            "Test Agent",
		Description:     "Test description",
		Color:           "primary",
		Hidden:          false,
		Disabled:        false,
	}

	if agent.MaxTokens != 4096 {
		t.Errorf("MaxTokens = %d, want %d", agent.MaxTokens, 4096)
	}
	if agent.ReasoningEffort != "medium" {
		t.Errorf("ReasoningEffort = %q, want %q", agent.ReasoningEffort, "medium")
	}
	if agent.Mode != AgentModeAgent {
		t.Errorf("Mode = %q, want %q", agent.Mode, AgentModeAgent)
	}
}

func TestProviderStruct(t *testing.T) {
	provider := Provider{
		APIKey:   "test-key",
		BaseURL:  "https://api.example.com",
		Disabled: false,
	}

	if provider.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want %q", provider.APIKey, "test-key")
	}
	if provider.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL = %q, want %q", provider.BaseURL, "https://api.example.com")
	}
}

func TestAgentOutputStruct(t *testing.T) {
	output := AgentOutput{
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]string{"type": "string"},
			},
		},
	}

	if output.Schema == nil {
		t.Error("Schema should not be nil")
	}
	if output.Schema["type"] != "object" {
		t.Errorf("Schema type = %v, want %q", output.Schema["type"], "object")
	}
}

func TestSessionProviderConfigStruct(t *testing.T) {
	config := SessionProviderConfig{
		Type: ProviderSQLite,
	}

	if config.Type != ProviderSQLite {
		t.Errorf("Type = %q, want %q", config.Type, ProviderSQLite)
	}
}

func TestMySQLConfigStruct(t *testing.T) {
	config := MySQLConfig{
		DSN:      "user:pass@tcp(localhost:3306)/db",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
	}

	if config.Host != "localhost" {
		t.Errorf("Host = %q, want %q", config.Host, "localhost")
	}
	if config.Port != 3306 {
		t.Errorf("Port = %d, want %d", config.Port, 3306)
	}
	if config.Database != "testdb" {
		t.Errorf("Database = %q, want %q", config.Database, "testdb")
	}
}

func TestProviderType(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		expected string
	}{
		{"SQLite provider", ProviderSQLite, "sqlite"},
		{"MySQL provider", ProviderMySQL, "mysql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.provider) != tt.expected {
				t.Errorf("ProviderType = %q, want %q", tt.provider, tt.expected)
			}
		})
	}
}

func TestLSPConfigStruct(t *testing.T) {
	config := LSPConfig{
		Disabled:   false,
		Command:    "gopls",
		Args:       []string{"serve"},
		Extensions: []string{".go"},
	}

	if config.Command != "gopls" {
		t.Errorf("Command = %q, want %q", config.Command, "gopls")
	}
	if len(config.Extensions) != 1 {
		t.Errorf("Extensions length = %d, want %d", len(config.Extensions), 1)
	}
}

func TestTUIConfigStruct(t *testing.T) {
	config := TUIConfig{
		Theme: "dark",
	}

	if config.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", config.Theme, "dark")
	}
}

func TestShellConfigStruct(t *testing.T) {
	config := ShellConfig{
		Path: "/bin/bash",
		Args: []string{"-l"},
	}

	if config.Path != "/bin/bash" {
		t.Errorf("Path = %q, want %q", config.Path, "/bin/bash")
	}
}

func TestDataStruct(t *testing.T) {
	data := Data{
		Directory: "/data/opencode",
	}

	if data.Directory != "/data/opencode" {
		t.Errorf("Directory = %q, want %q", data.Directory, "/data/opencode")
	}
}

func TestSkillsConfigStruct(t *testing.T) {
	config := SkillsConfig{
		Paths: []string{"/custom/skills", "~/.config/skills"},
	}

	if len(config.Paths) != 2 {
		t.Errorf("Paths length = %d, want %d", len(config.Paths), 2)
	}
}

func TestPermissionConfigStruct(t *testing.T) {
	config := PermissionConfig{
		Rules: map[string]any{
			"bash": "ask",
			"edit": map[string]string{"*.go": "allow"},
		},
	}

	if config.Rules == nil {
		t.Error("Rules should not be nil")
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		WorkingDir:   "/project",
		Debug:        true,
		AutoCompact:  true,
		ContextPaths: []string{"/context"},
	}

	if cfg.WorkingDir != "/project" {
		t.Errorf("WorkingDir = %q, want %q", cfg.WorkingDir, "/project")
	}
	if !cfg.Debug {
		t.Error("Debug should be true")
	}
	if !cfg.AutoCompact {
		t.Error("AutoCompact should be true")
	}
}
