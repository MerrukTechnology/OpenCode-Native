// Package config manages application configuration from various sources.
package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/MerrukTechnology/OpenCode-Native/internal/llm/models"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"github.com/spf13/viper"
)

// MCPType defines the type of MCP (Model Control Protocol) server.
type MCPType string

// Supported MCP types
const (
	MCPStdio MCPType = "stdio"
	MCPSse   MCPType = "sse"
	MCPHttp  MCPType = "http"
)

// MCPServer defines the configuration for a Model Control Protocol server.
type MCPServer struct {
	Command string            `json:"command"`
	Env     []string          `json:"env"`
	Args    []string          `json:"args"`
	Type    MCPType           `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// AgentName is a string alias to allow flexibility
type AgentName = string

type AgentMode string

const (
	AgentModeAgent    AgentMode = "agent"
	AgentModeSubagent AgentMode = "subagent"
)

// Agent Constants
const (
	AgentCoder      AgentName = "coder"
	AgentSummarizer AgentName = "summarizer"
	AgentExplorer   AgentName = "explorer"   // Replaces Task
	AgentDescriptor AgentName = "descriptor" // Replaces Title
	AgentWorkhorse  AgentName = "workhorse"
	AgentHivemind   AgentName = "hivemind"

	// Deprecated names (kept for migration logic)
	AgentTask  AgentName = "task"
	AgentTitle AgentName = "title"
)

// providerDefinition defines which model to use for each specific agent
type providerDefinition struct {
	EnvKey          string
	CheckFunc       func() bool
	CoderModel      models.ModelID
	SummarizerModel models.ModelID
	ExplorerModel   models.ModelID
	DescriptorModel models.ModelID
	WorkhorseModel  models.ModelID
	HivemindModel   models.ModelID
	FallbackModel   models.ModelID
}

// Agent defines configuration for different LLM models and their token limits.
type Agent struct {
	Model           models.ModelID  `json:"model"`
	MaxTokens       int64           `json:"maxTokens"`
	ReasoningEffort string          `json:"reasoningEffort"`      // For openai models low,medium,high
	Permission      map[string]any  `json:"permission,omitempty"` // tool name -> "allow" | {"pattern": "action"}
	Tools           map[string]bool `json:"tools,omitempty"`      // e.g., {"skill": false}
	Mode            AgentMode       `json:"mode,omitempty"`       // "agent" or "subagent"
	Name            string          `json:"name,omitempty"`
	Native          bool            `json:"native,omitempty"`
	Description     string          `json:"description,omitempty"`
	Prompt          string          `json:"prompt,omitempty"`
	Color           string          `json:"color,omitempty"`
	Hidden          bool            `json:"hidden,omitempty"`
}

// Provider defines configuration for an LLM provider.
type Provider struct {
	APIKey   string            `json:"apiKey"`
	Disabled bool              `json:"disabled"`
	BaseURL  string            `json:"baseURL"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// Data defines storage configuration.
type Data struct {
	Directory string `json:"directory,omitempty"`
}

// LSPConfig defines configuration for Language Server Protocol integration.
type LSPConfig struct {
	Disabled       bool              `json:"disabled"`
	Command        string            `json:"command"`
	Args           []string          `json:"args"`
	Extensions     []string          `json:"extensions,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	Initialization any               `json:"initialization,omitempty"`
}

// TUIConfig defines the configuration for the Terminal User Interface.
type TUIConfig struct {
	Theme string `json:"theme,omitempty"`
}

// ShellConfig defines the configuration for the shell used by the bash tool.
type ShellConfig struct {
	Path string   `json:"path,omitempty"`
	Args []string `json:"args,omitempty"`
}

// ProviderType defines the type of session storage provider.
type ProviderType string

const (
	ProviderSQLite ProviderType = "sqlite"
	ProviderMySQL  ProviderType = "mysql"
)

// MySQLConfig defines MySQL-specific configuration.
type MySQLConfig struct {
	DSN                string `json:"dsn,omitempty"`
	Host               string `json:"host,omitempty"`
	Port               int    `json:"port,omitempty"`
	Database           string `json:"database,omitempty"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
	MaxConnections     int    `json:"maxConnections,omitempty"`
	MaxIdleConnections int    `json:"maxIdleConnections,omitempty"`
	ConnectionTimeout  int    `json:"connectionTimeout,omitempty"`
}

// SessionProviderConfig defines configuration for session storage.
type SessionProviderConfig struct {
	Type  ProviderType `json:"type,omitempty"`
	MySQL MySQLConfig  `json:"mysql,omitempty"`
}

// SkillsConfig defines configuration for skills.
type SkillsConfig struct {
	Paths []string `json:"paths,omitempty"` // Custom skill paths
}

// PermissionConfig defines permission configuration using Rules.
// Each tool key maps to either a simple string ("allow"/"deny"/"ask")
// or an object with glob pattern keys (e.g., {"*": "ask", "git *": "allow"}).
type PermissionConfig struct {
	// Rules maps tool names to permission logic.
	Rules map[string]any `json:"rules,omitempty"` // tool name -> "allow" | {"pattern": "action"}

	// Depricated: use Rules instead, Needed for backward compatibility.
	Skill map[string]string `json:"skill,omitempty"`
}

// Config is the main configuration structure for the application.
type Config struct {
	Data               Data                              `json:"data"`
	WorkingDir         string                            `json:"wd,omitempty"`
	MCPServers         map[string]MCPServer              `json:"mcpServers,omitempty"`
	Providers          map[models.ModelProvider]Provider `json:"providers,omitempty"`
	LSP                map[string]LSPConfig              `json:"lsp,omitempty"`
	Agents             map[AgentName]Agent               `json:"agents,omitempty"`
	Debug              bool                              `json:"debug,omitempty"`
	DebugLSP           bool                              `json:"debugLSP,omitempty"`
	ContextPaths       []string                          `json:"contextPaths,omitempty"`
	TUI                TUIConfig                         `json:"tui"`
	Shell              ShellConfig                       `json:"shell,omitempty"`
	AutoCompact        bool                              `json:"autoCompact,omitempty"`
	DisableLSPDownload bool                              `json:"disableLSPDownload,omitempty"`
	SessionProvider    SessionProviderConfig             `json:"sessionProvider,omitempty"`

	// Depricated: use Rules instead, Needed for backward compatibility.
	Skills     *SkillsConfig     `json:"skills,omitempty"`
	Permission *PermissionConfig `json:"permission,omitempty"`
}

// Application constants
const (
	defaultDataDirectory = ".opencode"
	defaultLogLevel      = "info"
	appName              = "opencode"

	MaxTokensFallbackDefault = 4096
)

var defaultContextPaths = []string{
	".github/copilot-instructions.md",
	".cursorrules",
	".cursor/rules/",
	"CLAUDE.md",
	"CLAUDE.local.md",
	"opencode.md",
	"opencode.local.md",
	"OpenCode.md",
	"OpenCode.local.md",
	"OPENCODE.md",
	"OPENCODE.local.md",
	"AGENTS.local.md",
	"AGENTS.md",
}

// Configurator interface for testability
type Configurator interface {
	WorkingDirectory() string
}

// Global configuration instance
var (
	cfg *Config
	mu  sync.RWMutex // Thread safety lock
)

// Reset clears the global configuration.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	cfg = nil
}

