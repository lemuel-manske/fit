package cmd

import (
	"os"

	"fit/fit/internal"

	"path/filepath"
)

// Add adds a file or directory to the index.
// If the path is a directory, it recursively adds all files in that directory.
func Add(dir string, path string) error {
	err := internal.ValidateRepoPath(path)
	if err != nil {
		return err
	}

	absPath := filepath.Join(dir, path)

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		files, err := os.ReadDir(absPath)
		if err != nil {
			return err
		}

		for _, file := range files {
			err := Add(dir, filepath.Join(path, file.Name()))
			if err != nil {
				return err
			}
		}

		return nil
	}

	fileContent, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	blobs := internal.NewBlobStore(dir)

	blobID, err := blobs.Put(fileContent)
	if err != nil {
		return err
	}

	index, err := internal.LoadIndex(dir)
	if err != nil {
		return err
	}

	index.Entries[path] = internal.IndexEntry{
		Blob: string(blobID),
	}

	err = internal.WriteIndex(dir, index)
	if err != nil {
		return err
	}

	return nil
}
