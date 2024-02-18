package daemon

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon/handler"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
)

type Server struct {
	*http.Server

	bus *transport.Connection
}

func Start(bus *transport.Connection) *Server {
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", address, port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler.Handler(bus),
	}

	go func() {
		slog.Info("starting daemon", "address", address, "port", port)

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("error: %v", err)
		}
	}()

	return &Server{
		Server: server,
		bus:    bus,
	}
}
