package services

import (
	"listener/domain/entities"
	"listener/domain/types"
	"time"

	"github.com/radovskyb/watcher"
)

func HandleEvent(event watcher.Event, connection *entities.Connector) {
	var eventMessage *entities.Event = nil

	now := time.Now().UTC()

	isFolderCreated := (watcher.Create == event.Op) && event.IsDir()
	isFileCreated := (watcher.Create == event.Op) && !event.IsDir()
	isFileSaved := (watcher.Write == event.Op) && !event.IsDir()
	isDirectoryRemoved := (watcher.Remove == event.Op) && event.IsDir()
	isFileRemoved := (watcher.Remove == event.Op) && !event.IsDir()
	isDirectoryRenamed := (watcher.Rename == event.Op) && event.IsDir()
	isFileRenamed := (watcher.Rename == event.Op) && !event.IsDir()
	isDirectoryMoved := (watcher.Move == event.Op) && event.IsDir()
	isFileMoved := (watcher.Move == event.Op) && !event.IsDir()

	if isFolderCreated {
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   nil,
			Type:      types.DirectoryCreatedEvent,
			CreatedAt: now,
		}

	} else if isFileCreated {
		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   []byte(content),
			Type:      types.FileCreatedEvent,
			CreatedAt: now,
		}

	} else if isFileSaved {
		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   []byte(content),
			Type:      types.FileWritedEvent,
			CreatedAt: now,
		}
	} else if isDirectoryRemoved {
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   []byte(nil),
			Type:      types.DirectoryRemovedEvent,
			CreatedAt: now,
		}
	} else if isFileRemoved {
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   []byte(nil),
			Type:      types.FileRemovedEvent,
			CreatedAt: now,
		}

	} else if isDirectoryRenamed {
		eventMessage = &entities.Event{
			Path:      &event.Path,
			OldPath:   &event.OldPath,
			Content:   []byte(nil),
			Type:      types.DirectoryRenamedEvent,
			CreatedAt: now,
		}

	} else if isFileRenamed {
		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			OldPath:   &event.OldPath,
			Content:   []byte(content),
			Type:      types.FileRenamedEvent,
			CreatedAt: now,
		}
	} else if isDirectoryMoved {
		eventMessage = &entities.Event{
			Path:      &event.Path,
			OldPath:   &event.OldPath,
			Content:   []byte(nil),
			Type:      types.DirectoryMovedEvent,
			CreatedAt: now,
		}
	} else if isFileMoved {
		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			OldPath:   &event.OldPath,
			Content:   []byte(content),
			Type:      types.FileMovedEvent,
			CreatedAt: now,
		}
	}

	if eventMessage != nil {
		connection.SendMessage(eventMessage)
	}
}