// Load initializes the configuration.
func Load(workingDir string, debug bool) (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		WorkingDir: workingDir,
		MCPServers: make(map[string]MCPServer),
		Providers:  make(map[models.ModelProvider]Provider),
		LSP:        make(map[string]LSPConfig),
		Agents:     make(map[AgentName]Agent),
	}

	configureViper()
	setDefaults(debug)

	// Read global config
	if err := readConfig(viper.ReadInConfig()); err != nil {
		return cfg, err
	}

	// Load and merge local config
	mergeLocalConfig(workingDir)

	// Map environment variables to viper
	mapEnvVarsToViper()

	// Apply configuration to the struct
	if err := viper.Unmarshal(cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 1. MIGRATION: Handle Upstream rename logic
	migrateOldAgentNames()

	// 2. DEFAULTS: Set provider defaults based on environment and available credentials
	setProviderDefaults()

	applyDefaultValues()

	// Initialize logging
	if err := initLogging(debug); err != nil {
		return cfg, err
	}

	// Validate configuration
	if err := Validate(); err != nil {
		return cfg, fmt.Errorf("config validation failed: %w", err)
	}

	// Override max tokens for descriptor agent
	if agent, ok := cfg.Agents[AgentDescriptor]; ok {
		agent.MaxTokens = 80
		cfg.Agents[AgentDescriptor] = agent
	}

	return cfg, nil
}

