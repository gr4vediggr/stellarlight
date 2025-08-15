package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gr4vediggr/stellarlight/internal/domain/galaxy"
)

type AssetFile struct {
	ResourceType string        `json:"resourceType"`
	Resources    []interface{} `json:"resources"`
}

type Assets struct {
	PlanetTypes map[uint32]*galaxy.PlanetType
	StarTypes   map[uint32]*galaxy.StarType
	// Add more types as needed
}

// Recursively loads all JSON assets from the given base directories and fills the Assets struct as maps.
// Later IDs overwrite earlier IDs.
func LoadAssetsFromDirs(baseDirs []string) (*Assets, error) {
	assets := &Assets{
		PlanetTypes: make(map[uint32]*galaxy.PlanetType),
		StarTypes:   make(map[uint32]*galaxy.StarType),
	}
	for _, base := range baseDirs {
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(path) != ".json" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read asset file %s: %w", path, err)
			}
			var af AssetFile
			if err := json.Unmarshal(data, &af); err != nil {
				return fmt.Errorf("failed to parse asset file %s: %w", path, err)
			}
			switch af.ResourceType {
			case "PlanetType":
				for _, r := range af.Resources {
					b, _ := json.Marshal(r)
					var pt galaxy.PlanetType
					if err := json.Unmarshal(b, &pt); err == nil {
						assets.PlanetTypes[pt.ID] = &pt // Overwrite by ID
					}
				}
			case "StarType":
				for _, r := range af.Resources {
					b, _ := json.Marshal(r)
					var st galaxy.StarType
					if err := json.Unmarshal(b, &st); err == nil {
						assets.StarTypes[st.ID] = &st // Overwrite by ID
					}
				}
				// Add more cases for other resource types as needed
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return assets, nil
}
