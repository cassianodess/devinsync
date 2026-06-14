package services

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func CheckFiles(filePath string) ([]os.DirEntry, error) {
	return os.ReadDir(filePath)
}

func GetFileContent(path string) []byte {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read file %s %v", path, err)
	}

	return contentBytes
}

func GetIgnoredFiled() []string {
	contentBytes, err := os.ReadFile("./.disignore")
	if err != nil {
		log.Fatal("error while read ignore files")
	}

	var ignoredFiles []string = strings.Split(strings.TrimSpace(string(contentBytes)), "\n")

	return ignoredFiles
}
