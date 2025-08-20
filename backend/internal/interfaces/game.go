package interfaces

import (
	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/users"
	"github.com/gr4vediggr/stellarlight/pkg/messages"
)

// GameSessionInterface represents a game session without import cycles
type GameSessionInterface interface {
	AddClient(client GameClientInterface)
	RemoveClient(userID uuid.UUID)
	GetID() uuid.UUID
	GetInviteCode() string
	ProcessCommand(cmd *events.ClientCommandWrapper)
}

// GameClientInterface represents a client connection interface
type GameClientInterface interface {
	GetUserID() uuid.UUID
	SendMessage(*messages.ServerMessage) error
	Disconnect()
}

// SessionManagerInterface manages game sessions
type SessionManagerInterface interface {
	GetPlayerSession(playerID uuid.UUID) (GameSessionInterface, error)
	CreateSession(creator *users.User) (GameSessionInterface, error)
}

type GameEngineInterface interface {
	StartGame()
	Stop()
	ProcessGameCommand(cmd *events.ClientCommandWrapper) error
}
