package services

import (
	"encoding/json"
	"listener/domain/entities"
	"listener/domain/types"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
)

func HandleEvent(event watcher.Event, connection *entities.Connector, isHost bool, syncManager *entities.SyncManager) {
	if syncManager.ShouldIgnore(*GetParsedPath(event.Path, isHost)) {
		return
	}

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

	var path *string = GetParsedPath(event.Path, isHost)
	var oldPath *string = GetParsedPath(event.OldPath, isHost)
	var content []byte = nil

	if !event.IsDir() {
		content = GetFileContent(event.Path)
	}

	if isFolderCreated {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			Type:      types.DirectoryCreatedEvent,
			CreatedAt: now,
		}

	} else if isFileCreated {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			Type:      types.FileCreatedEvent,
			CreatedAt: now,
		}

	} else if isFileSaved {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			Type:      types.FileWritedEvent,
			CreatedAt: now,
		}
	} else if isDirectoryRemoved {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			Type:      types.DirectoryRemovedEvent,
			CreatedAt: now,
		}
	} else if isFileRemoved {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			Type:      types.FileRemovedEvent,
			CreatedAt: now,
		}

	} else if isDirectoryRenamed {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			OldPath:   oldPath,
			Type:      types.DirectoryRenamedEvent,
			CreatedAt: now,
		}

	} else if isFileRenamed {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			OldPath:   oldPath,
			Type:      types.FileRenamedEvent,
			CreatedAt: now,
		}
	} else if isDirectoryMoved {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			OldPath:   oldPath,
			Type:      types.DirectoryMovedEvent,
			CreatedAt: now,
		}
	} else if isFileMoved {
		eventMessage = &entities.Event{
			Path:      path,
			Content:   content,
			OldPath:   oldPath,
			Type:      types.FileMovedEvent,
			CreatedAt: now,
		}
	}

	if eventMessage != nil {
		connection.SendMessage(eventMessage)
	}
}

