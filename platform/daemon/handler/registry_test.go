package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GnarloqGames/genesis-avalon-gateway/platform/daemon/model"
	"github.com/GnarloqGames/genesis-avalon-kit/registry"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/bmizerany/assert"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
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
		Definition: registry.BuildingBlueprintRequest{
			Name: "house",
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
	srv := httptest.NewServer(AddBlueprint())
	defer srv.Close()

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

	client := resty.New()
	patches := gomonkey.ApplyFunc(registry.SaveBuildingBlueprint, func(ctx context.Context, bbp registry.BuildingBlueprintRequest) error {
		if bbp.Version == errVersion {
			return fmt.Errorf("test error")
		}
		return nil
	})
	patches.ApplyFunc(registry.SaveResourceBlueprint, func(ctx context.Context, rbp registry.ResourceBlueprintRequest) error {
		if rbp.Version == errVersion {
			return fmt.Errorf("test error")
		}
		return nil
	})
	defer patches.Reset()

	for _, tt := range tests {
		tf := func(t *testing.T) {
			resp, err := client.NewRequest().
				SetHeader("Content-Type", tt.contentType).
				SetBody(tt.body).
				Post(srv.URL)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode())
		}

		t.Run(tt.label, tf)
	}
}
