package session

import (
	"github.com/google/uuid"
)

// GameClient represents a client connection interface that doesn't depend on websocket implementation
type GameClient interface {
	GetUserID() uuid.UUID
	SendMessage(messageType string, data interface{}) error
	Disconnect()
}

// Message represents a message to be sent to a client
type GameMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}
