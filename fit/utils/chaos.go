package utils

import (
	"fit/fit/internal"
)

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
