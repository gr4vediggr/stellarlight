package server

import (
	"net/http"

	"github.com/gr4vediggr/stellarlight/internal/server/objects"
	"github.com/gr4vediggr/stellarlight/pkg/packets"
)

type ClientInterfacer interface {
	ID() uint64

	Initialize(uint64)

	ProcessMessage(uint64, packets.Msg) error

	BroadcastMessage(msg packets.Msg)

	SocketSendAs(msg packets.Msg, id uint64)
	SocketSend(msg packets.Msg)

	WritePump()
	ReadPump()

	Close(reason string)
}

type GameLobby struct {
	Clients *objects.SharedCollection[ClientInterfacer]

	RegisterChan   chan ClientInterfacer
	UnregisterChan chan ClientInterfacer
	BroadcastChan  chan *packets.Packet
}

func NewGameLobby() *GameLobby {
	return &GameLobby{
		Clients:        objects.NewSharedCollection[ClientInterfacer](),
		BroadcastChan:  make(chan *packets.Packet, 256),
		RegisterChan:   make(chan ClientInterfacer),
		UnregisterChan: make(chan ClientInterfacer),
	}
}

func (l *GameLobby) Run() {
	for {
		select {
		case client := <-l.RegisterChan:
			client.Initialize(l.Clients.Add(client))
		case client := <-l.UnregisterChan:
			l.Clients.Remove(client.ID())
		case packet := <-l.BroadcastChan:

			l.Clients.ForEach(func(id uint64, client ClientInterfacer) {

				if client == nil {
					return
				}

				if id != packet.Sender {
					if err := client.ProcessMessage(packet.Sender, packet.Payload); err != nil {

					}
				}

			})
		}
	}
}

func (l *GameLobby) Close() {
	close(l.BroadcastChan)

	close(l.RegisterChan)
	close(l.UnregisterChan)
}

func (l *GameLobby) Serve(registerFunc func(l *GameLobby, writer http.ResponseWriter, r *http.Request) (ClientInterfacer, error), w http.ResponseWriter, r *http.Request) {
	client, err := registerFunc(l, w, r)
	if err != nil {
		http.Error(w, "Failed to register client", http.StatusInternalServerError)
		return
	}
	l.RegisterChan <- client

	go client.WritePump()
	go client.ReadPump()
}
