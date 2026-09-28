package fit

import (
	"fmt"
)

// Rm removes a file or directory from the index.
func Rm(dir string, path string) error {
	if IsEmpty(path) {
		return fmt.Errorf("path cannot be empty")
	}

	index, err := LoadIndex(dir)
	if err != nil {
		return err
	}

	// 1. If there are any staged intentions for this path, rm just cancels that intention.
	if _, exists := index.Entries[path]; exists {
		delete(index.Entries, path)

		return WriteIndex(dir, index)
	}

	// 2. If the path is in the HEAD, removing it means:
	// - delete it from the working tree
	// - stage the deletion
	tracked, err := HEADContains(dir, path)
	if err != nil {
		return err
	}

	if tracked {
		if err := RemoveFile(dir, path); err != nil {
			return err
		}

		index.Entries[path] = IndexEntry{
			Delete: true,
		}

		return WriteIndex(dir, index)
	}

	// 3. It's neither in the index nor in the HEAD, so if it exists in the working tree, just remove it.
	if FileExists(dir, path) {
		return RemoveFile(dir, path)
	}

	return fmt.Errorf("path %s does not exist", path)
}
