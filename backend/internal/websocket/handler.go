package websocket

import (
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/gr4vediggr/stellarlight/internal/auth"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	hub         *Hub
	authService *auth.AuthService
}

func NewHandler(hub *Hub, authService *auth.AuthService) *Handler {
	return &Handler{
		hub:         hub,
		authService: authService,
	}
}

func (h *Handler) HandleWebSocket(c echo.Context) error {
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

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to upgrade connection"})
	}

	// Create client
	client := &Client{
		hub:  h.hub,
		conn: conn,
		send: make(chan []byte, 256),
		user: user,
	}

	// Register client
	client.hub.register <- client

	// Start goroutines
	go client.writePump()
	go client.readPump()

	return c.JSON(http.StatusOK, map[string]string{"message": "WebSocket connection established"})
}