// migrateOldAgentNames handles backward compatibility for upstream changes
func migrateOldAgentNames() {
	if cfg.Agents == nil {
		cfg.Agents = make(map[AgentName]Agent)
	}
	// Migrate "task" -> "explorer"
	if agent, ok := cfg.Agents[AgentTask]; ok {
		if _, exists := cfg.Agents[AgentExplorer]; !exists {
			cfg.Agents[AgentExplorer] = agent
		}
		delete(cfg.Agents, AgentTask)
		logging.Warn("agent name 'task' is deprecated, migrated to 'explorer'")
	}
	// Migrate "title" -> "descriptor"
	if agent, ok := cfg.Agents[AgentTitle]; ok {
		if _, exists := cfg.Agents[AgentDescriptor]; !exists {
			cfg.Agents[AgentDescriptor] = agent
		}
		delete(cfg.Agents, AgentTitle)
		logging.Warn("agent name 'title' is deprecated, migrated to 'descriptor'")
	}
}

// initLogging handles the creation of log files and logger initialization
func initLogging(debug bool) error {
	defaultLevel := slog.LevelInfo
	if cfg.Debug || debug {
		defaultLevel = slog.LevelDebug
	}

	// Check for dev debug mode override
	if os.Getenv("OPENCODE_DEV_DEBUG") == "true" {
		loggingFile := filepath.Join(cfg.Data.Directory, "debug.log")
		messagesPath := filepath.Join(cfg.Data.Directory, "messages")

		if err := os.MkdirAll(cfg.Data.Directory, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Create messages directory if it doesn't exist
		if _, err := os.Stat(messagesPath); os.IsNotExist(err) {
			if err := os.MkdirAll(messagesPath, 0o755); err != nil {
				return fmt.Errorf("failed to create messages directory: %w", err)
			}
		}
		logging.MessageDir = messagesPath

		logFile, err := os.OpenFile(loggingFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		slog.SetDefault(slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: defaultLevel,
		})))
	} else {
		// Standard Logger
		slog.SetDefault(slog.New(slog.NewTextHandler(logging.NewWriter(), &slog.HandlerOptions{
			Level: defaultLevel,
		})))
	}
	return nil
}

