package galaxy

import (
	"log"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/utils"
)

type Galaxy struct {
	ID           uuid.UUID                           `json:"id"`          // Unique identifier for the galaxy
	Name         string                              `json:"name"`        // Name of the galaxy
	StarSystems  map[uuid.UUID]*StarSystem           `json:"starSystems"` // Map of star systems in the galaxy, keyed by their unique IDs
	Adjacency    map[uuid.UUID]map[uuid.UUID]float64 `json:"-"`           // Adjacency list for star systems, keyed by their unique IDs
	HeuristicMap map[uuid.UUID]map[uuid.UUID]float64 `json:"-"`           // Heuristic distances between star systems, keyed by their unique IDs
}

func (g *Galaxy) heuristicFunc(from, to uuid.UUID) float64 {
	if g.HeuristicMap == nil {
		g.BuildHeuristicMap()
	}
	if distances, ok := g.HeuristicMap[from]; ok {
		if distance, ok := distances[to]; ok {
			return distance
		}
	}
	return 999999 // Return a large value if no heuristic is found
}

type StarSystem struct {
	ID      uuid.UUID `json:"id"`      // Unique identifier for the star system
	OwnerID uuid.UUID `json:"ownerId"` // ID of the owner of the star system

	Name      string  `json:"name"`      // Name of the star system
	LocationX float64 `json:"locationX"` // X coordinate in the galaxy
	LocationY float64 `json:"locationY"` // Y coordinate in the galaxy

	ConnectedSystems []uuid.UUID `json:"connectedSystems"` // IDs of connected star systems

	Stars []Star `json:"stars"` // List of stars in the star system
}

func (s *StarSystem) DistanceSquared(other *StarSystem) float64 {
	return ((s.LocationX - other.LocationX) * (s.LocationX - other.LocationX)) +
		((s.LocationY - other.LocationY) * (s.LocationY - other.LocationY))
}

type Star struct {
	ID   uuid.UUID `json:"id"`   // Unique identifier for the star
	Name string    `json:"name"` // Name of the star
	Type StarType  `json:"type"` // Type of the star (e.g., red dwarf, yellow giant)

	LocationX float64  `json:"locationX"` // X coordinate in the star system
	LocationY float64  `json:"locationY"` // Y coordinate in the star system
	Size      float64  `json:"size"`      // Size of the star
	Planets   []Planet `json:"planets"`   // List of planets orbiting the star
}

type Planet struct {
	ID          uuid.UUID `json:"id"`          // Unique identifier for the planet
	Name        string    `json:"name"`        // Name of the planet
	Type        string    `json:"type"`        // Type of the planet (e.g., terrestrial, gas giant)
	Size        int       `json:"size"`        // Size of the planet
	OrbitRadius float64   `json:"orbitRadius"` // Distance from the star in astronomical units
	Angle       float64   `json:"angle"`       // Current angle in its orbit
}

func NewGalaxy(name string) *Galaxy {
	return &Galaxy{
		ID:          uuid.New(),
		Name:        name,
		StarSystems: make(map[uuid.UUID]*StarSystem),
		Adjacency:   make(map[uuid.UUID]map[uuid.UUID]float64),
	}
}

func (g *Galaxy) AddStarSystem(system *StarSystem) {
	system.ID = uuid.New()
	g.StarSystems[system.ID] = system
}

func (g *Galaxy) BuildAdjacencyList() {

	g.Adjacency = make(map[uuid.UUID]map[uuid.UUID]float64)
	for id := range g.StarSystems {
		g.Adjacency[id] = make(map[uuid.UUID]float64)
	}
	for _, system := range g.StarSystems {
		for _, connectedID := range system.ConnectedSystems {
			distance := system.DistanceSquared(g.StarSystems[connectedID])
			g.Adjacency[system.ID][connectedID] = distance
			g.Adjacency[connectedID][system.ID] = distance
		}
	}
}

func (g *Galaxy) AddHeuristicDistance(from, to uuid.UUID, distance float64) {
	if g.HeuristicMap == nil {
		g.HeuristicMap = make(map[uuid.UUID]map[uuid.UUID]float64)
	}
	if _, exists := g.HeuristicMap[from]; !exists {
		g.HeuristicMap[from] = make(map[uuid.UUID]float64)
	}
	g.HeuristicMap[from][to] = distance
}

func (g *Galaxy) BuildHeuristicMap() {
	log.Println("Building heuristic map for galaxy:", g.Name)
	g.HeuristicMap = utils.GenerateHeuristicMap(g.Adjacency)
	log.Println("Heuristic map built with", len(g.HeuristicMap), "entries")
}

func (g *Galaxy) FindPath(startID, endID uuid.UUID) []uuid.UUID {

	return utils.FindPath(g.Adjacency, startID, endID, g.heuristicFunc)
}
