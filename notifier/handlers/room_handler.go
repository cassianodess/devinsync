package handlers

import (
	"devinsync/domain/entities"
	"log"

	"github.com/gofiber/contrib/v3/websocket"
)

type RoomHandler struct {
	hub *entities.Hub
}

func NewRoomHandler(hub *entities.Hub) *RoomHandler {
	return &RoomHandler{
		hub: hub,
	}
}

func (this *RoomHandler) ConnectRoom(ctx *websocket.Conn) {
	roomID := ctx.Params("id")
	log.Printf("[%s] Entrou na sala [%s]", ctx.Conn.LocalAddr().String(), roomID)

	client := &entities.Client{
		Conn:   ctx,
		RoomID: roomID,
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
