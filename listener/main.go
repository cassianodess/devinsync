package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"listener/domain/entities"
	"listener/domain/types"
	"listener/services"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/radovskyb/watcher"
)

func main() {

	target := flag.String("target", "", "[host|guest]")
	roomID := flag.String("room", "", "room_id")
	workspace := flag.String("path", "", "./path/to/dir")
	flag.Parse()

	isInvalidTarget := strings.ToLower(strings.TrimSpace(*target)) != string(types.TargetHost) && strings.ToLower(strings.TrimSpace(*target)) != string(types.TargetGuest)
	if isInvalidTarget {
		flag.Usage()
		log.Fatal("invalid target [host|guest]")
	}

	var isHost bool = strings.ToLower(strings.TrimSpace(*target)) == string(types.TargetHost)
	if isHost && *workspace == "" {
		flag.Usage()
		log.Fatal("missing path")
	}

	w := watcher.New()
	defer w.Close()

	//ignored := services.GetIgnoredFiled()
	//if err := w.Ignore(ignored...); err != nil {
	//	log.Fatal("error while ignoring files in .disignore")
	//}
	//TODO: verify by .disignore
	w.IgnoreHiddenFiles(true)

	var baseURL string = os.Getenv("SERVER_URL")
	var targetURL string = "host"

	if !isHost {
		if strings.ToLower(strings.TrimSpace(*roomID)) == "" {
			flag.Usage()
			log.Fatal("missing flag room")
		}

		targetURL = fmt.Sprintf("join/%s", strings.TrimSpace(strings.ToLower(*roomID)))
	}

	wsConnection, wsConnectionErr := entities.NewConnector(fmt.Sprintf("%s/%s", baseURL, targetURL))
	if wsConnectionErr != nil {
		log.Fatal("error while conenction to server: ", wsConnectionErr)
	}
	defer wsConnection.Connection.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(
		signals,
		os.Interrupt,
		syscall.SIGTERM,
	)

	go func() {
		<-signals
		log.Println("shutting down...")
		if !isHost {
			CleanUpWorkspace()
		}
		w.Close()
		wsConnection.Connection.Close()
		os.Exit(0)
	}()

	go ListenServer(wsConnection, w, roomID)
	go ListenChanges(w, wsConnection)

	if isHost {
		_, err := services.CheckFiles(*workspace)
		if err != nil {
			log.Fatal("error while finding directories")
		}

		if err := w.AddRecursive(*workspace); err != nil {
			log.Fatalln(err)
		}
	}

	if err := w.Start(time.Millisecond * 500); err != nil {
		log.Fatalln(err)
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
					FileContet:    services.GetFileContent(filePath),
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

			guestWorkspacePath := fmt.Sprintf("%s/.devinsync/%s", homeDir, *roomID)
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
			services.HandleEvent(event, wsConnection)
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

	guestWorkspacePath := fmt.Sprintf("%s/.devinsync", homeDir)
	if err := os.RemoveAll(guestWorkspacePath); err != nil {
		log.Println("error while removing guest workspace: ", err)
	}

	log.Println("guest workspace deleted successfully: ", guestWorkspacePath)
}
