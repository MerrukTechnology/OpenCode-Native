package db

import (
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
)

func TestProvider_TypeAndDialect(t *testing.T) {
	tests := []struct {
		name            string
		provider        Provider
		expectedType    config.ProviderType
		expectedDialect string
	}{
		{
			name:            "SQLite provider",
			provider:        NewSQLiteProvider("/tmp/test"),
			expectedType:    config.ProviderSQLite,
			expectedDialect: "sqlite3",
		},
		{
			name:            "MySQL provider",
			provider:        NewMySQLProvider(config.MySQLConfig{}),
			expectedType:    config.ProviderMySQL,
			expectedDialect: "mysql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_Type", func(t *testing.T) {
			if tt.provider.Type() != tt.expectedType {
				t.Errorf("Type() = %v, want %v", tt.provider.Type(), tt.expectedType)
			}
		})
		t.Run(tt.name+"_Dialect", func(t *testing.T) {
			if tt.provider.Dialect() != tt.expectedDialect {
				t.Errorf("Dialect() = %v, want %v", tt.provider.Dialect(), tt.expectedDialect)
			}
		})
	}
}

func TestMySQLProvider_BuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   config.MySQLConfig
		expected string
	}{
		{
			name: "DSN provided directly",
			config: config.MySQLConfig{
				DSN: "user:pass@tcp(localhost:3306)/dbname?parseTime=true",
			},
			expected: "user:pass@tcp(localhost:3306)/dbname?parseTime=true",
		},
		{
			name: "Build from individual fields",
			config: config.MySQLConfig{
				Host:     "localhost",
				Port:     3306,
				Database: "opencode",
				Username: "testuser",
				Password: "testpass",
			},
			expected: "testuser:testpass@tcp(localhost:3306)/opencode?parseTime=true",
		},
		{
			name: "Custom port",
			config: config.MySQLConfig{
				Host:     "db.example.com",
				Port:     3307,
				Database: "mydb",
				Username: "admin",
				Password: "secret",
			},
			expected: "admin:secret@tcp(db.example.com:3307)/mydb?parseTime=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewMySQLProvider(tt.config)
			result := provider.buildDSN()
			if result != tt.expected {
				t.Errorf("buildDSN() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectType  config.ProviderType
		expectError bool
	}{
		{
			name: "SQLite provider",
			config: &config.Config{
				Data: config.Data{Directory: "/tmp/test"},
				SessionProvider: config.SessionProviderConfig{
					Type: config.ProviderSQLite,
				},
			},
			expectType:  config.ProviderSQLite,
			expectError: false,
		},
		{
			name: "MySQL provider",
			config: &config.Config{
				SessionProvider: config.SessionProviderConfig{
					Type: config.ProviderMySQL,
					MySQL: config.MySQLConfig{
						Host:     "localhost",
						Port:     3306,
						Database: "test",
						Username: "user",
						Password: "pass",
					},
				},
			},
			expectType:  config.ProviderMySQL,
			expectError: false,
		},
		{
			name: "Default to SQLite when type is empty",
			config: &config.Config{
				Data: config.Data{Directory: "/tmp/test"},
				SessionProvider: config.SessionProviderConfig{
					Type: "",
				},
			},
			expectType:  config.ProviderSQLite,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && provider.Type() != tt.expectType {
				t.Errorf("Provider type = %v, want %v", provider.Type(), tt.expectType)
			}
		})
	}
}
