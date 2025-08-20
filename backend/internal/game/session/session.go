package session

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/game/types"
	"github.com/gr4vediggr/stellarlight/internal/interfaces"
	"github.com/gr4vediggr/stellarlight/internal/users"
	"github.com/gr4vediggr/stellarlight/pkg/messages"
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
	HostID     uuid.UUID // ID of the host player
	// Players and connections
	players map[uuid.UUID]*types.Player
	clients map[uuid.UUID]interfaces.GameClientInterface
	mu      sync.RWMutex

	// Game engine
	engine interfaces.GameEngineInterface

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// NewGameSession creates a new game session
func NewGameSession(creatorUser *users.User) *GameSession {
	ctx, cancel := context.WithCancel(context.Background())

	session := &GameSession{
		ID:         uuid.New(),
		InviteCode: generateInviteCode(),
		State:      StateWaiting,
		CreatedAt:  time.Now(),
		HostID:     creatorUser.ID, // Set creator as host
		players:    make(map[uuid.UUID]*types.Player),
		clients:    make(map[uuid.UUID]interfaces.GameClientInterface),

		ctx:    ctx,
		cancel: cancel,
	}

	// Add creator as first player
	session.AddPlayer(creatorUser)
	return session
}

// GetID returns the session ID (implements interfaces.GameSessionInterface)
func (s *GameSession) GetID() uuid.UUID {
	return s.ID
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

	return nil
}

// AddClient connects a client to the session
func (s *GameSession) AddClient(client interfaces.GameClientInterface) {
	log.Println("adding client")
	s.mu.Lock()

	s.clients[client.GetUserID()] = client

	// Update player last seen
	if player, exists := s.players[client.GetUserID()]; exists {
		player.LastSeen = time.Now()
		player.IsActive = true
	}
	s.mu.Unlock()

	// Send lobby state to client
	s.broadcastLobbyState()
}

// RemoveClient disconnects a websocket client
func (s *GameSession) RemoveClient(userID uuid.UUID) {
	s.mu.Lock()

	delete(s.clients, userID)

	if player, exists := s.players[userID]; exists {
		player.IsActive = false
		player.LastSeen = time.Now()
	}
	s.mu.Unlock()
	s.broadcastLobbyState()
}

// ProcessCommand handles a command from a client
func (s *GameSession) ProcessCommand(cmd *events.ClientCommandWrapper) {
	// Validate command
	if err := s.validateCommand(cmd); err != nil {
		s.sendErrorToClient(cmd.PlayerID, err)
		return
	}

	if lc := cmd.Command.GetLobbyCommand(); lc != nil {
		s.handleLobbyCommand(cmd.PlayerID, lc)
	}

	if gc := cmd.Command.GetGameCommand(); gc != nil {
		s.engine.ProcessGameCommand(cmd)
	}

}

// Shutdown cleanly shuts down the game session
func (s *GameSession) Shutdown() {
	// Todo
}

func (s *GameSession) validateCommand(cmd *events.ClientCommandWrapper) error {
	// Validate player exists and is active
	s.mu.RLock()
	id := cmd.PlayerID

	player, exists := s.players[id]
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

func (s *GameSession) GetInviteCode() string {
	return s.InviteCode
}

func generateInviteCode() string {
	// Generate a short, human-readable invite code
	return "GAME" + uuid.New().String()[:8]
}

func (s *GameSession) handleLobbyCommand(playerID uuid.UUID, lobbyCmd *messages.LobbyCommand) {
	// handle all lobby commands

	// Example: handle player ready state
	if sr := lobbyCmd.GetSetReady(); sr != nil {
		s.handlePlayerReady(playerID, sr)
	} else if sc := lobbyCmd.GetSetColor(); sc != nil {
		s.handlePlayerColor(playerID, sc)

	} else if st := lobbyCmd.GetUpdateSettings(); st != nil {
		s.handleSettingsUpdate(playerID, st)
	}

	// broadcast new state
	s.broadcastLobbyState()
}

func (s *GameSession) handlePlayerReady(playerID uuid.UUID, data *messages.SetReadyCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ready := data.Ready
	player := s.players[playerID]
	player.Ready = ready

}

func (s *GameSession) handlePlayerColor(playerID uuid.UUID, data *messages.SetColorCommand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	color := data.Color
	player := s.players[playerID]
	player.Color = color

}

func (s *GameSession) handleSettingsUpdate(playerID uuid.UUID, data *messages.UpdateSettingsCommand) {
	// Update lobby settings (implementation needed)
	// For now, just broadcast updated lobby state
}

func (s *GameSession) handleStartGame(playerID uuid.UUID) {
	// Start the game if conditions are met
}

func (s *GameSession) handlePlayerLeave(playerID uuid.UUID) {
	// Remove player from session (implementation needed)
	// For now, just broadcast updated lobby state
}

func (s *GameSession) broadcastLobbyState() {
	lobbyState := s.createLobbyStateMessage()
	log.Println("broadcasting lobby")

	s.mu.RLock()
	defer s.mu.RUnlock()
	msg := &messages.ServerMessage{
		Message: &messages.ServerMessage_LobbyMessage{
			LobbyMessage: lobbyState,
		},
	}
	for _, client := range s.clients {

		go client.SendMessage(msg)
	}
	log.Println("done broadcasting lobby")

}

func (s *GameSession) createLobbyStateMessage() *messages.LobbyMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert session players to lobby players
	var lobbyPlayers []*messages.LobbyPlayer
	for playerID, player := range s.players {
		lobbyPlayer := &messages.LobbyPlayer{
			PlayerId:    playerID.String(),
			DisplayName: player.User.DisplayName,
			IsHost:      s.HostID == playerID,
			IsReady:     player.Ready,
			Color:       player.Color,
		}
		lobbyPlayers = append(lobbyPlayers, lobbyPlayer)
	}

	// Determine lobby status
	var status messages.LobbyStateMessage_LobbyStatus
	switch s.State {
	case StateWaiting:
		status = messages.LobbyStateMessage_WAITING
	case StateActive:
		status = messages.LobbyStateMessage_IN_GAME
	default:
		status = messages.LobbyStateMessage_WAITING
	}

	// Create lobby state
	lobbyStateMsg := &messages.LobbyStateMessage{
		SessionId:    s.ID.String(),
		InviteCode:   s.InviteCode,
		HostPlayerId: s.HostID.String(),
		Status:       status,
		Players:      lobbyPlayers,
		Settings:     nil, // TODO: Implement settings
	}

	return &messages.LobbyMessage{
		Content: &messages.LobbyMessage_LobbyState{
			LobbyState: lobbyStateMsg,
		},
	}
}

func (s *GameSession) sendErrorToClient(playerID uuid.UUID, err error) {
	s.mu.RLock()
	client, exists := s.clients[playerID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	// Create error message
	errorMsg := &messages.ErrorMessage{
		ErrorMessage: err.Error(),
	}

	// Send to the specific client
	if protobufClient, ok := client.(interface {
		SendErrorMessage(msg *messages.ErrorMessage, messageID string) error
	}); ok {
		protobufClient.SendErrorMessage(errorMsg, "")
	}
}
