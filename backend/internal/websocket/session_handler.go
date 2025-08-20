package websocket

import (
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/auth"
	"github.com/gr4vediggr/stellarlight/internal/interfaces"
	"github.com/labstack/echo/v4"
)

type SessionHandler struct {
	sessionManager interfaces.SessionManagerInterface
	authService    *auth.AuthService
}

func NewSessionHandler(sessionManager interfaces.SessionManagerInterface, authService *auth.AuthService) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
		authService:    authService,
	}
}

func (h *SessionHandler) HandleWebSocket(c echo.Context) error {
	// Get token from query parameter
	tokenString := c.QueryParam("token")
	if tokenString == "" {
		slog.Info("No token provided in WebSocket request")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Token required"})
	}

	// Validate token
	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			slog.Info("Expired token in WebSocket request", slog.String("error", err.Error()))
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Token expired",
				"code":  "TOKEN_EXPIRED",
			})
		}
		slog.Info("Invalid token provided in WebSocket request", slog.String("error", err.Error()))
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid token",
			"code":  "TOKEN_INVALID",
		})
	}

	// Get user
	user, err := h.authService.GetUserByID(c.Request().Context(), claims.UserID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
	}

	// Try to find existing session for this player
	gameSession, err := h.sessionManager.GetPlayerSession(user.ID)
	if err != nil {
		// No existing session, create a new one
		gameSession, err = h.sessionManager.CreateSession(user)
		if err != nil {
			slog.Error("Failed to create session", slog.String("error", err.Error()))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create game session"})
		}
		slog.Info("Created new session for player", slog.String("player", user.Email), slog.String("session_id", gameSession.GetID().String()))
	} else {
		slog.Info("Player rejoining existing session", slog.String("player", user.Email), slog.String("session_id", gameSession.GetID().String()))
	}

	// Upgrade connection
	conn, err := protobufUpgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upgrade connection"})
	}

	// Disconnect handler that removes client from their session
	disconnectHandler := func(userID uuid.UUID) {
		// Find and remove client from their session
		if session, err := h.sessionManager.GetPlayerSession(userID); err == nil {
			session.RemoveClient(userID)
			log.Printf("Client disconnected from session: %s (session: %s)", user.Email, session.GetID().String())
		}
	}

	// Create protobuf client
	client := NewProtobufClient(conn, user, gameSession, disconnectHandler)

	client.Start()

	gameSession.AddClient(client)
	log.Printf("Protobuf client connected to session: %s (session: %s)", user.Email, gameSession.GetID().String())
	return nil // Don't send JSON response after WebSocket upgrade
}
