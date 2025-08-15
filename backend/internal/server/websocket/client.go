package websocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/gr4vediggr/stellarlight/internal/server"
	"github.com/gr4vediggr/stellarlight/pkg/packets"
	"google.golang.org/protobuf/proto"
)

type WebsocketClient struct {
	id       uint64
	conn     *websocket.Conn
	logger   *log.Logger
	lobby    *server.GameLobby
	sendChan chan *packets.Packet
}

func NewWebsocketClient(lobby *server.GameLobby, w http.ResponseWriter, r *http.Request) (server.ClientInterfacer, error) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
		WriteBufferSize: 2048,
		ReadBufferSize:  2048,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &WebsocketClient{
		id:       0, // Will be set later
		conn:     conn,
		logger:   log.New(log.Writer(), "Client unkown: ", log.LstdFlags),
		lobby:    lobby,
		sendChan: make(chan *packets.Packet, 256),
	}, nil
}

func (c *WebsocketClient) ID() uint64 {
	return c.id
}

func (c *WebsocketClient) Initialize(id uint64) {
	c.id = id
	c.logger.SetPrefix(fmt.Sprint("Client ID: ", id, ": "))

	c.SocketSend(packets.NewId(c.id))
	c.logger.Printf("Client %d initialized", c.id)

}

func (c *WebsocketClient) ProcessMessage(senderId uint64, message packets.Msg) error {

	if message == nil {
		return nil
	}

	if senderId == c.id {
		c.BroadcastMessage(message)
		c.SocketSendAs(message, senderId)

	} else {
		c.SocketSendAs(message, senderId)
	}
	return nil
}

func (c *WebsocketClient) SocketSendAs(message packets.Msg, senderId uint64) {
	packet := &packets.Packet{
		Sender:  senderId,
		Payload: message,
	}

	select {
	case c.sendChan <- packet:
	default:
		c.logger.Printf("Failed to send message to client %d: channel full", c.id)
	}
}

func (c *WebsocketClient) SocketSend(message packets.Msg) {
	c.SocketSendAs(message, c.id)
}

func (c *WebsocketClient) BroadcastMessage(msg packets.Msg) {
	packet := &packets.Packet{
		Sender:  c.id,
		Payload: msg,
	}

	c.lobby.BroadcastChan <- packet
}

func (c *WebsocketClient) Close(reason string) {

	c.logger.Printf("Client %d disconnected: %s", c.id, reason)
	c.lobby.UnregisterChan <- c
	if err := c.conn.Close(); err != nil {
		c.logger.Printf("Error closing connection for client %d: %v", c.id, err)
	}
	if _, closed := <-c.sendChan; !closed {
		close(c.sendChan)
	}
}

func (c *WebsocketClient) ReadPump() {
	defer func() {
		c.logger.Printf("Client %d disconnected", c.id)
		c.Close("read pump closed")
	}()

	for {
		_, data, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("Unexpected error reading message: %v", err)
			}
			break

		}
		packet := &packets.Packet{}
		err = proto.Unmarshal(data, packet)
		if err != nil {
			c.logger.Printf("Error unmarshalling packet: %v", err)
			continue
		}

		if packet.Sender == 0 {
			packet.Sender = c.id
		}
		c.ProcessMessage(packet.Sender, packet.Payload)
	}

}

func (c *WebsocketClient) WritePump() {
	defer func() {
		c.logger.Printf("Client %d write pump closed", c.id)
		c.Close("write pump closed")
	}()

	for packet := range c.sendChan {
		writer, err := c.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			c.logger.Printf("Error getting writer for client %d: %v", c.id, err)
			return
		}
		data, err := proto.Marshal(packet)
		if err != nil {
			c.logger.Printf("Error marshalling packet for client %d: %v", c.id, err)
			continue
		}
		_, err = writer.Write(data)
		if err != nil {
			c.logger.Printf("Error writing message to client %d: %v", c.id, err)
			continue

		}

		if err := writer.Close(); err != nil {
			c.logger.Printf("error closing writer, dropping %T message: %v", packet, err)
			continue
		}
	}
}

func (c *WebsocketClient) PassToPeer(message packets.Msg, peerId uint64) {
	if peer, exists := c.lobby.Clients.Get(peerId); exists {
		peer.ProcessMessage(c.id, message)
	}
}
