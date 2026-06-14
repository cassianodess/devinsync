package entities

import (
	"github.com/gofiber/contrib/v3/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	RoomID string
	IsHost bool
}