func ListenServer(wsConnection *entities.Connector, w *watcher.Watcher, isHost bool, syncManager *entities.SyncManager) {
	for {

		var eventBody *entities.Event = &entities.Event{}
		if err := wsConnection.Connection.ReadJSON(eventBody); err != nil {
			log.Println("connection failed: ", err)
			CleanUpWorkspace()
			return
		}

		filePath := "N/A"
		if eventBody.Path != nil {
			filePath = *eventBody.Path
		}

		oldPath := "N/A"
		if eventBody.OldPath != nil {
			oldPath = *eventBody.OldPath
		}

		log.Printf("event received: %s", eventBody.Type)
		log.Printf("Path: %s", filePath)
		log.Printf("OldPath: %s", oldPath)

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
			guestWorkspacePath := *GetGuestWorkspacePath()

			var contentJson []entities.SnapshotSyncContent = []entities.SnapshotSyncContent{}
			if err := json.Unmarshal(eventBody.Content, &contentJson); err != nil {
				log.Fatalln("error while unmarshal content json: ", err)
			}

			for _, snapshot := range contentJson {
				snapshotPath := filepath.Join(guestWorkspacePath, snapshot.DirectoryPath)
				if err := CreateDirectory(snapshotPath); err != nil {
					log.Fatal("error while creating guest workspace: ", err)
					return
				}

				cleanedFilePath := filepath.Join(guestWorkspacePath, snapshot.DirectoryPath, snapshot.FileName)
				currentFile, err := os.Create(
					cleanedFilePath,
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

			ShowInstructions(guestWorkspacePath)

		case types.DirectoryCreatedEvent:
			syncManager.Ignore(*GetParsedPath(*eventBody.Path, isHost))

			if !isHost {
				*eventBody.Path = filepath.Join(*GetGuestWorkspacePath(), *eventBody.Path)
			}

			if err := CreateDirectory(*eventBody.Path); err != nil {
				log.Println("error while create folder", err)
				continue
			}

			log.Println("directory created event: ", *eventBody.Path)

		case types.FileWritedEvent, types.FileCreatedEvent:
			syncManager.Ignore(*GetParsedPath(*eventBody.Path, isHost))

			if !isHost {
				*eventBody.Path = filepath.Join(*GetGuestWorkspacePath(), *eventBody.Path)
			}

			if err := CreateFile(*eventBody.Path, eventBody.Content); err != nil {
				log.Println("error while write/create file", err)
				continue
			}

			log.Println("file writed/created: ", *eventBody.Path)

		case types.DirectoryRemovedEvent, types.FileRemovedEvent:
			syncManager.Ignore(*GetParsedPath(*eventBody.Path, isHost))
			if !isHost {
				*eventBody.Path = filepath.Join(*GetGuestWorkspacePath(), *eventBody.Path)
			}

			if err := DeleteDirectoryOrFile(*eventBody.Path); err != nil {
				log.Println("error while delete directory/file", err)
				continue
			}

			log.Println("directory deleted: ", *eventBody.Path)

		case types.DirectoryMovedEvent, types.DirectoryRenamedEvent:
			syncManager.Ignore(*GetParsedPath(*eventBody.Path, isHost))
			syncManager.Ignore(*GetParsedPath(*eventBody.OldPath, isHost))
			if !isHost {
				*eventBody.Path = filepath.Join(*GetGuestWorkspacePath(), *eventBody.Path)
				*eventBody.OldPath = filepath.Join(*GetGuestWorkspacePath(), *eventBody.OldPath)
			}

			log.Println("old: ", *eventBody.OldPath)
			log.Println("new: ", *eventBody.Path)

			if err := RenameOrMoveDirectoryOrFile(*eventBody.OldPath, *eventBody.Path); err != nil {
				log.Println("error while move/rename directory", err)
				continue
			}

			log.Printf("directory moved/renamed: from %s to %s", *eventBody.OldPath, *eventBody.Path)

		case types.FileMovedEvent, types.FileRenamedEvent:
			syncManager.Ignore(*GetParsedPath(*eventBody.Path, isHost))
			if !isHost {
				*eventBody.Path = filepath.Join(*GetGuestWorkspacePath(), *eventBody.Path)
				*eventBody.OldPath = filepath.Join(*GetGuestWorkspacePath(), *eventBody.OldPath)
			}

			if err := RenameOrMoveDirectoryOrFile(*eventBody.OldPath, *eventBody.Path); err != nil {
				log.Println("error while move/rename file", err)
				continue
			}

			log.Printf("file moved/renamed: from %s to %s", *eventBody.OldPath, *eventBody.Path)
		}

		log.Println()
	}

}

func ListenChanges(w *watcher.Watcher, wsConnection *entities.Connector, isHost bool, syncManager *entities.SyncManager) {
	for {
		select {
		case event := <-w.Event:
			HandleEvent(event, wsConnection, isHost, syncManager)
		case err := <-w.Error:
			log.Fatalln(err)
		case <-w.Closed:
			return
		}
	}
}

func CleanUpWorkspace() {
	guestWorkspacePath := *GetGuestWorkspacePath()
	if err := os.RemoveAll(guestWorkspacePath); err != nil {
		log.Println("error while removing guest workspace: ", err)
	}

	log.Println("guest workspace deleted successfully: ", guestWorkspacePath)
}

func ShowInstructions(path string) {
	log.Println()
	log.Println("**********************************")
	log.Println("WORKSPACE SETUP SUCCESSFULLY")
	log.Println("**********************************")
	log.Println()
	log.Println(">>> FOLLOW THE INSTRUCTIONS TO OPEN THIS WORKSPACE <<<")
	log.Println("> Open your terminal")
	log.Printf("> Open your text editor in [cd %s]", path)
	log.Println("> And then start to editing")
	log.Println()
}
