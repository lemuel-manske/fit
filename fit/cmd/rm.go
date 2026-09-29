package cmd

import (
	"fmt"
	"os"

	"fit/fit/internal"

	"path/filepath"
)

// Rm removes a file or directory from the index.
func Rm(dir string, path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	absPath := filepath.Join(dir, path)

	index, err := internal.LoadIndex(dir)
	if err != nil {
		return err
	}

	// 1. if there are any staged intentions for this path, rm just cancels that intention.
	if _, exists := index.Entries[path]; exists {
		delete(index.Entries, path)

		return internal.WriteIndex(dir, index)
	}

	// 2. if the path is in the HEAD, removing it means:
	// - delete it from the working tree
	// - stage the deletion
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

	// 3. it's neither in the index nor in the HEAD,
	// so if it exists in the working tree, just remove it.
	err = os.Remove(absPath)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	return fmt.Errorf("path %s does not exist", path)
}
