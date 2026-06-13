package main

import (
	"listener/services"
	"log"
	"os"
	"time"

	"github.com/radovskyb/watcher"
)

func main() {
	w := watcher.New()

	//w.SetMaxEvents(1)

	//w.FilterOps(watcher.Rename, watcher.Move)

	go func() {
		for {
			select {
			case event := <-w.Event:
				services.HandleEvent(event)
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	services.CheckFiles()

	if err := w.AddRecursive(os.Args[1]); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(time.Millisecond * 500); err != nil {
		log.Fatalln(err)
	}
}
