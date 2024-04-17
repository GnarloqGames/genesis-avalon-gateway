package main

import (
	"github.com/GnarloqGames/genesis-avalon-gateway/config"
	"github.com/GnarloqGames/genesis-avalon-kit/database/cockroach"
	"github.com/GnarloqGames/genesis-avalon-kit/registry/cache"
	"github.com/spf13/viper"
)

func setConfigs() {
	cockroachConfig()
	registryConfig()
}

func cockroachConfig() {
	cockroach.SetHostname(viper.GetString(config.FlagDatabaseHost))
	cockroach.SetUsername(viper.GetString(config.FlagDatabaseUsername))
	cockroach.SetPassword(viper.GetString(config.FlagDatabasePassword))
	cockroach.SetDatabase(viper.GetString(config.FlagDatabaseName))
}

func registryConfig() {
	cache.SetVersion(viper.GetString(config.FlagBlueprintVersion))
}
