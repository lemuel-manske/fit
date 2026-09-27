package fit

// Init initializes a new repository in the specified directory with the given repository name.
func Init(repoDir string, repoName string) error {
	config := &Config{
		FormatVersion:  FormatVersion,
		PeerID:         PeerID(NewUUID()),
		RepositoryID:   RepositoryID(NewUUID()),
		RepositoryName: repoName,
	}

	WriteConfig(repoDir, config)

	index := &Index{
		Version: 0,
		Entries: make(map[string]IndexEntry),
	}

	WriteIndex(repoDir, index)

	WriteHEAD(repoDir, "")

	InitFsBlobStore(repoDir)
	InitFsCommitStore(repoDir)

	return nil
}
