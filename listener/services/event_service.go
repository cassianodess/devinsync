package services

import (
	"fmt"
	"github.com/radovskyb/watcher"
)

func HandleEvent(event watcher.Event) {

	if watcher.Create == event.Op {
		if event.IsDir() {
			fmt.Printf("folder created: %v mode=%v, info=%v", event.IsDir(), event.Mode(), event.FileInfo)
		}

		if !event.IsDir() {
			fmt.Printf("file create: name= %s, path=%s, content=%s, mode=%v, info=%v", event.Name(), event.Path, GetFileContent(event.Path), event.Mode(), event.FileInfo)
		}
	} else if watcher.Write == event.Op {
		if !event.IsDir() {
			fmt.Printf("file writed: name= %s, path=%s, content=\n--\n%s\n--\n, mode=%v, info=%v", event.Name(), event.Path, GetFileContent(event.Path), event.Mode(), event.FileInfo)
		}
	} else if watcher.Remove == event.Op {
		fmt.Println("removed: ", event)
	} else if watcher.Rename == event.Op {
		fmt.Println("renamed: ", event)
	} else if watcher.Move == event.Op {
		fmt.Println("moved: ", event)
	}
}
