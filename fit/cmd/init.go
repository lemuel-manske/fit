package cmd

import (
	"fit/fit/internal"

	"github.com/google/uuid"
)

// InitNew initializes a new repository in the specified directory with a new configuration.
func InitNew(repoDir string, repoName string) error {
	config := &internal.Config{
		FormatVersion:  internal.FormatVersion,
		PeerID:         internal.PeerID(uuid.New().String()),
		RepositoryID:   internal.RepositoryID(uuid.New().String()),
		RepositoryName: repoName,
		URL:            RemoteURL,
	}

	return InitWith(repoDir, config)
}

// InitWith initializes an existing repository in the specified directory with the provided configuration.
func InitWith(repoDir string, config *internal.Config) error {
	err := internal.WriteFitDir(repoDir)
	if err != nil {
		return err
	}

	err = internal.WriteConfig(repoDir, config)
	if err != nil {
		return err
	}

	index := &internal.Index{
		Version: 0,
		Entries: make(map[string]internal.IndexEntry),
	}

	err = internal.WriteIndex(repoDir, index)
	if err != nil {
		return err
	}

	err = internal.WriteHEAD(repoDir, "")
	if err != nil {
		return err
	}

	// create initial .fit/blobs and .fit/commits directories
	_ = internal.NewBlobStore(repoDir)
	_ = internal.NewCommitStore(repoDir)

	return nil
}
