package main

import (
	"github.com/GnarloqGames/genesis-avalon-gateway/config"
	"github.com/GnarloqGames/genesis-avalon-kit/database/couchbase"
	"github.com/spf13/viper"
)

func setConfigs() {
	couchbaseConfig()
}

func couchbaseConfig() {
	couchbase.SetURL(viper.GetString(config.FlagCouchbaseURL))
	couchbase.SetScope(viper.GetString(config.FlagEnvironment))
	couchbase.SetBucket(viper.GetString(config.FlagCouchbaseBucket))
	couchbase.SetUsername(viper.GetString(config.FlagCouchbaseUsername))
	couchbase.SetPassword(viper.GetString(config.FlagCouchbasePassword))
}
