package fit

import (
	"os"

	"path/filepath"
)

const (
	headFile = "HEAD"
)

func WriteHEAD(dir string, commitID CommitID) error {
	path := filepath.Join(dir, headFile)

	return os.WriteFile(path, []byte(commitID), 0644)
}

func ReadHEAD(dir string) (CommitID, error) {
	path := filepath.Join(dir, headFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return CommitID(data), nil
}
