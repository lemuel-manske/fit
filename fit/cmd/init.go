package cmd

import (
	"os"

	"fit/fit/internal"

	"github.com/google/uuid"
)

// Init initializes a new repository in the specified directory with the given repository name.
func Init(repoDir string, repoName string) error {
	config := &internal.Config{
		FormatVersion:  internal.FormatVersion,
		PeerID:         internal.PeerID(uuid.New().String()),
		RepositoryID:   internal.RepositoryID(uuid.New().String()),
		RepositoryName: repoName,
	}

	err := os.MkdirAll(internal.MakePath(repoDir, ""), 0755)
	if err != nil {
		return err
	}

	internal.WriteConfig(repoDir, config)

	index := &internal.Index{
		Version: 0,
		Entries: make(map[string]internal.IndexEntry),
	}

	internal.WriteIndex(repoDir, index)

	internal.WriteHEAD(repoDir, "")

	internal.InitBlobStore(repoDir)
	internal.InitCommitStore(repoDir)

	return nil
}
