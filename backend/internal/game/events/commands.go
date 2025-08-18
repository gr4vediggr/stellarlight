package events

import "github.com/google/uuid"

// GameCommand represents a command from a player
type GameCommand struct {
	ID        uuid.UUID              `json:"id"`
	PlayerID  uuid.UUID              `json:"player_id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}
