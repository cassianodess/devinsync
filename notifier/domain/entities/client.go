package entities

import (
	"time"

	"github.com/gofiber/contrib/v3/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	RoomID string
}

type Event struct {
	Type      EventType  `json:"type"`
	Path      *string    `json:"path"`
	Content   *string    `json:"content"`
	CreatedAt *time.Time `json:"created_at"`
}

type EventType string

const (
	FileCreatedEvent EventType = "FILE_CREATED"
)
