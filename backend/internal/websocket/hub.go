package websocket

import (
	"context"
	"log"
	"log/slog"

	"github.com/gr4vediggr/stellarlight/internal/auth"
	"github.com/gr4vediggr/stellarlight/internal/game/lobby"
)

type Hub struct {
	clients      map[*Client]bool
	broadcast    chan []byte
	register     chan *Client
	unregister   chan *Client
	authService  *auth.AuthService
	lobbyManager *lobby.Manager
}

func NewHub(authService *auth.AuthService, lobbyManager *lobby.Manager) *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		authService:  authService,
		lobbyManager: lobbyManager,
	}
}

func (h *Hub) Run(ctx context.Context) {
	defer slog.Info("WebSocket hub stopped")

	slog.Info("WebSocket hub started")
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected: %s", client.user.Email)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.lobbyManager.RemovePlayer(client.user.ID)
				log.Printf("Client disconnected: %s", client.user.Email)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
