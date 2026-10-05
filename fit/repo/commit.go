package repo

import (
	"fmt"
	"time"

	"fit/fit/fs"
)

func CommitChanges(dir string, message string) (fs.CommitID, error) {
	err := fs.RequireInitialized(dir)
	if err != nil {
		return "", err
	}

	store := fs.NewCommitStore(dir)

	index, err := fs.LoadIndex(dir)
	if err != nil {
		return "", err
	}

	for path, entry := range index.Entries {
		if entry.Conflict {
			return "", fmt.Errorf("unresolved merge conflict: %s", path)
		}
	}

	headID, err := fs.ReadHEAD(dir)
	if err != nil {
		return "", err
	}

	files := make(map[string]fs.Hash)

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

		files[path] = fs.Hash(entry.Blob)
	}

	parents := []fs.CommitID{}

	if hasHEAD {
		parents = append(parents, headID)

		mergeInProgress := fs.MergeInProgress(dir)

		if mergeInProgress {
			mergeCommitID, err := fs.ReadMergeHEAD(dir)
			if err != nil {
				return "", err
			}

			parents = append(parents, mergeCommitID)
		}
	}

	config, err := fs.LoadConfig(dir)
	if err != nil {
		return "", err
	}

	commit := fs.Commit{
		Author: fs.Author{
			PeerID: config.PeerID,
		},
		Parents:      parents,
		Message:      message,
		RepositoryID: config.RepositoryID,
		Timestamp:    time.Now().UTC(),
		Files:        files,
	}

	commitID, err := store.Put(commit)
	if err != nil {
		return "", err
	}

	err = fs.WriteHEAD(dir, commitID)
	if err != nil {
		return "", err
	}

	clear(index.Entries)

	err = fs.WriteIndex(dir, index)
	if err != nil {
		return "", err
	}

	if err = fs.ClearMergeState(dir); err != nil {
		return "", err
	}

	return commitID, nil
}
