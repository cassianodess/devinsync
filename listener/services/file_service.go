package services

import (
	"fmt"
	"log"
	"os"
)


func CheckFiles() {
	if len(os.Args) < 2 {
		log.Fatalf("You must pass an path argument")
	}

	filePath := os.Args[1]
	if _, err := os.ReadDir(filePath); err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	} else {
		log.Println("Listening to: ", filePath)
	}
}

func GetFileContent(path string) string {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read file %s %v", path, err)
	}

	return string(contentBytes)
}
