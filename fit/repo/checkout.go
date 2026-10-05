package repo

import (
	"fmt"
	"os"

	"fit/fit/fs"

	"path/filepath"
)

// ForceCheckout checks out the specified commit, discarding any uncommitted changes.
func ForceCheckout(dir string, targetID fs.CommitID) error {
	return doCheckout(dir, targetID, true)
}

// Checkout checks out the specified commit.
func Checkout(dir string, targetID fs.CommitID) error {
	return doCheckout(dir, targetID, false)
}

func doCheckout(dir string, targetID fs.CommitID, force bool) error {
	err := fs.RequireInitialized(dir)
	if err != nil {
		return err
	}

	commitStore := fs.NewCommitStore(dir)
	blobStore := fs.NewBlobStore(dir)

	target, err := commitStore.Get(targetID)
	if err != nil {
		return err
	}

	current, err := fs.CurrentCommit(dir)
	if err != nil {
		return err
	}

	if !force {
		status, err := GetStatus(dir)
		if err != nil {
			return err
		}

		if len(status.StagedModified) > 0 ||
			len(status.StagedDeleted) > 0 ||
			len(status.Modified) > 0 ||
			len(status.Deleted) > 0 {
			return fmt.Errorf("working tree has uncommitted changes")
		}

		for _, path := range status.Untracked {
			if _, exists := target.Files[path]; exists {
				return fmt.Errorf(
					"checkout would overwrite untracked file: %s",
					path,
				)
			}
		}
	}

	contents := make(map[string][]byte, len(target.Files))

	// Validate everything first.
	for path, blobID := range target.Files {
		if err = fs.ValidateRepoPath(path); err != nil {
			return err
		}

		content, err := blobStore.Get(blobID)
		if err != nil {
			return err
		}

		contents[path] = content
	}

	for path := range current.Files {
		if _, exists := target.Files[path]; exists {
			continue
		}

		absPath := filepath.Join(dir, filepath.FromSlash(path))

		if err = os.RemoveAll(absPath); err != nil {
			return err
		}
	}

	for path, content := range contents {
		absPath := filepath.Join(dir, filepath.FromSlash(path))

		// Handles directory -> file transition.
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			if err = os.RemoveAll(absPath); err != nil {
				return err
			}
		}

		if err = os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			return err
		}

		if err = os.WriteFile(absPath, content, 0644); err != nil {
			return err
		}
	}

	return fs.WriteHEAD(dir, targetID)
}
