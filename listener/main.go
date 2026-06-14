package main

import (
	"flag"
	"listener/domain/entities"
	"listener/services"
	"log"
	"time"

	"github.com/radovskyb/watcher"
)

func main() {

	userPath := flag.String("path", "", "./path/to/dir")
	flag.Parse()
	if *userPath == "" {
		log.Fatal("missing required flag: -path")
	}

	w := watcher.New()

	wsConnection, wsConnectionErr := entities.NewConnector("ws://localhost:8080/ws/room/123")
	if wsConnectionErr != nil {
		log.Fatal("error while conenction to server: ", wsConnectionErr)
	}

	go ListenServer(wsConnection)
	go ListenChanges(w, wsConnection)

	services.CheckFiles()

	if err := w.AddRecursive(*userPath); err != nil {
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
			break
		}

		log.Printf("event received:\nevent=<%s>\ncontent=<%s>\npath=<%s>\ncreated_at=<%s>", eventBody.Type, string(*eventBody.Content), string(*eventBody.Path), eventBody.CreatedAt)

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
