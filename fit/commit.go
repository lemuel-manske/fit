package fit

func CommitChanges(dir string, message string) (CommitID, error) {
	store := NewCommitStore(dir)

	index, err := LoadIndex(dir)
	if err != nil {
		return "", err
	}

	files := make(map[string]Hash)

	for path, entry := range index.Entries {
		if entry.Delete {
			delete(files, path)

			continue
		}

		files[path] = Hash(entry.Blob)
	}

	commit := Commit{
		Message: message,
		Files:   files,
	}

	commitID, err := store.Put(commit)
	if err != nil {
		return "", err
	}

	WriteHEAD(dir, commitID)

	clear(index.Entries)

	WriteIndex(dir, index)

	return commitID, nil
}
