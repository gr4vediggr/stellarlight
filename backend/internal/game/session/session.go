package session

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/game/systems"
	"github.com/gr4vediggr/stellarlight/internal/game/types"
	"github.com/gr4vediggr/stellarlight/internal/users"
	"github.com/gr4vediggr/stellarlight/internal/websocket"
)

// GameSessionState represents the current state of a game session
type GameSessionState string

const (
	StateWaiting GameSessionState = "waiting" // Waiting for players
	StateActive  GameSessionState = "active"  // Game is running
	StatePaused  GameSessionState = "paused"  // Game is paused
	StateEnded   GameSessionState = "ended"   // Game has ended
)

// GameSession represents a single game instance (lobby + active game)
type GameSession struct {
	ID         uuid.UUID
	InviteCode string
	State      GameSessionState
	CreatedAt  time.Time

	// Players and connections
	players map[uuid.UUID]*types.Player
	clients map[uuid.UUID]*websocket.Client
	mu      sync.RWMutex

	// Event system
	eventBus *events.EventBus
	systems  []types.GameSystem

	// Game state
	worldState *types.WorldState

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker
}

// NewGameSession creates a new game session
func NewGameSession(creatorUser *users.User) *GameSession {
	ctx, cancel := context.WithCancel(context.Background())

	session := &GameSession{
		ID:         uuid.New(),
		InviteCode: generateInviteCode(),
		State:      StateWaiting,
		CreatedAt:  time.Now(),
		players:    make(map[uuid.UUID]*types.Player),
		clients:    make(map[uuid.UUID]*websocket.Client),
		eventBus:   events.NewEventBus(),
		worldState: types.NewWorldState(),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Add creator as first player
	session.AddPlayer(creatorUser)

	return session
}

// AddPlayer adds a player to the session
func (s *GameSession) AddPlayer(user *users.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.players[user.ID]; exists {
		return ErrPlayerAlreadyInSession
	}

	player := &types.Player{
		User:     user,
		EmpireID: uuid.New(),
		JoinedAt: time.Now(),
		LastSeen: time.Now(),
		IsActive: true,
	}

	s.players[user.ID] = player

	// Publish player joined event
	s.eventBus.Publish(&types.PlayerJoinedEvent{
		BaseEvent: types.BaseEvent{SessionID: s.ID},
		Player:    player,
	})

	return nil
}

// AddClient connects a websocket client to the session
func (s *GameSession) AddClient(client *websocket.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client.GetUserID()] = client

	// Update player last seen
	if player, exists := s.players[client.GetUserID()]; exists {
		player.LastSeen = time.Now()
		player.IsActive = true
	}

	// Send current game state to new client
	s.sendGameStateToClient(client)
}

// RemoveClient disconnects a websocket client
func (s *GameSession) RemoveClient(userID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, userID)

	if player, exists := s.players[userID]; exists {
		player.IsActive = false
		player.LastSeen = time.Now()
	}
}

// ProcessCommand handles a command from a client
func (s *GameSession) ProcessCommand(cmd *events.GameCommand) {
	// Validate command
	if err := s.validateCommand(cmd); err != nil {
		s.sendErrorToClient(cmd.PlayerID, err)
		return
	}

	// Convert command to event and publish
	event := s.commandToEvent(cmd)
	if event != nil {
		s.eventBus.Publish(event)
	}
}

// StartGame transitions from waiting to active state
func (s *GameSession) StartGame() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State != StateWaiting {
		return ErrInvalidStateTransition
	}

	// Initialize game systems
	s.initializeSystems()

	// Start game tick
	s.ticker = time.NewTicker(100 * time.Millisecond) // 10 TPS
	go s.gameLoop()

	s.State = StateActive

	// Publish game started event
	s.eventBus.Publish(&types.GameStartedEvent{
		BaseEvent: types.BaseEvent{SessionID: s.ID},
	})

	return nil
}

// gameLoop runs the main game tick
func (s *GameSession) gameLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.C:
			s.tick()
		}
	}
}

// tick executes one game tick
func (s *GameSession) tick() {
	// Publish tick event to all systems
	s.eventBus.Publish(&types.GameTickEvent{
		BaseEvent: types.BaseEvent{
			SessionID: s.ID,
		},
		Tick:      s.worldState.Turn,
		DeltaTime: 100 * time.Millisecond,
	})

	s.worldState.Turn++
}

// Shutdown cleanly shuts down the game session
func (s *GameSession) Shutdown() {
	s.cancel()
	if s.ticker != nil {
		s.ticker.Stop()
	}

	// Disconnect all clients
	s.mu.Lock()
	for _, client := range s.clients {
		client.Disconnect()
	}
	s.mu.Unlock()

	// Shutdown systems
	for _, system := range s.systems {
		system.Shutdown()
	}
}

// Helper methods
func (s *GameSession) initializeSystems() {
	// Initialize all game systems and subscribe them to events
	s.systems = []types.GameSystem{
		systems.NewEconomySystem(s.eventBus, s.worldState),
		systems.NewCombatSystem(s.eventBus, s.worldState),
		systems.NewClientUpdateSystem(s.eventBus, s.clients),
		// Add more systems as needed
	}

	for _, system := range s.systems {
		system.Initialize()
	}
}

func (s *GameSession) validateCommand(cmd *events.GameCommand) error {
	// Validate player exists and is active
	s.mu.RLock()
	player, exists := s.players[cmd.PlayerID]
	s.mu.RUnlock()

	if !exists {
		return ErrPlayerNotInSession
	}

	if !player.IsActive {
		return ErrPlayerNotActive
	}

	// Add more validation as needed
	return nil
}

func (s *GameSession) commandToEvent(cmd *events.GameCommand) events.GameEvent {
	// Convert commands to appropriate events
	switch cmd.Type {
	case "move_fleet":
		return &types.FleetMoveCommandEvent{
			BaseEvent: types.BaseEvent{SessionID: s.ID},
			PlayerID:  cmd.PlayerID,
			Data:      cmd.Data,
		}
	case "build_ship":
		return &types.BuildShipCommandEvent{
			BaseEvent: types.BaseEvent{SessionID: s.ID},
			PlayerID:  cmd.PlayerID,
			Data:      cmd.Data,
		}
	// Add more command types
	default:
		return nil
	}
}

func (s *GameSession) sendGameStateToClient(client *websocket.Client) {
	// Send current game state to client
	state := s.getGameStateForPlayer(client.GetUserID())
	client.SendMessage(&websocket.Message{
		Type: "game_state_update",
		Data: state,
	})
}

func (s *GameSession) sendErrorToClient(playerID uuid.UUID, err error) {
	s.mu.RLock()
	client, exists := s.clients[playerID]
	s.mu.RUnlock()

	if exists {
		client.SendMessage(&websocket.Message{
			Type:  "error",
			Error: err.Error(),
		})
	}
}

func (s *GameSession) getGameStateForPlayer(playerID uuid.UUID) interface{} {
	// Return filtered game state for specific player
	// This would include only what the player should see
	return map[string]interface{}{
		"session_id": s.ID,
		"state":      s.State,
		"turn":       s.worldState.Turn,
		// Add more state data
	}
}

func generateInviteCode() string {
	// Generate a short, human-readable invite code
	return "GAME" + uuid.New().String()[:8]
}
