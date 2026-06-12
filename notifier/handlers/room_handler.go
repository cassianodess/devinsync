package handlers

import (
	"devinsync/domain/entities"
	"log"
	"time"

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

func (this *RoomHandler) Create(ctx *websocket.Conn) {
	log.Println(ctx.Locals("allowed"))

	for {

		if _, _, err := ctx.ReadMessage(); err != nil {
			log.Println("read error: ", err)
			break
		}

		body := map[string]string{
			"Hello": "Olá",
			"World": "Mundo",
		}

		if err := ctx.WriteJSON(body); err != nil {
			log.Println("write json error: ", err)
		}

	}

}

func (this *RoomHandler) Join(ctx *websocket.Conn) {
	log.Println("Entrou no join")
	roomID := ctx.Params("id")

	client := &entities.Client{
		Conn:   ctx,
		RoomID: roomID,
	}

	this.hub.Register(roomID, client)
	defer this.hub.UnRegister(roomID, client)

	for {
		if _, _, err := ctx.ReadMessage(); err != nil {
			break
		}

		path := "./Desktop/text.txt"
		content := "Hello world!"
		now := time.Now().UTC()
		event := entities.Event{
			Type:      entities.FileCreatedEvent,
			Path:      &path,
			Content:   &content,
			CreatedAt: &now,
		}

		this.hub.Broadcast(roomID, event)
	}
}
