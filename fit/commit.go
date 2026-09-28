package fit

func CommitChanges(dir string, message string) (CommitID, error) {
	store := NewCommitStore(dir)

	index, err := LoadIndex(dir)
	if err != nil {
		return "", err
	}

	headID, err := ReadHEAD(dir)
	if err != nil {
		return "", err
	}

	files := make(map[string]Hash)

	hasHEAD := !IsEmpty(string(headID))

	if hasHEAD {
		parentCommit, err := store.Get(headID)
		if err != nil {
			return "", err
		}

		for path, hash := range parentCommit.Files {
			if _, ok := files[path]; !ok {
				files[path] = hash
			}
		}
	}

	for path, entry := range index.Entries {
		if entry.Delete {
			delete(files, path)

			continue
		}

		files[path] = Hash(entry.Blob)
	}

	parents := []CommitID{}

	if hasHEAD {
		parents = append(parents, headID)
	}

	commit := Commit{
		Parents: parents,
		Message: message,
		Files:   files,
	}

	commitID, err := store.Put(commit)
	if err != nil {
		return "", err
	}

	err = WriteHEAD(dir, commitID)
	if err != nil {
		return "", err
	}

	clear(index.Entries)

	err = WriteIndex(dir, index)
	if err != nil {
		return "", err
	}

	return commitID, nil
}
