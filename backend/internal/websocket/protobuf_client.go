package websocket

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/gr4vediggr/stellarlight/internal/game/events"
	"github.com/gr4vediggr/stellarlight/internal/interfaces"
	"github.com/gr4vediggr/stellarlight/internal/users"
	"github.com/gr4vediggr/stellarlight/pkg/messages"
	"google.golang.org/protobuf/proto"
)

type ClientDisconnectHandler func(userID uuid.UUID)

var protobufUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}
var (
	ErrChannelFull = errors.New("send channel is full")
)

// MessageType constants
const (
	MessageTypeGame    = "game"
	MessageTypeLobby   = "lobby"
	MessageTypeLoading = "loading"
	MessageTypeError   = "error"
	MessageTypePing    = "ping"
	MessageTypePong    = "pong"
)

// ProtobufWebSocketMessage wrapper for binary protobuf messages
type ProtobufWebSocketMessage struct {
	MessageID   string
	Timestamp   int64
	MessageType string
	Data        []byte
}

// ProtobufClient handles protobuf WebSocket communication
type ProtobufClient struct {
	conn              *websocket.Conn
	send              chan []byte
	user              *users.User
	sessionManager    interfaces.GameSessionInterface
	disconnectHandler ClientDisconnectHandler
}

// NewProtobufClient creates a new protobuf websocket client
func NewProtobufClient(conn *websocket.Conn, user *users.User, sessionManager interfaces.GameSessionInterface, disconnectHandler ClientDisconnectHandler) *ProtobufClient {
	return &ProtobufClient{
		conn:              conn,
		send:              make(chan []byte, 256),
		user:              user,
		sessionManager:    sessionManager,
		disconnectHandler: disconnectHandler,
	}
}

func (c *ProtobufClient) Start() {
	go c.writePump()
	go c.readPump()
}

func (c *ProtobufClient) readPump() {
	defer func() {
		if c.disconnectHandler != nil {
			c.disconnectHandler(c.user.ID)
		}
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB limit for protobuf messages

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		switch messageType {
		case websocket.BinaryMessage:
			c.handleProtobufMessage(messageBytes)
		case websocket.TextMessage:
			// Fallback for JSON error messages
			continue
		}
	}
}

func (c *ProtobufClient) writePump() {
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

			if err := c.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
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

func (c *ProtobufClient) handleProtobufMessage(data []byte) {
	// Directly decode the ClientCommand protobuf (no wrapper)
	log.Printf("ProtobufClient: Received direct protobuf message, data length: %d bytes", len(data))
	c.handleClientCommand(data)
}

func (c *ProtobufClient) handleClientCommand(data []byte) {
	log.Printf("ProtobufClient: Received raw client command, data length: %d bytes", len(data))
	// Decode client command
	var cmd messages.ClientCommand
	if err := proto.Unmarshal(data, &cmd); err != nil {
		log.Printf("ProtobufClient: Failed to unmarshal client command: %v", err)
		return
	}

	cmd.PlayerId = c.user.ID.String()
	cmd.Timestamp = time.Now().UnixNano()

	wrapped := events.ClientCommandWrapper{
		PlayerID: c.user.ID,
		Command:  &cmd,
	}

	c.sessionManager.ProcessCommand(&wrapped)

}

// Send methods for new message format
func (c *ProtobufClient) SendServerMessage(msg *messages.ServerMessage, messageID string) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	// Send protobuf directly without wrapper
	log.Printf("ProtobufClient: Sending server message directly, data length: %d bytes", len(data))
	select {
	case c.send <- data:
	default:
		log.Printf("Send buffer is full, dropping message")
	}
	return nil
}

func (c *ProtobufClient) SendLobbyMessage(lobbyMsg *messages.LobbyMessage, messageID string) error {
	serverMsg := &messages.ServerMessage{
		Timestamp: time.Now().UnixMilli(),
		MessageId: messageID,
		Message: &messages.ServerMessage_LobbyMessage{
			LobbyMessage: lobbyMsg,
		},
	}
	return c.SendServerMessage(serverMsg, messageID)
}

func (c *ProtobufClient) SendGameMessage(gameMsg *messages.GameMessage, messageID string) error {
	serverMsg := &messages.ServerMessage{
		Timestamp: time.Now().UnixMilli(),
		MessageId: messageID,
		Message: &messages.ServerMessage_GameMessage{
			GameMessage: gameMsg,
		},
	}
	return c.SendServerMessage(serverMsg, messageID)
}

func (c *ProtobufClient) SendError(code, message string) error {
	errorMsg := &messages.ErrorMessage{
		ErrorCode:    code,
		ErrorMessage: message,
		Context:      "",
		Details:      []string{},
	}

	serverMsg := &messages.ServerMessage{
		Timestamp: time.Now().UnixMilli(),
		Message: &messages.ServerMessage_ErrorMessage{
			ErrorMessage: errorMsg,
		},
	}

	return c.SendServerMessage(serverMsg, "")
}

// Cleanup methods
func (c *ProtobufClient) Disconnect() {
	if c.disconnectHandler != nil {
		c.disconnectHandler(c.user.ID)
	}

	close(c.send)
	c.conn.Close()
}

func (c *ProtobufClient) GetUserID() uuid.UUID {
	if c.user != nil {
		return c.user.ID
	}
	return uuid.Nil
}

// Legacy compatibility methods
func (c *ProtobufClient) SendMessage(msg *messages.ServerMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	// Send protobuf directly without wrapper
	log.Printf("ProtobufClient: Sending legacy message directly, data length: %d bytes", len(data))
	select {
	case c.send <- data:
	default:
		log.Printf("Send buffer is full, dropping message")
	}
	return nil

}
