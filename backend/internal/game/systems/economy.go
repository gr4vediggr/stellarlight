package systems

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/events"
	"github.com/gr4vediggr/stellarlight/internal/game/types"
)

// EconomySystem handles resource generation and management
type EconomySystem struct {
	name       string
	eventBus   *events.EventBus
	worldState *types.WorldState

	// Subscriptions
	subscriptions []func()
	mu            sync.RWMutex
}

func NewEconomySystem(eventBus *events.EventBus, worldState *types.WorldState) *EconomySystem {
	return &EconomySystem{
		name:          "EconomySystem",
		eventBus:      eventBus,
		worldState:    worldState,
		subscriptions: make([]func(), 0),
	}
}

func (s *EconomySystem) Initialize() error {
	log.Printf("Initializing %s", s.name)

	// Subscribe to relevant events
	s.subscriptions = append(s.subscriptions,
		s.eventBus.Subscribe("game_tick", s.handleGameTick),
		s.eventBus.Subscribe("build_ship_command", s.handleBuildShipCommand),
	)

	return nil
}

func (s *EconomySystem) Shutdown() error {
	log.Printf("Shutting down %s", s.name)

	// Unsubscribe from all events
	for _, unsubscribe := range s.subscriptions {
		unsubscribe()
	}
	s.subscriptions = nil

	return nil
}

func (s *EconomySystem) GetName() string {
	return s.name
}

func (s *EconomySystem) handleGameTick(event events.Event) {
	tickEvent := event.(*types.GameTickEvent)

	// Generate resources every 10 ticks (1 second at 10 TPS)
	if tickEvent.Tick%10 == 0 {
		s.generateResources()
	}
}

func (s *EconomySystem) handleBuildShipCommand(event events.Event) {
	buildEvent := event.(*types.BuildShipCommandEvent)

	// Get empire
	s.worldState.AcquireLock()
	empire, exists := s.worldState.Empires[buildEvent.PlayerID]
	s.worldState.ReleaseLock()

	if !exists {
		return
	}

	// Get ship cost (example)
	shipType, ok := buildEvent.Data["ship_type"].(string)
	if !ok {
		return
	}

	cost := s.getShipCost(shipType)

	// Check if empire can afford it
	if empire.CanAfford(cost) {
		if empire.SpendResources(cost) {
			// Create ship built event
			s.eventBus.Publish(&types.ShipBuiltEvent{
				BaseEvent: types.BaseEvent{
					SessionID: buildEvent.typesID,
					Type:      "ship_built",
					Timestamp: time.Now().UnixNano(),
				},
				PlayerID: buildEvent.PlayerID,
				ShipType: shipType,
				ShipID:   uuid.New(),
			})
		}
	}
}

func (s *EconomySystem) generateResources() {
	s.worldState.AcquireLock()
	defer s.worldState.ReleaseLock()

	for _, empire := range s.worldState.Empires {
		// Calculate resource generation based on empire's systems
		generation := s.calculateResourceGeneration(empire)
		empire.AddResources(generation)
	}
}

func (s *EconomySystem) calculateResourceGeneration(empire *types.EmpireState) types.ResourceState {
	// Example calculation - in real game this would be more complex
	return types.ResourceState{
		Credits:  10,
		Minerals: 5,
		Energy:   5,
		Research: 2,
	}
}

func (s *EconomySystem) getShipCost(shipType string) types.ResourceState {
	costs := map[string]types.ResourceState{
		"fighter":     {Credits: 100, Minerals: 50, Energy: 25},
		"cruiser":     {Credits: 500, Minerals: 200, Energy: 100},
		"dreadnought": {Credits: 2000, Minerals: 1000, Energy: 500},
	}

	if cost, exists := costs[shipType]; exists {
		return cost
	}

	// Default cost
	return types.ResourceState{Credits: 100, Minerals: 50, Energy: 25}
}
