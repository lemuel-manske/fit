package cmd

import (
	"os"

	"fit/fit/internal"

	"path/filepath"
)

// Add adds a file or directory to the index.
func Add(dir string, path string) error {
	err := internal.RequireInitialized(dir)
	if err != nil {
		return err
	}

	err = internal.ValidateRepoPath(path)
	if err != nil {
		return err
	}

	absPath := filepath.Join(dir, path)

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		workingTreeFiles, err := workingTreeFiles(absPath)
		if err != nil {
			return err
		}

		for fname := range workingTreeFiles {
			err := Add(dir, filepath.Join(path, fname))
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
