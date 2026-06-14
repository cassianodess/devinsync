package entities

import (
	"devinsync/domain/types"
	"time"
)

type Event struct {
	Type      types.EventType `json:"type"`
	Path      *string         `json:"path"`
	OldPath   *string         `json:"old_path,omitempty"`
	Content   []byte          `json:"content,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}
