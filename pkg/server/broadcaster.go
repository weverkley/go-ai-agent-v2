package server

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Broadcaster manages active WebSocket connections and broadcasts messages to them.
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

// NewBroadcaster creates a new Broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[*websocket.Conn]bool),
	}
}

// AddClient registers a new WebSocket client.
func (b *Broadcaster) AddClient(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[conn] = true
	log.Println("Client added. Total clients:", len(b.clients))
}

// RemoveClient unregisters a WebSocket client.
func (b *Broadcaster) RemoveClient(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[conn]; ok {
		conn.Close()
		delete(b.clients, conn)
		log.Println("Client removed. Total clients:", len(b.clients))
	}
}

// Broadcast sends a message to all registered clients.
func (b *Broadcaster) Broadcast(messageType int, data []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		if err := client.WriteMessage(messageType, data); err != nil {
			log.Printf("Error broadcasting to client: %v", err)
			// Assume the client is disconnected and remove it.
			// This is a bit simplistic; a more robust implementation would handle this better.
			go b.RemoveClient(client)
		}
	}
}