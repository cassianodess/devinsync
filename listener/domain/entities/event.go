package entities

import (
	"listener/domain/types"
	"time"
)

type Event struct {
	Type      types.EventType `json:"type"`
	Path      *string         `json:"path"`
	OldPath   *string         `json:"old_path,omitempty"`
	Content   []byte          `json:"content,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type SnapshotSyncContent struct {
	DirectoryPath string `json:"directory_path"`
	FileName      string `json:"file_name"`
	FileContet    []byte `json:"file_content"`
}
