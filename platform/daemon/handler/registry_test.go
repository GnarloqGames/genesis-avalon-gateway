package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/GnarloqGames/genesis-avalon-gateway/mocks/mockhttp"
	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon/model"
	"github.com/GnarloqGames/genesis-avalon-kit/registry"
	"github.com/bmizerany/assert"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	errVersion = "e"
)

var (
	buildingBody = model.BlueprintRequest{
		Kind:  "building",
		Force: true,
		Definition: registry.BuildingBlueprintRequest{
			Name: "house",
		},
	}

	buildingErrorBody = model.BlueprintRequest{
		Kind:  "building",
		Force: true,
		Definition: registry.BuildingBlueprintRequest{
			Name:    "house",
			Version: errVersion,
		},
	}

	resourceBody = model.BlueprintRequest{
		Kind:  "resource",
		Force: true,
		Definition: registry.ResourceBlueprintRequest{
			Name: "wood",
		},
	}

	invalidKind = model.BlueprintRequest{
		Kind:  "bogus",
		Force: true,
		Definition: registry.BuildingBlueprintRequest{
			Name: "house",
		},
	}

	missingKind = model.BlueprintRequest{
		Force: true,
		Definition: registry.BuildingBlueprintRequest{
			Name: "house",
		},
	}

	yamlBody = []byte(`---
kind: building
body:
    name: house
`)
)

func TestAddBlueprint(t *testing.T) {
	tests := []struct {
		label          string
		contentType    string
		expectedStatus int
		body           any
	}{
		{
			label:          "json/valid-building",
			contentType:    "application/json",
			expectedStatus: 200,
			body:           buildingBody,
		},
		{
			label:          "json/valid-resource",
			contentType:    "application/json",
			expectedStatus: 200,
			body:           resourceBody,
		},
		{
			label:          "json/invalid-kind",
			contentType:    "application/json",
			expectedStatus: 400,
			body:           invalidKind,
		},
		{
			label:          "json/save-error",
			contentType:    "application/json",
			expectedStatus: 500,
			body:           buildingErrorBody,
		},
		{
			label:          "json/missing-kind",
			contentType:    "application/json",
			expectedStatus: 400,
			body:           missingKind,
		},
		{
			label:          "yaml/valid",
			contentType:    "application/yaml",
			expectedStatus: 200,
			body:           yamlBody,
		},
		{
			label:          "yaml/invalid-media-type",
			contentType:    "bogus",
			expectedStatus: http.StatusUnsupportedMediaType,
			body:           yamlBody,
		},
	}

	expectation := mockhttp.Expect("/registry/blueprint",
		mockhttp.WithExpectedVisits(7),
		mockhttp.WithMethod(http.MethodPost),
		mockhttp.WithHandler(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("failed to read body", "error", err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			var (
				request     model.BlueprintRequest
				contentType = r.Header.Get("Content-Type")
			)

			switch contentType {
			case "application/json":
				err = json.Unmarshal(body, &request)
			case "application/yaml":
				err = yaml.Unmarshal(body, &request)
			default:
				err = NewErrInvalidMediaType(contentType)
			}

			if err != nil {
				if _, ok := err.(ErrInvalidMediaType); ok {
					http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
				} else {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				return
			}

			if request.Kind != "building" && request.Kind != "resource" {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if request.Definition.GetVersion() == errVersion {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	mockServer, err := mockhttp.New(expectation)
	require.NoError(t, err)
	defer mockServer.Close()

	client := resty.New()

	url := fmt.Sprintf("%s/registry/blueprint", mockServer.URL())

	for _, tt := range tests {
		tf := func(t *testing.T) {
			resp, err := client.R().
				SetHeader("Content-Type", tt.contentType).
				SetBody(tt.body).
				Post(url)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode())
		}

		t.Run(tt.label, tf)
	}

	fmt.Println(expectation.ActualVisits)
}
