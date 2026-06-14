package services

import (
	"fmt"
	"os"
)

func CheckFiles(filePath string) ([]os.DirEntry, error) {
	return os.ReadDir(filePath)
}

func GetFileContent(path string) string {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read file %s %v", path, err)
	}

	return string(contentBytes)
}
