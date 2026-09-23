package fit

import (
	"os"

	"path/filepath"
)

// WriteFile creates a new file with the given content.
// The last element of the elements slice is the content,
// the second to last is the filename, and the rest are directories.
// If the directories do not exist, they will be created.
// If the file already exists, it will be overwritten.
func WriteFile(elements... string) (string, error) {
	content := elements[len(elements)-1]
	filename := elements[len(elements)-2]

	dir := filepath.Join(elements[:len(elements)-2]...)
	path := filepath.Join(dir, filename)

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
	filename := elements[len(elements)-1]

	dir := filepath.Join(elements[:len(elements)-1]...)
	path := filepath.Join(dir, filename)

	err := os.Remove(path)
	if err != nil {
		return err
	}

	return nil
}

// ReadFile reads the content of a file and returns it as a byte slice.
func ReadFile(dir, file string) ([]byte, error) {
	path := filepath.Join(dir, file)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ListFiles(dir, path string) ([]string, error) {
	fullPath := filepath.Join(dir, path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, filepath.Join(path, entry.Name()))
	}

	return files, nil
}

func IsDir(dir, path string) bool {
	fullPath := filepath.Join(dir, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}
