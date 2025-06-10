package orationis

import (
	"encoding/json"
	"fmt"
	"os"
)

type Identifier struct {
	Name   string `json:"name"`      // from metadata.component.name
	Purl   string `json:"purl"`      // from each component
	Source string `json:"source"`    // file path
}

func ParseCycloneDX(path string) ([]Identifier, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("file read error: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(f, &data); err != nil {
		return nil, fmt.Errorf("json parse error: %w", err)
	}

	if data["bomFormat"] != "CycloneDX" {
		return nil, fmt.Errorf("not CycloneDX")
	}

	// Fetch top-level name from metadata.component.name
	meta, ok := data["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	component, ok := meta["component"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing metadata.component")
	}
	name, ok := component["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("missing metadata.component.name")
	}

	// Fetch components
	components, ok := data["components"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no components")
	}

	var out []Identifier
	for _, raw := range components {
		comp := raw.(map[string]interface{})
		purl, ok := comp["purl"].(string)
		if !ok || purl == "" {
			continue // skip components without purl
		}
		out = append(out, Identifier{
			Name:   name,
			Purl:   purl,
			Source: path,
		})
	}

	return out, nil
}
