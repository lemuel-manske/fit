package utils

import (
	"os"

	"path/filepath"
)

// WriteFile creates a new file with the given content.
func WriteFile(elements... string) (string, error) {
	content := elements[len(elements)-1]
	fileName := elements[len(elements)-2]

	dir := filepath.Join(elements[:len(elements)-2]...)
	path := filepath.Join(dir, fileName)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return path, nil
}

// ReadFile reads the content of a file and returns it as a byte slice.
func ReadFile(elements... string) ([]byte, error) {
	fileName := elements[len(elements)-1]

	dir := filepath.Join(elements[:len(elements)-1]...)
	path := filepath.Join(dir, fileName)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}
