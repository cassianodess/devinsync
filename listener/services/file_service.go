package services

import (
	"fmt"
	"listener/domain/constants"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CheckFiles(filePath string) ([]os.DirEntry, error) {
	return os.ReadDir(filePath)
}

func GetFileContent(path string) []byte {
	if path == "" {
		return nil
	}

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

func GetParsedPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("error while getting home directory: ", err)
	}

	guestWorkspacePath := filepath.Join(homeDir, constants.GUEST_WORKSPACE)
	if err := os.MkdirAll(guestWorkspacePath, 0775); err != nil {
		log.Fatal("error while creating guest workspace: ", err)
	}

	replacer := strings.NewReplacer(guestWorkspacePath, "")
	return replacer.Replace(path)
}
