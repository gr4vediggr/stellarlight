package systems

import (
	"log"
	"sync"

	"github.com/google/uuid"

	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/interfaces"
	"github.com/gr4vediggr/stellarlight/pkg/messages"
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

}

func (s *ClientUpdateSystem) handleFleetMoved(event events.GameEvent) {

}

func (s *ClientUpdateSystem) handleGameStateUpdate(event events.GameEvent) {

}

func (s *ClientUpdateSystem) handlePlayerJoined(event events.GameEvent) {

}

func (s *ClientUpdateSystem) handleGameStarted(event events.GameEvent) {

}

func (s *ClientUpdateSystem) SendToPlayer(playerID uuid.UUID, message *messages.ServerMessage) {
	s.mu.RLock()
	client, exists := s.clients[playerID]
	s.mu.RUnlock()

	if exists && client != nil {
		if err := client.SendMessage(message); err != nil {
			log.Printf("Failed to send message to player %s: %v", playerID, err)
		}
	}
}

func (s *ClientUpdateSystem) BroadcastToAll(message *messages.ServerMessage) {
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
		if err := client.SendMessage(message); err != nil {
			log.Printf("Failed to broadcast message to client: %v", err)
		}
	}
}
