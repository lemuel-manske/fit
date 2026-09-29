package cmd

import (
	"os"

	"fit/fit/internal"

	"path/filepath"
)

// Checkout checks out the specified commit in the given directory.
func Checkout(dir string, targetID internal.CommitID) error {
	commitStore := internal.NewCommitStore(dir)
	blobStore := internal.NewBlobStore(dir)

	target, err := commitStore.Get(targetID)
	if err != nil {
		return err
	}

	current, err := internal.CurrentCommit(dir)
	if err != nil {
		return err
	}

	// 1. validate all blobs BEFORE modifying anything
	contents := make(map[string][]byte)

	for path, blobID := range target.Files {
		content, err := blobStore.Get(blobID)
		if err != nil {
			return err
		}

		contents[path] = content
	}

	// 2. remove files that exist in the current HEAD but not in the target
	for path := range current.Files {
		absPath := filepath.Join(dir, path)

		if _, exists := target.Files[path]; !exists {
			if err := os.Remove(absPath); err != nil {
				return err
			}
		}
	}

	// 3. write snapshot target
	for path, content := range contents {
		absPath := filepath.Join(dir, path)

		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// 4. then move HEAD
	return internal.WriteHEAD(dir, targetID)
}
