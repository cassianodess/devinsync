package services

import (
	"fmt"
	"log"
	"os"
)

func CheckFiles(filePath string) {

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
