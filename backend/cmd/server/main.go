package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/server"
	"github.com/gr4vediggr/stellarlight/internal/server/websocket"
)

var (
	port = flag.Int("port", 8080, "Port to run the server on")
)

func main() {
	flag.Parse()

	// Generate a unique server ID
	serverID := uuid.New().String()
	log.Printf("Starting server with ID: %s on port %d", serverID, *port)

	hub := server.NewGameLobby()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(websocket.NewWebsocketClient, w, r)
	})
	go hub.Run()
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server is listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
