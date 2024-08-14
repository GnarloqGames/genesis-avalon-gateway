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
	Definition registry.Request `json:"body"`
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
		breq.Version = b.Version

		b.Definition = registry.Request(breq)
	case KindResource:
		var rreq registry.ResourceBlueprintRequest
		if err := json.Unmarshal(rawBody, &rreq); err != nil {
			return err
		}
		rreq.Version = b.Version

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
		breq.Version = b.Version

		b.Definition = registry.Request(breq)
	case KindResource:
		var rreq registry.ResourceBlueprintRequest
		if err := json.Unmarshal(rawBody, &rreq); err != nil {
			return err
		}
		rreq.Version = b.Version

		b.Definition = registry.Request(rreq)
	}

	return nil
}
