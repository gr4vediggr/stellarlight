package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/users"
)

// Player represents a player in the game session
type Player struct {
	User     *users.User
	EmpireID uuid.UUID
	JoinedAt time.Time
	LastSeen time.Time
	IsActive bool
}

// GameSystem interface that all game systems must implement
type GameSystem interface {
	Initialize() error
	Shutdown() error
	GetName() string
}

// Base event struct
type BaseEvent struct {
	SessionID uuid.UUID `json:"session_id"`
	Type      string    `json:"type"`
	Timestamp int64     `json:"timestamp"`
}

func (e *BaseEvent) GetSessionID() uuid.UUID { return e.SessionID }
func (e *BaseEvent) GetType() string         { return e.Type }
func (e *BaseEvent) GetTimestamp() int64     { return e.Timestamp }

// System Events
type GameTickEvent struct {
	BaseEvent
	Tick      int           `json:"tick"`
	DeltaTime time.Duration `json:"delta_time"` // in milliseconds
}

type GameStartedEvent struct {
	BaseEvent
}

type PlayerJoinedEvent struct {
	BaseEvent
	Player *Player `json:"player"`
}

// Command Events (converted from commands)
type FleetMoveCommandEvent struct {
	BaseEvent
	PlayerID uuid.UUID              `json:"player_id"`
	Data     map[string]interface{} `json:"data"`
}

type BuildShipCommandEvent struct {
	BaseEvent
	PlayerID uuid.UUID              `json:"player_id"`
	Data     map[string]interface{} `json:"data"`
}

// System-generated Events
type FleetMovedEvent struct {
	BaseEvent
	FleetID     uuid.UUID `json:"fleet_id"`
	FromSystem  uuid.UUID `json:"from_system"`
	ToSystem    uuid.UUID `json:"to_system"`
	ArrivalTime int64     `json:"arrival_time"`
}

type ShipBuiltEvent struct {
	BaseEvent
	PlayerID uuid.UUID `json:"player_id"`
	SystemID uuid.UUID `json:"system_id"`
	ShipType string    `json:"ship_type"`
	ShipID   uuid.UUID `json:"ship_id"`
}

// Client Update Events
type PlayerStateUpdateEvent struct {
	BaseEvent
	PlayerID uuid.UUID   `json:"player_id"`
	State    interface{} `json:"state"`
}

type GameStateUpdateEvent struct {
	BaseEvent
	UpdateType string      `json:"update_type"`
	Data       interface{} `json:"data"`
}
