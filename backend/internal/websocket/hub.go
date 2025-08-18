package websocket

import (
	"log"
	"log/slog"
)

type Hub struct {
	clients        map[*Client]bool
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	done           chan struct{}
	sessionManager SessionManagerInterface
}

func NewHub(sessionManager SessionManagerInterface) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		broadcast:      make(chan []byte),
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

func (h *Hub) Stop() {
	close(h.done)
	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
	slog.Info("WebSocket hub stopped")
}
