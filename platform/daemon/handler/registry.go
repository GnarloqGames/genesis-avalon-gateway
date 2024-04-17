package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon/model"
	"github.com/GnarloqGames/genesis-avalon-kit/registry"
	"github.com/GnarloqGames/genesis-avalon-kit/registry/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
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
		case "current":
			store = cache.GetLoadedBlueprints(r.Context())
		default:
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		render.JSON(w, r, store)
	}

	return http.HandlerFunc(fn)
}

func AddBlueprint() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest(r)
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
			insertErr = registry.SaveBuildingBlueprint(r.Context(), req.Definition.(registry.BuildingBlueprintRequest), req.Force)
		case model.KindResource:
			insertErr = registry.SaveResourceBlueprint(r.Context(), req.Definition.(registry.ResourceBlueprintRequest), req.Force)
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

type bodyDecoder interface {
	Decode(v any) error
}

func decodeRequest(r *http.Request) (*model.BlueprintRequest, error) {
	contentType := r.Header.Get("Content-Type")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var rr model.BlueprintRequest

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

	return &rr, nil
}
