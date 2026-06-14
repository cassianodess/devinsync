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

	if isFolderCreated {
		fmt.Printf("folder created: %v mode=%v, info=%v", event.IsDir(), event.Mode(), event.FileInfo)
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   nil,
			Type:      types.DirectoryCreatedEvent,
			CreatedAt: &now,
		}

	} else if isFileCreated {
		fmt.Printf("file create: name= %s, path=%s, content=%s, mode=%v, info=%v", event.Name(), event.Path, GetFileContent(event.Path), event.Mode(), event.FileInfo)

		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   &content,
			Type:      types.FileCreatedEvent,
			CreatedAt: &now,
		}

	} else if isFileSaved {
		fmt.Printf("file writed: name= %s, path=%s, content=\n--\n%s\n--\n, mode=%v, info=%v", event.Name(), event.Path, GetFileContent(event.Path), event.Mode(), event.FileInfo)

		content := string(GetFileContent(event.Path))
		eventMessage = &entities.Event{
			Path:      &event.Path,
			Content:   &content,
			Type:      types.FileWritedEvent,
			CreatedAt: &now,
		}
	} else if watcher.Remove == event.Op {
		fmt.Println("removed: ", event)
	} else if watcher.Rename == event.Op {
		fmt.Println("renamed: ", event)
	} else if watcher.Move == event.Op {
		fmt.Println("moved: ", event)
	}

	if eventMessage != nil {
		connection.SendMessage(eventMessage)
	}
}
