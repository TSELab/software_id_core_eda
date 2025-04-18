package orationis

import (
    "encoding/json"
    "fmt"
    "os"
)

type Identifier struct {
    Name     string `json:"name"`
    Version  string `json:"version"`
    Purl     string `json:"purl"`
    Type     string `json:"type"`
    BomRef   string `json:"bomRef"`
    Source   string `json:"source_file"`
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

    var out []Identifier
    components, ok := data["components"].([]interface{})
    if !ok {
        return nil, fmt.Errorf("no components")
    }

    for _, raw := range components {
        comp := raw.(map[string]interface{})
        out = append(out, Identifier{
            Name:     comp["name"].(string),
            Version:  getOrDefault(comp, "version", "UNKNOWN"),
            Purl:     getOrDefault(comp, "purl", "MISSING"),
            Type:     getOrDefault(comp, "type", "library"),
            BomRef:   getOrDefault(comp, "bom-ref", "UNKNOWN"),
            Source:   path,
        })
    }

    return out, nil
}

func getOrDefault(m map[string]interface{}, key, fallback string) string {
    if v, ok := m[key]; ok {
        return v.(string)
    }
    return fallback
}