// configureViper sets up viper's configuration paths and environment variable handling.
func configureViper() {
	viper.SetConfigName(fmt.Sprintf(".%s", appName))
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(fmt.Sprintf("$XDG_CONFIG_HOME/%s", appName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/%s", appName))
	viper.SetEnvPrefix(strings.ToUpper(appName))
	viper.AutomaticEnv()
}

// setDefaults configures default values for configuration options.
func setDefaults(debug bool) {
	viper.SetDefault("data.directory", defaultDataDirectory)
	viper.SetDefault("contextPaths", defaultContextPaths)
	viper.SetDefault("tui.theme", "opencode")
	viper.SetDefault("autoCompact", true)

	// LSP download control
	if v := os.Getenv("OPENCODE_DISABLE_LSP_DOWNLOAD"); v == "true" || v == "1" {
		viper.Set("disableLSPDownload", true)
	}

	// Shell defaults
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/bash"
	}
	viper.SetDefault("shell.path", shellPath)
	viper.SetDefault("shell.args", []string{"-l"})

	// Session provider defaults
	viper.SetDefault("sessionProvider.type", "sqlite")
	viper.SetDefault("sessionProvider.mysql.port", 3306)
	viper.SetDefault("sessionProvider.mysql.maxConnections", 10)
	viper.SetDefault("sessionProvider.mysql.maxIdleConnections", 5)
	viper.SetDefault("sessionProvider.mysql.connectionTimeout", 30)

	// Environment variable overrides for session provider
	if providerType := os.Getenv("OPENCODE_SESSION_PROVIDER_TYPE"); providerType != "" {
		viper.Set("sessionProvider.type", providerType)
	}
	if mysqlDSN := os.Getenv("OPENCODE_MYSQL_DSN"); mysqlDSN != "" {
		viper.Set("sessionProvider.mysql.dsn", mysqlDSN)
	}

	if debug {
		viper.SetDefault("debug", true)
		viper.Set("log.level", "debug")
	} else {
		viper.SetDefault("debug", false)
		viper.SetDefault("log.level", defaultLogLevel)
	}
}

// mapEnvVarsToViper maps standard Env Vars to the internal Viper config structure.
func mapEnvVarsToViper() {
	envMap := map[string]string{
		"ANTHROPIC_API_KEY":  "providers.anthropic.apiKey",
		"OPENAI_API_KEY":     "providers.openai.apiKey",
		"GEMINI_API_KEY":     "providers.gemini.apiKey",
		"GROQ_API_KEY":       "providers.groq.apiKey",
		"OPENROUTER_API_KEY": "providers.openrouter.apiKey",
		"XAI_API_KEY":        "providers.xai.apiKey",
		"DEEPSEEK_API_KEY":   "providers.deepseek.apiKey",
		"QWEN_API_KEY":       "providers.qwen.apiKey",
		//"MOONSHOT_API_KEY":   "providers.moonshot.apiKey",
		"VERTEXAI_PROJECT":  "providers.vertexai.project",
		"VERTEXAI_LOCATION": "providers.vertexai.location",
	}

	for envKey, configKey := range envMap {
		if val := os.Getenv(envKey); val != "" {
			viper.SetDefault(configKey, val)
		}
	}
}

// setProviderDefaults configures LLM provider defaults.
func setProviderDefaults() {
	agentTypes := []AgentName{
		AgentCoder,
		AgentSummarizer,
		AgentExplorer,
		AgentDescriptor,
		AgentWorkhorse,
		AgentHivemind,
	}

	for _, agent := range agentTypes {
		existing, exists := cfg.Agents[agent]
		if !exists || existing.Model == "" {
			setDefaultModelForAgent(agent)
		}
	}
}

// hasAWSCredentials checks if AWS credentials are available in the environment.
func hasAWSCredentials() bool {
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}
	if os.Getenv("AWS_PROFILE") != "" || os.Getenv("AWS_DEFAULT_PROFILE") != "" {
		return true
	}
	if os.Getenv("AWS_REGION") != "" || os.Getenv("AWS_DEFAULT_REGION") != "" {
		return true
	}
	if os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI") != "" ||
		os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI") != "" {
		return true
	}
	return false
}

// hasVertexAICredentials checks if VertexAI credentials are available in the environment.
func hasVertexAICredentials() bool {
	if os.Getenv("VERTEXAI_PROJECT") != "" && os.Getenv("VERTEXAI_LOCATION") != "" {
		return true
	}
	if os.Getenv("GOOGLE_CLOUD_PROJECT") != "" && (os.Getenv("GOOGLE_CLOUD_REGION") != "" || os.Getenv("GOOGLE_CLOUD_LOCATION") != "") {
		return true
	}
	return false
}

