package gen_test

import (
	"testing"

	"github.com/gr4vediggr/stellarlight/internal/domain/galaxy"
	"github.com/gr4vediggr/stellarlight/internal/gen"
	"github.com/gr4vediggr/stellarlight/internal/render"
	"github.com/gr4vediggr/stellarlight/internal/resource"
)

func TestGenerateGalaxy(t *testing.T) {

	// Load base assets

	assets, err := resource.LoadAssetsFromDirs([]string{
		"../../assets/",
	})
	if err != nil {
		t.Fatalf("Failed to load assets: %v", err)
	}

	planetTypesSlice := make([]*galaxy.PlanetType, 0, len(assets.PlanetTypes))
	for _, pt := range assets.PlanetTypes {
		planetType := pt // create a new variable to take the address
		planetTypesSlice = append(planetTypesSlice, planetType)
	}

	starTypesSlice := make([]*galaxy.StarType, 0, len(assets.StarTypes))
	for _, st := range assets.StarTypes {
		starType := st // create a new variable to take the address
		starTypesSlice = append(starTypesSlice, starType)
	}

	builder := gen.GalaxyBuilder{
		PlanetTypes: planetTypesSlice,
		StarTypes:   starTypesSlice,
	}

	config := gen.GalaxyGenerationConfig{
		NumStarSystems: 500,      // Number of star systems to generate
		Shape:          "spiral", // Shape of the galaxy
	}

	galaxy, err := builder.GenerateGalaxy(config)

	if err != nil {
		t.Fatalf("Failed to generate galaxy: %v", err)
	}

	if len(galaxy.StarSystems) != config.NumStarSystems {
		t.Errorf("Expected %d star systems, got %d", config.NumStarSystems, len(galaxy.StarSystems))
	}

	for _, system := range galaxy.StarSystems {
		if len(system.Stars) == 0 {
			t.Error("Expected at least one star in each star system")
		}
	}
	galaxy.BuildAdjacencyList()
	galaxy.BuildHeuristicMap()
	render.RenderGalaxyToImage(galaxy, "generated_galaxy.png")
	t.Error("Test completed, check generated_galaxy.png for visual output")

}
