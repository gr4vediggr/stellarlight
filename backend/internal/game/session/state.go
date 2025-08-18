package session

import (
	"sync"

	"github.com/google/uuid"
)

// GalaxyState represents the state of the game galaxy
type GalaxyState struct {
	Systems map[uuid.UUID]*StarSystemState `json:"systems"`
	mu      sync.RWMutex
}

// StarSystemState represents a star system in the game
type StarSystemState struct {
	ID        uuid.UUID            `json:"id"`
	Name      string               `json:"name"`
	Position  Coordinates          `json:"position"`
	Owner     *uuid.UUID           `json:"owner,omitempty"`
	Planets   []*PlanetState       `json:"planets"`
	Fleets    map[uuid.UUID]*Fleet `json:"fleets"`
	Resources ResourceState        `json:"resources"`
	Buildings []BuildingState      `json:"buildings"`

	mu sync.RWMutex
}

// EmpireState represents a player's empire
type EmpireState struct {
	ID           uuid.UUID                  `json:"id"`
	PlayerID     uuid.UUID                  `json:"player_id"`
	Name         string                     `json:"name"`
	Color        string                     `json:"color"`
	HomeSystem   uuid.UUID                  `json:"home_system"`
	Systems      []uuid.UUID                `json:"systems"`
	TotalFleets  map[uuid.UUID]*Fleet       `json:"total_fleets"`
	Resources    ResourceState              `json:"resources"`
	Technologies map[string]TechnologyLevel `json:"technologies"`

	mu sync.RWMutex
}

// Coordinates represents a position in 2D space
type Coordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// PlanetState represents a planet
type PlanetState struct {
	ID         uuid.UUID     `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Size       int           `json:"size"`
	Population int64         `json:"population"`
	Resources  ResourceState `json:"resources"`
}

// Fleet represents a collection of ships
type Fleet struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Owner       uuid.UUID      `json:"owner"`
	Ships       map[string]int `json:"ships"` // ship_type -> count
	Location    uuid.UUID      `json:"location"`
	Destination *uuid.UUID     `json:"destination,omitempty"`
	ArrivalTime *int64         `json:"arrival_time,omitempty"`
}

// ResourceState represents resources
type ResourceState struct {
	Credits    int64 `json:"credits"`
	Minerals   int64 `json:"minerals"`
	Energy     int64 `json:"energy"`
	Research   int64 `json:"research"`
	Population int64 `json:"population"`
}

// BuildingState represents a building on a planet
type BuildingState struct {
	Type     string    `json:"type"`
	Level    int       `json:"level"`
	PlanetID uuid.UUID `json:"planet_id"`
}

// TechnologyLevel represents research progress
type TechnologyLevel struct {
	Level       int   `json:"level"`
	Progress    int64 `json:"progress"`
	Researching bool  `json:"researching"`
}

// State constructors and methods
func NewGalaxyState() *GalaxyState {
	return &GalaxyState{
		Systems: make(map[uuid.UUID]*StarSystemState),
	}
}

func NewEmpireState(playerID uuid.UUID, name string) *EmpireState {
	return &EmpireState{
		ID:           uuid.New(),
		PlayerID:     playerID,
		Name:         name,
		Systems:      make([]uuid.UUID, 0),
		TotalFleets:  make(map[uuid.UUID]*Fleet),
		Resources:    ResourceState{Credits: 1000, Minerals: 500, Energy: 500},
		Technologies: make(map[string]TechnologyLevel),
	}
}

// Galaxy operations
func (g *GalaxyState) AddSystem(system *StarSystemState) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Systems[system.ID] = system
}

func (g *GalaxyState) GetSystem(id uuid.UUID) (*StarSystemState, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	system, exists := g.Systems[id]
	return system, exists
}

// Empire operations
func (e *EmpireState) AddResources(resources ResourceState) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.Resources.Credits += resources.Credits
	e.Resources.Minerals += resources.Minerals
	e.Resources.Energy += resources.Energy
	e.Resources.Research += resources.Research
	e.Resources.Population += resources.Population
}

func (e *EmpireState) CanAfford(cost ResourceState) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.Resources.Credits >= cost.Credits &&
		e.Resources.Minerals >= cost.Minerals &&
		e.Resources.Energy >= cost.Energy
}

func (e *EmpireState) SpendResources(cost ResourceState) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.CanAfford(cost) {
		return false
	}

	e.Resources.Credits -= cost.Credits
	e.Resources.Minerals -= cost.Minerals
	e.Resources.Energy -= cost.Energy

	return true
}

// System operations
func (s *StarSystemState) SetOwner(empireID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Owner = &empireID
}

func (s *StarSystemState) AddFleet(fleet *Fleet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Fleets[fleet.ID] = fleet
}

func (s *StarSystemState) RemoveFleet(fleetID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Fleets, fleetID)
}
