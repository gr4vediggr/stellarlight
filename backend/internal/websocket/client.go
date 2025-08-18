package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gr4vediggr/stellarlight/internal/users"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	user *users.User
}

type Message struct {
	Type       string      `json:"type"`
	Data       interface{} `json:"data,omitempty"`
	GameID     string      `json:"gameId,omitempty"`
	InviteCode string      `json:"inviteCode,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case "create_game":
		gameID, inviteCode := c.hub.lobbyManager.CreateGame(c.user.ID)
		c.sendMessage(&Message{
			Type:       "game_joined",
			GameID:     gameID,
			InviteCode: inviteCode,
		})

	case "join_game":
		if err := c.hub.lobbyManager.JoinGame(msg.InviteCode, c.user.ID); err != nil {
			c.sendMessage(&Message{
				Type:  "error",
				Error: err.Error(),
			})
			return
		}
		c.sendMessage(&Message{
			Type:       "game_joined",
			GameID:     msg.InviteCode, // Simplified for demo
			InviteCode: msg.InviteCode,
		})

	case "leave_game":
		c.hub.lobbyManager.RemovePlayer(c.user.ID)
		c.sendMessage(&Message{
			Type: "game_left",
		})

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (c *Client) sendMessage(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		delete(c.hub.clients, c)
	}
}
