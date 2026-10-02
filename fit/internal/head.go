package internal

import (
	"os"
)

func WriteHEAD(dir string, commitID CommitID) error {
	path := HEADPath(dir)

	return os.WriteFile(path, []byte(commitID), 0644)
}

func ReadHEAD(dir string) (CommitID, error) {
	path := HEADPath(dir)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return CommitID(data), nil
}

func HEADContains(dir, path string) (bool, error) {
	headID, err := ReadHEAD(dir)
	if err != nil {
		return false, err
	}

	if headID == "" {
		return false, nil
	}

	store := NewCommitStore(dir)

	commit, err := store.Get(headID)
	if err != nil {
		return false, err
	}

	_, exists := commit.Files[path]
	return exists, nil
}
