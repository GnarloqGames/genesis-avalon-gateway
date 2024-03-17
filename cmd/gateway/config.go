package main

import (
	"github.com/GnarloqGames/genesis-avalon-gateway/config"
	"github.com/GnarloqGames/genesis-avalon-kit/database/cockroach"
	"github.com/GnarloqGames/genesis-avalon-kit/database/couchbase"
	"github.com/GnarloqGames/genesis-avalon-kit/registry"
	"github.com/spf13/viper"
)

func setConfigs() {
	couchbaseConfig()
	cockroachConfig()
	registryConfig()
}

func couchbaseConfig() {
	couchbase.SetURL(viper.GetString(config.FlagCouchbaseURL))
	couchbase.SetScope(viper.GetString(config.FlagEnvironment))
	couchbase.SetBucket(viper.GetString(config.FlagCouchbaseBucket))
	couchbase.SetUsername(viper.GetString(config.FlagCouchbaseUsername))
	couchbase.SetPassword(viper.GetString(config.FlagCouchbasePassword))
}

func cockroachConfig() {
	cockroach.SetHostname(viper.GetString(config.FlagDatabaseHost))
	cockroach.SetUsername(viper.GetString(config.FlagDatabaseUsername))
	cockroach.SetPassword(viper.GetString(config.FlagDatabasePassword))
	cockroach.SetDatabase(viper.GetString(config.FlagCockroachDatabase))
}

func registryConfig() {
	registry.SetVersion(viper.GetString(config.FlagBlueprintVersion))
}
