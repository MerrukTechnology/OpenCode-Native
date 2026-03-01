package config

import (
	"os"
	"strings"
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/models"
)

// Helper to test string fields in structs
func testStringField[T any](t *testing.T, obj T, getField func(T) string, expected, fieldName string) {
	t.Helper()
	if got := getField(obj); got != expected {
		t.Errorf("%s = %q, want %q", fieldName, got, expected)
	}
}

// Helper to test bool fields in structs
func testBoolField[T any](t *testing.T, obj T, getField func(T) bool, expected bool, fieldName string) {
	t.Helper()
	if got := getField(obj); got != expected {
		t.Errorf("%s = %v, want %v", fieldName, got, expected)
	}
}

// Helper to test int fields in structs
func testIntField[T any](t *testing.T, obj T, getField func(T) int, expected int, fieldName string) {
	t.Helper()
	if got := getField(obj); got != expected {
		t.Errorf("%s = %d, want %d", fieldName, got, expected)
	}
}

// Helper to test string slice fields in structs
func testStringSliceField[T any](t *testing.T, obj T, getField func(T) []string, expected []string, fieldName string) {
	t.Helper()
	if got := getField(obj); len(got) != len(expected) {
		t.Errorf("%s length = %d, want %d", fieldName, len(got), len(expected))
	}
}

// Helper to test map fields in structs
func testMapField[T any](t *testing.T, obj T, getField func(T) map[string]any, fieldName string) {
	t.Helper()
	if got := getField(obj); got == nil {
		t.Error(fieldName + " should not be nil")
	}
}

