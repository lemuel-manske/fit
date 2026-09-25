package fit

import (
	"os"

	"path/filepath"
)

// WriteFile creates a new file with the given content.
// The last element of the elements slice is the content,
// the second to last is the fileName, and the rest are directories.
// If the directories do not exist, they will be created.
// If the file already exists, it will be overwritten.
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

// RemoveFile removes a file from the filesystem.
func RemoveFile(elements... string) error {
	fileName := elements[len(elements)-1]

	dir := filepath.Join(elements[:len(elements)-1]...)
	path := filepath.Join(dir, fileName)

	err := os.Remove(path)
	if err != nil {
		return err
	}

	return nil
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

// ListFiles lists all files in a directory and returns their paths relative to the given directory.
func ListFiles(elements... string) ([]string, error) {
	fileName := elements[len(elements)-1]

	dir := filepath.Join(elements[:len(elements)-1]...)
	path := filepath.Join(dir, fileName)

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, filepath.Join(fileName, entry.Name()))
	}

	return files, nil
}

// IsDir checks if the given path is a directory.
func IsDir(elements... string) bool {
	fileName := elements[len(elements)-1]

	dir := filepath.Join(elements[:len(elements)-1]...)
	path := filepath.Join(dir, fileName)

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
