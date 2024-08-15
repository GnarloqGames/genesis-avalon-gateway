package model

import (
	"encoding/json"

	"github.com/GnarloqGames/genesis-avalon-kit/registry"
	"gopkg.in/yaml.v3"
)

const (
	KindBuilding = "building"
	KindResource = "resource"
)

type BlueprintRequest struct {
	Kind       string           `json:"kind"`
	Version    string           `json:"version"`
	Force      bool             `json:"force"`
	Definition registry.Request `json:"body"`
}

type BlueprintBatchRequest struct {
	Version   string                              `json:"version"`
	Force     bool                                `json:"force"`
	Buildings []registry.BuildingBlueprintRequest `json:"buildings"`
	Resources []registry.ResourceBlueprintRequest `json:"resources"`
}

func (b *BlueprintRequest) UnmarshalJSON(d []byte) error {
	tmp := make(map[string]interface{})
	if err := json.Unmarshal(d, &tmp); err != nil {
		return err
	}

	if kind, ok := getString(tmp, "kind"); ok {
		b.Kind = kind
	}

	if version, ok := getString(tmp, "version"); ok {
		b.Version = version
	}

	rawBody := []byte("{}")

	if body, ok := tmp["body"].(map[string]interface{}); ok {
		if raw, err := json.Marshal(body); err != nil {
			return err
		} else {
			rawBody = raw
		}
	}

	switch b.Kind {
	case KindBuilding:
		var breq registry.BuildingBlueprintRequest
		if err := json.Unmarshal(rawBody, &breq); err != nil {
			return err
		}

		b.Definition = registry.Request(breq)
	case KindResource:
		var rreq registry.ResourceBlueprintRequest
		if err := json.Unmarshal(rawBody, &rreq); err != nil {
			return err
		}

		b.Definition = registry.Request(rreq)
	}

	return nil
}

func (b *BlueprintRequest) UnmarshalYAML(x *yaml.Node) error {
	tmp := make(map[string]interface{})
	if err := x.Decode(&tmp); err != nil {
		return err
	}

	if kind, ok := getString(tmp, "kind"); ok {
		b.Kind = kind
	}

	if version, ok := getString(tmp, "version"); ok {
		b.Version = version
	}

	rawBody := []byte("{}")

	if body, ok := tmp["body"].(map[string]interface{}); ok {
		if raw, err := json.Marshal(body); err != nil {
			return err
		} else {
			rawBody = raw
		}
	}

	switch b.Kind {
	case KindBuilding:
		var breq registry.BuildingBlueprintRequest
		if err := json.Unmarshal(rawBody, &breq); err != nil {
			return err
		}

		b.Definition = registry.Request(breq)
	case KindResource:
		var rreq registry.ResourceBlueprintRequest
		if err := json.Unmarshal(rawBody, &rreq); err != nil {
			return err
		}

		b.Definition = registry.Request(rreq)
	}

	return nil
}
