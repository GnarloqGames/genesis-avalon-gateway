package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/auth/claims"
	"github.com/GnarloqGames/genesis-avalon-kit/registry/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func ReloadBlueprints() http.HandlerFunc {
	logger := slog.Default().With("context", "Build")
	fn := func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(auth.ClaimsContext).(*claims.Claims)
		if !ok || claims == nil {
			logger.Error("failed to read claims from context")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		if !claims.HasRole("dev.avalon.cool:cache:reload") {
			logger.Error("user doesn't have correct permissions", "role", "dev.avalon.cool:can-build", "user_id", claims.Subject)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		version := strings.TrimPrefix(chi.URLParam(r, "version"), "v")

		cache.SetVersion(version)

		if err := cache.Load(r.Context()); err != nil {
			logger.Error("failed to reload cache", "error", err, "version", version)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		render.JSON(w, r, map[string]string{"status": "OK"})
	}

	return fn
}
