package websocket

import (
	"log"
	"log/slog"
)

// Hub manages WebSocket connections and routes them to game sessions
// It's now simplified to focus only on connection lifecycle
type Hub struct {
	clients        map[*Client]bool
	register       chan *Client
	unregister     chan *Client
	done           chan struct{}
	sessionManager SessionManagerInterface
}

func NewHub(sessionManager SessionManagerInterface) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		done:           make(chan struct{}),
		sessionManager: sessionManager,
	}
}

func (h *Hub) Run() {
	defer slog.Info("WebSocket hub stopped")

	slog.Info("WebSocket hub started")
	for {
		select {
		case <-h.done:
			return
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected: %s", client.user.Email)

			// Set session manager reference and connect to game session
			client.sessionManager = h.sessionManager
			if err := h.sessionManager.ConnectClient(client); err != nil {
				log.Printf("Failed to connect client to session: %v", err)
				// Don't disconnect the client, they might not be in a game yet
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Disconnect from game session
				if h.sessionManager != nil {
					h.sessionManager.DisconnectClient(client.user.ID)
				}

				log.Printf("Client disconnected: %s", client.user.Email)
			}
		}
	}
}

func (h *Hub) Stop() {
	close(h.done)
	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
	slog.Info("WebSocket hub stopped")
}
