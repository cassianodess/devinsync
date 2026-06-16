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

func GetParsedPath(path string, isHost bool) *string {
	if strings.TrimSpace(path) == "" {
		return nil
	}

	if isHost {
		return &path
	}

	guestWorkspacePath := GetGuestWorkspacePath()
	replacer := strings.NewReplacer(*guestWorkspacePath, "")
	source := replacer.Replace(path)
	return &source
}

func GetGuestWorkspacePath() *string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("error while getting home directory: ", err)
		return nil
	}

	guestWorkspacePath := filepath.Join(homeDir, constants.GUEST_WORKSPACE)
	return &guestWorkspacePath

}

func CreateDirectory(path string) error {
	return os.MkdirAll(path, 0775)
}

func CreateFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

func DeleteDirectoryOrFile(path string) error {
	return os.RemoveAll(path)
}

func RenameOrMoveDirectoryOrFile(oldPath string, newPath string) error {
	return os.Rename(oldPath, newPath)
}