// readConfig handles the result of reading a configuration file.
func readConfig(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		return nil
	}
	return fmt.Errorf("failed to read config: %w", err)
}

// mergeLocalConfig loads and merges configuration from the local directory.
func mergeLocalConfig(workingDir string) {
	local := viper.New()
	local.SetConfigName(fmt.Sprintf(".%s", appName))
	local.SetConfigType("json")
	local.AddConfigPath(workingDir)
	if err := local.ReadInConfig(); err == nil {
		viper.MergeConfigMap(local.AllSettings())
	}
}

// applyDefaultValues sets default values for configuration fields that need processing.
func applyDefaultValues() {
	for k, v := range cfg.MCPServers {
		if v.Type == "" {
			v.Type = MCPStdio
			cfg.MCPServers[k] = v
		}
	}
}

// setDefaultModelForAgent sets default models based on available providers
func setDefaultModelForAgent(agent AgentName) bool {
	definitions := []providerDefinition{
		// 1. Google Cloud VertexAI
		{
			CheckFunc:       hasVertexAICredentials,
			CoderModel:      models.VertexAIGemini30Pro,
			SummarizerModel: models.VertexAIGemini30Pro,
			ExplorerModel:   models.VertexAIGemini30Flash,
			DescriptorModel: models.VertexAIGemini30Flash,
			WorkhorseModel:  models.VertexAIGemini30Pro,
			HivemindModel:   models.VertexAIGemini30Pro,
			FallbackModel:   models.VertexAIGemini30Pro,
		},
		// 2. Anthropic
		{
			EnvKey:          "ANTHROPIC_API_KEY",
			CoderModel:      models.Claude45Sonnet1M,
			SummarizerModel: models.Claude45Sonnet1M,
			ExplorerModel:   models.Claude45Sonnet1M,
			DescriptorModel: models.Claude45Sonnet1M,
			WorkhorseModel:  models.Claude45Sonnet1M,
			HivemindModel:   models.Claude45Sonnet1M,
			FallbackModel:   models.Claude45Sonnet1M,
		},
		// 3. OpenAI
		{
			EnvKey:          "OPENAI_API_KEY",
			CoderModel:      models.GPT5,
			SummarizerModel: models.GPT5,
			ExplorerModel:   models.O4Mini,
			DescriptorModel: models.O4Mini,
			WorkhorseModel:  models.GPT5,
			HivemindModel:   models.GPT5,
			FallbackModel:   models.GPT5,
		},
		// 4. Google Gemini (API)
		{
			EnvKey:          "GEMINI_API_KEY",
			CoderModel:      models.Gemini30Pro,
			SummarizerModel: models.Gemini30Pro,
			ExplorerModel:   models.Gemini30Flash,
			DescriptorModel: models.Gemini30Flash,
			WorkhorseModel:  models.Gemini30Pro,
			HivemindModel:   models.Gemini30Pro,
			FallbackModel:   models.Gemini30Pro,
		},
		// 5. DeepSeek
		{
			EnvKey:          "DEEPSEEK_API_KEY",
			CoderModel:      models.DeepSeekReasoner,
			SummarizerModel: models.DeepSeekReasoner,
			ExplorerModel:   models.DeepSeekChat,
			DescriptorModel: models.DeepSeekChat,
			WorkhorseModel:  models.DeepSeekReasoner,
			HivemindModel:   models.DeepSeekReasoner,
			FallbackModel:   models.DeepSeekReasoner,
		},
		// 6. Groq
		{
			EnvKey:          "GROQ_API_KEY",
			CoderModel:      models.QWENQwq,
			SummarizerModel: models.QWENQwq,
			ExplorerModel:   models.QWENQwq,
			DescriptorModel: models.QWENQwq,
			WorkhorseModel:  models.QWENQwq,
			HivemindModel:   models.QWENQwq,
			FallbackModel:   models.QWENQwq,
		},
		// 7. OpenRouter
		{
			EnvKey:          "OPENROUTER_API_KEY",
			CoderModel:      models.OpenRouterClaude37Sonnet,
			SummarizerModel: models.OpenRouterClaude37Sonnet,
			ExplorerModel:   models.OpenRouterClaude37Sonnet,
			DescriptorModel: models.OpenRouterClaude35Haiku,
			WorkhorseModel:  models.OpenRouterClaude37Sonnet,
			HivemindModel:   models.OpenRouterClaude37Sonnet,
			FallbackModel:   models.OpenRouterClaude37Sonnet,
		},
		// 8. xAI
		{
			EnvKey:          "XAI_API_KEY",
			CoderModel:      models.XAIGrokCodeFast1,
			SummarizerModel: models.XAIGrok41FastReasoning,
			ExplorerModel:   models.XAIGrok41FastReasoning,
			DescriptorModel: models.XAIGrok41FastNonReasoning,
			WorkhorseModel:  models.XAIGrokCodeFast1,
			HivemindModel:   models.XAIGrok41FastReasoning,
			FallbackModel:   models.XAIGrokCodeFast1,
		},
		// 9. AWS Bedrock
		{
			CheckFunc:       hasAWSCredentials,
			CoderModel:      models.BedrockClaude45Sonnet,
			SummarizerModel: models.BedrockClaude45Sonnet,
			ExplorerModel:   models.BedrockClaude45Sonnet,
			DescriptorModel: models.BedrockClaude45Sonnet,
			WorkhorseModel:  models.BedrockClaude45Sonnet,
			HivemindModel:   models.BedrockClaude45Sonnet,
			FallbackModel:   models.BedrockClaude45Sonnet,
		},
	}

	for _, def := range definitions {
		available := false
		if def.CheckFunc != nil {
			available = def.CheckFunc()
		} else if def.EnvKey != "" {
			available = os.Getenv(def.EnvKey) != ""
		}

		if available {
			configureAgent(agent, def)
			return true
		}
	}

	return false
}

