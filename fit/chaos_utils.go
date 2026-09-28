package fit

func CorruptCommit(dir string, commitID CommitID, content string) error {
	commitsDir := MakePath(dir, commitsDir)

	_, err := WriteFile(commitsDir, string(commitID), content)
	if err != nil {
		return err
	}

	return nil
}

func CorruptBlob(dir string, blobID Hash, content string) error {
	blobsDir := MakePath(dir, blobsDir)

	_, err := WriteFile(blobsDir, string(blobID), content)
	if err != nil {
		return err
	}

	return nil
}
