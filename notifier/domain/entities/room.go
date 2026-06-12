package entities

import (
	"sync"
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
		this.clients[roomID] = make(map[*Client]bool)
	}

	this.clients[roomID][client] = true
}

func (this *Hub) UnRegister(roomID string, client *Client) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if clients, ok := this.clients[roomID]; ok {
		delete(clients, client)

		if len(clients) == 0 {
			delete(this.clients, roomID)
		}
	}
}

func (this *Hub) Broadcast(roomID string, payload any) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	clients := this.clients[roomID]

	for client := range clients {
		client.Conn.WriteJSON(payload)
	}
}
