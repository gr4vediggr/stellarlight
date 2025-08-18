package interfaces

import (
	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/users"
)

// GameSessionInterface represents a game session without import cycles
type GameSessionInterface interface {
	AddClient(client GameClientInterface)
	RemoveClient(userID uuid.UUID)
	GetID() uuid.UUID
	ProcessCommand(cmd *events.GameCommand)
	StartGame() error
}

// GameClientInterface represents a client connection interface
type GameClientInterface interface {
	GetUserID() uuid.UUID
	SendMessage(messageType string, data interface{}) error
	Disconnect()
}

// SessionManagerInterface manages game sessions
type SessionManagerInterface interface {
	ProcessCommand(playerID uuid.UUID, cmd *events.GameCommand) error
	GetPlayerSession(playerID uuid.UUID) (GameSessionInterface, error)
	CreateSession(creator *users.User) (GameSessionInterface, error)
}
