package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/interfaces"
	"github.com/gr4vediggr/stellarlight/internal/users"
)

// SessionManager manages all active game sessions
type SessionManager struct {
	sessions       map[uuid.UUID]*GameSession // sessionID -> session
	playerSessions map[uuid.UUID]uuid.UUID    // playerID -> sessionID
	inviteCodes    map[string]uuid.UUID       // inviteCode -> sessionID
	mu             sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:       make(map[uuid.UUID]*GameSession),
		playerSessions: make(map[uuid.UUID]uuid.UUID),
		inviteCodes:    make(map[string]uuid.UUID),
	}
}

// CreateSession creates a new game session
func (sm *SessionManager) CreateSession(creator *users.User) (interfaces.GameSessionInterface, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if player is already in a session
	if existingSessionID, exists := sm.playerSessions[creator.ID]; exists {
		if session, exists := sm.sessions[existingSessionID]; exists {
			// Player can rejoin if session is still active
			if session.State != StateEnded {
				return session, ErrPlayerAlreadyInSession
			}
			// Clean up ended session
			sm.cleanupSession(existingSessionID)
		}
	}

	// Create new session
	session := NewGameSession(creator)

	// Register session
	sm.sessions[session.ID] = session
	sm.playerSessions[creator.ID] = session.ID
	sm.inviteCodes[session.InviteCode] = session.ID

	return session, nil
}

// JoinSession allows a player to join an existing session
func (sm *SessionManager) JoinSession(player *users.User, inviteCode string) (interfaces.GameSessionInterface, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if player is already in a session
	if existingSessionID, exists := sm.playerSessions[player.ID]; exists {
		if session, exists := sm.sessions[existingSessionID]; exists {
			if session.State != StateEnded {
				return session, ErrPlayerAlreadyInSession
			}
			sm.cleanupSession(existingSessionID)
		}
	}

	// Find session by invite code
	sessionID, exists := sm.inviteCodes[inviteCode]
	if !exists {
		return nil, ErrSessionNotFound
	}

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	// Check if session can accept new players
	if session.State != StateWaiting {
		return nil, ErrInvalidStateTransition
	}

	// Add player to session
	if err := session.AddPlayer(player); err != nil {
		return nil, err
	}

	// Track player's session
	sm.playerSessions[player.ID] = session.ID

	return session, nil
}

// GetPlayerSession returns the session a player is currently in
func (sm *SessionManager) GetPlayerSession(playerID uuid.UUID) (interfaces.GameSessionInterface, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionID, exists := sm.playerSessions[playerID]
	if !exists {
		return nil, ErrPlayerNotInSession
	}

	session, exists := sm.sessions[sessionID]
	if !exists {
		// Cleanup stale reference
		delete(sm.playerSessions, playerID)
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// GetSession returns a session by ID
func (sm *SessionManager) GetSession(sessionID uuid.UUID) (interfaces.GameSessionInterface, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// GetSessionByInviteCode returns a session by invite code
func (sm *SessionManager) GetSessionByInviteCode(inviteCode string) (interfaces.GameSessionInterface, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionID, exists := sm.inviteCodes[inviteCode]
	if !exists {
		return nil, ErrSessionNotFound
	}

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// LeaveSession removes a player from their current session
func (sm *SessionManager) LeaveSession(playerID uuid.UUID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID, exists := sm.playerSessions[playerID]
	if !exists {
		return ErrPlayerNotInSession
	}

	session, exists := sm.sessions[sessionID]
	if !exists {
		delete(sm.playerSessions, playerID)
		return ErrSessionNotFound
	}

	// Remove player from session
	session.RemoveClient(playerID)
	delete(sm.playerSessions, playerID)

	// If this was the last player, clean up the session
	session.mu.RLock()
	playerCount := len(session.players)
	session.mu.RUnlock()

	if playerCount == 0 {
		sm.cleanupSession(sessionID)
	}

	return nil
}

// ProcessCommand forwards a command to the appropriate session
func (sm *SessionManager) ProcessCommand(playerID uuid.UUID, cmd *events.GameCommand) error {
	session, err := sm.GetPlayerSession(playerID)
	if err != nil {
		return err
	}

	// Ensure command is from the correct player
	cmd.PlayerID = playerID
	cmd.Timestamp = time.Now().UnixNano()

	session.ProcessCommand(cmd)
	return nil
}

// CleanupExpiredSessions removes old sessions
func (sm *SessionManager) CleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	expiredSessions := make([]uuid.UUID, 0)

	for sessionID, session := range sm.sessions {
		// Clean up sessions that have been ended for more than 1 hour
		if session.State == StateEnded && time.Since(session.CreatedAt) > time.Hour {
			expiredSessions = append(expiredSessions, sessionID)
		}

		// Clean up waiting sessions older than 24 hours
		if session.State == StateWaiting && time.Since(session.CreatedAt) > 24*time.Hour {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	for _, sessionID := range expiredSessions {
		sm.cleanupSession(sessionID)
	}
}

// cleanupSession removes a session and all its references (must be called with lock held)
func (sm *SessionManager) cleanupSession(sessionID uuid.UUID) {
	session, exists := sm.sessions[sessionID]
	if !exists {
		return
	}

	// Shutdown the session
	session.Shutdown()

	// Remove from sessions
	delete(sm.sessions, sessionID)

	// Remove invite code
	delete(sm.inviteCodes, session.InviteCode)

	// Remove player references
	for playerID := range session.players {
		delete(sm.playerSessions, playerID)
	}
}

// GetActiveSessions returns a list of all active sessions (for admin/monitoring)
func (sm *SessionManager) GetActiveSessions() []interfaces.GameSessionInterface {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]interfaces.GameSessionInterface, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}