// configureAgent applies the specific configuration for a single agent type
func configureAgent(agent AgentName, def providerDefinition) {
	var selectedModel models.ModelID
	var maxTokens int64

	switch agent {
	case AgentCoder:
		selectedModel = def.CoderModel
	case AgentSummarizer:
		selectedModel = def.SummarizerModel
	case AgentExplorer:
		selectedModel = def.ExplorerModel
	case AgentDescriptor:
		selectedModel = def.DescriptorModel
		maxTokens = 80 // Hard limit
	case AgentWorkhorse:
		selectedModel = def.WorkhorseModel
	case AgentHivemind:
		selectedModel = def.HivemindModel
	}

	if selectedModel == "" {
		selectedModel = def.FallbackModel
	}

	if maxTokens == 0 {
		if info, ok := models.SupportedModels[selectedModel]; ok && info.DefaultMaxTokens > 0 {
			maxTokens = info.DefaultMaxTokens
		} else {
			maxTokens = MaxTokensFallbackDefault
		}
	}

	reasoningEffort := ""
	if info, ok := models.SupportedModels[selectedModel]; ok && info.CanReason {
		reasoningEffort = "medium"
	}

	cfg.Agents[agent] = Agent{
		Model:           selectedModel,
		MaxTokens:       maxTokens,
		ReasoningEffort: reasoningEffort,
	}
}

