package handler

import (
	"net/http"
	"time"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth"
	"github.com/GnarloqGames/genesis-avalon-kit/proto"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Handler(bus *transport.Connection) http.Handler {
	r := chi.NewRouter()

	r.Use(LoggingMiddleware())
	r.Use(auth.Middleware())

	r.Post("/build", Build(bus))

	return r
}

func Build(bus *transport.Connection) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(auth.ClaimsContext).(*auth.Claims)
		if !ok || claims == nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if !claims.HasRole("dev.avalon.cool:can-build") {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		context, err := structpb.NewStruct(map[string]any{
			"owner": claims.Subject,
		})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		req := &proto.BuildRequest{
			Header: &proto.RequestHeader{
				TraceID:   "",
				Timestamp: timestamppb.Now(),
			},
			Name:     "house",
			Duration: "10s",
			Context:  context,
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
