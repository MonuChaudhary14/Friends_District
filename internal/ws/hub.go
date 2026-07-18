package ws

import (
	"encoding/json"
	"log"

	"district-friends/internal/models"
)

type BroadcastMessage struct {
	RoomID  uint
	Message models.Message
}

// Hub maintains the set of active clients grouped by room
type Hub struct {
	// Registered clients grouped by room ID
	rooms map[uint]map[*Client]bool

	// Inbound messages from the clients
	Broadcast chan BroadcastMessage

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan BroadcastMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		rooms:      make(map[uint]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			roomClients, ok := h.rooms[client.RoomID]
			if !ok {
				roomClients = make(map[*Client]bool)
				h.rooms[client.RoomID] = roomClients
			}
			roomClients[client] = true
			log.Printf("Client registered to room %d. Total clients in room: %d", client.RoomID, len(roomClients))

		case client := <-h.Unregister:
			if roomClients, ok := h.rooms[client.RoomID]; ok {
				if _, ok := roomClients[client]; ok {
					delete(roomClients, client)
					close(client.send)
					if len(roomClients) == 0 {
						delete(h.rooms, client.RoomID)
					}
					log.Printf("Client unregistered from room %d", client.RoomID)
				}
			}

		case broadcastMsg := <-h.Broadcast:
			roomClients, ok := h.rooms[broadcastMsg.RoomID]
			if ok {
				payload, err := json.Marshal(broadcastMsg.Message)
				if err != nil {
					log.Printf("Failed to marshal broadcast message: %v", err)
					continue
				}
				for client := range roomClients {
					select {
					case client.send <- payload:
					default:
						// If the client's send buffer is full, remove them
						close(client.send)
						delete(roomClients, client)
					}
				}
			}
		}
	}
}
