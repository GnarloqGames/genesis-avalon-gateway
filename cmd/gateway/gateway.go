package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GnarloqGames/genesis-avalon-gateway/logging"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon"
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
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

		bus, err := initMessageBus(cmd)
		if err != nil {
			return err
		}
		defer bus.Close()

		s := daemon.Start(bus)

		<-stopChan
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

	startCmd.Flags().String("host", "127.0.0.1", "host to bind listener to")
	startCmd.Flags().Uint16("port", uint16(8080), "port to bind listener to")
	startCmd.Flags().String("nats-address", "127.0.0.1:4222", "NATS address")
	startCmd.Flags().String("nats-encoding", "json", "NATS encoding")

	rootCmd.PersistentFlags().String("log-level", "info", "log level (default is info)")
	rootCmd.PersistentFlags().String("log-kind", "text", "log kind (text or json, default is text)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/gatewayd/config.yaml)")
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

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		slog.Info("loaded config file", "file", viper.ConfigFileUsed())
	}

	if err := logging.Init(); err != nil {
		slog.Error("failed to create logger", "error", err.Error())
	}
}

func initMessageBus(cmd *cobra.Command) (*transport.Connection, error) {
	natsAddress, err := cmd.Flags().GetString("nats-address")
	if err != nil {
		natsAddress = defaultNatsAddress
	}

	natsEncoder, err := cmd.Flags().GetString("nats-encoder")
	if err != nil {
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