// =============================================================================
// GetProviderAPIKey Tests
// =============================================================================

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
			if tt.envKey != "" {
				originalValue := os.Getenv(tt.envKey)
				defer os.Setenv(tt.envKey, originalValue)

				// Set test value
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

// =============================================================================
// ValidateSessionProvider Tests
// =============================================================================

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
			defer func() { cfg = nil }()

			err := validateSessionProvider()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error message %q does not contain %q", err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// =============================================================================
// Constant Value Tests - Unified table-driven tests
// =============================================================================

func TestTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		// MCPType constants
		{"MCPType Stdio", string(MCPStdio), "stdio"},
		{"MCPType SSE", string(MCPSse), "sse"},
		{"MCPType HTTP", string(MCPHttp), "http"},
		// AgentMode constants
		{"AgentMode Agent", string(AgentModeAgent), "agent"},
		{"AgentMode Subagent", string(AgentModeSubagent), "subagent"},
		// ProviderType constants
		{"ProviderType SQLite", string(ProviderSQLite), "sqlite"},
		{"ProviderType MySQL", string(ProviderMySQL), "mysql"},
		// AgentName constants
		{"AgentName Coder", AgentCoder, "coder"},
		{"AgentName Summarizer", AgentSummarizer, "summarizer"},
		{"AgentName Explorer", AgentExplorer, "explorer"},
		{"AgentName Descriptor", AgentDescriptor, "descriptor"},
		{"AgentName Workhorse", AgentWorkhorse, "workhorse"},
		{"AgentName Hivemind", AgentHivemind, "hivemind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Struct Field Tests - Unified table-driven tests
// =============================================================================

func TestMCPServerStruct(t *testing.T) {
	server := MCPServer{
		Command: "test-command",
		Env:     []string{"KEY=value"},
		Args:    []string{"--flag"},
		Type:    MCPStdio,
		URL:     "http://localhost:8080",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}

	testStringField(t, server, func(s MCPServer) string { return s.Command }, "test-command", "Command")
	testStringSliceField(t, server, func(s MCPServer) []string { return s.Env }, []string{"KEY=value"}, "Env")
	testStringSliceField(t, server, func(s MCPServer) []string { return s.Args }, []string{"--flag"}, "Args")
	testStringField(t, server, func(s MCPServer) string { return string(s.Type) }, "stdio", "Type")
	testStringField(t, server, func(s MCPServer) string { return s.URL }, "http://localhost:8080", "URL")
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

	// MaxTokens is int64, so we use direct comparison
	if agent.MaxTokens != 4096 {
		t.Errorf("MaxTokens = %d, want %d", agent.MaxTokens, 4096)
	}
	testStringField(t, agent, func(a Agent) string { return a.ReasoningEffort }, "medium", "ReasoningEffort")
	testStringField(t, agent, func(a Agent) string { return string(a.Mode) }, "agent", "Mode")
}

func TestProviderStruct(t *testing.T) {
	provider := Provider{
		APIKey:   "test-key",
		BaseURL:  "https://api.example.com",
		Disabled: false,
	}

	testStringField(t, provider, func(p Provider) string { return p.APIKey }, "test-key", "APIKey")
	testStringField(t, provider, func(p Provider) string { return p.BaseURL }, "https://api.example.com", "BaseURL")
	testBoolField(t, provider, func(p Provider) bool { return p.Disabled }, false, "Disabled")
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

	testMapField(t, output, func(o AgentOutput) map[string]any { return o.Schema }, "Schema")

	if output.Schema["type"] != "object" {
		t.Errorf("Schema type = %v, want %q", output.Schema["type"], "object")
	}
}

func TestSessionProviderConfigStruct(t *testing.T) {
	config := SessionProviderConfig{
		Type: ProviderSQLite,
	}

	testStringField(t, config, func(c SessionProviderConfig) string { return string(c.Type) }, "sqlite", "Type")
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

	testStringField(t, config, func(c MySQLConfig) string { return c.Host }, "localhost", "Host")
	testIntField(t, config, func(c MySQLConfig) int { return c.Port }, 3306, "Port")
	testStringField(t, config, func(c MySQLConfig) string { return c.Database }, "testdb", "Database")
	testStringField(t, config, func(c MySQLConfig) string { return c.Username }, "testuser", "Username")
	testStringField(t, config, func(c MySQLConfig) string { return c.Password }, "testpass", "Password")
}

func TestLSPConfigStruct(t *testing.T) {
	config := LSPConfig{
		Disabled:   false,
		Command:    "gopls",
		Args:       []string{"serve"},
		Extensions: []string{".go"},
	}

	testStringField(t, config, func(c LSPConfig) string { return c.Command }, "gopls", "Command")
	testStringSliceField(t, config, func(c LSPConfig) []string { return c.Extensions }, []string{".go"}, "Extensions")
}

func TestTUIConfigStruct(t *testing.T) {
	config := TUIConfig{
		Theme: "dark",
	}

	testStringField(t, config, func(c TUIConfig) string { return c.Theme }, "dark", "Theme")
}

func TestShellConfigStruct(t *testing.T) {
	config := ShellConfig{
		Path: "/bin/bash",
		Args: []string{"-l"},
	}

	testStringField(t, config, func(c ShellConfig) string { return c.Path }, "/bin/bash", "Path")
	testStringSliceField(t, config, func(c ShellConfig) []string { return c.Args }, []string{"-l"}, "Args")
}

func TestDataStruct(t *testing.T) {
	data := Data{
		Directory: "/data/opencode",
	}

	testStringField(t, data, func(d Data) string { return d.Directory }, "/data/opencode", "Directory")
}

func TestSkillsConfigStruct(t *testing.T) {
	config := SkillsConfig{
		Paths: []string{"/custom/skills", "~/.config/skills"},
	}

	testStringSliceField(t, config, func(c SkillsConfig) []string { return c.Paths }, []string{"/custom/skills", "~/.config/skills"}, "Paths")
}

func TestPermissionConfigStruct(t *testing.T) {
	config := PermissionConfig{
		Rules: map[string]any{
			"bash": "ask",
			"edit": map[string]string{"*.go": "allow"},
		},
	}

	testMapField(t, config, func(c PermissionConfig) map[string]any { return c.Rules }, "Rules")
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		WorkingDir:   "/project",
		Debug:        true,
		AutoCompact:  true,
		ContextPaths: []string{"/context"},
	}

	testStringField(t, cfg, func(c Config) string { return c.WorkingDir }, "/project", "WorkingDir")
	testBoolField(t, cfg, func(c Config) bool { return c.Debug }, true, "Debug")
	testBoolField(t, cfg, func(c Config) bool { return c.AutoCompact }, true, "AutoCompact")
	testStringSliceField(t, cfg, func(c Config) []string { return c.ContextPaths }, []string{"/context"}, "ContextPaths")
}
