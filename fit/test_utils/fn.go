package test_utils

import (
	"os"

	"fit/fit/internal"

	"path/filepath"
)

// WriteFile creates a new file with the given content.
func WriteFile(elements... string) (string, error) {
	content := elements[len(elements)-1]
	fileName := elements[len(elements)-2]

	dir := filepath.Join(elements[:len(elements)-2]...)
	path := filepath.Join(dir, fileName)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return path, nil
}

// ReadFile reads the content of a file and returns it as a byte slice.
func ReadFile(elements... string) ([]byte, error) {
	fileName := elements[len(elements)-1]

	dir := filepath.Join(elements[:len(elements)-1]...)
	path := filepath.Join(dir, fileName)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CorruptCommit overwrites the content of a commit file with the given content.
func CorruptCommit(dir string, commitID internal.CommitID, content string) error {
	commitsDir := internal.CommitsPath(dir)

	_, err := WriteFile(commitsDir, string(commitID), content)
	if err != nil {
		return err
	}

	return nil
}

// CorruptBlob overwrites the content of a blob file with the given content.
func CorruptBlob(dir string, blobID internal.Hash, content string) error {
	blobsDir := internal.BlobsPath(dir)

	_, err := WriteFile(blobsDir, string(blobID), content)
	if err != nil {
		return err
	}

	return nil
}
