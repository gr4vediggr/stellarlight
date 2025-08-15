package resource_test

import (
	"testing"

	"github.com/gr4vediggr/stellarlight/internal/resource"
)

func TestLoadAssetsFromDirs(t *testing.T) {
	assets, err := resource.LoadAssetsFromDirs([]string{"../../assets/"})
	if err != nil {
		t.Fatalf("Failed to load assets: %v", err)
	}

	if len(assets.PlanetTypes) == 0 {
		t.Error("Expected at least one planet type, got none")
	}

	if len(assets.StarTypes) == 0 {
		t.Error("Expected at least one star type, got none")
	}

	for _, pt := range assets.PlanetTypes {
		if pt.MinSize >= pt.MaxSize {
			t.Errorf("PlanetType %s has invalid size range: %f - %f", pt.Name, pt.MinSize, pt.MaxSize)
		}
	}
}
