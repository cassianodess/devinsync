package main

import (
	"flag"
	"fmt"
	"listener/domain/entities"
	"listener/domain/types"
	"listener/services"
	"log"
	"os"
	"os/signal"
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
			services.CleanUpWorkspace()
		}
		w.Close()
		wsConnection.Connection.Close()
		os.Exit(0)
	}()

	syncManager := entities.NewSyncManager()
	go services.ListenServer(wsConnection, w, syncManager)
	go services.ListenChanges(w, wsConnection, isHost, syncManager)

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
