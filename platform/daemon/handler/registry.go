package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/GnarloqGames/genesis-avalon-gateway/config"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon/model"
	"github.com/GnarloqGames/genesis-avalon-kit/registry"
	"github.com/GnarloqGames/genesis-avalon-kit/registry/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type ErrInvalidMediaType struct {
	mediaType string
}

func (e ErrInvalidMediaType) Error() string {
	return fmt.Sprintf("invalid media type: %s", e.mediaType)
}

func NewErrInvalidMediaType(t string) ErrInvalidMediaType {
	return ErrInvalidMediaType{
		mediaType: t,
	}
}

func GetBlueprints() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		version := chi.URLParam(r, "version")

		var store map[string]any

		switch version {
		case "current", viper.GetString(config.FlagBlueprintVersion):
			store = cache.GetLoadedBlueprints(r.Context())
		default:
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		render.JSON(w, r, store)
	}

	return http.HandlerFunc(fn)
}

func GetBlueprint() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		version := chi.URLParam(r, "version")
		kind := chi.URLParam(r, "kind")
		slug := chi.URLParam(r, "slug")

		version = strings.TrimPrefix(version, "v")

		// If we're trying to get currently deployed version, try cache first
		if version == "current" || version == viper.GetString(config.FlagBlueprintVersion) {
			var (
				bp any
				ok bool
			)

			switch kind {
			case "building":
				bp, ok = cache.GetBuildingBlueprint(r.Context(), slug)
			case "resource":
				bp, ok = cache.GetResourceBlueprint(r.Context(), slug)
			}

			if !ok {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			render.JSON(w, r, bp)

			return
		}

		switch kind {
		case "building":
			bp, err := registry.GetBuildingBlueprint(r.Context(), version, slug)
			if err != nil {
				slog.Debug("failed to get building blueprint", "version", version, "slug", slug)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			render.JSON(w, r, bp)
		case "resource":
			bp, err := registry.GetResourceBlueprint(r.Context(), version, slug)
			if err != nil {
				slog.Debug("failed to get resource blueprint", "version", version, "slug", slug)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			render.JSON(w, r, bp)
		}
	}

	return fn
}

func AddBlueprintBatch() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest[*model.BlueprintBatchRequest](r)

		if err != nil {
			slog.Error("failed to decode blueprint request", "error", err)

			if _, ok := err.(ErrInvalidMediaType); ok {
				http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			return
		}

		for _, building := range req.Buildings {
			if err := registry.SaveBuildingBlueprint(r.Context(), req.Version, building, req.Force); err != nil {
				slog.Error("failed to save blueprint",
					"error", err,
					"kind", "building",
					"name", building.Name)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}
		}

		for _, resource := range req.Resources {
			if err := registry.SaveResourceBlueprint(r.Context(), req.Version, resource, req.Force); err != nil {
				slog.Error("failed to save blueprint",
					"error", err,
					"kind", "resource",
					"name", resource.Name)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}
		}

		render.JSON(w, r, map[string]string{
			"status": "OK",
		})
	}

	return http.HandlerFunc(fn)
}

func AddBlueprint() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest[*model.BlueprintRequest](r)
		if err != nil {
			slog.Error("failed to decode blueprint request", "error", err)

			if _, ok := err.(ErrInvalidMediaType); ok {
				http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			return
		}

		var insertErr error

		switch req.Kind {
		case model.KindBuilding:
			def := req.Definition.(registry.BuildingBlueprintRequest)
			insertErr = registry.SaveBuildingBlueprint(r.Context(), req.Version, def, false)
		case model.KindResource:
			def := req.Definition.(registry.ResourceBlueprintRequest)
			insertErr = registry.SaveResourceBlueprint(r.Context(), req.Version, def, false)
		case "":
			slog.Info("error: missing kind field")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		default:
			slog.Debug("error: invalid kind field", "kind", req.Kind)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		if insertErr != nil {
			slog.Error("failed to insert blueprint", "error", insertErr, "kind", req.Kind)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		render.JSON(w, r, map[string]interface{}{"status": "OK"})
	}

	return http.HandlerFunc(fn)
}

type RequestInto interface {
	*model.BlueprintRequest
}

func decodeRequest[T *model.BlueprintRequest | *model.BlueprintBatchRequest](r *http.Request) (T, error) {
	contentType := r.Header.Get("Content-Type")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var rr T

	switch contentType {
	case "application/json":
		err = json.Unmarshal(body, &rr)
	case "application/yaml":
		err = yaml.Unmarshal(body, &rr)
	default:
		return nil, NewErrInvalidMediaType(contentType)
	}

	if err != nil {
		return nil, err
	}

	return rr, nil
}
