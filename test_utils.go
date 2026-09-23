package fit

import (
	"os"

	"path/filepath"
)

func NewFile(dir, filename, content string) string {
	path := filepath.Join(dir, filename)

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}

	return path
}

func WriteFile(path, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
}
