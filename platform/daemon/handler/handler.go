package handler

import (
	"net/http"
	"time"

	"github.com/GnarloqGames/genesis-avalon-kit/proto"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Handler(bus *transport.Connection) http.Handler {
	r := chi.NewRouter()

	r.Use(LoggingMiddleware())

	r.Post("/build", Build(bus))

	return r
}

func Build(bus *transport.Connection) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		req := &proto.BuildRequest{
			Header: &proto.RequestHeader{
				TraceID:   "",
				Timestamp: timestamppb.Now(),
			},
			Name:     "house",
			Duration: "10s",
		}

		var res proto.BuildResponse
		if err := bus.Request("build", req, &res, 10*time.Second); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("OK")) //nolint:errcheck
	}

	return http.HandlerFunc(fn)
}
