package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GnarloqGames/genesis-avalon-gateway/config"
	"github.com/GnarloqGames/genesis-avalon-gateway/logging"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon"
	"github.com/GnarloqGames/genesis-avalon-kit/database/couchbase"
	"github.com/GnarloqGames/genesis-avalon-kit/observability"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

const (
	defaultNatsAddress string = "127.0.0.1:4222"
	defaultNatsEncoder string = "protobuf"
)

var rootCmd = &cobra.Command{
	Use:   "gatewayd",
	Short: "Gateway daemon for routing requests between clients and the game servers",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the gateway daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		otelShutdown, err := observability.Setup(cmdContext)
		if err != nil {
			return fmt.Errorf("failed to set up observability providers: %w", err)
		}
		defer func() {
			if err := otelShutdown(cmdContext); err != nil {
				slog.Error("failed to shut down all otel providers", "error", err)
			}
		}()

		if err := auth.InitProvider(); err != nil {
			return fmt.Errorf("oidc: %w", err)
		}

		bus, err := initMessageBus()
		if err != nil {
			return fmt.Errorf("message bus: %w", err)
		}
		defer bus.Close()

		// Try connecting to Couchbase to catch issues at runtime
		if _, err := couchbase.Get(); err != nil {
			return fmt.Errorf("couchbase: %w", err)
		}

		s := daemon.Start(bus)

		<-cmdContext.Done()
		slog.Info("shutting down daemon")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			slog.Error("daemon shutdown failed", "error", err.Error())
		}

		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("host", "127.0.0.1", "host to bind listener to")
	rootCmd.PersistentFlags().Uint16("port", uint16(8080), "port to bind listener to")
	rootCmd.PersistentFlags().String(config.FlagNatsAddress, "127.0.0.1:4222", "NATS address")
	rootCmd.PersistentFlags().String(config.FlagNatsEncoding, "json", "NATS encoding")
	rootCmd.PersistentFlags().String(config.FlagEnvironment, "development", "environment")
	rootCmd.PersistentFlags().String(config.FlagLogLevel, "info", "log level (default is info)")
	rootCmd.PersistentFlags().String(config.FlagLogKind, "text", "log kind (text or json, default is text)")
	rootCmd.PersistentFlags().String(config.FlagOidcProvider, "", "OIDC provider URL")
	rootCmd.PersistentFlags().String(config.FlagOidcClientId, "", "OIDC client ID")
	rootCmd.PersistentFlags().String(config.FlagCouchbaseURL, "127.0.0.1", "Couchbase host")
	rootCmd.PersistentFlags().String(config.FlagCouchbaseBucket, "default", "Couchbase bucket")
	rootCmd.PersistentFlags().String(config.FlagCouchbaseUsername, "", "Couchbase username")
	rootCmd.PersistentFlags().String(config.FlagCouchbasePassword, "", "Couchbase password")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/gatewayd/config.yaml)")

	envPrefix := "AVALOND"
	bindFlags := map[string]string{
		config.FlagEnvironment:       config.EnvEnvironment,
		config.FlagLogLevel:          config.EnvLogLevel,
		config.FlagLogKind:           config.EnvLogKind,
		config.FlagGatewayHost:       config.EnvGatewayHost,
		config.FlagGatewayPort:       config.EnvGatewayPort,
		config.FlagNatsAddress:       config.EnvNatsAddress,
		config.FlagNatsEncoding:      config.EnvNatsEncoding,
		config.FlagOidcProvider:      config.EnvOidcProvider,
		config.FlagOidcClientId:      config.EnvOidcClientId,
		config.FlagCouchbaseURL:      config.EnvCouchbaseURL,
		config.FlagCouchbaseBucket:   config.EnvCouchbaseBucket,
		config.FlagCouchbaseUsername: config.EnvCouchbaseUsername,
		config.FlagCouchbasePassword: config.EnvCouchbasePassword,
	}

	for flag, env := range bindFlags {
		if err := viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag)); err != nil {
			slog.Warn("failed to bind flag", "error", err, "name", flag)
		}

		env = fmt.Sprintf("%s_%s", envPrefix, env)
		if err := viper.BindEnv(flag, env); err != nil {
			slog.Warn("failed to bind env", "error", err, "flag", flag, "env", env)
		}
	}

	viper.SetDefault("log-level", "info")
	viper.SetDefault("log-kind", "text")
	viper.SetDefault("author", "Alfred Dobradi <alfreddobradi@gmail.com>")
	viper.SetDefault("license", "MIT")

	rootCmd.AddCommand(startCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./")
		viper.AddConfigPath("/etc/gatewayd")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err == nil {
		slog.Info("loaded config file", "file", viper.ConfigFileUsed())
	}

	if err := logging.Init(); err != nil {
		slog.Error("failed to create logger", "error", err.Error())
	}

	setConfigs()
}

func initMessageBus() (*transport.Connection, error) {
	natsAddress := viper.GetString(config.FlagNatsAddress)
	if natsAddress == "" {
		natsAddress = defaultNatsAddress
	}

	natsEncoder := viper.GetString(config.FlagNatsEncoding)
	if natsEncoder == "" {
		natsEncoder = defaultNatsEncoder
	}

	encoder := transport.ParseEncoder(natsEncoder)
	config := transport.DefaultConfig
	config.URL = natsAddress
	config.Encoder = encoder

	bus, err := transport.NewEncodedConn(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	slog.Info("established connection to NATS", "address", natsAddress, "encoding", natsEncoder)

	return bus, nil
}
