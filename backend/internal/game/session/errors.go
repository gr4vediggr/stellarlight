package session

import "errors"

// Session errors
var (
	ErrPlayerAlreadyInSession = errors.New("player already in session")
	ErrPlayerNotInSession     = errors.New("player not in session")
	ErrPlayerNotActive        = errors.New("player not active")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionFull            = errors.New("session is full")
	ErrInvalidCommand         = errors.New("invalid command")
)
