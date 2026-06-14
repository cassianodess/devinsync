package services

import (
	"fmt"
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
		fmt.Println("directory created: ", event.Path)
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   nil,
			Type:      types.DirectoryCreatedEvent,
			CreatedAt: now,
		}

	} else if isFileCreated {
		fmt.Println("file created", event.Path)

		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   []byte(content),
			Type:      types.FileCreatedEvent,
			CreatedAt: now,
		}

	} else if isFileSaved {
		fmt.Println("file writed", event.Path)

		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   []byte(content),
			Type:      types.FileWritedEvent,
			CreatedAt: now,
		}
	} else if isDirectoryRemoved {
		fmt.Println("directory removed old path:", event.OldPath)

		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   []byte(nil),
			Type:      types.DirectoryRemovedEvent,
			CreatedAt: now,
		}
	} else if isFileRemoved {
		fmt.Println("file removed old path: ", event.OldPath)

		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   []byte(nil),
			Type:      types.FileRemovedEvent,
			CreatedAt: now,
		}

	} else if isDirectoryRenamed {

		fmt.Println("directory renamed old path:", event.OldPath)
		eventMessage = &entities.Event{
			Path:      &event.Path,
			OldPath:   &event.OldPath,
			Content:   []byte(nil),
			Type:      types.DirectoryRenamedEvent,
			CreatedAt: now,
		}

	} else if isFileRenamed {
		fmt.Println("file removed old path: ", event.OldPath)

		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			OldPath:   &event.OldPath,
			Content:   []byte(content),
			Type:      types.FileRenamedEvent,
			CreatedAt: now,
		}
	} else if isDirectoryMoved {

		fmt.Printf("directory moved from %s to %s: ", event.OldPath, event.Path)
		eventMessage = &entities.Event{
			Path:      &event.Path,
			OldPath:   &event.OldPath,
			Content:   []byte(nil),
			Type:      types.DirectoryMovedEvent,
			CreatedAt: now,
		}
	} else if isFileMoved {
		fmt.Printf("file moved from %s to %s: ", event.OldPath, event.Path)
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
