package entities

import (
	"devinsync/domain/types"
	"time"
)

type Event struct {
	Type      types.EventType `json:"type"`
	Path      *string         `json:"path"`
	Content   *string         `json:"content"`
	CreatedAt *time.Time      `json:"created_at"`
}
