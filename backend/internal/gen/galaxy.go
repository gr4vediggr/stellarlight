package gen

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/domain/galaxy"
	"github.com/gr4vediggr/stellarlight/internal/utils"
)

type GalaxyShape string

const (
	SpiralGalaxy GalaxyShape = "spiral"
)

type GalaxyGenerationConfig struct {
	NumStarSystems         int         `json:"numStarSystems"`         // Number of star systems to generate
	Shape                  GalaxyShape `json:"shape"`                  // Shape of the galaxy (e.g., spiral, elliptical)
	HyperlaneDensity       float64     `json:"hyperlaneDensity"`       // Density of hyperlanes in the galaxy, 0.0 to 1.0
	MaxHyperlanesPerSystem int         `json:"maxHyperlanesPerSystem"` // Maximum number of hyperlanes per star system

}

type GalaxyBuilder struct {
	StarTypes   []*galaxy.StarType   // List of star types to use for generation
	PlanetTypes []*galaxy.PlanetType // List of planet types to use for generation
}

func (b GalaxyBuilder) GenerateGalaxy(config GalaxyGenerationConfig) (*galaxy.Galaxy, error) {
	g := galaxy.NewGalaxy("Generated Galaxy")

	// Creating shape and connections
	var t *Triangulation
	var err error
	var genFunc func() []Point
	switch config.Shape {
	case "spiral":
		genFunc = func() []Point {
			return GenerateSpiralGalaxyPointsWithInterarm(config.NumStarSystems, 2, 0.5, 0.2, 2, 0.01, 0.02)
		}
	default:
		return nil, fmt.Errorf("unknown galaxy shape: %s", config.Shape)

	}

	maxDistance := 4. / math.Sqrt(float64(config.NumStarSystems))

	t, err = GenerateValidGalaxy(BaseGenerationConfig{
		NumPoints:    config.NumStarSystems,
		MaxDistance:  maxDistance,
		MaxDegree:    config.MaxHyperlanesPerSystem,
		RemoveChance: 1 - config.HyperlaneDensity,
	}, genFunc)

	if err != nil {
		return nil, fmt.Errorf("failed to generate galaxy shape: %w", err)
	}
	ids := make([]uuid.UUID, 0, config.NumStarSystems)
	for i := 0; i < config.NumStarSystems; i++ {
		// Randomly select a point from the triangulation
		location := t.Points[i]

		system := b.GenerateStarSystem(location.X, location.Y)
		g.AddStarSystem(&system)
		ids = append(ids, system.ID)
	}

	// Connect star systems based on triangulation edges
	for _, edge := range t.Edges {

		system1 := g.StarSystems[ids[edge.A]]
		system2 := g.StarSystems[ids[edge.B]]
		system1.ConnectedSystems = append(system1.ConnectedSystems, system2.ID)
		system2.ConnectedSystems = append(system2.ConnectedSystems, system1.ID)
	}

	return g, err
}

func (b GalaxyBuilder) GenerateStarSystem(locationX, locationY float64) galaxy.StarSystem {
	id := uuid.New()
	// Create a new star system with a unique ID and default properties
	starSystem := galaxy.StarSystem{
		ID:        id,
		Name:      "Star System " + id.String(),
		LocationX: locationX, // Default position, can be randomized later
		LocationY: locationY,
		Stars:     []galaxy.Star{},
	}

	// Generate a random star for the system
	selectedType := utils.WeightedRandomChoice(b.StarTypes)
	star := b.GenerateStar(locationX, locationY, *selectedType)
	starSystem.Stars = append(starSystem.Stars, star)
	return starSystem
}

func (b GalaxyBuilder) GenerateStar(locationX, locationY float64, starType galaxy.StarType) galaxy.Star {
	id := uuid.New()
	s := galaxy.Star{
		ID:        id,
		Name:      "Star " + id.String(),
		Type:      starType,
		LocationX: locationX,
		LocationY: locationY,
		Size:      utils.RandomFloat(starType.MinSize, starType.MaxSize),
		Planets:   []galaxy.Planet{},
	}

	// Generate planets for the star
	for i := 0; i < starType.MaxPlanets; i++ {

		if i >= starType.MinPlanets && rand.Float64() > starType.PlanetChance {
			// If min planets is reached and chance condition fails, skip planet generation
			break
		}

		planetType := utils.WeightedRandomChoice(b.PlanetTypes)

		planet := GeneratePlanet(s, i+1, *planetType)
		s.Planets = append(s.Planets, planet)
	}

	return s
}

func GeneratePlanet(star galaxy.Star, planetNumber int, planetType galaxy.PlanetType) galaxy.Planet {
	id := uuid.New()

	orbitRadius := star.Size + float64(planetNumber+1)*star.Size
	angle := rand.Float64() * 360 // Random angle in degrees

	return galaxy.Planet{
		ID:          id,
		Name:        fmt.Sprintf("%s %d", star.Name, planetNumber),
		Type:        planetType.Name,
		OrbitRadius: orbitRadius,
		Angle:       angle,
	}

}
