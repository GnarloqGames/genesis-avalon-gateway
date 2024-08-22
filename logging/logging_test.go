package logging_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/GnarloqGames/genesis-avalon-gateway/config"
	"github.com/GnarloqGames/genesis-avalon-gateway/logging"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	tests := []struct {
		kind        string
		level       string
		expectError bool
	}{
		{
			kind:        "json",
			level:       "debug",
			expectError: false,
		},
		{
			kind:        "bogus",
			level:       "debug",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			viper.Set(config.FlagLogKind, tt.kind)
			viper.Set(config.FlagLogLevel, tt.level)
			initErr := logging.Init()
			if tt.expectError {
				require.Error(t, initErr)
				return
			}

			require.NoError(t, initErr)

			defaultLogger := slog.Default()

			require.IsType(t, &slog.JSONHandler{}, defaultLogger.Handler())
		}
		t.Run(fmt.Sprintf("%s-%s", tt.level, tt.kind), tf)
	}
}
