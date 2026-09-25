package fit

import (
	"fmt"
)

// Rm removes a file or directory from the index.
// If the path is a directory, it recursively removes all files in that directory.
// If the file is staged, it will be removed from the index.
// If the file is not staged, it will be removed from the working directory.
func Rm(fitDir string, path string) error {
	index := LoadIndex(fitDir)

	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if IsDir(fitDir, path) {
		files, err := ListFiles(fitDir, path)
		if err != nil {
			return err
		}

		for _, file := range files {
			err := Rm(fitDir, file)
			if err != nil {
				return err
			}
		}

		return nil
	}

	_, exists := index.Entries[path]
	if exists {
		delete(index.Entries, path)
		WriteIndex(fitDir, index)
	} else {
		err := RemoveFile(fitDir, path)
		if err != nil {
			return err
		}
	}

	return nil
}
