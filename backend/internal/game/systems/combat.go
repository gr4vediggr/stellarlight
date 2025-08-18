package systems

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/events"
	"github.com/gr4vediggr/stellarlight/internal/game/types"
)

// CombatSystem handles fleet movement and combat
type CombatSystem struct {
	name       string
	eventBus   *events.EventBus
	worldState *types.WorldState

	subscriptions []func()
	mu            sync.RWMutex
}

func NewCombatSystem(eventBus *events.EventBus, worldState *types.WorldState) *CombatSystem {
	return &CombatSystem{
		name:          "CombatSystem",
		eventBus:      eventBus,
		worldState:    worldState,
		subscriptions: make([]func(), 0),
	}
}

func (s *CombatSystem) Initialize() error {
	log.Printf("Initializing %s", s.name)

	s.subscriptions = append(s.subscriptions,
		s.eventBus.Subscribe("game_tick", s.handleGameTick),
		s.eventBus.Subscribe("fleet_move_command", s.handleFleetMoveCommand),
	)

	return nil
}

func (s *CombatSystem) Shutdown() error {
	log.Printf("Shutting down %s", s.name)

	for _, unsubscribe := range s.subscriptions {
		unsubscribe()
	}
	s.subscriptions = nil

	return nil
}

func (s *CombatSystem) GetName() string {
	return s.name
}

func (s *CombatSystem) handleGameTick(event events.Event) {
	// Check for arriving fleets every tick
	s.processFleetArrivals()
}

func (s *CombatSystem) handleFleetMoveCommand(event events.Event) {
	moveEvent := event.(*types.FleetMoveCommandEvent)

	fleetIDStr, ok := moveEvent.Data["fleet_id"].(string)
	if !ok {
		return
	}

	fleetID, err := uuid.Parse(fleetIDStr)
	if err != nil {
		return
	}

	targetSystemIDStr, ok := moveEvent.Data["target_system"].(string)
	if !ok {
		return
	}

	targetSystemID, err := uuid.Parse(targetSystemIDStr)
	if err != nil {
		return
	}

	// Find the fleet and move it
	if fleet := s.findFleet(fleetID, moveEvent.PlayerID); fleet != nil {
		s.moveFleet(fleet, targetSystemID, moveEvent.typesID)
	}
}

func (s *CombatSystem) findFleet(fleetID, playerID uuid.UUID) *types.Fleet {
	s.worldState.AcquireLock()
	defer s.worldState.ReleaseLock()

	empire, exists := s.worldState.Empires[playerID]
	if !exists {
		return nil
	}

	if fleet, exists := empire.TotalFleets[fleetID]; exists {
		return fleet
	}

	return nil
}

func (s *CombatSystem) moveFleet(fleet *types.Fleet, targetSystemID, typesID uuid.UUID) {
	// Calculate travel time (simplified)
	travelTime := s.calculateTravelTime(fleet.Location, targetSystemID)
	arrivalTime := time.Now().Add(travelTime).Unix()

	// Update fleet destination
	fleet.Destination = &targetSystemID
	fleet.ArrivalTime = &arrivalTime

	// Publish fleet moved event
	s.eventBus.Publish(&types.FleetMovedEvent{
		BaseEvent: types.BaseEvent{
			SessionID: typesID,
			Type:      "fleet_moved",
			Timestamp: time.Now().UnixNano(),
		},
		FleetID:     fleet.ID,
		FromSystem:  fleet.Location,
		ToSystem:    targetSystemID,
		ArrivalTime: arrivalTime,
	})
}

func (s *CombatSystem) processFleetArrivals() {
	currentTime := time.Now().Unix()

	s.worldState.AcquireLock()
	defer s.worldState.ReleaseLock()

	for _, empire := range s.worldState.Empires {
		for _, fleet := range empire.TotalFleets {
			if fleet.ArrivalTime != nil && *fleet.ArrivalTime <= currentTime {
				s.processFleetArrival(fleet)
			}
		}
	}
}

func (s *CombatSystem) processFleetArrival(fleet *types.Fleet) {
	if fleet.Destination == nil {
		return
	}

	// Remove fleet from old system
	if oldSystem, exists := s.worldState.Galaxy.GetSystem(fleet.Location); exists {
		oldSystem.RemoveFleet(fleet.ID)
	}

	// Add fleet to new system
	fleet.Location = *fleet.Destination
	if newSystem, exists := s.worldState.Galaxy.GetSystem(*fleet.Destination); exists {
		newSystem.AddFleet(fleet)
	}

	// Clear destination and arrival time
	fleet.Destination = nil
	fleet.ArrivalTime = nil

	// TODO: Check for combat if enemy fleets present
}

func (s *CombatSystem) calculateTravelTime(from, to uuid.UUID) time.Duration {
	// Simplified travel time calculation
	// In a real game, this would consider distance, fleet speed, etc.
	return 30 * time.Second
}