// validateAgent validates model IDs and providers.
func validateAgent(cfg *Config, name AgentName, agent Agent) error {
	// Check if model exists
	model, modelExists := models.SupportedModels[agent.Model]
	if !modelExists {
		logging.Warn("unsupported model configured, reverting to default", "agent", name, "model", agent.Model)
		if setDefaultModelForAgent(name) {
			return nil
		}
		return fmt.Errorf("no valid provider available for agent %s", name)
	}

	// Check if provider for the model is configured
	provider := model.Provider
	providerCfg, providerExists := cfg.Providers[provider]

	if !providerExists {
		apiKey := getProviderAPIKey(provider)
		if apiKey == "" {
			logging.Warn("provider not configured for model, reverting default", "agent", name)
			if setDefaultModelForAgent(name) {
				return nil
			}
			return fmt.Errorf("no valid provider available for agent %s", name)
		}
		// Add provider from env
		if cfg.Providers == nil {
			cfg.Providers = make(map[models.ModelProvider]Provider)
		}
		cfg.Providers[provider] = Provider{APIKey: apiKey}
	} else if providerCfg.Disabled || providerCfg.APIKey == "" {
		logging.Warn("provider disabled/empty, reverting default", "agent", name)
		if setDefaultModelForAgent(name) {
			return nil
		}
		return fmt.Errorf("no valid provider available for agent %s", name)
	}

	// Update max tokens if invalid
	if agent.MaxTokens <= 0 {
		updatedAgent := cfg.Agents[name]
		if model.DefaultMaxTokens > 0 {
			updatedAgent.MaxTokens = model.DefaultMaxTokens
		} else {
			updatedAgent.MaxTokens = MaxTokensFallbackDefault
		}
		cfg.Agents[name] = updatedAgent
	}

	// Reasoning effort checks
	if model.CanReason && (provider == models.ProviderOpenAI || provider == models.ProviderLocal) {
		if agent.ReasoningEffort == "" {
			updatedAgent := cfg.Agents[name]
			updatedAgent.ReasoningEffort = "medium"
			cfg.Agents[name] = updatedAgent
		}
	}

	return nil
}

// Validate checks if the configuration is valid.
func Validate() error {
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	if err := validateSessionProvider(); err != nil {
		return fmt.Errorf("session provider validation failed: %w", err)
	}

	for name, agent := range cfg.Agents {
		if err := validateAgent(cfg, name, agent); err != nil {
			return err
		}
	}

	// Validate providers
	for provider, providerCfg := range cfg.Providers {
		if providerCfg.APIKey == "" && !providerCfg.Disabled {
			fmt.Printf("provider has no API key, marking as disabled %s", provider)
			logging.Warn("provider has no API key, marking as disabled", "provider", provider)
			providerCfg.Disabled = true
			cfg.Providers[provider] = providerCfg
		}
	}

	// Validate LSP configurations
	for language, lspConfig := range cfg.LSP {
		if lspConfig.Command == "" && !lspConfig.Disabled && len(lspConfig.Extensions) == 0 {
			logging.Warn("LSP configuration has no command, marking as disabled", "language", language)
			lspConfig.Disabled = true
			cfg.LSP[language] = lspConfig
		}
	}

	return nil
}

// validateSessionProvider validates the session provider configuration.
func validateSessionProvider() error {
	providerType := cfg.SessionProvider.Type
	if providerType == "" {
		providerType = ProviderSQLite
	}

	// Validate provider type
	if providerType != ProviderSQLite && providerType != ProviderMySQL {
		return fmt.Errorf("invalid session provider type: %s (must be 'sqlite' or 'mysql')", providerType)
	}

	// Validate MySQL configuration if MySQL is selected
	if providerType == ProviderMySQL {
		mysql := cfg.SessionProvider.MySQL

		// If DSN is provided, it takes precedence over individual fields
		if mysql.DSN == "" {
			// Validate individual connection fields
			if mysql.Host == "" {
				return fmt.Errorf("MySQL host is required when using MySQL session provider (or provide DSN)")
			}
			if mysql.Database == "" {
				return fmt.Errorf("MySQL database is required when using MySQL session provider (or provide DSN)")
			}
			if mysql.Username == "" {
				return fmt.Errorf("MySQL username is required when using MySQL session provider (or provide DSN)")
			}
			if mysql.Password == "" {
				return fmt.Errorf("MySQL password is required when using MySQL session provider (or provide DSN)")
			}
		}
	}

	return nil
}

