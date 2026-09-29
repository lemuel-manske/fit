package cmd

import (
	"fit/fit/internal"
)

func CommitChanges(dir string, message string) (internal.CommitID, error) {
	store := internal.NewCommitStore(dir)

	index, err := internal.LoadIndex(dir)
	if err != nil {
		return "", err
	}

	headID, err := internal.ReadHEAD(dir)
	if err != nil {
		return "", err
	}

	files := make(map[string]internal.Hash)

	hasHEAD := headID != ""

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

		files[path] = internal.Hash(entry.Blob)
	}

	parents := []internal.CommitID{}

	if hasHEAD {
		parents = append(parents, headID)
	}

	commit := internal.Commit{
		Parents: parents,
		Message: message,
		Files:   files,
	}

	commitID, err := store.Put(commit)
	if err != nil {
		return "", err
	}

	err = internal.WriteHEAD(dir, commitID)
	if err != nil {
		return "", err
	}

	clear(index.Entries)

	err = internal.WriteIndex(dir, index)
	if err != nil {
		return "", err
	}

	return commitID, nil
}
