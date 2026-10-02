package cmd

import (
	"fmt"
	"os"

	"fit/fit/internal"

	"path/filepath"
)

// Rm removes a file or directory from the index.
func Rm(dir string, path string) error {
	if err := internal.ValidateRepoPath(path); err != nil {
		return err
	}

	absPath := filepath.Join(dir, path)

	index, err := internal.LoadIndex(dir)
	if err != nil {
		return err
	}

	if _, exists := index.Entries[path]; exists {
		delete(index.Entries, path)

		return internal.WriteIndex(dir, index)
	}

	tracked, err := internal.HEADContains(dir, path)
	if err != nil {
		return err
	}

	if tracked {
		if err = os.Remove(absPath); err != nil {
			return err
		}

		index.Entries[path] = internal.IndexEntry{
			Delete: true,
		}

		return internal.WriteIndex(dir, index)
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
