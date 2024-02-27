package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon/handler/middleware"
	"github.com/GnarloqGames/genesis-avalon-kit/database/couchbase"
	"github.com/GnarloqGames/genesis-avalon-kit/proto"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Handler(bus *transport.Connection) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.Logging())
	r.Use(auth.Middleware())

	r.Post("/build", Build(bus))
	r.Get("/buildings", ListBuildings())

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

func ListBuildings() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(auth.ClaimsContext).(*auth.Claims)
		if !ok || claims == nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		db, err := couchbase.Get()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		buildings, err := db.GetBuildings(claims.Subject)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response := map[string]any{
			"count":     len(buildings),
			"buildings": buildings,
		}

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(response); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	return http.HandlerFunc(fn)
}

// func getClaims(w http.ResponseWriter, r *http.Request) *auth.Claims {
// 	claims, ok := r.Context().Value(auth.ClaimsContext).(*auth.Claims)
// 	if !ok || claims == nil {
// 		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
// 		return nil
// 	}

// 	return claims
// }
