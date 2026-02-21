package app

import (
	"testing"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestApp_ActiveAgentName(t *testing.T) {
	tests := []struct {
		name             string
		primaryAgentKeys []config.AgentName
		activeAgentIdx   int
		expected         config.AgentName
	}{
		{
			name:             "empty keys returns coder",
			primaryAgentKeys: []config.AgentName{},
			activeAgentIdx:   0,
			expected:         config.AgentCoder,
		},
		{
			name:             "single agent",
			primaryAgentKeys: []config.AgentName{config.AgentCoder},
			activeAgentIdx:   0,
			expected:         config.AgentCoder,
		},
		{
			name:             "multiple agents returns first",
			primaryAgentKeys: []config.AgentName{config.AgentCoder, config.AgentExplorer},
			activeAgentIdx:   0,
			expected:         config.AgentCoder,
		},
		{
			name:             "multiple agents returns second",
			primaryAgentKeys: []config.AgentName{config.AgentCoder, config.AgentExplorer},
			activeAgentIdx:   1,
			expected:         config.AgentExplorer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				PrimaryAgentKeys: tt.primaryAgentKeys,
				ActiveAgentIdx:   tt.activeAgentIdx,
			}

			result := app.ActiveAgentName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApp_SwitchAgent(t *testing.T) {
	tests := []struct {
		name             string
		primaryAgentKeys []config.AgentName
		initialIdx       int
		expectedIdx      int
	}{
		{
			name:             "single agent stays same",
			primaryAgentKeys: []config.AgentName{config.AgentCoder},
			initialIdx:       0,
			expectedIdx:      0,
		},
		{
			name:             "two agents switches from first to second",
			primaryAgentKeys: []config.AgentName{config.AgentCoder, config.AgentExplorer},
			initialIdx:       0,
			expectedIdx:      1,
		},
		{
			name:             "two agents switches from second to first (wraps)",
			primaryAgentKeys: []config.AgentName{config.AgentCoder, config.AgentExplorer},
			initialIdx:       1,
			expectedIdx:      0,
		},
		{
			name:             "three agents cycles correctly",
			primaryAgentKeys: []config.AgentName{config.AgentCoder, config.AgentExplorer, config.AgentHivemind},
			initialIdx:       2,
			expectedIdx:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				PrimaryAgentKeys: tt.primaryAgentKeys,
				ActiveAgentIdx:   tt.initialIdx,
			}

			result := app.SwitchAgent()
			assert.Equal(t, tt.primaryAgentKeys[tt.expectedIdx], result)
			assert.Equal(t, tt.expectedIdx, app.ActiveAgentIdx)
		})
	}
}
