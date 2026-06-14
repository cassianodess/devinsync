package main

import (
	"flag"
	"fmt"
	"listener/domain/entities"
	"listener/domain/types"
	"listener/services"
	"log"
	"os"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
)

func main() {

	target := flag.String("target", "", "[host|guest]")
	roomID := flag.String("room", "", "room_id")
	workspace := flag.String("path", "", "./path/to/dir")
	flag.Parse()

	if strings.ToLower(strings.TrimSpace(*target)) != string(types.TargetHost) && strings.ToLower(strings.TrimSpace(*target)) != string(types.TargetGuest) {
		flag.Usage()
		log.Fatal("invalid target [host|guest]")
	}

	var isHost bool = strings.ToLower(strings.TrimSpace(*target)) == string(types.TargetHost)

	if isHost && *workspace == "" {
		flag.Usage()
		log.Fatal("missing path")
	}

	w := watcher.New()

	//ignored := services.GetIgnoredFiled()
	//if err := w.Ignore(ignored...); err != nil {
	//	log.Fatal("error while ignoring files in .disignore")
	//}
	//TODO: verify by .disignore
	w.IgnoreHiddenFiles(true)

	var baseURL string = "ws://localhost:8080/ws"
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

	go ListenServer(wsConnection, w)
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

func ListenServer(wsConnection *entities.Connector, w *watcher.Watcher) {
	for {

		var eventBody *entities.Event = &entities.Event{}
		if err := wsConnection.Connection.ReadJSON(eventBody); err != nil {
			log.Fatal("connection failed: ", err)
		}

		path := ""
		if eventBody.Path != nil {
			path = *eventBody.Path
		}

		log.Printf("event received:\nevent=<%s>\ncontent=<%s>\npath=<%s>\ncreated_at=<%s>", eventBody.Type, string(eventBody.Content), path, eventBody.CreatedAt)

		switch eventBody.Type {
		case types.SnapshotCreateEvent:
			fmt.Println("SnapshotCreateEvent")
			eventMessage := &entities.Event{
				Content:   []byte(nil),
				Type:      types.SnapshotSyncEvent,
				CreatedAt: time.Now().UTC(),
			}
			if err := wsConnection.SendMessage(eventMessage); err != nil {
				log.Fatal("error while send SnapshotSyncEvent: ", err)
			}
		case types.SnapshotSyncEvent:
			fmt.Println("SnapshotSyncEvent")
			//CRIA O DIRETORIO BASE
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal("error while getting home directory: ", err)
			}

			guestWorkspacePath := fmt.Sprintf("%s/.devinsync/123", homeDir)
			if err := os.MkdirAll(guestWorkspacePath, 0775); err != nil {
				log.Fatal("error while creating guest workspace: ", err)
			}

			//Cria os demais diretorios e arquivos

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
