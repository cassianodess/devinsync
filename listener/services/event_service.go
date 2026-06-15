package services

import (
	"encoding/json"
	"fmt"
	"listener/domain/constants"
	"listener/domain/entities"
	"listener/domain/types"
	"log"
	"os"
	"path/filepath"
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

func ListenServer(wsConnection *entities.Connector, w *watcher.Watcher, roomID *string) {
	for {

		var eventBody *entities.Event = &entities.Event{}
		if err := wsConnection.Connection.ReadJSON(eventBody); err != nil {
			log.Println("connection failed: ", err)
			CleanUpWorkspace()
			return
		}

		log.Println("event received: ", eventBody.Type)

		switch eventBody.Type {
		case types.RoomCreatedEvent:
			log.Printf("room has been created: [%s]\n", string(eventBody.Content))

		case types.HostDisconnectedEvent:
			log.Println("host has leaf the room")
			CleanUpWorkspace()
			return

		case types.SnapshotCreateEvent:

			var contentJson []entities.SnapshotSyncContent = []entities.SnapshotSyncContent{}

			for filePath, currentFile := range w.WatchedFiles() {
				if currentFile.IsDir() {
					continue
				}

				content := entities.SnapshotSyncContent{
					DirectoryPath: filepath.Dir(filePath),
					FileName:      filepath.Base(filePath),
					FileContet:    GetFileContent(filePath),
				}
				contentJson = append(contentJson, content)
			}

			content, err := json.Marshal(contentJson)
			if err != nil {
				log.Fatalln("error while marshal content json: ", err)
			}

			eventMessage := &entities.Event{
				Content:   []byte(content),
				Type:      types.SnapshotSyncEvent,
				CreatedAt: time.Now().UTC(),
			}
			if err := wsConnection.SendMessage(eventMessage); err != nil {
				log.Fatal("error while send SnapshotSyncEvent: ", err)
			}
		case types.SnapshotSyncEvent:
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal("error while getting home directory: ", err)
			}

			guestWorkspacePath := fmt.Sprintf("%s/%s/%s", homeDir, constants.GUEST_WORKSPACE, *roomID)
			if err := os.MkdirAll(guestWorkspacePath, 0775); err != nil {
				log.Fatal("error while creating guest workspace: ", err)
			}

			var contentJson []entities.SnapshotSyncContent = []entities.SnapshotSyncContent{}

			if err := json.Unmarshal(eventBody.Content, &contentJson); err != nil {
				log.Fatalln("error while unmarshal content json: ", err)
			}

			for _, snapshot := range contentJson {
				snapshotPath := fmt.Sprintf("%s/%s", guestWorkspacePath, snapshot.DirectoryPath)
				if err := os.MkdirAll(snapshotPath, 0775); err != nil {
					log.Fatal("error while creating guest workspace: ", err)
				}

				currentFile, err := os.Create(
					fmt.Sprintf("%s/%s/%s", guestWorkspacePath, snapshot.DirectoryPath, snapshot.FileName),
				)
				if err != nil {
					log.Println("error while creating guest workspace files: ", err)
					continue
				}
				if _, err := currentFile.Write(snapshot.FileContet); err != nil {
					currentFile.Close()
					log.Println("error while write file: ", err)
					continue
				}
				currentFile.Close()

			}

			if err := w.AddRecursive(guestWorkspacePath); err != nil {
				log.Fatalln(err)
			}

			log.Println("guest workspace setup successfully")

		case types.DirectoryCreatedEvent:
			log.Println("directory created event: ", *eventBody.Path)

		case types.FileCreatedEvent:
			log.Println("file created event: ", *eventBody.Path)

		case types.DirectoryWritedEvent:
			log.Println("directory writed: ", *eventBody.Path)

		case types.FileWritedEvent:
			log.Println("file writed: ", *eventBody.Path)

		case types.DirectoryRemovedEvent:
			log.Println("directory removed: ", *eventBody.Path)

		case types.FileRemovedEvent:
			log.Println("file removed: ", *eventBody.Path)

		case types.DirectoryRenamedEvent:
			log.Println("directory renamed: ", *eventBody.Path)

		case types.FileRenamedEvent:
			log.Println("file renamed: ", *eventBody.Path)

		case types.DirectoryMovedEvent:
			log.Printf("directory moved from %s to %s: ", *eventBody.OldPath, *eventBody.Path)

		case types.FileMovedEvent:
			log.Printf("file moved from %s to %s: ", *eventBody.OldPath, *eventBody.Path)
		}

	}

}

func ListenChanges(w *watcher.Watcher, wsConnection *entities.Connector) {
	for {
		select {
		case event := <-w.Event:
			HandleEvent(event, wsConnection)
		case err := <-w.Error:
			log.Fatalln(err)
		case <-w.Closed:
			return
		}
	}
}

func CleanUpWorkspace() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("error while getting home dir: ", err)
	}

	guestWorkspacePath := fmt.Sprintf("%s/%s", homeDir, constants.GUEST_WORKSPACE)
	if err := os.RemoveAll(guestWorkspacePath); err != nil {
		log.Println("error while removing guest workspace: ", err)
	}

	log.Println("guest workspace deleted successfully: ", guestWorkspacePath)
}