// getProviderAPIKey gets the API key from environment.
func getProviderAPIKey(provider models.ModelProvider) string {
	switch provider {
	case models.ProviderAnthropic:
		return os.Getenv("ANTHROPIC_API_KEY")
	case models.ProviderOpenAI:
		return os.Getenv("OPENAI_API_KEY")
	case models.ProviderGemini:
		return os.Getenv("GEMINI_API_KEY")
	case models.ProviderGrok:
		return os.Getenv("GROQ_API_KEY")
	case models.ProviderOpenRouter:
		return os.Getenv("OPENROUTER_API_KEY")
	case models.ProviderXAI:
		return os.Getenv("XAI_API_KEY")
	case models.ProviderDeepSeek:
		return os.Getenv("DEEPSEEK_API_KEY")
	case models.ProviderBedrock:
		if hasAWSCredentials() {
			return "aws-credentials-available"
		}
	case models.ProviderVertexAI:
		if hasVertexAICredentials() {
			return "vertex-ai-credentials-available"
		}
	}
	return ""
}

// updateCfgFile updates the configuration file.
func updateCfgFile(updateCfg func(config *Config)) error {
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	configFile := viper.ConfigFileUsed()
	var configData []byte
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(homeDir, fmt.Sprintf(".%s.json", appName))
		configData = []byte(`{}`)
	} else {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		configData = data
	}

	var userCfg *Config
	if err := json.Unmarshal(configData, &userCfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	updateCfg(userCfg)

	updatedData, err := json.MarshalIndent(userCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configFile, updatedData, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// Get returns the current configuration.
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return cfg
}

// WorkingDirectory returns the current working directory.
func WorkingDirectory() string {
	mu.Lock()
	defer mu.Unlock()
	if cfg == nil {
		panic("config not loaded")
	}
	return cfg.WorkingDir
}

func (c *Config) WorkingDirectory() string {
	return WorkingDirectory()
}

// UpdateAgentModel updates an agent's model in the config.
func UpdateAgentModel(agentName AgentName, modelID models.ModelID) error {
	mu.Lock()
	defer mu.Unlock()

	if cfg == nil {
		panic("config not loaded")
	}
	existingAgentCfg := cfg.Agents[agentName]
	model, ok := models.SupportedModels[modelID]
	if !ok {
		return fmt.Errorf("model %s not supported", modelID)
	}
	maxTokens := existingAgentCfg.MaxTokens
	if model.DefaultMaxTokens > 0 {
		maxTokens = model.DefaultMaxTokens
	}
	newAgentCfg := Agent{
		Model:           modelID,
		MaxTokens:       maxTokens,
		ReasoningEffort: existingAgentCfg.ReasoningEffort,
		// Preserve upstream fields
		Mode:        existingAgentCfg.Mode,
		Name:        existingAgentCfg.Name,
		Description: existingAgentCfg.Description,
		Prompt:      existingAgentCfg.Prompt,
		Color:       existingAgentCfg.Color,
		Hidden:      existingAgentCfg.Hidden,
		// Preserve rules/permissions
		Permission: existingAgentCfg.Permission,
		Tools:      existingAgentCfg.Tools,
	}
	cfg.Agents[agentName] = newAgentCfg

	if err := validateAgent(cfg, agentName, newAgentCfg); err != nil {
		cfg.Agents[agentName] = existingAgentCfg
		return fmt.Errorf("failed to update agent model: %w", err)
	}

	return updateCfgFile(func(config *Config) {
		if config.Agents == nil {
			config.Agents = make(map[AgentName]Agent)
		}
		config.Agents[agentName] = newAgentCfg
	})
}

// UpdateTheme updates the theme.
func UpdateTheme(themeName string) error {
	mu.Lock()
	defer mu.Unlock()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	cfg.TUI.Theme = themeName
	return updateCfgFile(func(config *Config) {
		config.TUI.Theme = themeName
	})
}
