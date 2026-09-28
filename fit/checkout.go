package fit

// Checkout checks out the specified commit in the given directory.
func Checkout(dir string, targetID CommitID) error {
	commitStore := NewCommitStore(dir)
	blobStore := NewBlobStore(dir)

	target, err := commitStore.Get(targetID)
	if err != nil {
		return err
	}

	current, err := CurrentCommit(dir)
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
		if _, exists := target.Files[path]; !exists {
			if err := RemoveFile(dir, path); err != nil {
				return err
			}
		}
	}

	// 3. write snapshot target
	for path, content := range contents {
		if _, err := WriteFile(dir, path, string(content)); err != nil {
			return err
		}
	}

	// 4. then move HEAD
	return WriteHEAD(dir, targetID)
}
