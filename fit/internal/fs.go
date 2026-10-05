package internal

import (
	"os"

	"path/filepath"
)

// WriteFileAtomic writes data to a temporary file and renames it into place.
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

	err = file.Chmod(perm)
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	err = file.Close()
	file = nil
	if err != nil {
		return err
	}

	err = os.Rename(tempPath, path)
	if err != nil {
		return err
	}

	parent, err := os.Open(dir)
	if err != nil {
		return err
	}

	defer parent.Close()

	return parent.Sync()
}
