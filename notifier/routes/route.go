package routes

import (
	"devinsync/domain/entities"
	"devinsync/handlers"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

func Init(app *fiber.App) {

	hub := entities.NewHub()

	globalGroup := app.Group("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	roomHandler := handlers.NewRoomHandler(hub)
	roomGroup := globalGroup.Group("/room")
	roomGroup.Get("/:id", websocket.New(roomHandler.ConnectRoom))

}
