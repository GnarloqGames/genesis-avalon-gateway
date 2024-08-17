package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth/claims"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth/provider"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon/handler/middleware"
	"github.com/GnarloqGames/genesis-avalon-kit/database"
	"github.com/GnarloqGames/genesis-avalon-kit/proto"
	"github.com/GnarloqGames/genesis-avalon-kit/transport"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Handler(bus *transport.Connection, verifier provider.TokenVerifier) http.Handler {
	meters, err := newMeters()
	if err != nil {
		slog.Error("failed to create meters", "error", err)
	}

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
	r.Use(middleware.Metrics(meters, []string{"/favicon.ico", "/metrics"}))
	r.Use(middleware.Tracing([]string{"/favicon.ico", "/metrics"}))
	r.Use(middleware.Logging([]string{"/favicon.ico", "/metrics"}))

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(rr chi.Router) {
		rr.Use(auth.Middleware(verifier))
		rr.Post("/build", Build(bus))
		rr.Get("/buildings", ListBuildings())
	})

	r.Group(func(rr chi.Router) {
		rr.Use(auth.Middleware(verifier))

		rr.Post("/registry/reload/{version}", ReloadBlueprints())
	})

	r.Post("/registry/blueprint", AddBlueprint())
	r.Post("/registry/blueprints", AddBlueprintBatch())
	r.Get("/registry/blueprint/{version}/{kind}/{slug}", GetBlueprint())
	r.Get("/registry/blueprint/{version}", GetBlueprints())

	return r
}

func Build(bus *transport.Connection) http.HandlerFunc {
	logger := slog.Default().With("context", "Build")
	fn := func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(auth.ClaimsContext).(*claims.Claims)
		if !ok || claims == nil {
			logger.Error("failed to read claims from context")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		if !claims.HasRole("dev.avalon.cool:inventory:write") {
			logger.Error("user doesn't have correct permissions", "role", "dev.avalon.cool:inventory:write", "user_id", claims.Subject)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		src := map[string]any{"owner": claims.Subject}
		context, err := structpb.NewStruct(src)

		if err != nil {
			logger.Error("failed to create new protobuf struct", "error", err, "source", src)
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

		_, err = transport.Request(bus, "build", req, &res, 10*time.Second)
		if err != nil {
			logger.Error("build request failed", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if res.Header.Status == proto.Status_ERROR {
			logger.Error("received error in response", "error", res.Header.Error)
			http.Error(w, res.Header.Error, http.StatusInternalServerError)

			return
		}

		render.JSON(w, r, map[string]string{"status": "OK"})
	}

	return http.HandlerFunc(fn)
}

func ListBuildings() http.HandlerFunc {
	logger := slog.Default().With("context", "ListBuildings")
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		ctx, span := otel.Tracer("test").Start(ctx, "ListBuildings")
		defer span.End()

		claims, ok := r.Context().Value(auth.ClaimsContext).(*claims.Claims)
		if !ok || claims == nil {
			logger.Error("failed to read claims from context")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		db, err := database.Get()
		if err != nil {
			logger.Error("failed to get database connection", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		id, err := uuid.Parse(claims.Subject)
		if err != nil {
			logger.Error("failed to parse owner ID", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		buildings, err := db.GetBuildings(ctx, id)
		if err != nil {
			logger.Error("failed to fetch buildings", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		response := map[string]any{
			"count":     len(buildings),
			"buildings": buildings,
		}

		render.JSON(w, r, response)
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

func newMeters() (middleware.Meters, error) {
	meter := otel.Meter("gateway")

	reqCounter, err := meter.Int64UpDownCounter(
		"http.request.num",
		metric.WithDescription("Number of requests per path"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return middleware.Meters{}, fmt.Errorf("failed to create request counter: %w", err)
	}

	reqLength, err := meter.Float64Histogram(
		"http.request.length",
		metric.WithDescription("Request length"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return middleware.Meters{}, fmt.Errorf("failed to create request length histogram: %w", err)
	}

	return middleware.Meters{
		RequestCounter: reqCounter,
		RequestLength:  reqLength,
	}, nil
}
