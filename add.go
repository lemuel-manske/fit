package fit

import (
	"fmt"
)

// Add adds a file or directory to the index.
// If the path is a directory, it recursively adds all files in that directory.
func Add(dir string, path string) error {
	index := LoadIndex(dir)

	store := NewFsBlobStore(dir)

	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if IsDir(dir, path) {
		files, err := ListFiles(dir, path)
		if err != nil {
			return err
		}

		for _, file := range files {
			err := Add(dir, file)
			if err != nil {
				return err
			}
		}

		return nil
	}

	fileContent, err := ReadFile(dir, path)
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

	WriteIndex(dir, index)

	return nil
}
