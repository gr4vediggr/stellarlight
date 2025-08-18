package systems

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"

	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/game/types"
	"github.com/gr4vediggr/stellarlight/internal/interfaces"
)

// ClientUpdateSystem handles sending updates to connected clients
type ClientUpdateSystem struct {
	name     string
	eventBus *events.EventBus
	clients  map[uuid.UUID]interfaces.GameClientInterface

	subscriptions []func()
	mu            sync.RWMutex
}

func NewClientUpdateSystem(eventBus *events.EventBus, clients map[uuid.UUID]interfaces.GameClientInterface) *ClientUpdateSystem {
	return &ClientUpdateSystem{
		name:          "ClientUpdateSystem",
		eventBus:      eventBus,
		clients:       clients,
		subscriptions: make([]func(), 0),
	}
}

func (s *ClientUpdateSystem) Initialize() error {
	log.Printf("Initializing %s", s.name)

	// Subscribe to all events that should be sent to clients
	s.subscriptions = append(s.subscriptions,
		s.eventBus.Subscribe("ship_built", s.handleShipBuilt),
		s.eventBus.Subscribe("fleet_moved", s.handleFleetMoved),
		s.eventBus.Subscribe("game_state_update", s.handleGameStateUpdate),
		s.eventBus.Subscribe("player_joined", s.handlePlayerJoined),
		s.eventBus.Subscribe("game_started", s.handleGameStarted),
	)

	return nil
}

func (s *ClientUpdateSystem) Shutdown() error {
	log.Printf("Shutting down %s", s.name)

	for _, unsubscribe := range s.subscriptions {
		unsubscribe()
	}
	s.subscriptions = nil

	return nil
}

func (s *ClientUpdateSystem) GetName() string {
	return s.name
}

func (s *ClientUpdateSystem) handleShipBuilt(event events.GameEvent) {
	shipEvent := event.(*types.ShipBuiltEvent)

	// Send to the player who built the ship
	s.sendToPlayer(shipEvent.PlayerID, "ship_built", map[string]interface{}{
		"ship_type": shipEvent.ShipType,
		"ship_id":   shipEvent.ShipID,
		"system_id": shipEvent.SystemID,
	})
}

func (s *ClientUpdateSystem) handleFleetMoved(event events.GameEvent) {
	fleetEvent := event.(*types.FleetMovedEvent)

	// Send to all players (they can filter based on visibility)
	s.broadcastToAll("fleet_moved", map[string]interface{}{
		"fleet_id":     fleetEvent.FleetID,
		"from_system":  fleetEvent.FromSystem,
		"to_system":    fleetEvent.ToSystem,
		"arrival_time": fleetEvent.ArrivalTime,
	})
}

func (s *ClientUpdateSystem) handleGameStateUpdate(event events.GameEvent) {
	updateEvent := event.(*types.GameStateUpdateEvent)

	s.broadcastToAll("game_state_update", map[string]interface{}{
		"update_type": updateEvent.UpdateType,
		"data":        updateEvent.Data,
	})
}

func (s *ClientUpdateSystem) handlePlayerJoined(event events.GameEvent) {
	playerEvent := event.(*types.PlayerJoinedEvent)

	s.broadcastToAll("player_joined", map[string]interface{}{
		"player_id":    playerEvent.Player.User.ID,
		"display_name": playerEvent.Player.User.DisplayName,
		"empire_id":    playerEvent.Player.EmpireID,
	})
}

func (s *ClientUpdateSystem) handleGameStarted(event events.GameEvent) {
	gameEvent := event.(*types.GameStartedEvent)

	s.broadcastToAll("game_started", map[string]interface{}{
		"session_id": gameEvent.SessionID,
	})
}

func (s *ClientUpdateSystem) sendToPlayer(playerID uuid.UUID, messageType string, data interface{}) {
	s.mu.RLock()
	client, exists := s.clients[playerID]
	s.mu.RUnlock()

	if exists && client != nil {
		if err := client.SendMessage(messageType, data); err != nil {
			log.Printf("Failed to send message to player %s: %v", playerID, err)
		}
	}
}

func (s *ClientUpdateSystem) broadcastToAll(messageType string, data interface{}) {
	s.mu.RLock()
	clients := make([]interfaces.GameClientInterface, 0, len(s.clients))
	for _, client := range s.clients {
		if client != nil {
			clients = append(clients, client)
		}
	}
	s.mu.RUnlock()

	// Send to all clients
	for _, client := range clients {
		if err := client.SendMessage(messageType, data); err != nil {
			log.Printf("Failed to broadcast message to client: %v", err)
		}
	}
}

func (s *ClientUpdateSystem) sendGameStateToPlayer(playerID uuid.UUID, gameState interface{}) {
	data, err := json.Marshal(gameState)
	if err != nil {
		log.Printf("Failed to marshal game state: %v", err)
		return
	}

	s.sendToPlayer(playerID, "full_game_state", json.RawMessage(data))
}
