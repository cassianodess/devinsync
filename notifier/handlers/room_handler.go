package handlers

import (
	"devinsync/domain/entities"
	"log"
	"strings"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/google/uuid"
)

type RoomHandler struct {
	hub *entities.Hub
}

func NewRoomHandler(hub *entities.Hub) *RoomHandler {
	return &RoomHandler{
		hub: hub,
	}
}

func (this *RoomHandler) HostRoom(ctx *websocket.Conn) {
	roomID := strings.TrimSpace(strings.ToLower(uuid.New().String()))
	log.Printf("[%s] Criou a sala id: [%s]", ctx.Conn.LocalAddr().String(), roomID)

	client := &entities.Client{
		Conn:   ctx,
		RoomID: roomID,
		IsHost: true,
	}

	this.hub.Register(roomID, client)
	defer this.hub.UnRegister(roomID, client)

	for {
		var eventBody *entities.Event = &entities.Event{}
		if err := ctx.ReadJSON(eventBody); err != nil {
			break
		}

		this.hub.Broadcast(roomID, eventBody)

	}
}

func (this *RoomHandler) JoinRoom(ctx *websocket.Conn) {
	roomID := strings.TrimSpace(strings.ToLower(ctx.Params("id")))
	log.Printf("[%s] Entrou na sala [%s]", ctx.Conn.LocalAddr().String(), roomID)

	client := &entities.Client{
		Conn:   ctx,
		RoomID: roomID,
		IsHost: false,
	}

	this.hub.Register(roomID, client)
	defer this.hub.UnRegister(roomID, client)

	for {
		var eventBody *entities.Event = &entities.Event{}
		if err := ctx.ReadJSON(eventBody); err != nil {
			break
		}

		this.hub.Broadcast(roomID, eventBody)

	}
}
