package internal

import (
	"os"
)

func MergeInProgress(dir string) bool {
	_, err := os.Stat(MergeHEADPath(dir))
	return err == nil
}

func ReadMergeHEAD(dir string) (CommitID, error) {
	path := MergeHEADPath(dir)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return CommitID(data), nil
}

func WriteMergeHEAD(dir string, commitID CommitID) error {
	path := MergeHEADPath(dir)

	return WriteFileAtomic(path, []byte(commitID), 0644)
}

func ReadMergeBase(dir string) (CommitID, error) {
	path := MergeBasePath(dir)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return CommitID(data), nil
}

func WriteMergeBase(dir string, commitID CommitID) error {
	path := MergeBasePath(dir)

	return WriteFileAtomic(path, []byte(commitID), 0644)
}

func ClearMergeState(dir string) error {
	err := os.Remove(MergeHEADPath(dir))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.Remove(MergeBasePath(dir))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
