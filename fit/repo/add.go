package repo

import (
	"os"

	"fit/fit/fs"

	"path/filepath"
)

// Add adds a file or directory to the index.
func Add(dir string, path string) error {
	err := fs.RequireInitialized(dir)
	if err != nil {
		return err
	}

	if err = fs.ValidateRepoPath(path); err != nil {
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

			if err = Add(dir, filepath.Join(path, fname)); err != nil {
				return err
			}
		}

		return nil
	}

	fileContent, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	blobs := fs.NewBlobStore(dir)

	blobID, err := blobs.Put(fileContent)
	if err != nil {
		return err
	}

	index, err := fs.LoadIndex(dir)
	if err != nil {
		return err
	}

	index.Entries[path] = fs.IndexEntry{
		Blob: string(blobID),
	}

	if err = fs.WriteIndex(dir, index); err != nil {
		return err
	}

	return nil
}
