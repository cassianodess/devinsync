package main

import (
	"flag"
	"fmt"
	"listener/domain/entities"
	"listener/domain/types"
	"listener/services"
	"log"
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

	go ListenServer(wsConnection)
	go ListenChanges(w, wsConnection)

	//if !isHost {
	//	*workspace = fmt.Sprintf("$HOME/.devinsync/%s", *roomID)
	//}

	services.CheckFiles(*workspace)

	if err := w.AddRecursive(*workspace); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(time.Millisecond * 500); err != nil {
		log.Fatalln(err)
	}
}

func ListenServer(wsConnection *entities.Connector) {
	for {

		var eventBody *entities.Event = &entities.Event{}
		if err := wsConnection.Connection.ReadJSON(eventBody); err != nil {
			log.Fatal("connection failed: ", err)
		}

		log.Printf("event received:\nevent=<%s>\ncontent=<%s>\npath=<%s>\ncreated_at=<%s>", eventBody.Type, string(eventBody.Content), eventBody.Path, eventBody.CreatedAt)

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
