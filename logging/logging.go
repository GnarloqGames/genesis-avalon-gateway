package logging

import (
	"log/slog"

	"github.com/GnarloqGames/genesis-avalon-gateway/config"
	"github.com/GnarloqGames/genesis-avalon-kit/logging"
	"github.com/spf13/viper"
)

func Init() error {
	kind := viper.GetString(config.FlagLogKind)
	level := viper.GetString(config.FlagLogLevel)

	l, err := logging.Logger(
		logging.WithKind(kind),
		logging.WithLevel(level),
	)
	if err != nil {
		return err
	}

	slog.SetDefault(l)

	return nil
}
