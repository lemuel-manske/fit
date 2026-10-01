package cmd

import (
	"os"

	"fit/fit/internal"

	"path/filepath"
)

// Checkout checks out the specified commit.
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

	contents := make(map[string][]byte)

	for path, blobID := range target.Files {
		content, err := blobStore.Get(blobID)
		if err != nil {
			return err
		}

		contents[path] = content
	}

	for path := range current.Files {
		absPath := filepath.Join(dir, path)

		if _, exists := target.Files[path]; !exists {
			if err := os.Remove(absPath); err != nil {
				return err
			}
		}
	}

	for path, content := range contents {
		absPath := filepath.Join(dir, path)

		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return internal.WriteHEAD(dir, targetID)
}
