package fs

import (
	"os"

	"path/filepath"
)

// WriteFileAtomic writes files atomically.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	file, err := os.CreateTemp(dir, ".fit-write-*")
	if err != nil {
		return err
	}

	tempPath := file.Name()
	defer os.Remove(tempPath)

	// close on early errors; the explicit close below must succeed before rename.
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	if err = file.Chmod(perm); err != nil {
		return err
	}

	if err = file.Sync(); err != nil {
		return err
	}

	err = file.Close()
	file = nil
	if err != nil {
		return err
	}

	if err = os.Rename(tempPath, path); err != nil {
		return err
	}

	parent, err := os.Open(dir)
	if err != nil {
		return err
	}

	defer parent.Close()

	return parent.Sync()
}

// WriteFile creates a new file with the given content.
func WriteFile(elements ...string) (string, error) {
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
func ReadFile(elements ...string) ([]byte, error) {
	fileName := elements[len(elements)-1]

	dir := filepath.Join(elements[:len(elements)-1]...)
	path := filepath.Join(dir, fileName)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}
