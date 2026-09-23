package fit

import (
	"fmt"
)

// Rm removes a file or directory from the index.
// If the path is a directory, it recursively removes all files in that directory.
func Rm(dir string, path string) error {
	index := LoadIndex(dir)

	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if IsDir(dir, path) {
		files, err := ListFiles(dir, path)
		if err != nil {
			return err
		}

		for _, file := range files {
			err := Rm(dir, file)
			if err != nil {
				return err
			}
		}

		return nil
	}

	_, exists := index.Entries[path]
	if exists {
		delete(index.Entries, path)
		WriteIndex(dir, index)
	} else {
		err := RemoveFile(dir, path)
		if err != nil {
			return err
		}
	}

	return nil
}
