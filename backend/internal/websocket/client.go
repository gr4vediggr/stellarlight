package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/users"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

var (
	ErrChannelFull = errors.New("send channel is full")
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	user *users.User

	// Session integration
	sessionManager SessionManagerInterface
}

type SessionManagerInterface interface {
	ProcessCommand(playerID uuid.UUID, cmd *events.GameCommand) error
	DisconnectClient(userID uuid.UUID)
	ConnectClient(client *Client) error
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
	// Convert websocket message to game command
	if msg.Type != "" {
		cmd := &events.GameCommand{
			ID:        uuid.New(),
			PlayerID:  c.user.ID, // Set immediately
			Type:      msg.Type,
			Data:      make(map[string]interface{}),
			Timestamp: time.Now().UnixNano(),
		}

		// Copy relevant data
		if msg.Data != nil {
			if dataMap, ok := msg.Data.(map[string]interface{}); ok {
				cmd.Data = dataMap
			}
		}

		// Add any additional fields
		if msg.GameID != "" {
			cmd.Data["game_id"] = msg.GameID
		}
		if msg.InviteCode != "" {
			cmd.Data["invite_code"] = msg.InviteCode
		}

		// Forward to session manager
		if c.sessionManager != nil {
			if err := c.sessionManager.ProcessCommand(c.user.ID, cmd); err != nil {
				c.sendError(err.Error())
			}
		}
	}
}

func (c *Client) SendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		// Channel is full, close connection
		close(c.send)
		return ErrChannelFull
	}
}

func (c *Client) sendError(errorMsg string) {
	c.SendMessage(&Message{
		Type:  "error",
		Error: errorMsg,
	})
}

func (c *Client) Disconnect() {
	if c.sessionManager != nil {
		c.sessionManager.DisconnectClient(c.user.ID)
	}

	c.hub.unregister <- c
	c.conn.Close()
}

func (c *Client) GetUserID() uuid.UUID {
	if c.user != nil {
		return c.user.ID
	}
	return uuid.Nil
}
