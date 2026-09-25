package fit

import (
	"fmt"
)

// Add adds a file or directory to the index.
// If the path is a directory, it recursively adds all files in that directory.
func Add(fitDir string, path string) error {
	index := LoadIndex(fitDir)

	store := NewFsBlobStore(fitDir)

	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if IsDir(fitDir, path) {
		files, err := ListFiles(fitDir, path)
		if err != nil {
			return err
		}

		for _, file := range files {
			err := Add(fitDir, file)
			if err != nil {
				return err
			}
		}

		return nil
	}

	fileContent, err := ReadFile(fitDir, path)
	if err != nil {
		return err
	}

	blobID, err := store.Put(fileContent)
	if err != nil {
		return err
	}

	index.Entries[path] = IndexEntry{
		Blob: string(blobID),
	}

	WriteIndex(fitDir, index)

	return nil
}
