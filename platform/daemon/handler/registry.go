package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/GnarloqGames/genesis-avalon-kit/registry"
	"github.com/go-chi/render"
)

func Blueprints() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		store := registry.GetLoadedBlueprints(r.Context())
		render.JSON(w, r, store)
	}

	return http.HandlerFunc(fn)
}

func AddBuildingBlueprint() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		var blueprint registry.BuildingBlueprintRequest

		if err := decoder.Decode(&blueprint); err != nil {
			slog.Error("failed to decode building blueprint", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		if err := registry.SaveBuildingBlueprint(r.Context(), blueprint); err != nil {
			slog.Error("failed to save building blueprint", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		render.JSON(w, r, struct{ Result string }{Result: "OK"})
	}

	return http.HandlerFunc(fn)
}

func AddResourceBlueprint() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		var blueprint registry.ResourceBlueprintRequest

		if err := decoder.Decode(&blueprint); err != nil {
			slog.Error("failed to decode resource blueprint", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		if err := registry.SaveResourceBlueprint(r.Context(), blueprint); err != nil {
			slog.Error("failed to save resource blueprint", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		render.JSON(w, r, struct{ Result string }{Result: "OK"})
	}

	return http.HandlerFunc(fn)
}
