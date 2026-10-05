package repo

import (
	"fmt"
	"os"

	"fit/fit/fs"

	"path/filepath"
)

// Remove removes a file or directory from the index.
func Remove(dir string, path string) error {
	err := fs.RequireInitialized(dir)
	if err != nil {
		return err
	}

	if err := fs.ValidateRepoPath(path); err != nil {
		return err
	}

	absPath := filepath.Join(dir, path)

	index, err := fs.LoadIndex(dir)
	if err != nil {
		return err
	}

	if index.Entries[path].Conflict {
		if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
			return err
		}

		index.Entries[path] = fs.IndexEntry{Delete: true}

		return fs.WriteIndex(dir, index)
	}

	if _, exists := index.Entries[path]; exists {
		delete(index.Entries, path)

		return fs.WriteIndex(dir, index)
	}

	tracked, err := fs.HEADContains(dir, path)
	if err != nil {
		return err
	}

	if tracked {
		if err = os.Remove(absPath); err != nil {
			return err
		}

		index.Entries[path] = fs.IndexEntry{
			Delete: true,
		}

		return fs.WriteIndex(dir, index)
	}

	err = os.Remove(absPath)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	return fmt.Errorf("path %s does not exist", path)
}
