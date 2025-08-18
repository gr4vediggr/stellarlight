package events

type EventType string

// Event represents a generic event in the system
const (
	EventTypeGameTick    EventType = "game_tick"
	EventTypePlayerJoin  EventType = "player_join"
	EventTypePlayerLeave EventType = "player_leave"
	EventTypeShipBuilt   EventType = "ship_built"
	EventTypeFleetMoved  EventType = "fleet_moved"
	EventTypeGameStarted EventType = "game_started"
)
