package entities

import (
	"devinsync/domain/types"
	"log"
	"sync"
	"time"
)

type Hub struct {
	mutex   sync.RWMutex
	clients map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]bool),
	}
}

func (this *Hub) Register(roomID string, client *Client) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.clients[roomID] == nil {
		if !client.IsHost {
			log.Println("only host can create rooms")
			client.Conn.Close()
			return
		}
		this.clients[roomID] = make(map[*Client]bool)
	}

	this.clients[roomID][client] = true

	if client.IsHost {
		roomCreatedEvent := &Event{
			Type:      types.RoomCreatedEvent,
			Content:   []byte(roomID),
			CreatedAt: time.Now().UTC(),
		}
		if err := client.Conn.WriteJSON(roomCreatedEvent); err != nil {
			log.Println("error while write room created")
		}
	} else {
		snapshotEvent := &Event{
			Type:      types.SnapshotCreateEvent,
			Content:   []byte(nil),
			CreatedAt: time.Now().UTC(),
		}
		this.SendToHost(client, snapshotEvent)
	}
}

func (this *Hub) UnRegister(roomID string, client *Client) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if client.IsHost {
		log.Printf("host [%s] has leafted room [%s]", client.Conn.LocalAddr().String(), roomID)
	}
	if clients, ok := this.clients[roomID]; ok {
		delete(clients, client)

		if len(clients) == 0 {
			delete(this.clients, roomID)
		}
	}
}

func (this *Hub) Broadcast(client *Client, payload any) {
	clients := this.clients[client.RoomID]

	for otherClient := range clients {
		if otherClient.Conn != client.Conn {
			otherClient.Conn.WriteJSON(payload)
		}
	}
}

func (this *Hub) SendToHost(client *Client, payload any) {
	clients := this.clients[client.RoomID]

	for otherClient := range clients {
		if otherClient.IsHost {
			otherClient.Conn.WriteJSON(payload)
		}
	}
}
