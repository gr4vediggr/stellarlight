package events

import (
	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/pkg/messages"
)

// GameCommand represents a command from a player
type GameCommand struct {
	ID        uuid.UUID              `json:"id"`
	PlayerID  uuid.UUID              `json:"player_id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

type ClientCommandWrapper struct {
	PlayerID uuid.UUID
	Command  *messages.ClientCommand
}
