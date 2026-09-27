package fit

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

	return nil
}
