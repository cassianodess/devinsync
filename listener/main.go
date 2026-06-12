package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

func main() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	CheckFiles()
	go HandleChanges(watcher)
	log.Println("Listening...")

	if err = watcher.Add(os.Args[1]); err != nil {
		log.Fatal(err)
	}

	<-make(chan struct{})

}

func CheckFiles() {
	if len(os.Args) < 2 {
		log.Fatalf("You must pass an path argument")
	}

	filePath := os.Args[1]
	if _, err := os.ReadDir(filePath); err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}
}

func HandleChanges(watcher *fsnotify.Watcher) {
	for {
		select {

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			//log.Println("bateu event:", event, "\n")
			path := event.Name

			if event.Has(fsnotify.Create) {
				log.Println("file created: ", path)
				fmt.Println("content:")
				fmt.Println(GetFileContent(path))
			} else if event.Has(fsnotify.Write) {
				log.Println("file saved: ", path)
				fmt.Println("content: ", GetFileContent(path))
			} else if event.Has(fsnotify.Rename) {
				log.Println("file renamed", path)
			} else if event.Has(fsnotify.Remove) {
				log.Println("file removed", event)
			} else {
				//noting
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error: ", err)
		}

	}
}

func GetFileContent(path string) string {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read file %s %v", path, err)
	}

	return string(contentBytes)
}
