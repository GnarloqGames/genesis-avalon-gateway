package daemon

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/GnarloqGames/genesis-avalon-gateway/config"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth/provider"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon/handler"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/spf13/viper"
)

type Server struct {
	*http.Server

	bus *transport.Connection
}

func Start(bus *transport.Connection, verifier provider.TokenVerifier) *Server {
	host := viper.GetString(config.FlagGatewayHost)
	port := viper.GetUint16(config.FlagGatewayPort)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler.Handler(bus, verifier),
	}

	go func() {
		slog.Info("starting daemon", "address", host, "port", port)

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("error: %v", err)
		}
	}()

	return &Server{
		Server: server,
		bus:    bus,
	}
}
