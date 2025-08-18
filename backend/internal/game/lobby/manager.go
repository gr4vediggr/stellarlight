package lobby

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/google/uuid"
)

type Manager struct {
	lobbies       map[string]*Lobby
	playerLobbies map[uuid.UUID]string // playerID -> lobbyID
	mu            sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		lobbies:       make(map[string]*Lobby),
		playerLobbies: make(map[uuid.UUID]string),
	}
}

func (m *Manager) CreateGame(playerID uuid.UUID) (string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove player from any existing lobby
	if existingLobbyID, exists := m.playerLobbies[playerID]; exists {
		if lobby, found := m.lobbies[existingLobbyID]; found {
			lobby.RemovePlayer(playerID)
		}
	}

	// Create new lobby
	lobby := NewLobby()
	inviteCode := m.generateInviteCode()

	m.lobbies[inviteCode] = lobby
	lobby.AddPlayer(playerID)
	m.playerLobbies[playerID] = inviteCode

	return lobby.ID, inviteCode
}

func (m *Manager) JoinGame(inviteCode string, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lobby, exists := m.lobbies[inviteCode]
	if !exists {
		return errors.New("game not found")
	}

	// Remove player from any existing lobby
	if existingLobbyID, exists := m.playerLobbies[playerID]; exists {
		if existingLobby, found := m.lobbies[existingLobbyID]; found {
			existingLobby.RemovePlayer(playerID)
		}
	}

	lobby.AddPlayer(playerID)
	m.playerLobbies[playerID] = inviteCode
	return nil
}

func (m *Manager) RemovePlayer(playerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if lobbyID, exists := m.playerLobbies[playerID]; exists {
		if lobby, found := m.lobbies[lobbyID]; found {
			lobby.RemovePlayer(playerID)
			// Clean up empty lobby
			if len(lobby.Players) == 0 {
				delete(m.lobbies, lobbyID)
			}
		}
		delete(m.playerLobbies, playerID)
	}
}

func (m *Manager) generateInviteCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
