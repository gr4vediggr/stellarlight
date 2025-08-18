package lobby

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Lobby struct {
	ID         string
	Players    map[uuid.UUID]bool
	MaxPlayers int
	CreatedAt  time.Time
	mu         sync.RWMutex
}

func NewLobby() *Lobby {
	return &Lobby{
		ID:         uuid.New().String(),
		Players:    make(map[uuid.UUID]bool),
		MaxPlayers: 24,
		CreatedAt:  time.Now(),
	}
}

func (l *Lobby) AddPlayer(playerID uuid.UUID) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.Players) >= l.MaxPlayers {
		return false
	}

	l.Players[playerID] = true
	return true
}

func (l *Lobby) RemovePlayer(playerID uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.Players, playerID)
}

func (l *Lobby) GetPlayerCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return len(l.Players)
}

func (l *Lobby) HasPlayer(playerID uuid.UUID) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.Players[playerID]
}
